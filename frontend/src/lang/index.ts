import { createI18n } from 'vue-i18n';

type LocaleMessage = Record<string, unknown>;
type LocaleLoader = () => Promise<{ default: LocaleMessage }>;

const DEFAULT_LOCALE = 'en';
const STORAGE_KEY = 'lang';

const LOCALE_LOADERS: Record<string, LocaleLoader> = {
    zh: () => import('./modules/zh'),
    'zh-Hant': () => import('./modules/zh-Hant'),
    en: () => import('./modules/en'),
    'pt-BR': () => import('./modules/pt-br'),
    ja: () => import('./modules/ja'),
    ru: () => import('./modules/ru'),
    ms: () => import('./modules/ms'),
    ko: () => import('./modules/ko'),
    tr: () => import('./modules/tr'),
    'es-ES': () => import('./modules/es-es'),
};

const getStoredLocale = () => {
    if (typeof window === 'undefined') return DEFAULT_LOCALE;
    return localStorage.getItem(STORAGE_KEY) || DEFAULT_LOCALE;
};

const initialLocale = getStoredLocale();

const i18n = createI18n({
    legacy: false,
    missingWarn: false,
    locale: initialLocale,
    fallbackLocale: DEFAULT_LOCALE,
    globalInjection: true,
    messages: {
        [initialLocale]: {},
    },
    warnHtmlMessage: false,
});

const loadedLocales = new Set<string>();

export const loadLocaleMessages = async (locale: string) => {
    const targetLocale = LOCALE_LOADERS[locale] ? locale : DEFAULT_LOCALE;
    if (loadedLocales.has(targetLocale)) {
        return targetLocale;
    }
    const loader = LOCALE_LOADERS[targetLocale];
    if (!loader) {
        return targetLocale;
    }
    const messagesModule = await loader();
    const messages = messagesModule.default || {};
    i18n.global.setLocaleMessage(targetLocale, messages);
    loadedLocales.add(targetLocale);
    return targetLocale;
};

export const ensureFallbackLocale = async () => {
    const fallback = i18n.global.fallbackLocale.value || DEFAULT_LOCALE;
    if (typeof fallback === 'string') {
        await loadLocaleMessages(fallback);
    }
};

export const setActiveLocale = async (locale: string) => {
    const loaded = await loadLocaleMessages(locale);
    i18n.global.locale.value = loaded;
    if (typeof window !== 'undefined') {
        localStorage.setItem(STORAGE_KEY, loaded);
    }
    return loaded;
};

export default i18n;
