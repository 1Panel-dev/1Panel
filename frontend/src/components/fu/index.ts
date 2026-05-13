import { type App } from 'vue';

import FuInputRwSwitch from './FuInputRwSwitch.vue';
import FuDropdownItem from './FuDropdownItem.vue';
import FuReadWriteSwitch from './FuReadWriteSwitch.vue';
import FuSelectRwSwitch from './FuSelectRwSwitch.vue';
import FuStep from './FuStep';
import FuSteps from './FuSteps';
import FuTable from './FuTable';
import FuTableColumnSelect from './FuTableColumnSelect.vue';
import FuTableOperations from './FuTableOperations.vue';
import FuTablePagination from './FuTablePagination.vue';

const components = [
    FuTable,
    FuTableOperations,
    FuTablePagination,
    FuDropdownItem,
    FuInputRwSwitch,
    FuReadWriteSwitch,
    FuSelectRwSwitch,
    FuSteps,
    FuStep,
    FuTableColumnSelect,
];

export default {
    install(app: App) {
        components.forEach((component) => {
            app.component((component as any).name, component as any);
        });
    },
};
