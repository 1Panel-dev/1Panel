import { createI18n } from 'vue-i18n';
import zh from './modules/zh';
import zhHant from './modules/zh-Hant';
import en from './modules/en';
import ptBr from './modules/pt-br';
import ja from './modules/ja';
import ru from './modules/ru';
import ms from './modules/ms';
import ko from './modules/ko';
import tr from './modules/tr';

const i18n = createI18n({
    legacy: false,
    missingWarn: false,
    locale: localStorage.getItem('lang') || 'en',
    fallbackLocale: 'en',
    globalInjection: true,
    messages: {
        zh,
        'zh-Hant': zhHant,
        en,
        'pt-BR': ptBr,
        ja,
        ru,
        ms,
        ko,
        tr,
    },
    warnHtmlMessage: false,
});

export default i18n;
