import { type App } from 'vue';
import LayoutContent from './layout-content/index.vue';
import RouterButton from './router-button/index.vue';
import ComplexTable from './complex-table/index.vue';
import ErrPrompt from './error-prompt/index.vue';
import OpDialog from './del-dialog/index.vue';
import Tooltip from '@/components/tooltip/index.vue';
import CopyButton from '@/components/copy-button/index.vue';
export default {
    install(app: App) {
        app.component(LayoutContent.name, LayoutContent);
        app.component(RouterButton.name, RouterButton);
        app.component(ComplexTable.name, ComplexTable);
        app.component(ErrPrompt.name, ErrPrompt);
        app.component(OpDialog.name, OpDialog);
        app.component(Tooltip.name, Tooltip);
        app.component(CopyButton.name, CopyButton);
    },
};
