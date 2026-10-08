import { Firewall } from '@/api/interface/firewall';
import { inferAddressFamily, isValidAddressForFamily, normalizePortRange } from '../utils/validation';

export const normalizeForwardRuleImport = (value: unknown): Firewall.RuleForward => {
    const item = value as Partial<Firewall.RuleForward>;
    if (!item || typeof item.targetIP !== 'string' || typeof item.protocol !== 'string') {
        throw new Error('invalid forwarding rule');
    }
    let targetIP = item.targetIP.trim();
    const family = item.family?.trim().toLowerCase() || inferAddressFamily(targetIP);
    const protocol = item.protocol.trim().toLowerCase();
    if ((family !== 'ipv4' && family !== 'ipv6') || !isValidAddressForFamily(family, targetIP, false)) {
        throw new Error('invalid forwarding address');
    }
    if (!['tcp', 'udp', 'tcp/udp', 'udp/tcp'].includes(protocol)) {
        throw new Error('invalid forwarding protocol');
    }
    if (item.interface != null && typeof item.interface !== 'string') {
        throw new Error('invalid forwarding interface');
    }
    if (family === 'ipv6') targetIP = new URL(`http://[${targetIP}]/`).hostname.slice(1, -1);
    const inboundInterface = item.interface?.trim() || '';
    return {
        operation: 'add',
        family,
        protocol,
        port: normalizePortRange(item.port),
        targetIP,
        targetPort: normalizePortRange(item.targetPort),
        interface: ['all', '*'].includes(inboundInterface) ? '' : inboundInterface,
    };
};

export const buildForwardRuleExport = (rules: Firewall.RuleForward[]) =>
    rules.map((rule) => ({
        family: rule.family,
        protocol: rule.protocol,
        port: rule.port,
        targetIP: rule.targetIP,
        targetPort: rule.targetPort,
        interface: rule.interface,
    }));

export const parseForwardRuleImport = (content: string): Firewall.RuleForward[] | undefined => {
    let parsed = JSON.parse(content);
    if (parsed?.subsystem === 'forwarding' && Array.isArray(parsed.forwarding)) {
        parsed = parsed.forwarding.map((rule) => ({
            family: rule.Family,
            protocol: rule.Protocol,
            port: rule.Port,
            targetIP: rule.TargetIP,
            targetPort: rule.TargetPort,
            interface: rule.Interface,
        }));
    }
    if (!Array.isArray(parsed)) return;
    return parsed;
};
