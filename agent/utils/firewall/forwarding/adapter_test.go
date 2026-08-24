package forwarding

import "testing"

func TestRuleIdentityIncludesEveryIdentityField(t *testing.T) {
	base := Rule{Family: FamilyIPv4, Protocol: "tcp", Port: "8080", TargetIP: "127.0.0.1", TargetPort: "80", Interface: "eth0"}
	variants := []Rule{
		{Family: FamilyIPv6, Protocol: base.Protocol, Port: base.Port, TargetIP: base.TargetIP, TargetPort: base.TargetPort, Interface: base.Interface},
		{Family: base.Family, Protocol: "udp", Port: base.Port, TargetIP: base.TargetIP, TargetPort: base.TargetPort, Interface: base.Interface},
		{Family: base.Family, Protocol: base.Protocol, Port: "8081", TargetIP: base.TargetIP, TargetPort: base.TargetPort, Interface: base.Interface},
		{Family: base.Family, Protocol: base.Protocol, Port: base.Port, TargetIP: "127.0.0.2", TargetPort: base.TargetPort, Interface: base.Interface},
		{Family: base.Family, Protocol: base.Protocol, Port: base.Port, TargetIP: base.TargetIP, TargetPort: "81", Interface: base.Interface},
		{Family: base.Family, Protocol: base.Protocol, Port: base.Port, TargetIP: base.TargetIP, TargetPort: base.TargetPort, Interface: "eth1"},
	}
	for _, variant := range variants {
		if variant.Identity() == base.Identity() {
			t.Fatalf("identity collision for %#v", variant)
		}
	}
}
