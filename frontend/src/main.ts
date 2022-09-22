import { createApp } from 'vue';
import App from './App.vue';
import '@/styles/reset.scss';
import '@/styles/common.scss';
import '@/assets/iconfont/iconfont.css';
import '@/assets/iconfont/iconfont.js';
import ElementPlus from 'element-plus';
import Fit2CloudPlus from 'fit2cloud-ui-plus';
import * as Icons from '@element-plus/icons-vue';
import '@/styles/element-dark.scss';
import '@/styles/element.scss';
import 'element-plus/dist/index.css';
import 'element-plus/theme-chalk/dark/css-vars.css';
import 'fit2cloud-ui-plus/src/styles/index.scss';
import directives from '@/directives/index';
import router from '@/routers/index';
import I18n from '@/lang/index';
import pinia from '@/store/index';
import SvgIcon from './components/svg-icon/svg-icon.vue';
import VMdPreview from '@kangc/v-md-editor/lib/preview';
import '@kangc/v-md-editor/lib/style/preview.css';
import githubTheme from '@kangc/v-md-editor/lib/theme/github.js';
import '@kangc/v-md-editor/lib/theme/style/github.css';
import hljs from 'highlight.js';

VMdPreview.use(githubTheme, {
    hljs,
});

const app = createApp(App);
app.component('SvgIcon', SvgIcon);
app.use(ElementPlus);
app.use(Fit2CloudPlus);
Object.keys(Icons).forEach((key) => {
    app.component(key, Icons[key as keyof typeof Icons]);
});

app.use(router);
app.use(I18n);
app.use(pinia);
app.use(directives);
app.use(VMdPreview);
app.mount('#app');
