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

import directives from '@/directives/index';
import router from '@/routers/index';
import i18n from '@/lang/index';
import pinia from '@/store/index';
import SvgIcon from './components/svg-icon/svg-icon.vue';
import Components from '@/components';

import ElementPlus from 'element-plus';
import Fit2CloudPlus from 'fit2cloud-ui-plus';
import * as Icons from '@element-plus/icons-vue';
import VueDiff from 'vue-diff';
import 'vue-diff/dist/index.css';
import yaml from 'highlight.js/lib/languages/yaml';
VueDiff.hljs.registerLanguage('yaml', yaml);

const app = createApp(App);
app.component('SvgIcon', SvgIcon);
app.use(ElementPlus);
app.use(VueDiff);
app.use(Fit2CloudPlus, { locale: i18n.global.messages.value[localStorage.getItem('lang') || 'zh'] });

Object.keys(Icons).forEach((key) => {
    app.component(key, Icons[key as keyof typeof Icons]);
});

app.use(router);
app.use(i18n);
app.use(pinia);
app.use(directives);
app.use(Components);
app.mount('#app');
