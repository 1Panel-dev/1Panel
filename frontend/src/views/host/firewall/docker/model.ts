import { Firewall } from '@/api/interface/firewall';
import { isValidAddressForFamily } from '@/views/host/firewall/utils/validation';

type DockerGuardEndpointIdentity = Pick<Firewall.DockerGuardEndpoint, 'family' | 'hostIP' | 'hostPort' | 'protocol'>;

export const dockerGuardEndpointKey = (endpoint: DockerGuardEndpointIdentity): string =>
    `${endpoint.family}|${endpoint.hostIP}|${endpoint.hostPort}|${endpoint.protocol}`;

export const isValidDockerGuardSource = (family: Firewall.DockerGuardEndpoint['family'], value: string): boolean =>
    isValidAddressForFamily(family, value);

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
    if (policy.mode === 'deny_sources' && sources.length === 0) return;
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
