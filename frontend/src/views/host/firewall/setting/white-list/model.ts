export type WhiteListFamily = 'ipv4' | 'ipv6';
export type WhiteListProtocol = 'tcp' | 'udp';

export interface WhiteListRule {
    family: WhiteListFamily;
    protocol: WhiteListProtocol;
    port: string;
}

const normalizePort = (value: string): string => {
    const input = value.trim().replace(':', '-');
    const parts = input.split('-');
    if (parts.length > 2 || parts.some((part) => !/^\d+$/.test(part))) {
        throw new Error('invalid port');
    }
    const start = Number(parts[0]);
    const end = parts.length === 2 ? Number(parts[1]) : start;
    if (start < 1 || start > 65535 || end < 1 || end > 65535 || start > end) {
        throw new Error('invalid port');
    }
    return start === end ? String(start) : `${start}-${end}`;
};

export const normalizeWhiteListRule = (rule: WhiteListRule): WhiteListRule => {
    const family = rule.family?.toLowerCase() as WhiteListFamily;
    const protocol = rule.protocol?.toLowerCase() as WhiteListProtocol;
    if (!['ipv4', 'ipv6'].includes(family) || !['tcp', 'udp'].includes(protocol)) {
        throw new Error('invalid whitelist rule');
    }
    return { family, protocol, port: normalizePort(rule.port) };
};

export const whiteListRuleKey = (rule: WhiteListRule): string => {
    const normalized = normalizeWhiteListRule(rule);
    return `${normalized.family}/${normalized.protocol}/${normalized.port}`;
};

export const parseWhiteList = (value: string): WhiteListRule[] => {
    const input = value?.trim();
    if (!input) return [];

    let rules: WhiteListRule[];
    if (input.startsWith('[')) {
        rules = JSON.parse(input) as WhiteListRule[];
    } else {
        rules = input
            .split(/[\s,;]+/)
            .filter(Boolean)
            .map((item) => {
                const parts = item.split('/');
                if (parts.length === 3) {
                    return { family: parts[0], port: parts[1], protocol: parts[2] } as WhiteListRule;
                }
                return {
                    family: 'ipv4',
                    port: parts[0],
                    protocol: parts[1] || 'tcp',
                } as WhiteListRule;
            });
    }

    const result: WhiteListRule[] = [];
    const seen = new Set<string>();
    for (const item of rules) {
        const rule = normalizeWhiteListRule(item);
        const key = whiteListRuleKey(rule);
        if (seen.has(key)) continue;
        seen.add(key);
        result.push(rule);
    }
    return result;
};

export const serializeWhiteList = (rules: WhiteListRule[]): string => {
    return JSON.stringify(rules.map((rule) => normalizeWhiteListRule(rule)));
};

export const whiteListRuleCount = (value: string): number => {
    try {
        return parseWhiteList(value).length;
    } catch {
        return 0;
    }
};

export const whiteListRulesOverlap = (left: WhiteListRule, right: WhiteListRule): boolean => {
    const first = normalizeWhiteListRule(left);
    const second = normalizeWhiteListRule(right);
    if (first.family !== second.family || first.protocol !== second.protocol) return false;
    const range = (port: string) => {
        const [start, end = start] = port.split('-').map(Number);
        return { start, end };
    };
    const leftRange = range(first.port);
    const rightRange = range(second.port);
    return leftRange.start <= rightRange.end && rightRange.start <= leftRange.end;
};
