import { createPinia } from 'pinia';
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate';
import GlobalStore from './modules/global';
import MenuStore from './modules/menu';
import TabsStore from './modules/tabs';

const pinia = createPinia();
pinia.use(piniaPluginPersistedstate);

export { GlobalStore, MenuStore, TabsStore };

export default pinia;
