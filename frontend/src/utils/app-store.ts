import i18n from '@/lang';

export function getLanguage() {
    return localStorage.getItem('lang') || 'zh';
}

function normalizeAppLocaleKey(language: string) {
    const localeMap: Record<string, string> = {
        tw: 'zh-hant',
        'zh-Hant': 'zh-hant',
        'pt-BR': 'pt-br',
        'es-ES': 'es-es',
    };
    return localeMap[language] || language.toLowerCase();
}

export function getLabel(row: any) {
    const language = localStorage.getItem('lang') || 'zh';
    const lang = normalizeAppLocaleKey(language);
    if (row.label && typeof row.label[lang] === 'string' && row.label[lang] !== '') {
        return row.label[lang];
    }
    if (language == 'zh' || language == 'tw') {
        return row.labelZh;
    }
    return row.labelEn;
}

export function getDescription(row: any) {
    const language = localStorage.getItem('lang') || 'zh';
    const lang = normalizeAppLocaleKey(language);
    if (row.description && typeof row.description[lang] === 'string' && row.description[lang] !== '') {
        return row.description[lang];
    }
    return '';
}

export function getDBName(length: number): string {
    const chars = 'abcdefhijkmnprstwxyz2345678';
    let result = '';
    for (let i = 0; i < length; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
}

export function getRuntimeLabel(type: string) {
    if (type == 'appstore') {
        return i18n.global.t('menu.apps');
    }
    if (type == 'local') {
        return i18n.global.t('commons.table.local');
    }
    return '';
}
