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
                :key="item.index"
                v-for="item in terminalTabs"
                :closable="true"
                :label="item.title"
                :name="item.index"
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
                <Terminal
                    :style="{
                        height: cmdPanelVisible
                            ? `calc(100vh - ${loadHeightWithPanel()})`
                            : `calc(100vh - ${loadHeight()})`,
                        'background-color': `var(--panel-logs-bg-color)`,
                    }"
                    :ref="'t-' + item.index"
                    :key="item.Refresh"
                ></Terminal>

                <transition name="el-fade-in">
                    <div
                        v-show="cmdPanelVisible"
                        class="mb-2 border-b border-[var(--el-border-color)] pb-2 w-full bg-[var(--el-bg-color)]"
                    >
                        <el-tabs v-model="activeGroupTab" type="card" class="command-tabs">
                            <el-tab-pane
                                v-for="group in commandTree"
                                :key="group.value"
                                :label="''"
                                :name="group.value"
                            >
                                <template #label>
                                    <span class="group-tab-label">
                                        <span v-if="group.label.length <= 6">{{ group.label }}</span>
                                        <el-tooltip v-else :content="group.label" placement="top">
                                            <span>{{ group.label.substring(0, 6) }}...</span>
                                        </el-tooltip>
                                    </span>
                                </template>
                                <div class="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-2 p-0">
                                    <el-tag
                                        v-for="cmd in group.children"
                                        :key="cmd.value"
                                        class="command-tag"
                                        @click="executeCommand(cmd.value)"
                                        type="info"
                                        effect="plain"
                                    >
                                        <div class="flex items-center justify-between w-full gap-1.5">
                                            <span class="command-tag-name" :title="cmd.label">
                                                {{
                                                    cmd.label.length > 8 ? cmd.label.substring(0, 8) + '...' : cmd.label
                                                }}
                                            </span>
                                            <el-popover placement="top" :width="320" trigger="hover">
                                                <template #reference>
                                                    <el-icon class="command-preview-icon">
                                                        <InfoFilled />
                                                    </el-icon>
                                                </template>
                                                <div class="command-preview">
                                                    <div class="command-preview-name">
                                                        <strong>{{ cmd.label }}</strong>
                                                    </div>
                                                    <div class="command-preview-value">{{ cmd.value }}</div>
                                                </div>
                                            </el-popover>
                                        </div>
                                    </el-tag>
                                </div>
                            </el-tab-pane>
                        </el-tabs>
                    </div>
                </transition>

                <div class="flex items-center gap-3 w-full py-2 flex-wrap">
                    <el-button
                        @click="cmdPanelVisible = !cmdPanelVisible"
                        type="primary"
                        class="min-w-[120px] max-w-[150px] shrink-0"
                    >
                        {{ $t('terminal.quickCommand') }}
                        <el-icon class="ml-1">
                            <component :is="cmdPanelVisible ? 'ArrowUp' : 'ArrowDown'" />
                        </el-icon>
                    </el-button>
                    <el-input
                        v-model="batchVal"
                        @keyup.enter.exact="batchInput"
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
                    <el-popover ref="popoverRef" width="250px" trigger="hover" virtual-triggering persistent>
                        <div class="ml-2.5">
                            <el-button link type="primary" @click="onNewSsh">{{ $t('terminal.createConn') }}</el-button>
                        </div>
                        <div class="ml-2.5">
                            <el-button link type="primary" @click="onNewLocal">
                                {{ $t('terminal.localhost') }}
                            </el-button>
                        </div>
                        <div class="search-button">
                            <el-input
                                v-model="hostFilterInfo"
                                class="mt-1.5 w-[90%]"
                                clearable
                                suffix-icon="Search"
                                :placeholder="$t('commons.button.search')"
                            ></el-input>
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
                        >
                            <template #default="{ node, data }">
                                <span class="custom-tree-node">
                                    <span v-if="node.label === 'Default'">{{ $t('commons.table.default') }}</span>
                                    <div v-else>
                                        <span v-if="node.label.length <= 25">
                                            <a @click="onClickConn(node, data)">{{ node.label }}</a>
                                        </span>
                                        <el-tooltip v-else :content="node.label" placement="right">
                                            <span>
                                                <a @click="onClickConn(node, data)">
                                                    {{ node.label.substring(0, 22) }}...
                                                </a>
                                            </span>
                                        </el-tooltip>
                                    </div>
                                </span>
                            </template>
                        </el-tree>
                    </el-popover>
                </template>
            </el-tab-pane>
            <div v-if="terminalTabs.length === 0">
                <el-empty
                    :style="{ height: `calc(100vh - ${loadEmptyHeight()})`, 'background-color': '#000' }"
                    :description="$t('terminal.emptyTerminal')"
                ></el-empty>
            </div>
        </el-tabs>
        <el-tooltip :content="loadTooltip()" placement="top">
            <el-button
                @click="toggleFullscreen"
                v-if="!mobile"
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
import { ref, getCurrentInstance, watch, nextTick, computed, onMounted } from 'vue';
import Terminal from '@/components/terminal/index.vue';
import HostDialog from '@/views/terminal/terminal/host-create.vue';
import type Node from 'element-plus/es/components/tree/src/model/node';
import { ElTree } from 'element-plus';
import screenfull from 'screenfull';
import i18n from '@/lang';
import { Host } from '@/api/interface/host';
import { getHostTree, testByID, testLocalConn } from '@/api/modules/terminal';
import { GlobalStore } from '@/store';
import router from '@/routers';
import { getCommandTree } from '@/api/modules/command';
import { getAgentSettingByKey } from '@/api/modules/setting';

const dialogRef = ref();
const ctx = getCurrentInstance() as any;
const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});

const toggleFullscreen = () => {
    if (screenfull.isEnabled) {
        screenfull.toggle();
    }
};
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (globalStore.isFullScreen ? 'quitFullscreen' : 'fullscreen'));
};

let timer: NodeJS.Timer | null = null;
const terminalValue = ref();
const terminalTabs = ref([]) as any;
let tabIndex = 0;

const commandTree = ref();
const cmdPanelVisible = ref(false);
const activeGroupTab = ref('');
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
    globalStore.isFullScreen = false;
    loadCommandTree();
    loadHostTree();
    if (terminalTabs.value.length === 0) {
        await getAgentSettingByKey('LocalSSHConnShow').then((res) => {
            if (res.data === 'Enable') {
                onNewLocal();
            }
        });
    }
    timer = setInterval(() => {
        syncTerminal();
    }, 1000 * 5);
    if (!mobile.value) {
        screenfull.on('change', () => {
            globalStore.isFullScreen = screenfull.isFullscreen;
        });
    }
};

const cleanTimer = () => {
    clearInterval(Number(timer));
    timer = null;
    for (const terminal of terminalTabs.value) {
        if (ctx && ctx.refs[`t-${terminal.index}`][0]) {
            terminal.status = ctx.refs[`t-${terminal.index}`][0].onClose();
        }
    }
};

const loadHeight = () => {
    return globalStore.openMenuTabs ? '230px' : '190px';
};
const loadHeightWithPanel = () => {
    return globalStore.openMenuTabs ? '470px' : '430px';
};
const loadEmptyHeight = () => {
    return globalStore.openMenuTabs ? '201px' : '156px';
};
const loadFullScreenHeight = () => {
    return globalStore.openMenuTabs ? '105px' : '60px';
};

const handleTabsRemove = (targetName: string, action: 'remove' | 'add') => {
    if (action !== 'remove') {
        return;
    }
    if (ctx) {
        ctx.refs[`t-${targetName}`] && ctx.refs[`t-${targetName}`][0].onClose();
    }
    const tabs = terminalTabs.value;
    let activeName = terminalValue.value;
    if (activeName === targetName) {
        tabs.forEach((tab: any, index: any) => {
            if (tab.index === targetName) {
                const nextTab = tabs[index + 1] || tabs[index - 1];
                if (nextTab) {
                    activeName = nextTab.index;
                }
            }
        });
    }
    terminalValue.value = activeName;
    terminalTabs.value = tabs.filter((tab: any) => tab.index !== targetName);
};

const loadHostTree = async () => {
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
    if (commandTree.value.length > 0) {
        activeGroupTab.value = commandTree.value[0].value;
    }
};

const executeCommand = (command: string) => {
    if (!ctx) {
        return;
    }
    if (isBatch.value) {
        for (const tab of terminalTabs.value) {
            ctx.refs[`t-${tab.index}`] && ctx.refs[`t-${tab.index}`][0].sendMsg(command + '\n');
        }
    } else {
        ctx.refs[`t-${terminalValue.value}`] && ctx.refs[`t-${terminalValue.value}`][0].sendMsg(command + '\n');
    }
};

function batchInput() {
    if (batchVal.value === '' || !ctx) {
        return;
    }
    if (isBatch.value) {
        for (const tab of terminalTabs.value) {
            ctx.refs[`t-${tab.index}`] && ctx.refs[`t-${tab.index}`][0].sendMsg(batchVal.value + '\n');
        }
        batchVal.value = '';
        return;
    }
    ctx.refs[`t-${terminalValue.value}`] && ctx.refs[`t-${terminalValue.value}`][0].sendMsg(batchVal.value + '\n');
    batchVal.value = '';
}

function beforeLeave(activeName: string) {
    if (activeName === 'newTabs') {
        return false;
    }
}

const onNewSsh = () => {
    dialogRef.value!.acceptParams({ isLocal: false });
};
const onNewLocal = async () => {
    const res = await testLocalConn();
    if (!res.data) {
        dialogRef.value!.acceptParams({ isLocal: true });
        return;
    }
    terminalTabs.value.push({
        index: tabIndex,
        title: i18n.global.t('terminal.localhost'),
        wsID: 0,
        status: 'online',
        latency: 0,
    });
    terminalValue.value = tabIndex;
    nextTick(() => {
        ctx.refs[`t-${terminalValue.value}`] &&
            ctx.refs[`t-${terminalValue.value}`][0].acceptParams({
                endpoint: '/api/v2/hosts/terminal',
                initCmd: initCmd.value,
                error: '',
            });
        initCmd.value = '';
    });
    tabIndex++;
};

const onClickConn = (node: Node, data: Tree) => {
    if (node.level === 1) {
        return;
    }
    onConnTerminal(node.label, data.id);
};

const onReconnect = async (item: any) => {
    if (ctx) {
        ctx.refs[`t-${item.index}`] && ctx.refs[`t-${item.index}`][0].onClose();
    }
    item.Refresh = !item.Refresh;
    if (item.wsID === 0) {
        const res = await testLocalConn();
        nextTick(() => {
            ctx.refs[`t-${item.index}`] &&
                ctx.refs[`t-${item.index}`][0].acceptParams({
                    endpoint: '/api/v2/hosts/terminal',
                    initCmd: initCmd.value,
                    error: res.data ? '' : 'Failed to set up the connection. Please check the host information',
                });
            initCmd.value = '';
        });
        syncTerminal();
        return;
    }

    const res = await testByID(item.wsID);
    nextTick(() => {
        ctx.refs[`t-${item.index}`] &&
            ctx.refs[`t-${item.index}`][0].acceptParams({
                endpoint: '/api/v2/core/hosts/terminal',
                args: `id=${item.wsID}`,
                initCmd: initCmd.value,
                error: res.data ? '' : 'Failed to set up the connection. Please check the host information',
            });
        initCmd.value = '';
    });
    syncTerminal();
};

const onConnTerminal = async (title: string, wsID: number) => {
    const res = await testByID(wsID);
    terminalTabs.value.push({
        index: tabIndex,
        title: title,
        wsID: wsID,
        status: res.data ? 'online' : 'closed',
        latency: 0,
    });
    terminalValue.value = tabIndex;
    nextTick(() => {
        ctx.refs[`t-${terminalValue.value}`] &&
            ctx.refs[`t-${terminalValue.value}`][0].acceptParams({
                endpoint: '/api/v2/core/hosts/terminal',
                args: `id=${wsID}`,
                initCmd: initCmd.value,
                error: res.data ? '' : 'Authentication failed. Please check the host information!',
            });
        initCmd.value = '';
    });
    tabIndex++;
};

function syncTerminal() {
    for (const terminal of terminalTabs.value) {
        if (ctx && ctx.refs[`t-${terminal.index}`][0]) {
            terminal.status = ctx.refs[`t-${terminal.index}`][0].isWsOpen() ? 'online' : 'closed';
            terminal.latency = ctx.refs[`t-${terminal.index}`][0].getLatency();
        }
    }
}

const changeFullScreen = () => {
    globalStore.isFullScreen = screenfull.isFullscreen;
};

defineExpose({
    acceptParams,
    cleanTimer,
});

onBeforeUnmount(() => {
    document.removeEventListener('fullscreenchange', changeFullScreen);
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

.command-tabs {
    :deep(.el-tabs__header) {
        margin-bottom: 0;
        background-color: var(--el-bg-color);
    }
    :deep(.el-tabs__content) {
        height: 180px;
        overflow-y: auto;
        overflow-x: hidden;
        background-color: var(--el-bg-color);
    }
    :deep(.el-tabs__item) {
        min-width: 80px;
        max-width: 110px;
        text-align: center;
        display: flex;
        align-items: center;
        justify-content: center;
        overflow: hidden;
        padding: 0 8px;
    }
}
.group-tab-label {
    width: 90px;
    display: inline-block;
    text-align: center;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 14px;
}

.command-tag {
    cursor: pointer;
    height: auto;
    padding: 8px 12px;
    transition: all 0.3s;
    border-radius: 4px;
    white-space: nowrap;
    border: 1px solid transparent;

    &:hover {
        border-color: var(--el-color-primary);
    }
}

.command-tag-name {
    font-weight: 500;
    font-size: 13px;
    flex: 1;
    text-align: left;
}

.command-preview-icon {
    font-size: 14px;
    opacity: 0.6;
    transition: opacity 0.3s;
    cursor: help;
    flex-shrink: 0;
    display: flex;
    align-items: center;

    &:hover {
        opacity: 1;
    }
}

.command-preview {
    .command-preview-name {
        font-size: 13px;
        margin-bottom: 6px;
        color: var(--el-text-color-primary);
        word-break: break-word;
    }

    .command-preview-value {
        font-size: 12px;
        font-family: monospace;
        padding: 8px;
        background-color: var(--el-fill-color-light);
        border-radius: 4px;
        color: var(--el-text-color-regular);
        word-break: break-all;
        white-space: pre-wrap;
    }
}

.command-tag-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: inline-block;
}
</style>
