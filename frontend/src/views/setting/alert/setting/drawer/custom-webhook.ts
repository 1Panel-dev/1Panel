export const CUSTOM_WEBHOOK_TYPE = 'custom';
export const CUSTOM_WEBHOOK_SCHEMA_VERSION = 1 as const;
export const CUSTOM_WEBHOOK_LIMITS = {
    url: 8192,
    headers: 64,
    headerName: 256,
    headerValue: 16 * 1024,
    body: 256 * 1024,
    formFields: 128,
} as const;

export type CustomWebhookPreset = 'genericJson' | 'slack' | 'discord' | 'teamsWorkflows' | 'custom';
export type CustomWebhookBodyType = 'json' | 'form' | 'text';
export type CustomWebhookSecretAction = 'keep' | 'replace' | 'clear';

export interface CustomWebhookSecretView {
    configured: boolean;
    masked?: string;
    value?: string;
}

export interface CustomWebhookSecretDraft extends CustomWebhookSecretView {
    masked: string;
    action: CustomWebhookSecretAction;
    value: string;
    originalValue?: string;
}

export interface CustomWebhookFormField {
    uid: string;
    key: string;
    value: string;
}

export interface CustomWebhookFormFieldView {
    uid?: string;
    key: string;
    value: string;
}

export interface CustomWebhookHeaderView {
    uid: string;
    key: string;
    secret: boolean;
    value?: string;
    configured?: boolean;
    masked?: string;
}

export interface CustomWebhookHeaderDraft {
    uid: string;
    key: string;
    secret: boolean;
    configured: boolean;
    masked: string;
    action: CustomWebhookSecretAction;
    value: string;
    originalValue?: string;
}

export interface CustomWebhookBody {
    type: CustomWebhookBodyType;
    template: string;
    fields: CustomWebhookFormField[];
}

export interface CustomWebhookBodyView {
    type: CustomWebhookBodyType;
    template: string;
    fields: CustomWebhookFormFieldView[];
}

export interface CustomWebhookConfigView {
    schemaVersion: typeof CUSTOM_WEBHOOK_SCHEMA_VERSION;
    state?: 'legacy' | 'invalid';
    displayName: string;
    preset: CustomWebhookPreset;
    method?: 'POST';
    url: CustomWebhookSecretView | string;
    body: Partial<CustomWebhookBodyView>;
    headers: CustomWebhookHeaderView[];
}

export interface CustomWebhookDraft {
    schemaVersion: typeof CUSTOM_WEBHOOK_SCHEMA_VERSION;
    state?: 'legacy' | 'invalid';
    displayName: string;
    preset: CustomWebhookPreset;
    url: CustomWebhookSecretDraft;
    body: CustomWebhookBody;
    headers: CustomWebhookHeaderDraft[];
}

export interface CustomWebhookSecretMutation {
    action: CustomWebhookSecretAction;
    value?: string;
}

export interface CustomWebhookConfigUpdate {
    schemaVersion: typeof CUSTOM_WEBHOOK_SCHEMA_VERSION;
    displayName: string;
    preset: CustomWebhookPreset;
    method: 'POST';
    url: CustomWebhookSecretMutation;
    body: {
        type: CustomWebhookBodyType;
        template: string;
        fields: Array<{
            key: string;
            value: string;
        }>;
    };
    headers: Array<{
        uid: string;
        key: string;
        secret: boolean;
        action: CustomWebhookSecretAction;
        value?: string;
    }>;
}

export interface CustomWebhookValidationIssue {
    field: string;
    code:
        | 'displayNameRequired'
        | 'urlRequired'
        | 'urlInvalid'
        | 'bodyRequired'
        | 'jsonInvalid'
        | 'formFieldRequired'
        | 'formFieldDuplicate'
        | 'headerRequired'
        | 'headerInvalid'
        | 'headerDuplicate'
        | 'headerReserved'
        | 'headerMustBeSecret'
        | 'templateVariableInvalid'
        | 'secretRequired';
}

interface PresetDefinition {
    bodyType: CustomWebhookBodyType;
    template: string;
}

const PRESET_DEFINITIONS: Record<Exclude<CustomWebhookPreset, 'custom'>, PresetDefinition> = {
    genericJson: {
        bodyType: 'json',
        template: `{
  "schema_version": "1",
  "title": "{{title}}",
  "message": "{{message}}",
  "type": "{{type}}",
  "node_name": "{{nodeName}}",
  "timestamp": "{{timestamp}}"
}`,
    },
    slack: {
        bodyType: 'json',
        template: `{
  "text": "*{{title}}*\\n{{message}}"
}`,
    },
    discord: {
        bodyType: 'json',
        template: `{
  "content": "**{{title}}**\\n{{message}}",
  "allowed_mentions": { "parse": [] }
}`,
    },
    teamsWorkflows: {
        bodyType: 'json',
        template: `{
  "type": "message",
  "attachments": [
    {
      "contentType": "application/vnd.microsoft.card.adaptive",
      "contentUrl": null,
      "content": {
        "$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
        "type": "AdaptiveCard",
        "version": "1.4",
        "body": [
          { "type": "TextBlock", "weight": "Bolder", "text": "{{title}}" },
          { "type": "TextBlock", "wrap": true, "text": "{{message}}" }
        ]
      }
    }
  ]
}`,
    },
};

export const CUSTOM_WEBHOOK_VARIABLES = ['{{title}}', '{{message}}', '{{type}}', '{{nodeName}}', '{{timestamp}}'];

const RESERVED_HEADERS = new Set([
    'connection',
    'content-length',
    'content-type',
    'host',
    'proxy-connection',
    'proxy-authorization',
    'te',
    'trailer',
    'transfer-encoding',
    'upgrade',
]);

const CUSTOM_WEBHOOK_VARIABLE_TOKENS = new Set(CUSTOM_WEBHOOK_VARIABLES);

const HEADER_NAME_PATTERN = /^[!#$%&'*+.^_`|~0-9A-Za-z-]+$/;

export const contentTypeForBodyType = (type: CustomWebhookBodyType): string => {
    if (type === 'form') return 'application/x-www-form-urlencoded';
    if (type === 'text') return 'text/plain';
    return 'application/json';
};

export const createCustomWebhookUid = (): string => {
    if (typeof globalThis.crypto?.randomUUID === 'function') {
        return globalThis.crypto.randomUUID();
    }
    const bytes = new Uint8Array(16);
    if (typeof globalThis.crypto?.getRandomValues === 'function') {
        globalThis.crypto.getRandomValues(bytes);
    } else {
        for (let index = 0; index < bytes.length; index += 1) {
            bytes[index] = Math.floor(Math.random() * 256);
        }
    }
    bytes[6] = (bytes[6] & 0x0f) | 0x40;
    bytes[8] = (bytes[8] & 0x3f) | 0x80;
    const hex = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
    return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
};

export const isCustomWebhookSecretHeader = (key: string): boolean => {
    const normalized = key.trim().toLowerCase();
    const compact = normalized.replace(/[-_]/g, '');
    return (
        normalized === 'authorization' ||
        normalized === 'cookie' ||
        normalized.endsWith('-key') ||
        normalized.endsWith('_key') ||
        compact.includes('apikey') ||
        compact.includes('token') ||
        compact.includes('secret') ||
        compact.includes('signature')
    );
};

export const createCustomWebhookFormField = (): CustomWebhookFormField => ({
    uid: createCustomWebhookUid(),
    key: '',
    value: '',
});

export const createCustomWebhookHeader = (): CustomWebhookHeaderDraft => ({
    uid: createCustomWebhookUid(),
    key: '',
    secret: false,
    configured: false,
    masked: '',
    action: 'replace',
    value: '',
});

export const createDefaultCustomWebhookDraft = (): CustomWebhookDraft => {
    const draft: CustomWebhookDraft = {
        schemaVersion: CUSTOM_WEBHOOK_SCHEMA_VERSION,
        displayName: '',
        preset: 'genericJson',
        url: {
            configured: false,
            masked: '',
            action: 'replace',
            value: '',
        },
        body: {
            type: 'json',
            template: '',
            fields: [],
        },
        headers: [],
    };
    return applyCustomWebhookPreset(draft, 'genericJson');
};

const normalizePreset = (value: unknown): CustomWebhookPreset => {
    if (value === 'genericJson' || value === 'slack' || value === 'discord' || value === 'teamsWorkflows') {
        return value;
    }
    return 'custom';
};

const normalizeBodyType = (value: unknown): CustomWebhookBodyType => {
    if (value === 'form' || value === 'text') return value;
    return 'json';
};

export const maskCustomWebhookUrl = (value: string): string => {
    return value ? '******' : '';
};

export const hydrateCustomWebhookDraft = (raw: Partial<CustomWebhookConfigView> = {}): CustomWebhookDraft => {
    const fallback = createDefaultCustomWebhookDraft();
    const preset = normalizePreset(raw.preset);
    const rawUrl = raw.url;
    const hasSanitizedUrlView = typeof rawUrl === 'object' && rawUrl !== null;
    const urlConfigured = typeof rawUrl === 'string' ? Boolean(rawUrl) : Boolean(rawUrl?.configured);
    const urlValue = typeof rawUrl === 'string' ? rawUrl : '';
    const urlMasked = typeof rawUrl === 'string' ? '' : rawUrl?.masked || '';
    const bodyType = normalizeBodyType(raw.body?.type);
    const headers = Array.isArray(raw.headers)
        ? raw.headers.map((header) => {
              const secret = Boolean(header.secret) || isCustomWebhookSecretHeader(header.key || '');
              const secretValue = secret && typeof header.value === 'string' ? header.value : '';
              const configured = secret ? Boolean(header.configured ?? secretValue) : false;
              return {
                  uid: header.uid || createCustomWebhookUid(),
                  key: header.key || '',
                  secret,
                  configured,
                  masked: secret ? header.masked || (header.value ? '***' : '') : '',
                  action: secret
                      ? secretValue
                          ? ('keep' as const)
                          : configured
                            ? ('keep' as const)
                            : ('clear' as const)
                      : ('replace' as const),
                  value: secret ? secretValue : header.value || '',
                  ...(secretValue ? { originalValue: secretValue } : {}),
              };
          })
        : [];

    return {
        schemaVersion: CUSTOM_WEBHOOK_SCHEMA_VERSION,
        ...(raw.state === 'legacy' || raw.state === 'invalid' ? { state: raw.state } : {}),
        displayName: raw.displayName || '',
        preset,
        url: {
            configured: urlConfigured,
            masked: urlMasked,
            action:
                typeof rawUrl === 'string' && rawUrl
                    ? 'keep'
                    : urlConfigured
                      ? 'keep'
                      : hasSanitizedUrlView
                        ? 'clear'
                        : 'replace',
            value: urlValue,
            ...(urlValue ? { originalValue: urlValue } : {}),
        },
        body: {
            type: bodyType,
            template:
                bodyType === 'form'
                    ? ''
                    : raw.body?.template === undefined
                      ? bodyType === 'text'
                          ? '{{message}}'
                          : fallback.body.template
                      : raw.body.template,
            fields: Array.isArray(raw.body?.fields)
                ? raw.body.fields.map((field) => ({
                      uid: field.uid || createCustomWebhookUid(),
                      key: field.key || '',
                      value: field.value || '',
                  }))
                : [],
        },
        headers,
    };
};

export const applyCustomWebhookPreset = (
    source: CustomWebhookDraft,
    preset: CustomWebhookPreset,
): CustomWebhookDraft => {
    if (preset === 'custom') {
        return { ...source, preset: 'custom' };
    }
    const definition = PRESET_DEFINITIONS[preset];
    return {
        ...source,
        preset,
        body: {
            type: definition.bodyType,
            template: definition.template,
            fields: [],
        },
    };
};

export const isCustomWebhookPresetPristine = (draft: CustomWebhookDraft): boolean => {
    if (draft.preset === 'custom') return false;
    const expected = applyCustomWebhookPreset(draft, draft.preset);
    return (
        draft.body.type === expected.body.type &&
        draft.body.template === expected.body.template &&
        draft.body.fields.length === expected.body.fields.length
    );
};

export const updateCustomWebhookBodyTemplate = (draft: CustomWebhookDraft, template: string): CustomWebhookDraft => ({
    ...draft,
    body: { ...draft.body, template },
});

const serializeSecret = (secret: CustomWebhookSecretDraft): CustomWebhookSecretMutation => {
    if (secret.action === 'replace') {
        return { action: 'replace', value: secret.value.trim() };
    }
    return { action: secret.action };
};

export const serializeCustomWebhookDraft = (draft: CustomWebhookDraft): CustomWebhookConfigUpdate => ({
    schemaVersion: CUSTOM_WEBHOOK_SCHEMA_VERSION,
    displayName: draft.displayName.trim(),
    preset: draft.preset,
    method: 'POST',
    url: serializeSecret(draft.url),
    body: {
        type: draft.body.type,
        template: draft.body.type === 'form' ? '' : draft.body.template,
        fields:
            draft.body.type === 'form'
                ? draft.body.fields.map((field) => ({
                      key: field.key.trim(),
                      value: field.value,
                  }))
                : [],
    },
    headers: draft.headers.map((header) => ({
        uid: header.uid,
        key: header.key.trim(),
        secret: header.secret,
        action: header.secret ? header.action : 'replace',
        ...(header.secret && header.action !== 'replace' ? {} : { value: header.value }),
    })),
});

const isValidHttpUrl = (value: string): boolean => {
    try {
        const parsed = new URL(value);
        return (
            (parsed.protocol === 'http:' || parsed.protocol === 'https:') &&
            !parsed.username &&
            !parsed.password &&
            !parsed.hash
        );
    } catch {
        return false;
    }
};

const isJsonTemplateValid = (value: string): boolean => {
    try {
        JSON.parse(value);
        return true;
    } catch {
        return false;
    }
};

const hasTemplateJsonKey = (value: string): boolean => {
    try {
        const visit = (node: unknown): boolean => {
            if (Array.isArray(node)) return node.some(visit);
            if (node && typeof node === 'object') {
                return Object.entries(node).some(
                    ([key, child]) => key.includes('{{') || key.includes('}}') || visit(child),
                );
            }
            return false;
        };
        return visit(JSON.parse(value));
    } catch {
        return false;
    }
};

const hasInvalidTemplateVariable = (value: string): boolean => {
    let remainder = value;
    while (remainder) {
        const start = remainder.indexOf('{{');
        if (start < 0) return remainder.includes('}}');
        if (remainder.slice(0, start).includes('}}')) return true;
        const endOffset = remainder.slice(start + 2).indexOf('}}');
        if (endOffset < 0) return true;
        const end = start + 2 + endOffset + 2;
        if (!CUSTOM_WEBHOOK_VARIABLE_TOKENS.has(remainder.slice(start, end))) return true;
        remainder = remainder.slice(end);
    }
    return false;
};

export interface CustomWebhookTemplateSelection {
    start: number;
    end: number;
}

export interface CustomWebhookTemplateInsertion {
    template: string;
    cursor: number;
}

export const insertCustomWebhookVariable = (
    template: string,
    variable: string,
    bodyType: Exclude<CustomWebhookBodyType, 'form'>,
    selection?: CustomWebhookTemplateSelection,
): CustomWebhookTemplateInsertion => {
    if (selection && selection.start >= 0 && selection.end >= selection.start && selection.end <= template.length) {
        const next = template.slice(0, selection.start) + variable + template.slice(selection.end);
        return { template: next, cursor: selection.start + variable.length };
    }
    if (bodyType === 'text') {
        return { template: template + variable, cursor: template.length + variable.length };
    }

    const stringValuePattern = /("(?:message|text|content)"\s*:\s*")((?:\\.|[^"\\])*)(")/;
    const anyStringValuePattern = /(:\s*")((?:\\.|[^"\\])*)(")/;
    const match = stringValuePattern.exec(template) || anyStringValuePattern.exec(template);
    if (match) {
        const insertAt = match.index + match[0].length - 1;
        return {
            template: template.slice(0, insertAt) + variable + template.slice(insertAt),
            cursor: insertAt + variable.length,
        };
    }

    const trimmed = template.trim();
    if (trimmed === '{}') {
        const next = `{\n  "message": "${variable}"\n}`;
        return { template: next, cursor: next.indexOf(variable) + variable.length };
    }
    if (trimmed === '[]') {
        const next = `[\n  "${variable}"\n]`;
        return { template: next, cursor: next.indexOf(variable) + variable.length };
    }
    return { template, cursor: template.length };
};

export const validateCustomWebhookDraft = (
    draft: CustomWebhookDraft,
    options: { allowClearedUrl?: boolean } = {},
): CustomWebhookValidationIssue[] => {
    const issues: CustomWebhookValidationIssue[] = [];
    if (!draft.displayName.trim()) {
        issues.push({ field: 'displayName', code: 'displayNameRequired' });
    }
    if (draft.url.action === 'keep') {
        if (!draft.url.configured) issues.push({ field: 'url', code: 'urlRequired' });
    } else if (draft.url.action === 'replace') {
        if (!draft.url.value.trim()) {
            issues.push({ field: 'url', code: 'urlRequired' });
        } else if (!isValidHttpUrl(draft.url.value.trim())) {
            issues.push({ field: 'url', code: 'urlInvalid' });
        }
    } else if (!options.allowClearedUrl) {
        issues.push({ field: 'url', code: 'urlRequired' });
    }

    if (draft.body.type === 'form') {
        if (draft.body.fields.length === 0) {
            issues.push({ field: 'body', code: 'bodyRequired' });
        }
        const keys = new Set<string>();
        draft.body.fields.forEach((field, index) => {
            const key = field.key.trim();
            if (!key) {
                issues.push({ field: `body.fields.${index}`, code: 'formFieldRequired' });
                return;
            }
            if (key.includes('{{') || key.includes('}}')) {
                issues.push({ field: `body.fields.${index}`, code: 'templateVariableInvalid' });
            }
            const normalized = key.toLowerCase();
            if (keys.has(normalized)) {
                issues.push({ field: `body.fields.${index}`, code: 'formFieldDuplicate' });
            }
            keys.add(normalized);
            if (hasInvalidTemplateVariable(field.value)) {
                issues.push({ field: `body.fields.${index}`, code: 'templateVariableInvalid' });
            }
        });
    } else if (!draft.body.template.trim()) {
        issues.push({ field: 'body', code: 'bodyRequired' });
    } else if (draft.body.type === 'json' && !isJsonTemplateValid(draft.body.template)) {
        issues.push({ field: 'body', code: 'jsonInvalid' });
    } else if (draft.body.type === 'json' && hasTemplateJsonKey(draft.body.template)) {
        issues.push({ field: 'body', code: 'templateVariableInvalid' });
    }
    if (draft.body.type !== 'form' && hasInvalidTemplateVariable(draft.body.template)) {
        issues.push({ field: 'body', code: 'templateVariableInvalid' });
    }

    const headerKeys = new Set<string>();
    draft.headers.forEach((header, index) => {
        const key = header.key.trim();
        const normalized = key.toLowerCase();
        if (!key) {
            issues.push({ field: `headers.${index}`, code: 'headerRequired' });
        } else if (!HEADER_NAME_PATTERN.test(key)) {
            issues.push({ field: `headers.${index}`, code: 'headerInvalid' });
        } else if (RESERVED_HEADERS.has(normalized)) {
            issues.push({ field: `headers.${index}`, code: 'headerReserved' });
        } else if (headerKeys.has(normalized)) {
            issues.push({ field: `headers.${index}`, code: 'headerDuplicate' });
        }
        headerKeys.add(normalized);

        if (isCustomWebhookSecretHeader(key) && !header.secret) {
            issues.push({ field: `headers.${index}`, code: 'headerMustBeSecret' });
        }

        if (header.secret) {
            if (header.action === 'keep' && !header.configured) {
                issues.push({ field: `headers.${index}`, code: 'secretRequired' });
            }
            if (header.action === 'replace' && !header.value) {
                issues.push({ field: `headers.${index}`, code: 'secretRequired' });
            }
        }
    });
    return issues;
};

export const formatCustomWebhookSafeSummary = (
    config: Partial<CustomWebhookConfigView>,
    options: { includeUrl?: boolean } = {},
): string => {
    const bodyType = normalizeBodyType(config.body?.type).toUpperCase();
    const headerCount = Array.isArray(config.headers) ? config.headers.length : 0;
    const summary = `POST · ${bodyType} · ${headerCount} Headers`;
    if (options.includeUrl === false) return summary;
    const url = typeof config.url === 'string' ? maskCustomWebhookUrl(config.url) : config.url?.masked || '***';
    return `${summary} · ${url}`;
};

export const formatCustomWebhookDetails = (
    config: Partial<CustomWebhookConfigView> | string,
    missingUrl = '-',
): string => {
    const normalized = typeof config === 'string' ? { url: config } : config;
    const summary = formatCustomWebhookSafeSummary(normalized, { includeUrl: false });
    const url = typeof normalized.url === 'string' ? normalized.url : normalized.url?.value || '';
    return `${summary} · ${url || missingUrl}`;
};
