export const DEFAULT_AI_PREFIX = '#';

export const DEFAULT_AI_RISK_COMMANDS = [
    'rm -rf',
    'mkfs',
    'dd if=',
    'curl | sh',
    'wget | sh',
    'chmod -R 777 /',
    'shutdown',
    'reboot',
    'poweroff',
    'init 0',
    ':(){ :|:& };:',
];

export const parseRiskCommands = (value: string): string[] => {
    if (!value) {
        return [...DEFAULT_AI_RISK_COMMANDS];
    }
    try {
        const parsed = JSON.parse(value);
        if (!Array.isArray(parsed)) {
            return [...DEFAULT_AI_RISK_COMMANDS];
        }
        return parsed.map((item) => String(item).trim()).filter((item) => item.length > 0);
    } catch {
        return [...DEFAULT_AI_RISK_COMMANDS];
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
    return result.length > 0 ? result : [...DEFAULT_AI_RISK_COMMANDS];
};
