<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.selectNode')" @close="handleClose">
        <template #content>
            <div class="node-drawer">
                <el-input
                    v-model="keyword"
                    :placeholder="$t('xpack.node.searchNodePlaceholder')"
                    class="node-search-input"
                    clearable
                />
                <div class="node-grid">
                    <section v-for="group in nodeGroups" :key="group.key" class="node-section">
                        <div class="node-table">
                            <div class="node-table__header">
                                <div class="node-table__cell node-table__cell--node">
                                    <span class="node-table__group-title">{{ group.label }}</span>
                                    <span class="node-table__summary">{{ group.items.length }}</span>
                                </div>
                                <div class="node-table__cell node-table__cell--addr">{{ $t('home.ip') }}</div>
                            </div>
                            <div v-for="row in group.items" :key="row.name" class="node-table__row">
                                <div class="node-table__cell node-table__cell--node">
                                    <el-tooltip
                                        :content="
                                            row.isFavorite ? $t('website.cancelFavorite') : $t('website.favorite')
                                        "
                                        placement="left"
                                    >
                                        <el-button
                                            class="node-favorite-button"
                                            link
                                            v-if="isAdmin"
                                            :type="row.isFavorite ? 'warning' : 'info'"
                                            :icon="row.isFavorite ? 'StarFilled' : 'Star'"
                                            :loading="favoriteLoadingIDs.includes(row.id)"
                                            @click.stop="toggleFavorite(row)"
                                        />
                                    </el-tooltip>
                                    <div class="node-name-cell">
                                        <SvgIcon class="icon" iconName="p-zhuji" />
                                        <el-button
                                            class="node-name-button"
                                            link
                                            type="primary"
                                            @click="selectNode(row.name)"
                                        >
                                            {{ displayNodeName(row) }}
                                        </el-button>
                                        <el-tooltip
                                            v-if="row.status !== 'Healthy' || !row.isBound"
                                            :content="
                                                row.isBound
                                                    ? $t('xpack.node.nodeUnhealthy')
                                                    : $t('xpack.node.nodeUnbind')
                                            "
                                            placement="left"
                                        >
                                            <span class="node-status-wrap">
                                                <el-icon class="icon-status" type="danger">
                                                    <Warning />
                                                </el-icon>
                                            </span>
                                        </el-tooltip>
                                    </div>
                                </div>
                                <div class="node-table__cell node-table__cell--addr">
                                    <span class="node-table__addr-text">{{ row.addr }}</span>
                                </div>
                            </div>
                        </div>
                    </section>
                </div>
            </div>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { Setting } from '@/api/interface/setting';
import { updateNodeFavorite } from '@/api/modules/setting';
import DrawerPro from '@/components/drawer-pro/index.vue';
import i18n from '@/lang';
import { computed, ref, watch } from 'vue';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isAdmin } = useGlobalStore();

type NodeItem = Setting.NodeItem & {
    group?: string;
};

const props = defineProps<{
    modelValue: boolean;
    nodes: NodeItem[];
    masterAlias: string;
}>();

const emit = defineEmits<{
    (event: 'update:modelValue', value: boolean): void;
    (event: 'change', value: string): void;
    (event: 'favorite-change'): void;
}>();

const keyword = ref('');
const favoriteLoadingIDs = ref<number[]>([]);

const drawerVisible = computed({
    get: () => props.modelValue,
    set: (value: boolean) => emit('update:modelValue', value),
});

const filteredNodes = computed(() => {
    const searchValue = keyword.value.trim().toLowerCase();
    return props.nodes.filter((item) => {
        const group = displayGroupName(item);
        if (!searchValue) {
            return true;
        }
        return [item.name, group, item.addr].some((value) =>
            String(value || '')
                .toLowerCase()
                .includes(searchValue),
        );
    });
});

const sortedNodes = computed(() => {
    return [...filteredNodes.value];
});

const commonNodeLimit = 8;

const commonNodes = computed(() => {
    return sortedNodes.value.slice(0, commonNodeLimit);
});

const nodeGroups = computed(() => {
    const groupMap = new Map<string, { key: string; label: string; items: NodeItem[] }>();
    const commonNodeIDs = new Set(commonNodes.value.map((item) => item.id));
    for (const item of sortedNodes.value.filter((item) => !commonNodeIDs.has(item.id))) {
        const label = displayGroupName(item);
        if (!groupMap.has(label)) {
            groupMap.set(label, {
                key: label,
                label,
                items: [],
            });
        }
        groupMap.get(label)!.items.push(item);
    }
    const groups = Array.from(groupMap.values()).sort((a, b) => a.label.localeCompare(b.label));
    if (commonNodes.value.length > 0) {
        groups.unshift({
            key: 'common',
            label: i18n.global.t('xpack.node.commonNodes'),
            items: commonNodes.value,
        });
    }
    return groups;
});

const displayNodeName = (item: NodeItem) => {
    return item.name === 'local' ? props.masterAlias : item.name;
};

const displayGroupName = (item: NodeItem) => {
    if (item.groupBelong === 'Default') {
        return i18n.global.t('commons.table.default');
    }
    return item.group || item.groupBelong || i18n.global.t('commons.table.default');
};

const selectNode = (name: string) => {
    emit('change', name);
    emit('update:modelValue', false);
};

const toggleFavorite = async (row: NodeItem) => {
    if (favoriteLoadingIDs.value.includes(row.id)) {
        return;
    }
    favoriteLoadingIDs.value.push(row.id);
    try {
        await updateNodeFavorite(row.id, !row.isFavorite);
        emit('favorite-change');
    } finally {
        favoriteLoadingIDs.value = favoriteLoadingIDs.value.filter((id) => id !== row.id);
    }
};

const handleClose = () => {
    emit('update:modelValue', false);
};

watch(
    () => props.modelValue,
    (value) => {
        if (value) {
            keyword.value = '';
        }
    },
);
</script>

<style scoped lang="scss">
.node-drawer {
    display: flex;
    flex-direction: column;
    gap: 10px;
    height: calc(100vh - 104px);
    padding: 0 16px 16px;
    box-sizing: border-box;
}

.node-search-input {
    flex-shrink: 0;
}

.node-grid {
    display: grid;
    grid-template-columns: 1fr;
    align-content: start;
    gap: 8px;
    flex: 1;
    min-height: 0;
    overflow-y: auto;
}

.node-section {
    min-width: 0;
}

.node-table {
    overflow: hidden;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    background-color: var(--el-bg-color);
}

.node-table__header,
.node-table__row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(140px, 42%);
    align-items: stretch;
}

.node-table__header {
    min-height: 30px;
    background-color: var(--el-fill-color-light);
    color: var(--el-text-color-secondary);
    font-size: 12px;
    font-weight: 500;
}

.node-table__row {
    min-height: 34px;
    border-top: 1px solid var(--el-border-color-lighter);
    color: var(--el-text-color-regular);
    font-size: 12px;
}

.node-table__cell {
    min-width: 0;
    padding: 5px 8px;
    line-height: 20px;
}

.node-table__cell--node {
    display: flex;
    align-items: center;
    gap: 8px;
    overflow: hidden;
    white-space: nowrap;
}

.node-table__cell--addr {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    border-left: 1px solid var(--el-border-color-lighter);
    overflow: hidden;
}

.node-table__group-title {
    min-width: 0;
    overflow: hidden;
    color: var(--el-text-color-primary);
    font-size: 13px;
    font-weight: 700;
    text-overflow: ellipsis;
}

.node-table__summary {
    flex-shrink: 0;
    color: var(--el-text-color-secondary);
    font-size: 10px;
}

.node-table__addr-text {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.node-favorite-button {
    flex-shrink: 0;
    width: 20px;
    height: 20px;
    min-height: 20px;
    padding: 0;
}

.node-name-cell {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;

    .icon {
        flex-shrink: 0;
        font-size: 8px;
    }
}

.node-name-button {
    height: 20px;
    min-height: 20px;
    padding: 0;
    line-height: 20px;
}

.node-status-wrap {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    height: 20px;
}

.icon-status {
    font-size: 16px;
}

@media (max-width: 480px) {
    .node-table__header,
    .node-table__row {
        grid-template-columns: minmax(0, 1fr) minmax(96px, 34%);
    }

    .node-table__cell {
        padding: 5px 6px;
    }
}
</style>
