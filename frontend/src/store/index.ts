import { createPinia } from 'pinia';
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate';
import GlobalStore from './modules/global';
import MenuStore from './modules/menu';
import TabsStore from './modules/tabs';
import TerminalStore from './modules/terminal';
import TerminalSessionStore, { TerminalDockSessionStore } from './modules/terminal-session';
import ProcessStore from './modules/process';

const pinia = createPinia();
pinia.use(piniaPluginPersistedstate);

export {
    GlobalStore,
    MenuStore,
    TabsStore,
    TerminalStore,
    TerminalSessionStore,
    TerminalDockSessionStore,
    ProcessStore,
};

export default pinia;
