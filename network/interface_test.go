package network

import (
	"net"
	"testing"
)

func TestPhysicalInterfaceName(t *testing.T) {
	p := physicalInterface{logicalInterface{name: "testname"}}
	if p.Name() != "testname" {
		t.FailNow()
	}
}

func TestPhysicalInterfaceNetdev(t *testing.T) {
	p := physicalInterface{}
	if p.Netdev() != "" {
		t.FailNow()
	}
}

func TestPhysicalInterfaceLink(t *testing.T) {
	p := physicalInterface{}
	if p.Link() != "" {
		t.FailNow()
	}
}

func TestPhysicalInterfaceNetwork(t *testing.T) {
	p := physicalInterface{logicalInterface{
		name: "testname",
		children: []InterfaceGenerator{
			&bondInterface{
				logicalInterface{
					name: "testbond1",
				},
				nil,
			},
			&vlanInterface{
				logicalInterface{
					name: "testvlan1",
				},
				1,
				"",
			},
			&vlanInterface{
				logicalInterface{
					name: "testvlan2",
				},
				1,
				"",
			},
		},
	}}
	network := `[Match]
Name=testname

[Network]
Bond=testbond1
VLAN=testvlan1
VLAN=testvlan2
`
	if p.Network() != network {
		t.FailNow()
	}
}

func TestBondInterfaceName(t *testing.T) {
	b := bondInterface{logicalInterface{name: "testname"}, nil}
	if b.Name() != "testname" {
		t.FailNow()
	}
}

func TestBondInterfaceNetdev(t *testing.T) {
	b := bondInterface{logicalInterface{name: "testname"}, nil}
	netdev := `[NetDev]
Kind=bond
Name=testname
`
	if b.Netdev() != netdev {
		t.FailNow()
	}
}

func TestBondInterfaceLink(t *testing.T) {
	b := bondInterface{}
	if b.Link() != "" {
		t.FailNow()
	}
}

func TestBondInterfaceNetwork(t *testing.T) {
	b := bondInterface{
		logicalInterface{
			name: "testname",
			children: []InterfaceGenerator{
				&bondInterface{
					logicalInterface{
						name: "testbond1",
					},
					nil,
				},
				&vlanInterface{
					logicalInterface{
						name: "testvlan1",
					},
					1,
					"",
				},
				&vlanInterface{
					logicalInterface{
						name: "testvlan2",
					},
					1,
					"",
				},
			},
		},
		nil,
	}
	network := `[Match]
Name=testname

[Network]
DHCP=true
Bond=testbond1
VLAN=testvlan1
VLAN=testvlan2
`
	if b.Network() != network {
		t.FailNow()
	}
}

func TestVLANInterfaceName(t *testing.T) {
	v := vlanInterface{logicalInterface{name: "testname"}, 1, ""}
	if v.Name() != "testname" {
		t.FailNow()
	}
}

func TestVLANInterfaceNetdev(t *testing.T) {
	v := vlanInterface{logicalInterface{name: "testname"}, 1, ""}
	netdev := `[NetDev]
Kind=vlan
Name=testname

[VLAN]
Id=1
`
	if v.Netdev() != netdev {
		t.FailNow()
	}
}

func TestVLANInterfaceLink(t *testing.T) {
	v := vlanInterface{}
	if v.Link() != "" {
		t.FailNow()
	}
}

func TestVLANInterfaceNetwork(t *testing.T) {
	v := vlanInterface{
		logicalInterface{
			name: "testname",
			config: configMethodStatic{
				address: net.IPNet{
					IP:   []byte{192, 168, 1, 100},
					Mask: []byte{255, 255, 255, 0},
				},
				nameservers: []net.IP{
					[]byte{8, 8, 8, 8},
				},
				routes: []route{
					route{
						destination: net.IPNet{
							IP:   []byte{0, 0, 0, 0},
							Mask: []byte{0, 0, 0, 0},
						},
						gateway: []byte{1, 2, 3, 4},
					},
				},
			},
		},
		0,
		"",
	}
	network := `[Match]
Name=testname

[Network]
DNS=8.8.8.8

[Address]
Address=192.168.1.100/24

[Route]
Destination=0.0.0.0/0
Gateway=1.2.3.4
`
	if v.Network() != network {
		t.Log(v.Network())
		t.FailNow()
	}
}
