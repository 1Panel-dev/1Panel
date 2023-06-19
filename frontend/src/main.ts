import { createApp } from 'vue';
import App from './App.vue';

import '@/styles/index.scss';
import '@/styles/common.scss';
import '@/assets/iconfont/iconfont.css';
import '@/assets/iconfont/iconfont.js';
import '@/styles/style.css';

import directives from '@/directives/index';
import router from '@/routers/index';
import I18n from '@/lang/index';
import pinia from '@/store/index';
import SvgIcon from './components/svg-icon/svg-icon.vue';
import Components from '@/components';

import ElementPlus from 'element-plus';
import Fit2CloudPlus from 'fit2cloud-ui-plus';
import * as Icons from '@element-plus/icons-vue';
const app = createApp(App);
app.component('SvgIcon', SvgIcon);
app.use(ElementPlus);

app.use(Fit2CloudPlus, { locale: I18n.global.messages.value[localStorage.getItem('lang') || 'zh'] });

Object.keys(Icons).forEach((key) => {
    app.component(key, Icons[key as keyof typeof Icons]);
});

app.use(router);
app.use(I18n);
app.use(pinia);
app.use(directives);
app.use(Components);
app.mount('#app');
