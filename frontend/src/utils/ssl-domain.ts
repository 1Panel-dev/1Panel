interface CertificateDomains {
    primaryDomain?: string;
    domains?: string;
}

export function normalizeDomain(value = ''): string {
    let host = value.trim().toLowerCase();
    if (host.startsWith('[')) {
        host = host.slice(1, host.indexOf(']'));
    } else if ((host.match(/:/g) || []).length === 1) {
        host = host.replace(/:\d+$/, '');
    }
    host = host.replace(/\.$/, '');
    if (!host) return '';
    try {
        // URL normalizes internationalized hostnames and equivalent IPv6 spellings.
        const normalized = new URL(`http://${host.includes(':') ? `[${host}]` : host}`).hostname;
        return normalized.replace(/^\[|\]$/g, '');
    } catch {
        return '';
    }
}

export function domainMatches(pattern: string, target: string): boolean {
    pattern = normalizeDomain(pattern);
    target = normalizeDomain(target);
    if (!pattern || !target) return false;
    if (pattern === target) return true;
    if (!pattern.startsWith('*.') || target.includes(':') || /^\d+(\.\d+){3}$/.test(target)) return false;
    return target.endsWith(pattern.slice(1)) && target.split('.').length === pattern.split('.').length;
}

export function getCertificateDomains(ssl: CertificateDomains): string[] {
    return [
        ...new Set(
            [ssl.primaryDomain || '', ...(ssl.domains || '').split(/[,\s]+/)].map(normalizeDomain).filter(Boolean),
        ),
    ];
}

export function certificateDomainMismatch(domains: string[], ssl?: CertificateDomains): boolean {
    if (!ssl) return false;
    const patterns = getCertificateDomains(ssl);
    const hosts = domains.map(normalizeDomain).filter(Boolean);
    if (!patterns.length || !hosts.length) return false;
    return hosts.some((host) => !patterns.some((pattern) => domainMatches(pattern, host)));
}
