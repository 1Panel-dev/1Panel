import { Firewall } from '@/api/interface/firewall';

export const buildDockerPolicyExport = (policies: Firewall.DockerGuardPolicy[]) =>
    policies.map((policy) => ({
        family: policy.family,
        hostIP: policy.hostIP,
        hostPort: policy.hostPort,
        protocol: policy.protocol,
        mode: policy.mode,
        sources: [...policy.sources],
    }));

export const parseDockerPolicyImport = (content: string): unknown[] => {
    let parsed: unknown = JSON.parse(content);
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        const backup = parsed as {
            subsystem?: string;
            docker?: { Policies?: Record<string, unknown>[] };
        };
        if (backup.subsystem !== 'docker') throw new Error();
        parsed = backup.docker?.Policies?.map((policy) => ({
            family: policy.Family,
            hostIP: policy.HostIP,
            hostPort: policy.HostPort,
            protocol: policy.Protocol,
            mode: policy.Mode,
            sources: policy.Sources || [],
        }));
    }
    if (!Array.isArray(parsed)) throw new Error();
    return parsed;
};
