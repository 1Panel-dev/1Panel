import { toUnicode } from 'punycode';

export function deepCopy<T>(obj: any): T {
    let newObj: any;
    try {
        newObj = obj.push ? [] : {};
    } catch (error) {
        newObj = {};
    }
    for (let attr in obj) {
        if (typeof obj[attr] === 'object') {
            newObj[attr] = deepCopy(obj[attr]);
        } else {
            newObj[attr] = obj[attr];
        }
    }
    return newObj;
}

export function debounce(func: Function, wait: number) {
    let timeout: NodeJS.Timeout | null = null;
    return function executedFunction(...args: any[]) {
        const later = () => {
            if (timeout) clearTimeout(timeout);
            func(...args);
        };
        if (timeout) clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

export function isJson(str: string) {
    try {
        if (typeof JSON.parse(str) === 'object') {
            return true;
        }
    } catch {
        return false;
    }
}

export function getErrorMessage(error: unknown): string {
    if (typeof error === 'string') {
        return error;
    }
    if (error instanceof Error) {
        return error.message || String(error);
    }
    if (error && typeof error === 'object') {
        const err = error as Record<string, any>;
        return (
            err.message ||
            err.msg ||
            err.error?.message ||
            err.response?.data?.message ||
            err.response?.data?.msg ||
            err.data?.message ||
            err.data?.msg ||
            JSON.stringify(error)
        );
    }
    return String(error);
}

function compareBySort(a: any, b: any) {
    return a.sort - b.sort;
}

export function sortMenu(arr: any[]) {
    if (!arr || arr.length === 0) {
        return;
    }
    arr.forEach((item) => {
        if (item.children && Array.isArray(item.children)) {
            item.children.sort(compareBySort);
        }
    });
    arr.sort(compareBySort);
}

export function isPunycoded(domain: string): boolean {
    return domain.includes('xn--');
}

export function GetPunyCodeDomain(domain: string): string {
    if (!domain || typeof domain !== 'string') {
        return '';
    }
    const lowerDomain = domain.toLowerCase();
    if (!lowerDomain.includes('xn--')) {
        return '';
    }
    try {
        const decoded = toUnicode(domain);
        if (decoded === domain) {
            return '';
        }
        return decoded;
    } catch (error) {
        return '';
    }
}
