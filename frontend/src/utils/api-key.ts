import type { APIKey } from '../api/interface/api-key';

export const ANY_API_KEY_IP = '0.0.0.0/0\n::/0';

export const normalizeAPIKeyIPs = (value: string) => {
    return value
        .split(/[\s,]+/)
        .filter(Boolean)
        .join('\n');
};

export const isAnyAPIKeyIP = (value: string) => {
    const entries = new Set(normalizeAPIKeyIPs(value).split('\n'));
    return entries.has('0.0.0.0/0') && entries.has('::/0');
};

export const getAPIKeyStatus = (item: APIKey.Item, now = Date.now()): APIKey.Status => {
    if (item.status !== 'Revoked' && item.expiresAt && new Date(item.expiresAt).getTime() <= now) {
        return 'Expired';
    }
    return item.status;
};

export const canBindAPIKey = (item: APIKey.Item, now = Date.now()) => {
    return (
        item.status === 'Enable' &&
        item.allowAppBinding &&
        (!item.expiresAt || new Date(item.expiresAt).getTime() > now)
    );
};

export const defaultAPIKeyExpiry = (now = Date.now()) => {
    return new Date(now + 90 * 24 * 60 * 60 * 1000).toISOString();
};
