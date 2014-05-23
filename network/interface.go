package network

import (
	"fmt"
	"strconv"
	"strings"
)

type InterfaceGenerator interface {
	Name() string
	Netdev() string
	Link() string
	Network() string
}

type logicalInterface struct {
	name     string
	config   configMethod
	children []InterfaceGenerator
}

type physicalInterface struct {
	logicalInterface
}

func (p *physicalInterface) Name() string {
	return p.name
}

func (p *physicalInterface) Netdev() string {
	return ""
}

func (p *physicalInterface) Link() string {
	return ""
}

func (p *physicalInterface) Network() string {
	config := fmt.Sprintf("[Match]\nName=%s\n\n[Network]\n", p.name)

	for _, child := range p.children {
		switch iface := child.(type) {
		case *vlanInterface:
			config += fmt.Sprintf("VLAN=%s\n", iface.name)
		case *bondInterface:
			config += fmt.Sprintf("Bond=%s\n", iface.name)
		}
	}

	return config
}

type bondInterface struct {
	logicalInterface
	slaves []string
}

func (b *bondInterface) Name() string {
	return b.name
}

func (b *bondInterface) Netdev() string {
	return fmt.Sprintf("[NetDev]\nKind=bond\nName=%s\n", b.name)
}

func (b *bondInterface) Link() string {
	return ""
}

func (b *bondInterface) Network() string {
	config := fmt.Sprintf("[Match]\nName=%s\n\n[Network]\nDHCP=true\n", b.name)

	for _, child := range b.children {
		switch iface := child.(type) {
		case *vlanInterface:
			config += fmt.Sprintf("VLAN=%s\n", iface.name)
		case *bondInterface:
			config += fmt.Sprintf("Bond=%s\n", iface.name)
		}
	}

	return config
}

type vlanInterface struct {
	logicalInterface
	id        int
	rawDevice string
}

func (v *vlanInterface) Name() string {
	return v.name
}

func (v *vlanInterface) Netdev() string {
	return fmt.Sprintf("[NetDev]\nKind=vlan\nName=%s\n\n[VLAN]\nId=%d\n", v.name, v.id)
}

func (v *vlanInterface) Link() string {
	return ""
}

func (v *vlanInterface) Network() string {
	config := fmt.Sprintf("[Match]\nName=%s\n\n[Network]\n", v.name)
	switch conf := v.config.(type) {
	case configMethodStatic:
		for _, nameserver := range conf.nameservers {
			config += fmt.Sprintf("DNS=%s\n", nameserver)
		}
		if conf.address.IP != nil {
			config += fmt.Sprintf("\n[Address]\nAddress=%s\n", conf.address.String())
		}
		for _, route := range conf.routes {
			config += fmt.Sprintf("\n[Route]\nDestination=%s\nGateway=%s\n", route.destination.String(), route.gateway)
		}
	}

	return config
}

func buildInterfaces(stanzas []*stanzaInterface) []InterfaceGenerator {
	bondStanzas := make(map[string]*stanzaInterface)
	physicalStanzas := make(map[string]*stanzaInterface)
	vlanStanzas := make(map[string]*stanzaInterface)
	for _, iface := range stanzas {
		switch iface.kind {
		case interfaceBond:
			bondStanzas[iface.name] = iface
		case interfacePhysical:
			physicalStanzas[iface.name] = iface
		case interfaceVLAN:
			vlanStanzas[iface.name] = iface
		}
	}

	physicals := make(map[string]*physicalInterface)
	for _, p := range physicalStanzas {
		if p.name == "lo" {
			continue
		}
		physicals[p.name] = &physicalInterface{
			logicalInterface{
				name:     p.name,
				config:   p.configMethod,
				children: []InterfaceGenerator{},
			},
		}
	}

	bonds := make(map[string]*bondInterface)
	for _, b := range bondStanzas {
		var slaves []string
		if s, ok := b.options["bond-slaves"]; ok {
			slaves = s
		}
		bonds[b.name] = &bondInterface{
			logicalInterface{
				name:     b.name,
				config:   b.configMethod,
				children: []InterfaceGenerator{},
			},
			slaves,
		}
	}

	vlans := make(map[string]*vlanInterface)
	for _, v := range vlanStanzas {
		var rawDevice string
		id, _ := strconv.Atoi(strings.Split(v.name, ".")[1])
		if device, ok := v.options["vlan_raw_device"]; ok && len(device) == 1 {
			rawDevice = device[0]
		}
		vlans[v.name] = &vlanInterface{
			logicalInterface{
				name:     v.name,
				config:   v.configMethod,
				children: []InterfaceGenerator{},
			},
			id,
			rawDevice,
		}
	}

	for _, vlan := range vlans {
		if physical, ok := physicals[vlan.rawDevice]; ok {
			physical.children = append(physical.children, vlan)
		}
		if bond, ok := bonds[vlan.rawDevice]; ok {
			bond.children = append(bond.children, vlan)
		}
	}

	for _, bond := range bonds {
		for _, slave := range bond.slaves {
			if physical, ok := physicals[slave]; ok {
				physical.children = append(physical.children, bond)
			}
			if bond, ok := bonds[slave]; ok {
				bond.children = append(bond.children, bond)
			}
		}
	}

	interfaces := make([]InterfaceGenerator, 0, len(physicals)+len(bonds)+len(vlans))
	for _, physical := range physicals {
		interfaces = append(interfaces, physical)
	}
	for _, bond := range bonds {
		interfaces = append(interfaces, bond)
	}
	for _, vlan := range vlans {
		interfaces = append(interfaces, vlan)
	}

	return interfaces
}
