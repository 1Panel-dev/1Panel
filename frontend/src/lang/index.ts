import { createI18n } from 'vue-i18n';
import zh from './modules/zh';
import tw from './modules/tw';
import en from './modules/en';
import ptBr from './modules/pt-br';
import ja from './modules/ja';
import ru from './modules/ru';
import ms from './modules/ms';
import ko from './modules/ko';

const i18n = createI18n({
    legacy: false,
    locale: localStorage.getItem('lang') || 'en',
    fallbackLocale: 'en',
    globalInjection: true,
    messages: {
        zh,
        tw,
        en,
        'pt-BR': ptBr,
        ja,
        ru,
        ms,
        ko,
    },
    warnHtmlMessage: false,
});

export default i18n;
