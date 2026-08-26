export const footerNavigationKeys = ['learnMore', 'forum', 'documentation', 'project'] as const;

export type FooterNavigationKey = (typeof footerNavigationKeys)[number];

export interface FooterNavigationLinkSetting {
    visible: boolean;
    url: string;
}

export type FooterNavigationLinks = Record<FooterNavigationKey, FooterNavigationLinkSetting>;

export interface FooterNavigationSetting {
    customized: boolean;
    links: FooterNavigationLinks;
}

export interface FooterNavigationSettingEditor {
    validate: () => Promise<boolean>;
    save: () => Promise<void>;
    restoreDefaults: () => Promise<void>;
    reload: () => Promise<boolean>;
    isDirty: () => boolean;
}

export const createDefaultFooterNavigationLinks = (isIntl: boolean, docsUrl: string): FooterNavigationLinks => ({
    learnMore: {
        visible: true,
        url: isIntl ? 'https://1panel.pro/pricing' : 'https://1panel.cn/versions.html',
    },
    forum: {
        visible: true,
        url: isIntl ? 'https://github.com/1Panel-dev/1Panel/discussions' : 'https://bbs.fit2cloud.com/c/1p/7',
    },
    documentation: {
        visible: true,
        url: docsUrl.endsWith('/') ? docsUrl : `${docsUrl}/`,
    },
    project: {
        visible: true,
        url: 'https://github.com/1Panel-dev/1Panel',
    },
});

const controlCharacterPattern = /[\u0000-\u001f\u007f-\u009f]/u;
const schemeAuthorityPattern = /^[a-z][a-z\d+.-]*:\/\/([^/?#]*)/i;
const percentBytePattern = /^[\da-f]{2}$/i;

const containsControlCharacter = (value: string) => controlCharacterPattern.test(value);

const decodePercentEscapes = (value: string): string | null => {
    if (!value.includes('%')) {
        return value;
    }

    const encoder = new TextEncoder();
    const bytes: number[] = [];
    let segmentStart = 0;
    for (let index = 0; index < value.length; index += 1) {
        if (value[index] !== '%') {
            continue;
        }
        bytes.push(...encoder.encode(value.slice(segmentStart, index)));
        const escapedByte = value.slice(index + 1, index + 3);
        if (!percentBytePattern.test(escapedByte)) {
            return null;
        }
        bytes.push(Number.parseInt(escapedByte, 16));
        index += 2;
        segmentStart = index + 1;
    }
    bytes.push(...encoder.encode(value.slice(segmentStart)));
    return new TextDecoder().decode(Uint8Array.from(bytes));
};

export const isSafeExternalUrl = (value: unknown): value is string => {
    if (typeof value !== 'string') {
        return false;
    }
    if (containsControlCharacter(value)) {
        return false;
    }

    const normalizedURL = value.trim();
    if (normalizedURL.includes('\\')) {
        return false;
    }
    const authority = normalizedURL.match(schemeAuthorityPattern)?.[1];
    if (!authority || authority.endsWith(':') || authority.includes('@') || authority.includes('%')) {
        return false;
    }

    const decodedURL = decodePercentEscapes(normalizedURL);
    if (decodedURL === null || containsControlCharacter(decodedURL)) {
        return false;
    }

    try {
        const url = new URL(normalizedURL);
        const port = url.port ? Number(url.port) : null;
        return (
            (url.protocol === 'http:' || url.protocol === 'https:') &&
            Boolean(url.hostname) &&
            !url.username &&
            !url.password &&
            (port === null || (Number.isInteger(port) && port >= 1 && port <= 65535))
        );
    } catch {
        return false;
    }
};

export const mergeFooterNavigationLinks = (
    setting: FooterNavigationSetting | null,
    defaults: FooterNavigationLinks,
): FooterNavigationLinks => {
    if (!setting?.customized || !setting.links) {
        return defaults;
    }

    return footerNavigationKeys.reduce((links, key) => {
        const customized = setting.links[key];
        links[key] = {
            visible: typeof customized?.visible === 'boolean' ? customized.visible : defaults[key].visible,
            url: isSafeExternalUrl(customized?.url) ? customized.url : defaults[key].url,
        };
        return links;
    }, {} as FooterNavigationLinks);
};
