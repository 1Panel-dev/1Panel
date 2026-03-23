export const DEFAULT_AI_PREFIX = '@ai';
export const AI_PREFIX_OPTIONS = ['@ai', '#ai', '/ai'] as const;

export const parseRiskCommands = (value: string): string[] => {
    if (!value) {
        return [];
    }
    try {
        const parsed = JSON.parse(value);
        if (!Array.isArray(parsed)) {
            return [];
        }
        return parsed.map((item) => String(item).trim()).filter((item) => item.length > 0);
    } catch {
        return [];
    }
};

export const normalizeRiskCommands = (riskCommands: string[]): string[] => {
    const seen = new Set<string>();
    const result: string[] = [];
    for (const command of riskCommands) {
        const normalized = command.trim();
        if (!normalized || seen.has(normalized)) {
            continue;
        }
        seen.add(normalized);
        result.push(normalized);
    }
    return result;
};
