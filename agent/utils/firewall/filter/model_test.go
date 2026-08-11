package filter

import (
	"errors"
	"testing"
)

func TestScopeValidateMVP(t *testing.T) {
	tests := []struct {
		name    string
		scope   Scope
		key     string
		wantErr error
	}{
		{
			name:  "iptables basic chain",
			scope: Scope{Provider: "IPTABLES", Family: FamilyIPv4, Table: "FILTER", Chain: "1panel_basic", Direction: DirectionInput},
			key:   "iptables:ipv4:filter:1PANEL_BASIC:input",
		},
		{
			name:  "firewalld public",
			scope: Scope{Provider: ProviderFirewalld, Family: FamilyInet, Zone: "PUBLIC", Direction: DirectionInput},
			key:   "firewalld:public:input",
		},
		{
			name:  "ufw incoming default chain",
			scope: Scope{Provider: ProviderUFW, Family: FamilyIPv6, Direction: DirectionInput},
			key:   "ufw:incoming:ipv6",
		},
		{
			name:    "firewalld private unsupported",
			scope:   Scope{Provider: ProviderFirewalld, Family: FamilyInet, Zone: "private", Direction: DirectionInput},
			wantErr: ErrUnsupportedScope,
		},
		{
			name:    "iptables external chain unsupported",
			scope:   Scope{Provider: ProviderIptables, Family: FamilyIPv4, Table: "filter", Chain: "DOCKER", Direction: DirectionInput},
			wantErr: ErrUnsupportedScope,
		},
		{
			name:    "ufw output unsupported",
			scope:   Scope{Provider: ProviderUFW, Family: FamilyIPv4, Direction: Direction("output")},
			wantErr: ErrInvalidScope,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.scope.ValidateMVP()
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("expected %v, got %v", test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("validate scope: %v", err)
			}
			if got := test.scope.Key(); got != test.key {
				t.Fatalf("expected key %q, got %q", test.key, got)
			}
		})
	}
}

func TestCapabilitiesSupportsMVPScope(t *testing.T) {
	capabilities := Capabilities{Scopes: MVPScopePatterns()}
	if !capabilities.SupportsScope(Scope{Provider: ProviderUFW, Family: FamilyIPv4, Direction: DirectionInput}) {
		t.Fatal("expected UFW incoming IPv4 scope to be supported")
	}
	if capabilities.SupportsScope(Scope{Provider: ProviderUFW, Family: FamilyIPv6, Chain: "outgoing", Direction: Direction("output")}) {
		t.Fatal("did not expect UFW outgoing scope to be supported")
	}
	if capabilities.SupportsScope(Scope{Provider: ProviderFirewalld, Family: FamilyInet, Zone: "private", Direction: DirectionInput}) {
		t.Fatal("did not expect firewalld private zone to be supported")
	}
}

func TestFirewalldFamiliesSharePublicExecutionScope(t *testing.T) {
	ipv4 := Scope{Provider: ProviderFirewalld, Family: FamilyIPv4, Zone: "public", Direction: DirectionInput}
	ipv6 := Scope{Provider: ProviderFirewalld, Family: FamilyIPv6, Zone: "public", Direction: DirectionInput}
	if ipv4.Key() != ipv6.Key() || ipv4.Key() != "firewalld:public:input" {
		t.Fatalf("firewalld public pipeline was split by family: ipv4=%q ipv6=%q", ipv4.Key(), ipv6.Key())
	}
}
