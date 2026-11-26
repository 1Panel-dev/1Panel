import { checkIpV4V6 } from '@/utils/validate';

export const validNetworkIP = (value: string) => !!value && !value.includes('%') && !checkIpV4V6(value);

export const validNetworkCIDR = (value: string) => {
    const parts = value.split('/');
    return (
        parts.length === 2 &&
        validNetworkIP(parts[0]) &&
        /^(0|[1-9][0-9]{0,2})$/.test(parts[1]) &&
        Number(parts[1]) <= (parts[0].includes(':') ? 128 : 32)
    );
};
