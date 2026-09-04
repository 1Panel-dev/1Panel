import { Firewall } from '@/api/interface/firewall';
import i18n from '@/lang';
import { isValidAddressForFamily } from '@/views/host/firewall/utils/validation';

type DockerGuardEndpointIdentity = Pick<Firewall.DockerGuardEndpoint, 'family' | 'hostIP' | 'hostPort' | 'protocol'>;

export const dockerGuardEndpointKey = (endpoint: DockerGuardEndpointIdentity): string =>
    `${endpoint.family}|${endpoint.hostIP}|${endpoint.hostPort}|${endpoint.protocol}`;

export const isDockerGuardRuntimeEndpoint = (endpoint: Firewall.DockerGuardEndpoint): boolean =>
    !endpoint.containerState || ['running', 'paused', 'restarting'].includes(endpoint.containerState);

export const dockerGuardManagementTarget = (
    endpoint: Firewall.DockerGuardEndpoint,
): NonNullable<Firewall.DockerGuardEndpoint['managementTarget']> => {
    if (endpoint.managementTarget) return endpoint.managementTarget;
    if (endpoint.trafficPath === 'forward') return 'container_guard';
    if (endpoint.trafficPath === 'input') return 'host_firewall';
    return 'needs_diagnosis';
};

export const dockerGuardEndpointManagementMessage = (endpoint: Firewall.DockerGuardEndpoint): string => {
    if (endpoint.managementReason) {
        return i18n.global.t(`firewall.dockerTrafficPathReason.${endpoint.managementReason}`);
    }
    return i18n.global.t('firewall.dockerTrafficPathUnknown');
};

export const isValidDockerGuardSource = (family: Firewall.DockerGuardEndpoint['family'], value: string): boolean =>
    isValidAddressForFamily(family, value);

export const dockerGuardFamilyStatusMessage = (
    base: Firewall.DockerGuardBase,
    family: 'IPv4' | 'IPv6',
    status: Firewall.DockerGuardFamilyStatus,
): string => {
    if (status.reason) {
        return i18n.global.t(`firewall.dockerGuardStatusReason.${status.reason}`, [family]);
    }
    const state = i18n.global.t(
        !status.initialized
            ? 'firewall.notInitialized'
            : !status.bound
              ? 'commons.status.unbind'
              : 'firewall.notEffective',
    );
    const chainName = base.backend === 'nftables' ? 'NFT_1PANEL_DOCKER' : '1PANEL_DOCKER';
    return i18n.global.t('firewall.familyChainIssue', [family, chainName, state]);
};

export const dockerGuardEndpointStatusMessage = (
    base: Firewall.DockerGuardBase,
    endpoint: Firewall.DockerGuardEndpoint,
): string => {
    if (!endpoint.policyUUID || endpoint.effective) return '';
    const target = dockerGuardManagementTarget(endpoint);
    if (target === 'host_firewall') return i18n.global.t('firewall.dockerInputPolicyNotEffective');
    if (target === 'needs_diagnosis') return dockerGuardEndpointManagementMessage(endpoint);
    const ipv6 = endpoint.family === 'ipv6';
    return dockerGuardFamilyStatusMessage(base, ipv6 ? 'IPv6' : 'IPv4', ipv6 ? base.ipv6 : base.ipv4);
};

export const normalizeDockerGuardPolicy = (value: unknown): Firewall.DockerGuardPolicy | undefined => {
    if (!value || typeof value !== 'object') return;
    const policy = value as Partial<Firewall.DockerGuardPolicy>;
    if (!['ipv4', 'ipv6'].includes(String(policy.family))) return;
    if (!['tcp', 'udp'].includes(String(policy.protocol))) return;
    if (!['deny_sources', 'allow_sources', 'deny_all'].includes(String(policy.mode))) return;
    if (typeof policy.hostIP !== 'string' || !Number.isInteger(policy.hostPort)) return;
    if (policy.hostPort! < 1 || policy.hostPort! > 65535) return;
    if (!isValidAddressForFamily(policy.family!, policy.hostIP, false)) return;
    if (!Array.isArray(policy.sources) || !policy.sources.every((source) => typeof source === 'string')) return;
    const sources = policy.sources.map((source) => source.trim()).filter(Boolean);
    if (policy.mode !== 'deny_all' && sources.length === 0) return;
    if (policy.mode !== 'deny_all' && !sources.every((source) => isValidDockerGuardSource(policy.family!, source))) {
        return;
    }
    if (policy.description !== undefined && typeof policy.description !== 'string') return;
    return {
        family: policy.family!,
        hostIP: policy.hostIP,
        hostPort: policy.hostPort!,
        protocol: policy.protocol!,
        mode: policy.mode!,
        sources: policy.mode === 'deny_all' ? [] : [...new Set(sources)],
        description: policy.description?.trim() || '',
    };
};
