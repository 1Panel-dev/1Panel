export interface AgentProviderLogo {
    src?: string;
    mark: string;
    background: string;
    color: string;
    borderColor?: string;
    source?: string;
}

const asset = (name: string) => new URL(`../assets/images/ai-providers/${name}`, import.meta.url).href;

const providerLogos: Record<string, AgentProviderLogo> = {
    ollama: {
        src: asset('ollama.webp'),
        mark: 'OL',
        background: '#ffffff',
        color: '#111827',
        borderColor: '#dcdfe6',
        source: 'https://github.com/ollama/ollama/blob/main/docs/ollama-logo.svg',
    },
    vllm: {
        src: asset('vllm.svg'),
        mark: 'v',
        background: '#ffffff',
        color: '#4338ca',
        borderColor: '#dcdfe6',
        source: 'https://vllm.ai/vLLM-Logo.svg',
    },
    deepseek: {
        src: asset('deepseek.png'),
        mark: 'DS',
        background: '#ffffff',
        color: '#4d6bfe',
        borderColor: '#dcdfe6',
        source: 'https://cdn.deepseek.com/logo.png',
    },
    'bailian-coding-plan': {
        src: asset('aliyun.webp'),
        mark: 'QW',
        background: '#ffffff',
        color: '#ff6a00',
        borderColor: '#dcdfe6',
        source: 'https://www.alibabacloud.com/',
    },
    'ark-coding-plan': {
        src: asset('volcengine.png'),
        mark: 'AR',
        background: '#ffffff',
        color: '#1664ff',
        borderColor: '#dcdfe6',
        source: 'https://www.volcengine.com/',
    },
    zai: {
        src: asset('zai.webp'),
        mark: 'Z',
        background: '#ffffff',
        color: '#111827',
        borderColor: '#dcdfe6',
        source: 'https://z.ai/',
    },
    minimax: {
        src: asset('minimax.ico'),
        mark: 'MM',
        background: '#ffffff',
        color: '#2563eb',
        borderColor: '#dcdfe6',
        source: 'https://www.minimax.io/',
    },
    xiaomi: {
        src: asset('xiaomi.ico'),
        mark: 'MI',
        background: '#ffffff',
        color: '#ff6900',
        borderColor: '#dcdfe6',
        source: 'https://www.mi.com/favicon.ico',
    },
    kimi: {
        src: asset('kimi.svg'),
        mark: 'KM',
        background: '#ffffff',
        color: '#111827',
        borderColor: '#dcdfe6',
        source: 'https://moonshotai.github.io/Branding-Guide/',
    },
    'kimi-coding': {
        src: asset('kimi.svg'),
        mark: 'KC',
        background: '#ffffff',
        color: '#111827',
        borderColor: '#dcdfe6',
        source: 'https://moonshotai.github.io/Branding-Guide/',
    },
    openai: {
        src: asset('openai.svg'),
        mark: 'AI',
        background: '#ffffff',
        color: '#111827',
        borderColor: '#dcdfe6',
        source: 'https://openai.com/brand/',
    },
    openrouter: {
        src: asset('openrouter.png'),
        mark: 'OR',
        background: '#ffffff',
        color: '#101828',
        borderColor: '#dcdfe6',
        source: 'https://openrouter.ai/',
    },
    anthropic: {
        src: asset('anthropic.png'),
        mark: 'A',
        background: '#ffffff',
        color: '#141413',
        borderColor: '#dcdfe6',
        source: 'https://www.anthropic.com/',
    },
    gemini: {
        src: asset('gemini.webp'),
        mark: 'G',
        background: '#ffffff',
        color: '#1a73e8',
        borderColor: '#dcdfe6',
        source: 'https://gemini.google.com/',
    },
    moonshot: {
        src: asset('kimi.svg'),
        mark: 'MS',
        background: '#ffffff',
        color: '#111827',
        borderColor: '#dcdfe6',
        source: 'https://moonshotai.github.io/Branding-Guide/',
    },
};

const fallbackLogo: AgentProviderLogo = {
    mark: 'AI',
    background: '#f5f7fa',
    color: '#606266',
    borderColor: '#dcdfe6',
};

const normalizeProvider = (provider?: string) =>
    String(provider || '')
        .trim()
        .toLowerCase();

const buildFallbackMark = (provider?: string, displayName?: string) => {
    const source = String(displayName || provider || '').trim();
    const asciiWords = source.match(/[a-zA-Z0-9]+/g) || [];
    if (asciiWords.length >= 2) {
        return `${asciiWords[0][0]}${asciiWords[1][0]}`.toUpperCase();
    }
    if (asciiWords.length === 1) {
        return asciiWords[0].slice(0, 2).toUpperCase();
    }
    return fallbackLogo.mark;
};

export const getAgentProviderLogo = (provider?: string, displayName?: string): AgentProviderLogo => {
    const logo = providerLogos[normalizeProvider(provider)];
    if (logo) {
        return logo;
    }
    return {
        ...fallbackLogo,
        mark: buildFallbackMark(provider, displayName),
    };
};
