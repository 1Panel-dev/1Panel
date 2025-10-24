import { createApp } from 'vue';
import App from './App.vue';

import '@/styles/index.scss';
import '@/styles/common.scss';
import '@/assets/iconfont/iconfont.css';
import '@/assets/iconfont/iconfont.js';
import '@/styles/style.css';

const styleModule = import.meta.glob('xpack/styles/index.scss');
for (const path in styleModule) {
    styleModule[path]?.();
}

import router from '@/routers/index';
import i18n, { ensureFallbackLocale, loadLocaleMessages } from '@/lang/index';
import pinia from '@/store/index';
import SvgIcon from './components/svg-icon/svg-icon.vue';
import Components from '@/components';

import ElementPlus from 'element-plus';
import Fit2CloudPlus from 'fit2cloud-ui-plus';
import * as Icons from '@element-plus/icons-vue';

import directives from '@/directives/index';

const bootstrap = async () => {
    const currentLocale = i18n.global.locale.value;

    await Promise.all([loadLocaleMessages(currentLocale), ensureFallbackLocale()]);

    const app = createApp(App);
    app.component('SvgIcon', SvgIcon);
    app.use(ElementPlus);
    app.use(Fit2CloudPlus, { locale: i18n.global.getLocaleMessage(currentLocale) });

    Object.keys(Icons).forEach((key) => {
        app.component(key, Icons[key as keyof typeof Icons]);
    });

    app.use(router);
    app.use(i18n);
    app.use(pinia);
    app.use(Components);
    app.use(directives);

    app.mount('#app');
};

bootstrap();
