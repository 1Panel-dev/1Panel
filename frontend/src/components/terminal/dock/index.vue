<template>
    <!-- Right edge handle: open a terminal from any page without leaving it. Hidden on the terminal page itself. -->
    <div v-if="!onTerminalPage" class="terminal-dock-handle" @click="show">
        <el-badge
            :value="store.entries.length"
            :hidden="store.entries.length === 0"
            type="primary"
            class="terminal-dock-badge"
        >
            <svg-icon iconName="p-terminal2" class="terminal-dock-icon" />
        </el-badge>
        <span class="terminal-dock-label">{{ $t('menu.terminal') }}</span>
    </div>

    <DialogPro
        v-model="open"
        :title="$t('menu.terminal')"
        size="w-70"
        :show-close="false"
        :modal="false"
        @opened="claim"
    >
        <!-- minimize keeps sessions alive; X closes them all (with confirm) -->
        <template #header>
            <div class="flex items-center">
                <span class="el-dialog__title flex-1">{{ $t('menu.terminal') }}</span>
                <el-tooltip :content="$t('terminal.minimize')" placement="top">
                    <el-button link icon="Minus" @click="open = false" />
                </el-tooltip>
                <el-tooltip :content="$t('terminal.closeAllSessions')" placement="top">
                    <el-button link icon="Close" @click="closeAll" />
                </el-tooltip>
            </div>
        </template>

        <template #content>
            <div class="terminal-dock-toolbar">
                <el-tabs v-model="active" type="card" closable class="terminal-dock-tabs" @tab-remove="store.remove">
                    <el-tab-pane v-for="item in store.entries" :key="item.key" :name="item.key">
                        <template #label>
                            <span
                                class="terminal-status-dot"
                                :class="item.status === 'online' ? 'is-online' : 'is-offline'"
                            ></span>
                            <span class="terminal-tab-title" :title="item.title">{{ item.title }}</span>
                        </template>
                    </el-tab-pane>
                </el-tabs>

                <el-popover trigger="click" width="280px" @before-enter="loadHosts">
                    <template #reference>
                        <el-button type="primary" plain icon="Plus" size="small">
                            {{ $t('terminal.createConn') }}
                        </el-button>
                    </template>
                    <el-button link class="w-full" @click="connect(0, $t('terminal.localhost'))">
                        <el-icon class="mr-1"><House /></el-icon>
                        {{ $t('terminal.localhost') }}
                    </el-button>
                    <template v-if="!isNodeAdmin">
                        <el-divider class="my-1" />
                        <el-input
                            v-model="hostFilter"
                            size="small"
                            clearable
                            :placeholder="$t('commons.button.search')"
                            class="mb-1"
                        />
                        <el-tree
                            ref="treeRef"
                            node-key="id"
                            default-expand-all
                            :expand-on-click-node="false"
                            :data="hostTree"
                            :filter-node-method="filterHost"
                            :empty-text="$t('terminal.noHost')"
                            class="terminal-dock-tree"
                        >
                            <template #default="{ node, data }">
                                <span v-if="node.level === 1" class="text-xs font-medium">
                                    {{ node.label === 'Default' ? $t('commons.table.default') : node.label }}
                                </span>
                                <a
                                    v-else
                                    class="text-xs hover:text-[var(--el-color-primary)] truncate"
                                    :title="node.label"
                                    @click="connect(data.id, node.label)"
                                >
                                    {{ node.label }}
                                </a>
                            </template>
                        </el-tree>
                    </template>
                </el-popover>
            </div>

            <div v-if="store.entries.length === 0" class="terminal-dock-empty">
                {{ $t('terminal.emptyTerminal') }}
            </div>
            <div
                v-for="item in store.entries"
                v-show="item.key === active"
                :key="item.key"
                class="terminal-dock-slot"
                :ref="(el: any) => onSlot(item.key, el)"
                @click="store.instances[item.key]?.refit()"
            ></div>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import i18n from '@/lang';
import { ElTree } from 'element-plus';
import { TerminalSessionStore } from '@/store';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { getHostTree, testByID, testLocalConn } from '@/api/modules/terminal';
import { MsgError } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { Host } from '@/api/interface/host';

const store = TerminalSessionStore();
const { isNodeAdmin } = useGlobalStore();
const route = useRoute();
const onTerminalPage = computed(() => route.path.startsWith('/terminal'));

const open = ref(false);
const active = ref('');
let timer: ReturnType<typeof setInterval> | null = null;

const show = async () => {
    if (!store.find(active.value)) active.value = store.entries[0]?.key || '';
    open.value = true;
    await nextTick();
    claim();
    store.sync();
    timer = setInterval(store.sync, 5000);
};

// park releases every slot so the Terminals go back to the off-screen host.
const park = () => {
    if (timer) clearInterval(timer);
    timer = null;
    claim();
};
watch(open, (value) => {
    if (!value) park();
});

const slotEls: Record<string, HTMLElement> = {};
const onSlot = (key: string, el: HTMLElement | null) => {
    if (el) slotEls[key] = el;
    else delete slotEls[key];
};
// Slots not ours (the terminal page's) are left alone.
const claim = () => {
    for (const item of store.entries) {
        if (open.value && item.key === active.value) {
            store.setSlot(item.key, slotEls[item.key] || null);
        } else if (store.slots[item.key] && store.slots[item.key] === slotEls[item.key]) {
            store.setSlot(item.key, null);
        }
    }
};
watch(active, () => nextTick(claim));
watch(
    () => store.entries.length,
    () => {
        if (!store.find(active.value)) active.value = store.entries[0]?.key || '';
    },
);

const hostTree = ref<Array<Host.HostTree>>([]);
const treeRef = ref<InstanceType<typeof ElTree>>();
const hostFilter = ref('');
const loadHosts = async () => {
    if (isNodeAdmin.value) return;
    const res = await getHostTree({});
    hostTree.value = res.data;
};
watch(hostFilter, (v) => treeRef.value?.filter(v));
const filterHost = (value: string, data: any) => !value || data.label.toLowerCase().includes(value.toLowerCase());

const connect = async (wsID: number, title: string) => {
    if (wsID === 0) {
        const res = await testLocalConn();
        if (!res.data) {
            MsgError(i18n.global.t('terminal.connLocalErr'));
            return;
        }
        active.value = await store.open({ title, wsID });
        return;
    }
    const res = await testByID(wsID);
    active.value = await store.open({
        title,
        wsID,
        error: res.data ? '' : 'Authentication failed. Please check the host information!',
    });
};

const closeAll = async () => {
    if (store.entries.length > 0) {
        await ElMessageBox.confirm(
            i18n.global.t('terminal.closeAllConfirm'),
            i18n.global.t('terminal.closeAllSessions'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'warning',
            },
        );
        store.closeAll();
    }
    open.value = false;
};

// the terminal page claims the slots itself; give ours up when navigating there
watch(onTerminalPage, (v) => {
    if (v) open.value = false;
});
</script>

<style scoped lang="scss">
.terminal-dock-handle {
    position: fixed;
    right: 0;
    bottom: 96px;
    z-index: 100;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 10px 6px;
    border-radius: 8px 0 0 8px;
    background: var(--el-bg-color);
    box-shadow: var(--el-box-shadow-light);
    color: var(--el-color-primary);
    cursor: pointer;
    user-select: none;
    transition: background-color 0.2s;

    &:hover {
        background-color: var(--el-fill-color-light);
    }
}

.terminal-dock-icon {
    width: 20px;
    height: 20px;
}

.terminal-dock-badge {
    --el-badge-size: 14px;
    --el-badge-font-size: 10px;
    --el-badge-padding: 4px;
    --el-badge-radius: 7px;
}

.terminal-dock-label {
    writing-mode: vertical-rl;
    font-size: 12px;
    letter-spacing: 2px;
}

.terminal-dock-toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
}

.terminal-dock-tabs {
    min-width: 0;
    flex: 1;

    :deep(.el-tabs__header) {
        margin-bottom: 0;
    }

    :deep(.el-tabs__item) {
        max-width: 220px;
        height: 34px;
        padding: 0 12px;
    }
}

.terminal-status-dot {
    width: 7px;
    height: 7px;
    flex: 0 0 auto;
    margin-right: 7px;
    border-radius: 50%;

    &.is-online {
        background-color: var(--el-color-success);
    }

    &.is-offline {
        background-color: var(--el-color-danger);
    }
}

.terminal-tab-title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.terminal-dock-tree {
    max-height: 40vh;
    overflow: auto;
}

.terminal-dock-slot,
.terminal-dock-empty {
    height: 60vh;
    overflow: hidden;
    border-radius: 6px;
}

.terminal-dock-slot {
    background-color: var(--panel-logs-bg-color);
}

.terminal-dock-empty {
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--el-fill-color-extra-light);
    color: var(--el-text-color-secondary);
}

@media (max-width: 768px) {
    .terminal-dock-slot,
    .terminal-dock-empty {
        height: 70vh;
    }
}
</style>
