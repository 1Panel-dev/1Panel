package constant

const (
	FirewallProviderFirewalld = "firewalld"
	FirewallProviderUFW       = "ufw"
	FirewallProviderIptables  = "iptables"
	FirewallProviderNftables  = "nftables"

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
	FirewallPortWhiteListValue = "80/tcp,443/tcp,443/udp"
)

const (
	FirewallSystemAcceptedPortSourcePrefix = "accepted-port:"
	FirewallRuleCheckVersion               = 1

	FirewallRuleOriginCreated = "created"
	FirewallRuleOriginAdopted = "adopted"

	FirewallRuleSourceUser     = "user"
	FirewallRuleSourceImported = "imported"
	FirewallRuleSourcePanel    = "panel"
	FirewallRuleSourceSecurity = "security"
	FirewallRuleSourceApp      = "application"
)
