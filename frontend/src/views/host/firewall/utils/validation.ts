export type FirewallAddressFamily = 'ipv4' | 'ipv6' | 'inet';

export const splitTagValues = (values: string[]): string[] => [
    ...new Set(
        values
            .flatMap((value) => value.split(/[,，;；\s]+/))
            .map((value) => value.trim())
            .filter(Boolean),
    ),
];

export const isValidIPv4Address = (value: string): boolean => {
    const parts = value.split('.');
    return (
        parts.length === 4 &&
        parts.every((part) => /^(0|[1-9]\d{0,2})$/.test(part) && Number(part) >= 0 && Number(part) <= 255)
    );
};

export const isValidIPv6Address = (value: string): boolean => {
    if (!value.includes(':') || value.includes(':::') || (value.match(/::/g) || []).length > 1) return false;

    let normalized = value;
    if (normalized.includes('.')) {
        const separator = normalized.lastIndexOf(':');
        if (separator < 0 || !isValidIPv4Address(normalized.slice(separator + 1))) return false;
        normalized = `${normalized.slice(0, separator)}:0:0`;
    }

    const compressed = normalized.includes('::');
    const [left = '', right = ''] = normalized.split('::');
    const groups = [...(left ? left.split(':') : []), ...(right ? right.split(':') : [])];
    if (groups.some((group) => !/^[\da-f]{1,4}$/i.test(group))) return false;
    return compressed ? groups.length < 8 : groups.length === 8;
};

export const inferAddressFamily = (value: string): Exclude<FirewallAddressFamily, 'inet'> =>
    value.includes(':') ? 'ipv6' : 'ipv4';

export const isValidIPOrCIDR = (value: string): boolean => {
    const slash = value.indexOf('/');
    if (slash !== value.lastIndexOf('/')) return false;
    const address = slash === -1 ? value : value.slice(0, slash);
    const prefix = slash === -1 ? undefined : value.slice(slash + 1);
    const family = inferAddressFamily(address);
    if (!(family === 'ipv6' ? isValidIPv6Address(address) : isValidIPv4Address(address))) return false;
    if (prefix === undefined) return true;
    if (!/^\d{1,3}$/.test(prefix)) return false;
    const prefixNumber = Number(prefix);
    return prefixNumber >= 0 && prefixNumber <= (family === 'ipv6' ? 128 : 32);
};

export const isValidAddressForFamily = (
    family: Exclude<FirewallAddressFamily, 'inet'>,
    value: string,
    allowCIDR = true,
): boolean => {
    if (!allowCIDR && value.includes('/')) return false;
    return isValidIPOrCIDR(value) && inferAddressFamily(value.split('/')[0]) === family;
};

export const formatHostAddress = (value?: string, family?: FirewallAddressFamily): string => {
    const address = value?.trim() || '';
    if (!address) return '';
    const addressFamily = family === 'inet' || !family ? inferAddressFamily(address.split('/')[0]) : family;
    const hostPrefix = addressFamily === 'ipv6' ? '/128' : '/32';
    if (!address.endsWith(hostPrefix)) return address;
    const host = address.slice(0, -hostPrefix.length);
    const validHost = addressFamily === 'ipv6' ? isValidIPv6Address(host) : isValidIPv4Address(host);
    return validHost ? host : address;
};

export const formatHostAddressList = (values: string[] = [], family?: FirewallAddressFamily): string =>
    values
        .map((value) => formatHostAddress(value, family))
        .filter(Boolean)
        .join(', ');

export const normalizePortRange = (value: string): string => {
    const matched = value.trim().match(/^(\d+)(?:[:-](\d+))?$/);
    if (!matched) throw new Error('invalid port range');
    const start = Number(matched[1]);
    const end = matched[2] === undefined ? start : Number(matched[2]);
    if (start < 1 || start > 65535 || end < start || end > 65535) {
        throw new Error('invalid port range');
    }
    return start === end ? String(start) : `${start}-${end}`;
};

export const isValidPortRange = (value: string): boolean => {
    try {
        normalizePortRange(value);
        return true;
    } catch {
        return false;
    }
};
