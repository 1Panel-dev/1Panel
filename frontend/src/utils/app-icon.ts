export function isRemoteAppIcon(icon: string): boolean {
    if (
        !/^https?:\/\/[^/?#]/i.test(icon) ||
        /^https?:\/\/[^/?#]*@/i.test(icon) ||
        Array.from(icon).some(
            (char) => char <= ' ' || (char >= '\u007f' && char <= '\u009f') || /\s/.test(char) || char === '\\',
        ) ||
        /%(?![\da-f]{2})/i.test(icon)
    ) {
        return false;
    }
    try {
        const url = new URL(icon);
        return !!url.hostname && !url.username && !url.password;
    } catch {
        return false;
    }
}

export function resolveAppIconUrl(appKey: string, node: string | undefined, icon: string, baseURL: string): string {
    if (isRemoteAppIcon(icon)) {
        return icon;
    }
    const params = node ? `?operateNode=${encodeURIComponent(node)}` : '';
    return `${baseURL}/apps/icon/${encodeURIComponent(appKey)}${params}`;
}
