import i18n from '@/lang';
import { compareVersion } from '@/utils/version';
import type { AI } from '@/api/interface/ai';

const openclawHTTPSVersion = '2026.3.13';
const openclawHTTPVersion = '2026.3.23';
const openclawDefaultAccessHost = '127.0.0.1';

export const isAgentType = (value: unknown): value is AI.AgentType =>
    typeof value === 'string' && ['openclaw', 'copaw', 'hermes-agent', 'deepseek-harness', 'octop'].includes(value);

export const isDeepSeekAccessHost = (value: string): boolean => {
    if (/^[0-9.]+$/.test(value)) {
        return (
            value.split('.').length === 4 &&
            value.split('.').every((part) => /^(0|[1-9]\d{0,2})$/.test(part) && +part <= 255)
        );
    }
    return (
        value.length > 0 &&
        value.length <= 253 &&
        value === value.trim() &&
        value.split('.').every((part) => /^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$/.test(part))
    );
};

export const isOctopPassword = (value: string): boolean =>
    [...value].length >= 8 &&
    value === value.trim() &&
    !/[\r\n\0]/.test(value) &&
    /\p{L}/u.test(value) &&
    /\p{Nd}/u.test(value) &&
    ![
        'password',
        'password1',
        'password12',
        'password123',
        '12345678',
        '123456789',
        'qwerty123',
        'admin123',
        'welcome1',
        'letmein1',
        'changeme1',
        'octop123',
        'abc12345',
        'iloveyou1',
    ].includes(value.toLowerCase());

export const getAgentProviderDisplayName = (provider: string, displayName?: string): string => {
    if (provider === 'custom' || displayName === 'Custom') {
        return i18n.global.t('container.custom');
    }
    if (provider === 'vllm' || displayName === 'vLLM') {
        return 'vLLM';
    }
    return displayName || provider;
};

export const isAgentAccountVerificationSkipped = (provider?: string): boolean =>
    ['vllm', 'ollama', 'llmman', 'kimi-coding'].includes((provider || '').toLowerCase());

export const supportsAgentModelConfig = (agentType: string): boolean => {
    return agentType === 'openclaw' || agentType === 'hermes-agent';
};

export const supportsAgentToken = (agentType: string): boolean => {
    return agentType === 'openclaw';
};

export const supportsAgentDashboardAuth = (agentType: string): boolean => {
    return ['hermes-agent', 'copaw', 'deepseek-harness', 'octop'].includes(agentType);
};

export const isOpenclawHTTPSWindowVersion = (version: string): boolean => {
    return compareVersion(version, openclawHTTPSVersion) && !compareVersion(version, openclawHTTPVersion);
};

export const isOpenclawCurrentHTTPVersion = (version: string): boolean => {
    return compareVersion(version, openclawHTTPVersion);
};

export const getOpenclawAccessScheme = (version: string): 'http' | 'https' => {
    return isOpenclawHTTPSWindowVersion(version) ? 'https' : 'http';
};

export const buildDefaultAllowedOrigin = (systemIP: string, port?: number | string, version?: string): string => {
    const target = String(systemIP || '').trim() || openclawDefaultAccessHost;
    if (!port) {
        return '';
    }
    const host = target.includes(':') && !target.startsWith('[') && !target.endsWith(']') ? `[${target}]` : target;
    return `${getOpenclawAccessScheme(String(version || ''))}://${host}:${port}`;
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
