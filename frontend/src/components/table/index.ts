import type { App } from 'vue';

import FuTable from './Table';
import FuTableColumnSelect from './TableColumnSelect.vue';
import FuTableOperations from './TableOperations.vue';
import FuTablePagination from './TablePagination.vue';
import TableRefresh from './TableRefresh.vue';
import TableSearch from './TableSearch.vue';
import TableSetting from './TableSetting.vue';
import TableViewSwitch from './TableViewSwitch.vue';

const components = [
    FuTable,
    FuTableOperations,
    FuTablePagination,
    FuTableColumnSelect,
    TableSearch,
    TableSetting,
    TableRefresh,
    TableViewSwitch,
];

export default {
    install(app: App) {
        components.forEach((component) => {
            app.component((component as any).name, component as any);
        });
    },
};
