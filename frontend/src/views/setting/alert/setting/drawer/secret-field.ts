export type SecretAction = 'keep' | 'replace' | 'clear';

export interface SecretView {
    configured: boolean;
    masked?: string;
}

export interface SecretDraft extends SecretView {
    masked: string;
    action: SecretAction;
    value: string;
    originalValue?: string;
}

export const secretEditorValue = (draft: SecretDraft): string => {
    if (draft.action === 'keep' || draft.action === 'replace') return draft.value;
    return '';
};

export const secretEditorPlaceholder = (draft: SecretDraft, fallback = ''): string => {
    if (draft.action === 'keep' && draft.configured) return draft.masked || '******';
    return fallback;
};

export const replaceSecretDraft = <T extends SecretDraft>(draft: T, value: string): T => ({
    ...draft,
    action: 'replace',
    value,
});

export const keepSecretDraft = <T extends SecretDraft>(draft: T): T => ({
    ...draft,
    action: 'keep',
    value: draft.originalValue || '',
});

export const clearSecretDraft = <T extends SecretDraft>(draft: T): T => ({
    ...draft,
    action: 'clear',
    value: '',
});

export const rawSecretValue = (source: unknown): string => {
    if (typeof source === 'string') return source;
    if (!source || typeof source !== 'object') return '';
    const value = (source as Record<string, unknown>).value;
    return typeof value === 'string' ? value : '';
};

export const serializeLegacySecretValue = (value: string, trim = false): string => (trim ? value.trim() : value);

export interface LegacyEmailTestFields {
    host: string;
    port: number;
    sender: string;
    userName: string;
    password: string;
    displayName: string;
    encryption: string;
    recipient: string;
}

const rawStringValue = (value: unknown): string => (typeof value === 'string' ? value : '');

export const buildLegacyEmailTestFields = (config: Record<string, unknown>): LegacyEmailTestFields => ({
    host: rawStringValue(config.host),
    port: Number(config.port) || 0,
    sender: rawStringValue(config.sender),
    userName: rawStringValue(config.userName),
    password: rawSecretValue(config.password),
    displayName: rawStringValue(config.displayName),
    encryption: rawStringValue(config.encryption) || 'NONE',
    recipient: rawStringValue(config.recipient),
});

export const getAlertConfigDisplayName = (type: string, config: Record<string, unknown>): string => {
    const displayName = typeof config.displayName === 'string' ? config.displayName.trim() : '';
    if (displayName) return displayName;

    const sender = typeof config.sender === 'string' ? config.sender.trim() : '';
    if (sender) return sender;

    if (type === 'sms') {
        const phone = rawSecretValue(config.phone);
        if (phone) return phone;
    }

    if (Array.isArray(config.webhooks)) {
        return config.webhooks
            .map((webhook) =>
                webhook && typeof webhook === 'object' && typeof webhook.displayName === 'string'
                    ? webhook.displayName.trim()
                    : '',
            )
            .filter(Boolean)
            .join(', ');
    }

    return '';
};
