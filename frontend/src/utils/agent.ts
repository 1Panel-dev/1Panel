import i18n from '@/lang';

export const getAgentProviderDisplayName = (provider: string, displayName?: string): string => {
    if (provider === 'custom' || displayName === 'Custom') {
        return i18n.global.t('container.custom');
    }
    if (provider === 'vllm' || displayName === 'vLLM') {
        return 'vLLM';
    }
    return displayName || provider;
};

export const buildDefaultAllowedOrigin = (systemIP: string, port?: number | string): string => {
    const target = String(systemIP || '').trim();
    if (!target || !port) {
        return '';
    }
    const host = target.includes(':') && !target.startsWith('[') && !target.endsWith(']') ? `[${target}]` : target;
    return `https://${host}:${port}`;
};

export const normalizeAllowedOrigin = (value: string): string => {
    const target = String(value || '').trim();
    if (!target) {
        throw new Error(i18n.global.t('aiTools.agents.allowedOriginsInvalid'));
    }
    let parsed: URL;
    try {
        parsed = new URL(target);
    } catch (error) {
        throw new Error(i18n.global.t('aiTools.agents.allowedOriginsInvalid'));
    }
    if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
        throw new Error(i18n.global.t('aiTools.agents.allowedOriginsInvalid'));
    }
    if (parsed.username || parsed.password || parsed.search || parsed.hash) {
        throw new Error(i18n.global.t('aiTools.agents.allowedOriginsInvalid'));
    }
    if (parsed.pathname && parsed.pathname !== '/') {
        throw new Error(i18n.global.t('aiTools.agents.allowedOriginsInvalid'));
    }
    if (!parsed.host) {
        throw new Error(i18n.global.t('aiTools.agents.allowedOriginsInvalid'));
    }
    return `${parsed.protocol}//${parsed.host}`;
};

export const parseAllowedOriginsInput = (value: string): string[] => {
    const result: string[] = [];
    const seen = new Set<string>();
    for (const line of String(value || '').split(/\r?\n/)) {
        const target = line.trim();
        if (!target) {
            continue;
        }
        const normalized = normalizeAllowedOrigin(target);
        if (seen.has(normalized)) {
            continue;
        }
        seen.add(normalized);
        result.push(normalized);
    }
    return result;
};

export const validateAllowedOriginsInput = (value: string): string => {
    try {
        const parsed = parseAllowedOriginsInput(value);
        if (parsed.length === 0) {
            return i18n.global.t('aiTools.agents.allowedOriginsRequired');
        }
        return '';
    } catch (error: any) {
        return error?.message || i18n.global.t('aiTools.agents.allowedOriginsInvalid');
    }
};
