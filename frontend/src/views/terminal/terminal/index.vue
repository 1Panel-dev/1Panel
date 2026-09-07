<template>
    <div>
        <el-tabs
            type="card"
            class="terminal-tabs card-interval"
            style="background-color: var(--panel-terminal-tag-bg-color)"
            v-model="terminalValue"
            :before-leave="beforeLeave"
            @tab-change="quickCmd = ''"
            @edit="handleTabsRemove"
        >
            <el-tab-pane
                :key="item.key"
                v-for="item in store.entries"
                :closable="true"
                :label="item.title"
                :name="item.key"
            >
                <template #label>
                    <span class="custom-tabs-label">
                        <span
                            v-if="item.status === 'online'"
                            :style="`color: ${
                                item.latency < 100 ? '#69db7c' : item.latency < 300 ? '#f59f00' : '#d9480f'
                            }; display: inline-flex; align-items: center`"
                        >
                            <span>&nbsp;{{ item.latency }}&nbsp;ms&nbsp;</span>
                            <el-icon>
                                <circleCheck />
                            </el-icon>
                        </span>
                        <el-button
                            v-if="item.status === 'closed'"
                            icon="Refresh"
                            class="text-white"
                            size="default"
                            link
                            @click="onReconnect(item)"
                        />
                        <span v-if="item.title.length <= 20">&nbsp;{{ item.title }}&nbsp;</span>
                        <el-tooltip v-else :content="item.title" placement="top-start">
                            <span>&nbsp;{{ item.title.substring(0, 17) }}...&nbsp;</span>
                        </el-tooltip>
                    </span>
                </template>
                <!-- The Terminal itself is rendered by components/terminal/host.vue and teleported here. -->
                <div
                    class="terminal-slot"
                    :ref="(el: any) => onSlot(item.key, el)"
                    :style="{
                        height: `calc(100vh - ${loadHeight()})`,
                        'background-color': `var(--panel-logs-bg-color)`,
                    }"
                ></div>

                <div class="flex items-center gap-2 w-full py-2 flex-wrap">
                    <AiSetting v-if="!isMobile" class="shrink-0" />
                    <el-cascader
                        v-model="quickCmd"
                        :options="commandTree"
                        :props="quickCommandProps"
                        :show-all-levels="false"
                        filterable
                        clearable
                        class="quick-command-cascader min-w-[180px] max-w-[260px] shrink-0"
                        :placeholder="$t('terminal.quickCommand')"
                        @change="handleQuickCommandChange"
                    >
                        <template #default="{ data }">
                            <el-tooltip
                                v-if="!data.children?.length"
                                placement="right"
                                popper-class="command-detail-tooltip"
                            >
                                <template #content>
                                    <div class="command-detail-content">{{ data.value }}</div>
                                </template>
                                <div class="cascader-option">
                                    <span class="cascader-option-label">{{ data.label }}</span>
                                </div>
                            </el-tooltip>
                            <div v-else class="cascader-option">
                                <span class="cascader-option-label">{{ data.label }}</span>
                            </div>
                        </template>
                    </el-cascader>
                    <el-input
                        v-model="batchVal"
                        @keydown.enter.exact.prevent="batchInput"
                        type="textarea"
                        :autosize="{ minRows: 1, maxRows: 3 }"
                        class="flex-1 basis-[300px] min-w-[200px]"
                        placeholder=">"
                    ></el-input>
                    <el-checkbox
                        :label="$t('terminal.batchInput')"
                        v-model="isBatch"
                        class="shrink-0 whitespace-nowrap"
                    />
                </div>
            </el-tab-pane>
            <el-tab-pane :closable="false" name="newTabs">
                <template #label>
                    <el-button v-popover="popoverRef" class="tagButton" icon="Plus"></el-button>
                    <el-popover
                        ref="popoverRef"
                        width="320px"
                        trigger="hover"
                        virtual-triggering
                        persistent
                        :offset="-4"
                    >
                        <div class="p-2 space-y-2">
                            <div class="flex gap-2">
                                <button
                                    v-if="!isNodeAdmin"
                                    @click="onNewSsh"
                                    class="flex-1 flex flex-col items-center justify-center px-3 py-2.5 bg-[var(--el-fill-color-light)] hover:bg-[var(--panel-main-bg-color-9)] rounded transition-colors duration-200 cursor-pointer group border-0 outline-none"
                                >
                                    <el-icon
                                        class="text-xl mb-1 text-[var(--el-text-color-primary)] group-hover:text-[var(--el-color-primary)] transition-colors"
                                    >
                                        <Plus />
                                    </el-icon>
                                    <span
                                        class="text-xs text-[var(--el-text-color-primary)] group-hover:text-[var(--el-color-primary)] font-medium truncate w-full text-center transition-colors"
                                    >
                                        {{ $t('terminal.createConn') }}
                                    </span>
                                </button>
                                <button
                                    @click="onNewLocal"
                                    class="flex-1 flex flex-col items-center justify-center px-3 py-2.5 bg-[var(--el-fill-color-light)] hover:bg-[var(--panel-main-bg-color-9)] rounded transition-colors duration-200 cursor-pointer group border-0 outline-none"
                                >
                                    <el-icon
                                        class="text-xl mb-1 text-[var(--el-text-color-primary)] group-hover:text-[var(--el-color-primary)] transition-colors"
                                    >
                                        <House />
                                    </el-icon>
                                    <span
                                        class="text-xs text-[var(--el-text-color-primary)] group-hover:text-[var(--el-color-primary)] font-medium truncate w-full text-center transition-colors"
                                    >
                                        {{ $t('terminal.localhost') }}
                                    </span>
                                </button>
                            </div>
                            <template v-if="!isNodeAdmin">
                                <el-divider class="my-0" />

                                <div class="search-container px-1 py-1 bg-[var(--el-fill-color-light)] rounded">
                                    <el-input
                                        v-model="hostFilterInfo"
                                        class="w-full"
                                        clearable
                                        suffix-icon="Search"
                                        :placeholder="$t('commons.button.search')"
                                        size="small"
                                    >
                                        <template #prefix>
                                            <el-icon class="el-input__icon"><Search /></el-icon>
                                        </template>
                                    </el-input>
                                </div>
                                <el-tree
                                    ref="treeRef"
                                    :expand-on-click-node="false"
                                    node-key="id"
                                    :default-expand-all="true"
                                    :data="hostTree"
                                    :props="defaultProps"
                                    :filter-node-method="filterHost"
                                    :empty-text="$t('terminal.noHost')"
                                    class="host-tree"
                                >
                                    <template #default="{ node, data }">
                                        <span class="custom-tree-node w-full">
                                            <span
                                                v-if="node.label === 'Default'"
                                                class="text-xs font-medium text-[var(--el-text-color-primary)]"
                                            >
                                                {{ $t('commons.table.default') }}
                                            </span>
                                            <div v-else class="w-full min-w-0">
                                                <span v-if="node.label.length <= 22">
                                                    <a
                                                        @click="onClickConn(node, data)"
                                                        class="text-xs text-[var(--el-text-color-primary)] hover:text-[var(--el-color-primary)] transition-colors cursor-pointer block truncate"
                                                    >
                                                        {{ node.label }}
                                                    </a>
                                                </span>
                                                <el-tooltip v-else :content="node.label" placement="right">
                                                    <span>
                                                        <a
                                                            @click="onClickConn(node, data)"
                                                            class="text-xs text-[var(--el-text-color-primary)] hover:text-[var(--el-color-primary)] transition-colors cursor-pointer block truncate"
                                                        >
                                                            {{ node.label.substring(0, 30) }}...
                                                        </a>
                                                    </span>
                                                </el-tooltip>
                                            </div>
                                        </span>
                                    </template>
                                </el-tree>
                            </template>
                        </div>
                    </el-popover>
                </template>
            </el-tab-pane>
            <div v-if="store.entries.length === 0">
                <el-empty
                    :style="{ height: `calc(100vh - ${loadEmptyHeight()})`, 'background-color': '#000' }"
                    :description="$t('terminal.emptyTerminal')"
                ></el-empty>
            </div>
        </el-tabs>
        <el-tooltip :content="loadTooltip()" placement="top">
            <el-button
                @click="toggleFullscreen"
                v-if="!isMobile"
                class="bg-transparent border-0 absolute right-[50px] font-semibold text-sm"
                :style="{ top: loadFullScreenHeight() }"
                icon="FullScreen"
            ></el-button>
        </el-tooltip>

        <HostDialog
            ref="dialogRef"
            @on-conn-terminal="onConnTerminal"
            @on-new-local="onNewLocal"
            @load-host-tree="loadHostTree"
        />
    </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onBeforeUnmount, onActivated, onDeactivated } from 'vue';
import HostDialog from '@/views/terminal/terminal/host-create.vue';
import type Node from 'element-plus/es/components/tree/src/model/node';
import { ElTree } from 'element-plus';
import screenfull from 'screenfull';
import i18n from '@/lang';
import { Host } from '@/api/interface/host';
import { getHostTree, testByID, testLocalConn } from '@/api/modules/terminal';
import { useGlobalStore } from '@/composables/useGlobalStore';
import router from '@/routers';
import { getCommandTree } from '@/api/modules/command';
import { getAgentSettingInfo } from '@/api/modules/setting';
import AiSetting from '@/views/terminal/setting/ai/index.vue';
import { MsgWarning } from '@/utils/message';
import { TerminalSessionStore } from '@/store';

const { isFullScreen, isMobile, isNodeAdmin, openMenuTabs } = useGlobalStore();
const store = TerminalSessionStore();

const dialogRef = ref();

const toggleFullscreen = () => {
    if (screenfull.isEnabled) {
        screenfull.toggle();
    }
};
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (isFullScreen.value ? 'quitFullscreen' : 'fullscreen'));
};

let timer: ReturnType<typeof setInterval> | null = null;
const terminalValue = ref();

const commandTree = ref();
const quickCommandProps = {
    expandTrigger: 'hover' as const,
};
let quickCmd = ref();
let batchVal = ref();
let isBatch = ref<boolean>(false);

const popoverRef = ref();

const hostFilterInfo = ref('');
const hostTree = ref<Array<Host.HostTree>>();
const treeRef = ref<InstanceType<typeof ElTree>>();
const defaultProps = {
    label: 'label',
    children: 'children',
};
interface Tree {
    id: number;
    label: string;
    children?: Tree[];
}
const initCmd = ref('');

const acceptParams = async () => {
    isFullScreen.value = false;
    loadCommandTree();
    if (!isNodeAdmin.value) {
        loadHostTree();
    } else {
        hostTree.value = [];
    }
    if (store.entries.length === 0) {
        await openDefaultLocalConn();
    } else {
        // sessions kept alive while we were away: show them and re-fit to this container
        if (!store.find(terminalValue.value)) {
            terminalValue.value = store.entries[0].key;
        }
        await claim();
        store.sync();
    }
    timer = setInterval(store.sync, 1000 * 5);
    if (!isMobile.value) {
        screenfull.on('change', () => {
            isFullScreen.value = screenfull.isFullscreen;
        });
    }
};

const openDefaultLocalConn = async () => {
    if (isNodeAdmin.value) {
        onNewLocal();
        return;
    }
    await getAgentSettingInfo().then((res) => {
        if (res.data?.localSSHConnShow === 'Enable') {
            onNewLocal();
        }
    });
};

// Leaving the page keeps every session connected in the host; only the poll stops.
const cleanTimer = () => {
    clearInterval(Number(timer));
    timer = null;
};

// Slots are claimed explicitly, not from the ref callback: under a locked menu tab
// (keep-alive) the page keeps rendering while detached, and the dock takes the
// Terminals over meanwhile. Only a visible page owns its slots; release on leave
// and take them back on return, the same way the dock does.
const slotEls: Record<string, HTMLElement> = {};
const onSlot = (key: string, el: HTMLElement | null) => {
    if (el) slotEls[key] = el;
    else delete slotEls[key];
};
let pageVisible = true;
const claim = async () => {
    for (const item of store.entries) {
        if (pageVisible) {
            store.setSlot(item.key, slotEls[item.key] || null);
        } else if (store.slots[item.key] && store.slots[item.key] === slotEls[item.key]) {
            store.setSlot(item.key, null);
        }
    }
    if (!pageVisible) return;
    await nextTick();
    for (const item of store.entries) {
        store.instances[item.key]?.refit();
    }
};
watch(
    () => store.entries.length,
    () => nextTick(claim),
);
onActivated(() => {
    pageVisible = true;
    claim();
});
onDeactivated(() => {
    pageVisible = false;
    claim();
});

const loadHeight = () => {
    return openMenuTabs.value ? '250px' : '210px';
};
const loadEmptyHeight = () => {
    return openMenuTabs.value ? '201px' : '156px';
};
const loadFullScreenHeight = () => {
    return openMenuTabs.value ? '105px' : '60px';
};

const handleTabsRemove = async (targetName: string, action: 'remove' | 'add') => {
    if (action !== 'remove') {
        return;
    }
    if (!store.find(targetName)) {
        return;
    }
    const tabs = store.entries;
    let activeName = terminalValue.value;
    if (activeName === targetName) {
        tabs.forEach((tab, index) => {
            if (tab.key === targetName) {
                const nextTab = tabs[index + 1] || tabs[index - 1];
                if (nextTab) {
                    activeName = nextTab.key;
                }
            }
        });
    }
    terminalValue.value = activeName;
    store.remove(targetName);
};

const loadHostTree = async () => {
    if (isNodeAdmin.value) {
        hostTree.value = [];
        return;
    }
    const res = await getHostTree({});
    hostTree.value = res.data;
};
watch(hostFilterInfo, (val: any) => {
    treeRef.value!.filter(val);
});
const filterHost = (value: string, data: any) => {
    if (!value) return true;
    return data.label.includes(value);
};
const loadCommandTree = async () => {
    const res = await getCommandTree('command');
    commandTree.value = res.data || [];
    for (const item of commandTree.value) {
        if (item.label === 'Default') {
            item.label = i18n.global.t('commons.table.default');
        }
    }
};

const sendToTerminals = (command: string, all: boolean) => {
    const keys = all ? store.entries.map((e) => e.key) : [terminalValue.value];
    for (const key of keys) {
        store.instances[key]?.sendMsg(command);
    }
};

const handleQuickCommandChange = (val: Array<string>) => {
    if (!val?.length) {
        return;
    }
    sendToTerminals(val[val.length - 1] + '\n', isBatch.value);
    quickCmd.value = '';
};

function batchInput() {
    if (batchVal.value === '') {
        return;
    }
    sendToTerminals(batchVal.value + '\n', isBatch.value);
    batchVal.value = '';
}

function beforeLeave(activeName: string) {
    if (activeName === 'newTabs') {
        return false;
    }
}

const onNewSsh = () => {
    if (isNodeAdmin.value) {
        MsgWarning(i18n.global.t('terminal.nodeAdminLocalOnly'));
        return;
    }
    dialogRef.value!.acceptParams({ isLocal: false });
};

const connectionError = 'Failed to set up the connection. Please check the host information';

const openTab = async (title: string, wsID: number, error: string) => {
    const cmd = initCmd.value;
    initCmd.value = '';
    terminalValue.value = await store.open({ title, wsID, initCmd: cmd, error });
};

const onNewLocal = async () => {
    const res = await testLocalConn();
    if (!res.data) {
        dialogRef.value!.acceptParams({ isLocal: true });
        return;
    }
    await openTab(i18n.global.t('terminal.localhost'), 0, '');
};

const onClickConn = (node: Node, data: Tree) => {
    if (node.level === 1) {
        return;
    }
    onConnTerminal(node.label, data.id);
};

const onReconnect = async (item: any) => {
    const res = item.wsID === 0 ? await testLocalConn() : await testByID(item.wsID);
    const cmd = initCmd.value;
    initCmd.value = '';
    await store.reconnect(item.key, res.data ? '' : connectionError, cmd);
    store.sync();
};

const onConnTerminal = async (title: string, wsID: number) => {
    if (isNodeAdmin.value) {
        MsgWarning(i18n.global.t('terminal.nodeAdminLocalOnly'));
        return;
    }
    const res = await testByID(wsID);
    await openTab(title, wsID, res.data ? '' : 'Authentication failed. Please check the host information!');
};

const changeFullScreen = () => {
    isFullScreen.value = screenfull.isFullscreen;
};

defineExpose({
    acceptParams,
});

onBeforeUnmount(() => {
    document.removeEventListener('fullscreenchange', changeFullScreen);
    // parent refs are already null in the parent's onUnmounted, so leave-page cleanup lives here
    cleanTimer();
    pageVisible = false;
    claim();
});

onMounted(() => {
    if (router.currentRoute.value.query.path) {
        const path = String(router.currentRoute.value.query.path);
        initCmd.value = `cd "${path}" \n`;
    }
    document.addEventListener('fullscreenchange', changeFullScreen);
});
</script>

<style lang="scss" scoped>
.terminal-tabs {
    :deep(.el-tabs__header) {
        padding: 0;
        position: relative;
        margin: 0 0 3px 0;
    }
    :deep(.el-tabs__nav) {
        white-space: nowrap;
        position: relative;
        transition: transform var(--el-transition-duration);
        float: left;
        z-index: calc(var(--el-index-normal) + 1);
    }
    :deep(.el-tabs__item) {
        padding: 0;
    }
    :deep(.el-tabs__item.is-active) {
        color: var(--panel-terminal-tag-active-text-color);
        background-color: var(--panel-terminal-tag-active-bg-color);
    }
    :deep(.el-tabs__item:hover) {
        color: var(--panel-terminal-tag-hover-text-color);
    }
    :deep(.el-tabs__item.is-active:hover) {
        color: var(--panel-terminal-tag-active-text-color);
    }
}

.tagButton {
    border: 0;
    background-color: var(--el-tabs__item);
}

.terminal-slot {
    width: 100%;
}

.host-tree {
    max-height: 300px;
    overflow-y: auto;
}

.search-container {
    :deep(.el-input__wrapper) {
        border-radius: 6px;
        box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);

        &:hover {
            box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
        }

        &.is-focus {
            box-shadow: 0 0 0 2px var(--el-color-primary-light-3);
        }
    }
}

.vertical-tabs > .el-tabs__content {
    padding: 32px;
    color: #6b778c;
    font-size: 32px;
    font-weight: 600;
}
.el-tabs--top.el-tabs--card > .el-tabs__header .el-tabs__item:last-child {
    padding-right: 0px;
}
.el-input__wrapper {
    border-radius: 50px;
}

:deep(.el-textarea__inner) {
    border-radius: 4px;
    resize: none;
    min-height: 32px;
    transition: height 0.2s ease;
}

.quick-command-cascader {
    :deep(.el-input__wrapper) {
        border-radius: 6px;
    }
}

.cascader-option {
    width: 100%;
    display: flex;
    align-items: center;
    min-width: 0;
}

.cascader-option-label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

:deep(.command-detail-tooltip) {
    max-width: 420px;
}

.command-detail-content {
    white-space: pre-wrap;
    word-break: break-all;
    line-height: 1.5;
}
</style>
