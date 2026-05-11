import { ElTable } from 'element-plus';
import { cloneVNode, computed, defineComponent, h, onMounted, ref, watch, type PropType, type VNode } from 'vue';

import {
    FU_TABLE_STORAGE_PREFIX,
    flattenVNodes,
    isElTableColumnVNode,
    type FuTableColumnConfig,
    updateArrayInPlace,
} from './shared';

const buildColumnKey = (vnode: VNode, index: number) => {
    const props = (vnode.props || {}) as Record<string, any>;
    return String(props.columnKey || props.prop || props.label || `column-${index}`);
};

const buildColumnConfig = (vnode: VNode, index: number): FuTableColumnConfig => {
    const props = (vnode.props || {}) as Record<string, any>;
    return {
        key: buildColumnKey(vnode, index),
        label: String(props.label || props.prop || `column-${index}`),
        prop: props.prop,
        show: true,
        fixed: props.fixed ?? props.fix,
    };
};

const isConfigurableColumn = (vnode: VNode) => {
    if (!isElTableColumnVNode(vnode)) {
        return false;
    }
    const props = (vnode.props || {}) as Record<string, any>;
    if (['selection', 'index', 'expand'].includes(props.type)) {
        return false;
    }
    return Boolean(props.label || props.prop);
};

export default defineComponent({
    name: 'FuTable',
    props: {
        columns: {
            type: Array as PropType<FuTableColumnConfig[]>,
            default: undefined,
        },
        localKey: {
            type: String,
            default: '',
        },
        refresh: {
            type: Boolean,
            default: true,
        },
        refTable: {
            type: Object as PropType<{ value?: unknown } | null>,
            default: null,
        },
    },
    setup(props, { attrs, slots, expose }) {
        const refElTable = ref();
        const storageLoaded = ref(false);
        const storedColumns = ref<FuTableColumnConfig[]>([]);

        const slotChildren = computed(() => flattenVNodes(slots.default?.() || []));

        const slotColumns = computed(() =>
            slotChildren.value.filter(isConfigurableColumn).map((vnode, index) => buildColumnConfig(vnode, index)),
        );

        const ensureStoredColumns = () => {
            if (storageLoaded.value || !props.localKey || typeof window === 'undefined') {
                storageLoaded.value = true;
                return;
            }
            storageLoaded.value = true;
            try {
                const raw = localStorage.getItem(`${FU_TABLE_STORAGE_PREFIX}${props.localKey}`);
                if (!raw) {
                    return;
                }
                const parsed = JSON.parse(raw);
                if (Array.isArray(parsed)) {
                    storedColumns.value = parsed;
                }
            } catch {
                storedColumns.value = [];
            }
        };

        const syncColumns = (nextColumns: FuTableColumnConfig[]) => {
            if (!props.columns) {
                return;
            }
            ensureStoredColumns();
            const currentColumns = props.columns as FuTableColumnConfig[];
            const sourceColumns = currentColumns.length > 0 ? currentColumns : storedColumns.value;
            const sourceByKey = new Map(sourceColumns.map((column) => [column.key, column]));
            const sourceOrder = new Map(sourceColumns.map((column, index) => [column.key, index]));

            const mergedColumns = nextColumns
                .map((column, index) => ({
                    ...column,
                    show: sourceByKey.get(column.key)?.show ?? true,
                    fixed: sourceByKey.get(column.key)?.fixed ?? column.fixed,
                    __order: sourceOrder.get(column.key) ?? index,
                }))
                .sort((a, b) => a.__order - b.__order)
                .map((column) => {
                    const normalizedColumn = { ...column };
                    delete (normalizedColumn as Partial<typeof normalizedColumn> & { __order?: number }).__order;
                    return normalizedColumn;
                });

            updateArrayInPlace(currentColumns, mergedColumns);
        };

        const resolvedColumns = computed(() => {
            if (props.columns && props.columns.length > 0) {
                return props.columns;
            }
            if (storedColumns.value.length > 0) {
                const storageMap = new Map(storedColumns.value.map((column) => [column.key, column]));
                const orderMap = new Map(storedColumns.value.map((column, index) => [column.key, index]));
                return [...slotColumns.value]
                    .map((column, index) => ({
                        ...column,
                        show: storageMap.get(column.key)?.show ?? true,
                        __order: orderMap.get(column.key) ?? index,
                    }))
                    .sort((a, b) => a.__order - b.__order)
                    .map((column) => {
                        const normalizedColumn = { ...column };
                        delete (normalizedColumn as Partial<typeof normalizedColumn> & { __order?: number }).__order;
                        return normalizedColumn;
                    });
            }
            return slotColumns.value;
        });

        const orderedConfigurableChildren = computed(() => {
            const columnNodeMap = new Map(
                slotChildren.value
                    .filter(isConfigurableColumn)
                    .map((vnode, index) => [buildColumnKey(vnode, index), vnode]),
            );
            return resolvedColumns.value
                .filter((column) => column.show !== false)
                .map((column) => {
                    const vnode = columnNodeMap.get(column.key);
                    if (!vnode) {
                        return null;
                    }
                    return cloneVNode(vnode, {
                        key: column.key,
                    });
                })
                .filter((column): column is VNode => Boolean(column));
        });

        const renderedChildren = computed(() => {
            if (orderedConfigurableChildren.value.length === 0) {
                return slotChildren.value;
            }
            const children: VNode[] = [];
            let hasInsertedConfigurableColumns = false;
            for (const child of slotChildren.value) {
                if (isConfigurableColumn(child)) {
                    if (!hasInsertedConfigurableColumns) {
                        children.push(...orderedConfigurableChildren.value);
                        hasInsertedConfigurableColumns = true;
                    }
                    continue;
                }
                children.push(child);
            }
            return children;
        });

        watch(slotColumns, (nextColumns) => syncColumns(nextColumns), {
            immediate: true,
            deep: true,
        });

        watch(
            () => props.columns,
            (nextColumns) => {
                if (!props.localKey || !nextColumns || typeof window === 'undefined') {
                    return;
                }
                localStorage.setItem(`${FU_TABLE_STORAGE_PREFIX}${props.localKey}`, JSON.stringify(nextColumns));
            },
            {
                deep: true,
            },
        );

        onMounted(() => {
            ensureStoredColumns();
            if (props.refTable) {
                props.refTable.value = refElTable.value;
            }
        });

        expose({ refElTable });

        return () =>
            h(
                ElTable as any,
                {
                    ...attrs,
                    ref: refElTable,
                    class: ['fu-table', attrs.class],
                    headerRowClassName: 'fu-table-header',
                },
                {
                    ...slots,
                    default: () => renderedChildren.value,
                },
            );
    },
});
