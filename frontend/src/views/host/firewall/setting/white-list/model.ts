import { normalizePortRange } from '@/views/host/firewall/utils/validation';

import type { Firewall } from '@/api/interface/firewall';

export type WhiteListProtocol = 'tcp' | 'udp';
export type WhiteListType = 'panel' | 'ssh';
export type WhiteListRule = Firewall.PortWhitelist;

export const normalizeWhiteListRule = (rule: WhiteListRule): WhiteListRule => {
    const { sources } = rule;
    const protocol = (rule.protocol?.trim().toLowerCase() || (rule.type ? 'tcp' : '')) as WhiteListProtocol;
    if (!['tcp', 'udp'].includes(protocol)) throw new Error('invalid whitelist rule');
    if (rule.type) {
        if (!['panel', 'ssh'].includes(rule.type)) throw new Error('invalid whitelist rule');
        if (rule.port !== undefined && !/^\d+$/.test(rule.port.trim())) {
            throw new Error('invalid service port');
        }
        return {
            type: rule.type,
            protocol,
            port: rule.port === undefined ? undefined : normalizePortRange(rule.port),
            sources,
        };
    }
    return {
        protocol,
        port: normalizePortRange(rule.port || ''),
        sources,
    };
};

export const whiteListRuleKey = (rule: WhiteListRule): string =>
    rule.type ? `${rule.type}/${rule.protocol || 'tcp'}` : `custom/${rule.protocol}/${rule.port}`;
