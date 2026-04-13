export const toLink = (link: string) => {
    const ipv6Regex = /^https?:\/\/([a-f0-9:]+):(\d+)(\/?.*)?$/i;
    try {
        if (ipv6Regex.test(link)) {
            const match = link.match(ipv6Regex);
            if (match) {
                const ipv6 = match[1];
                const port = match[2];
                const path = match[3] || '';
                link = `${link.startsWith('https') ? 'https' : 'http'}://[${ipv6}]:${port}${path}`;
            }
        }
        window.open(link, '_blank');
    } catch (e) {}
};

export const preloadImage = (url: string): Promise<string> => {
    return new Promise((resolve) => {
        const img = new Image();
        img.onload = () => resolve(url);
        img.onerror = () => resolve(url);
        img.src = url;
    });
};
