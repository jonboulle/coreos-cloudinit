package network

import (
	"net"
	"reflect"
	"testing"
)

func TestSplitStanzasNoParent(t *testing.T) {
	lines := []string{"test"}
	_, err := splitStanzas(lines)
	if _, ok := err.(NoParentStanzaError); !ok {
		t.FailNow()
	}
}

func TestSplitStanzas(t *testing.T) {
	expect := [][]string{
		{
			"auto lo",
		},
		{
			"iface eth1",
			"option: 1",
		},
		{
			"mapping",
		},
		{
			"allow-",
		},
	}
	lines := make([]string, 0, 5)
	for _, stanza := range expect {
		for _, line := range stanza {
			lines = append(lines, line)
		}
	}

	stanzas, err := splitStanzas(lines)
	if err != nil {
		t.FailNow()
	}
	for i, stanza := range stanzas {
		if len(stanza) != len(expect[i]) {
			t.FailNow()
		}
		for j, line := range stanza {
			if line != expect[i][j] {
				t.FailNow()
			}
		}
	}
}

func TestParseStanzaNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.FailNow()
		}
	}()
	parseStanza(nil)
}

func TestParseStanzaMalformedStart(t *testing.T) {
	_, err := parseStanza([]string{"iface"})
	if _, ok := err.(MalformedStanzaStartError); !ok {
		t.FailNow()
	}
}

func TestParseStanzaAuto(t *testing.T) {
	_, err := parseStanza([]string{"auto a"})
	if err != nil {
		t.FailNow()
	}
}

func TestParseStanzaIface(t *testing.T) {
	_, err := parseStanza([]string{"iface a inet manual"})
	if err != nil {
		t.FailNow()
	}
}

func TestParseStanzaUnknownStanza(t *testing.T) {
	_, err := parseStanza([]string{"allow-?? unknown"})
	if _, ok := err.(UnknownStanzaError); !ok {
		t.FailNow()
	}
}

func TestParseAutoStanza(t *testing.T) {
	interfaces := []string{"test", "attribute"}
	stanza, err := parseAutoStanza(interfaces, nil)
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(stanza.interfaces, interfaces) {
		t.FailNow()
	}
}

func TestParseBondStanzaNoSlaves(t *testing.T) {
	bond, err := parseBondStanza("", nil, nil, map[string][]string{})
	if err != nil {
		t.FailNow()
	}
	if bond.options["slaves"] != nil {
		t.FailNow()
	}
}

func TestParseBondStanza(t *testing.T) {
	conf := configMethodManual{}
	options := map[string][]string{
		"bond-slaves": []string{"1", "2"},
	}
	bond, err := parseBondStanza("test", conf, nil, options)
	if err != nil {
		t.FailNow()
	}
	if bond.name != "test" {
		t.FailNow()
	}
	if bond.kind != interfaceBond {
		t.FailNow()
	}
	if bond.configMethod != conf {
		t.FailNow()
	}
	if !reflect.DeepEqual(bond.options["slaves"], options["bond-slaves"]) {
		t.FailNow()
	}
}

func TestParsePhysicalStanza(t *testing.T) {
	conf := configMethodManual{}
	options := map[string][]string{
		"a": []string{"1", "2"},
		"b": []string{"1"},
	}
	physical, err := parsePhysicalStanza("test", conf, nil, options)
	if err != nil {
		t.FailNow()
	}
	if physical.name != "test" {
		t.FailNow()
	}
	if physical.kind != interfacePhysical {
		t.FailNow()
	}
	if physical.configMethod != conf {
		t.FailNow()
	}
	if !reflect.DeepEqual(physical.options, options) {
		t.FailNow()
	}
}

func TestParseVLANStanzaVLANName(t *testing.T) {
	conf := configMethodManual{}
	options := map[string][]string{}
	vlan, err := parseVLANStanza("vlan25", conf, nil, options)
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(vlan.options["id"], []string{"25"}) {
		t.FailNow()
	}
}

func TestParseVLANStanzaDotName(t *testing.T) {
	conf := configMethodManual{}
	options := map[string][]string{}
	vlan, err := parseVLANStanza("eth.25", conf, nil, options)
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(vlan.options["id"], []string{"25"}) {
		t.FailNow()
	}
}

func TestParseVLANStanzaBadName(t *testing.T) {
	conf := configMethodManual{}
	options := map[string][]string{}
	_, err := parseVLANStanza("myvlan", conf, nil, options)
	if _, ok := err.(VLANNameError); !ok {
		t.FailNow()
	}
}

func TestParseVLANStanzaBadId(t *testing.T) {
	conf := configMethodManual{}
	options := map[string][]string{}
	_, err := parseVLANStanza("eth.vlan", conf, nil, options)
	if _, ok := err.(VLANNameError); !ok {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaMissingAttribute(t *testing.T) {
	_, err := parseInterfaceStanza([]string{}, nil)
	if _, ok := err.(InterfaceMissingAttributesError); !ok {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaInvalidConfigMethod(t *testing.T) {
	_, err := parseInterfaceStanza([]string{"eth", "inet", "invalid"}, nil)
	if _, ok := err.(InterfaceInvalidConfigMethodError); !ok {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticNoAddress(t *testing.T) {
	options := []string{"address 192.168.1.100"}
	_, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if _, ok := err.(MalformedStaticNetworkError); !ok {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticNoNetmask(t *testing.T) {
	options := []string{"netmask 255.255.255.0"}
	_, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if _, ok := err.(MalformedStaticNetworkError); !ok {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticAddress(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0"}
	expect := net.IPNet{
		IP:   net.IPv4(192, 168, 1, 100),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}

	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if !reflect.DeepEqual(static.address, expect) {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticInvalidAddress(t *testing.T) {
	options := []string{"address invalid", "netmask 255.255.255.0"}

	_, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if _, ok := err.(MalformedStaticNetworkError); !ok {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticInvalidNetmask(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask invalid"}

	_, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if _, ok := err.(MalformedStaticNetworkError); !ok {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticNoGateway(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0", "gateway"}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if len(static.routes) != 0 {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticGateway(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0", "gateway 192.168.1.1"}
	expect := []route{
		{
			destination: net.IPNet{
				IP:   net.IPv4(0, 0, 0, 0),
				Mask: net.IPv4Mask(0, 0, 0, 0),
			},
			gateway: net.IPv4(192, 168, 1, 1),
		},
	}

	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if !reflect.DeepEqual(static.routes, expect) {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticTwoGateways(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0", "gateway 192.168.1.1 192.168.1.2"}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if len(static.routes) != 0 {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticDNS(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0", "dns-nameservers 192.168.1.10 192.168.1.11 192.168.1.12"}
	expect := []net.IP{
		net.IPv4(192, 168, 1, 10),
		net.IPv4(192, 168, 1, 11),
		net.IPv4(192, 168, 1, 12),
	}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if !reflect.DeepEqual(static.nameservers, expect) {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticPostUpInvalid(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0", "post-up invalid"}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if len(static.routes) != 0 {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticPostUpEmpty(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0", "post-up route add"}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if len(static.routes) != 0 {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticPostUpBadNet(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0", "post-up route add -net"}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if len(static.routes) != 0 {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticPostUpBadGateway(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0", "post-up route add gw"}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if len(static.routes) != 0 {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticPostUpBadNetmask(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0", "post-up route add netmask"}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if len(static.routes) != 0 {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticPostUp(t *testing.T) {
	options := []string{
		"address 192.168.1.100",
		"netmask 255.255.255.0",
		"post-up route add gw 192.168.1.1 -net 192.168.1.0 netmask 255.255.255.0",
	}
	expect := []route{
		{
			destination: net.IPNet{
				IP:   net.IPv4(192, 168, 1, 0),
				Mask: net.IPv4Mask(255, 255, 255, 0),
			},
			gateway: net.IPv4(192, 168, 1, 1),
		},
	}

	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if !reflect.DeepEqual(static.routes, expect) {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaLoopback(t *testing.T) {
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "loopback"}, nil)
	if err != nil {
		t.FailNow()
	}
	if _, ok := iface.configMethod.(configMethodLoopback); !ok {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaManual(t *testing.T) {
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "manual"}, nil)
	if err != nil {
		t.FailNow()
	}
	if _, ok := iface.configMethod.(configMethodManual); !ok {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaPostUpOption(t *testing.T) {
	options := []string{
		"post-up",
		"post-up 1 2",
		"post-up 3 4",
	}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "manual"}, options)
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(iface.options["post-up"], []string{"1 2", "3 4"}) {
		t.Log(iface.options["post-up"])
		t.FailNow()
	}
}

func TestParseInterfaceStanzaPreDownOption(t *testing.T) {
	options := []string{
		"pre-down",
		"pre-down 3",
		"pre-down 4",
	}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "manual"}, options)
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(iface.options["pre-down"], []string{"3", "4"}) {
		t.Log(iface.options["pre-down"])
		t.FailNow()
	}
}

func TestParseInterfaceStanzaEmptyOption(t *testing.T) {
	options := []string{
		"test",
	}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "manual"}, options)
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(iface.options["test"], []string{}) {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaOptions(t *testing.T) {
	options := []string{
		"test1 1",
		"test2 2 3",
		"test1 5 6",
	}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "manual"}, options)
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(iface.options["test1"], []string{"5", "6"}) {
		t.Log(iface.options["test1"])
		t.FailNow()
	}
	if !reflect.DeepEqual(iface.options["test2"], []string{"2", "3"}) {
		t.Log(iface.options["test2"])
		t.FailNow()
	}
}

func TestParseInterfaceStazaBond(t *testing.T) {
	iface, err := parseInterfaceStanza([]string{"mybond", "inet", "manual"}, []string{"bond-slaves eth"})
	if err != nil {
		t.FailNow()
	}
	if iface.kind != interfaceBond {
		t.FailNow()
	}
}

func TestParseInterfaceStazaVLANName(t *testing.T) {
	iface, err := parseInterfaceStanza([]string{"eth0.1", "inet", "manual"}, nil)
	if err != nil {
		t.FailNow()
	}
	if iface.kind != interfaceVLAN {
		t.FailNow()
	}
}

func TestParseInterfaceStazaVLANOption(t *testing.T) {
	iface, err := parseInterfaceStanza([]string{"vlan1", "inet", "manual"}, []string{"vlan_raw_device eth"})
	if err != nil {
		t.FailNow()
	}
	if iface.kind != interfaceVLAN {
		t.FailNow()
	}
}

func TestParseStanzasNone(t *testing.T) {
	stanzas, err := parseStanzas(nil)
	if err != err {
		t.FailNow()
	}
	if len(stanzas) != 0 {
		t.FailNow()
	}
}

func TestParseStanzasBadSplit(t *testing.T) {
	_, err := parseStanzas([]string{""})
	if err == nil {
		t.FailNow()
	}
}

func TestParseStanzasBadInterface(t *testing.T) {
	_, err := parseStanzas([]string{"iface"})
	if err == nil {
		t.FailNow()
	}
}

func TestParseStanzas(t *testing.T) {
	lines := []string{
		"auto lo",
		"iface lo inet loopback",
		"iface eth1 inet manual",
		"iface eth2 inet manual",
		"iface eth3 inet manual",
		"auto eth1 eth3",
	}
	expect := []stanza{
		&stanzaAuto{
			interfaces: []string{"lo"},
		},
		&stanzaInterface{
			name:         "lo",
			kind:         interfacePhysical,
			auto:         true,
			configMethod: configMethodLoopback{},
			options:      map[string][]string{},
		},
		&stanzaInterface{
			name:         "eth1",
			kind:         interfacePhysical,
			auto:         true,
			configMethod: configMethodManual{},
			options:      map[string][]string{},
		},
		&stanzaInterface{
			name:         "eth2",
			kind:         interfacePhysical,
			auto:         false,
			configMethod: configMethodManual{},
			options:      map[string][]string{},
		},
		&stanzaInterface{
			name:         "eth3",
			kind:         interfacePhysical,
			auto:         true,
			configMethod: configMethodManual{},
			options:      map[string][]string{},
		},
		&stanzaAuto{
			interfaces: []string{"eth1", "eth3"},
		},
	}
	stanzas, err := parseStanzas(lines)
	if err != err {
		t.FailNow()
	}
	if !reflect.DeepEqual(stanzas, expect) {
		t.FailNow()
	}
}
