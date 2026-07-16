import { type App } from 'vue';
import LayoutContent from './layout-content/index.vue';
import RouterButton from './router-button/index.vue';
import ComplexTable from './table/complex/ComplexTable.vue';
import OpDialog from './del-dialog/index.vue';
import CopyButton from '@/components/copy-button/index.vue';
import MsgInfo from '@/components/msg-info/index.vue';
import DrawerPro from '@/components/drawer-pro/index.vue';
import DialogPro from '@/components/dialog-pro/index.vue';
import FuDropdownItem from '@/components/fu/FuDropdownItem.vue';
import FuInputRwSwitch from '@/components/fu/FuInputRwSwitch.vue';
import FuReadWriteSwitch from '@/components/fu/FuReadWriteSwitch.vue';
import FuSelectRwSwitch from '@/components/fu/FuSelectRwSwitch.vue';
import FuStep from '@/components/fu/FuStep';
import FuSteps from '@/components/fu/FuSteps';
import TableComponents from '@/components/table';
export default {
    install(app: App) {
        app.use(TableComponents);
        app.component(LayoutContent.name, LayoutContent);
        app.component(RouterButton.name, RouterButton);
        app.component(ComplexTable.name, ComplexTable);
        app.component(OpDialog.name, OpDialog);
        app.component(CopyButton.name, CopyButton);
        app.component(MsgInfo.name, MsgInfo);
        app.component(DrawerPro.name, DrawerPro);
        app.component(DialogPro.name, DialogPro);
        [FuDropdownItem, FuInputRwSwitch, FuReadWriteSwitch, FuSelectRwSwitch, FuSteps, FuStep].forEach((component) => {
            app.component((component as any).name, component as any);
        });
    },
};
