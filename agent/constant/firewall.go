package constant

const (
	FirewallProviderFirewalld   = "firewalld"
	FirewallProviderUFW         = "ufw"
	FirewallProviderIptables    = "iptables"
	FirewallProviderNftables    = "nftables"
	FirewallBackendNotInstalled = "backend_not_installed"

	FirewallFamilyIPv4 = "ipv4"
	FirewallFamilyIPv6 = "ipv6"
	FirewallFamilyInet = "inet"

	FirewallBasicBeforeChain = "1PANEL_BASIC_BEFORE"
	FirewallBasicChain       = "1PANEL_BASIC"
	FirewallBasicAfterChain  = "1PANEL_BASIC_AFTER"
)

const (
	FirewallSystemBackendKey         = "FirewallProvider"
	FirewallForwardingBackendKey     = "ForwardingBackend"
	FirewallDockerBackendKey         = "DockerFirewallBackend"
	FirewallDockerPortGuardStatusKey = "DockerPortGuardStatus"

	FirewallFilterInitializedKey     = "IptablesStatus"
	FirewallForwardingInitializedKey = "IptablesForwardStatus"
	FirewallPingStatusKey            = "BanPing"

	FirewallPortWhiteList      = "FirewallPortWhiteList"
	FirewallPortWhiteListValue = `[{"port":"80","protocol":"tcp","sources":["0.0.0.0/0","::/0"]},{"port":"443","protocol":"tcp","sources":["0.0.0.0/0","::/0"]},{"port":"443","protocol":"udp","sources":["0.0.0.0/0","::/0"]}]`
)

const (
	FirewallSystemAcceptedPortSourcePrefix = "accepted-port:"

	FirewallRuleOriginCreated = "created"
	FirewallRuleOriginAdopted = "adopted"

	FirewallRuleSourceUser     = "user"
	FirewallRuleSourceImported = "imported"
	FirewallRuleSourcePanel    = "panel"
	FirewallRuleSourceSecurity = "security"
	FirewallRuleSourceApp      = "application"
)
