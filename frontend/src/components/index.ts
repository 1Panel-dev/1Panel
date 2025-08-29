import { type App } from 'vue';
import LayoutContent from './layout-content/index.vue';
import RouterButton from './router-button/index.vue';
import ComplexTable from './complex-table/index.vue';
import OpDialog from './del-dialog/index.vue';
import TableSearch from './table-search/index.vue';
import TableSetting from './table-setting/index.vue';
import TableRefresh from './table-refresh/index.vue';
import CopyButton from '@/components/copy-button/index.vue';
import MsgInfo from '@/components/msg-info/index.vue';
import DrawerPro from '@/components/drawer-pro/index.vue';
import DialogPro from '@/components/dialog-pro/index.vue';
export default {
    install(app: App) {
        app.component(LayoutContent.name, LayoutContent);
        app.component(RouterButton.name, RouterButton);
        app.component(ComplexTable.name, ComplexTable);
        app.component(OpDialog.name, OpDialog);
        app.component(CopyButton.name, CopyButton);
        app.component(TableSearch.name, TableSearch);
        app.component(TableSetting.name, TableSetting);
        app.component(TableRefresh.name, TableRefresh);
        app.component(MsgInfo.name, MsgInfo);
        app.component(DrawerPro.name, DrawerPro);
        app.component(DialogPro.name, DialogPro);
    },
};
