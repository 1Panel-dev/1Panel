import { createI18n } from 'vue-i18n';
import zh from './modules/zh';
import tw from './modules/tw';
import en from './modules/en';

const i18n = createI18n({
    legacy: false,
    locale: localStorage.getItem('lang') || 'zh',
    globalInjection: true,
    messages: {
        zh,
        tw,
        en,
    },
    warnHtmlMessage: false,
});

export default i18n;
