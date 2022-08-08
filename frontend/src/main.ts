import { createApp } from 'vue';
import App from './App.vue';
import '@/styles/reset.scss';
import '@/styles/common.scss';
import '@/assets/iconfont/iconfont.scss';
import '@/assets/fonts/font.scss';
import ElementPlus from 'element-plus';
import * as Icons from '@element-plus/icons-vue';
import 'element-plus/dist/index.css';
import 'element-plus/theme-chalk/dark/css-vars.css';
import '@/styles/element-dark.scss';
import '@/styles/element.scss';
import directives from '@/directives/index';
import router from '@/routers/index';
import I18n from '@/lang/index';
import pinia from '@/store/index';
import { tr } from '@/utils/i18n/tr';
const app = createApp(App);
app.config.globalProperties.$tr = tr;
Object.keys(Icons).forEach((key) => {
    app.component(key, Icons[key as keyof typeof Icons]);
});

app.use(router)
    .use(I18n)
    .use(pinia)
    .use(directives)
    .use(ElementPlus)
    .mount('#app');
