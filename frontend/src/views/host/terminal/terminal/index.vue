<template>
    <div>
        <el-tabs
            type="card"
            class="terminal-tabs"
            style="background-color: var(--panel-terminal-tag-bg-color); margin-top: 20px"
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
                            style="color: white"
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
                        height: `calc(100vh - ${loadHeight()})`,
                        'background-color': `var(--panel-logs-bg-color)`,
                    }"
                    :ref="'t-' + item.index"
                    :key="item.Refresh"
                ></Terminal>
                <div class="flex w-full flex-col md:flex-row">
                    <el-select v-model="quickCmd" clearable filterable @change="quickInput">
                        <template #prefix>{{ $t('terminal.quickCommand') }}</template>
                        <el-option-group v-for="group in commandTree" :key="group.label" :label="group.label">
                            <el-option
                                v-for="(cmd, index) in group.children"
                                :key="index"
                                :label="cmd.name"
                                :value="cmd.command"
                            />
                        </el-option-group>
                    </el-select>
                    <el-input v-model="batchVal" @keyup.enter="batchInput">
                        <template #prepend>
                            <el-checkbox :label="$t('terminal.batchInput')" v-model="isBatch" />
                        </template>
                    </el-input>
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
                        <div class="search-button" style="float: none">
                            <el-input
                                v-model="hostFilterInfo"
                                style="margin-top: 5px; width: 90%"
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
                                    <span v-if="node.label === 'default'">{{ $t('website.default') }}</span>
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
                class="fullScreen"
                :style="{ top: loadFullScreenHeight() }"
                icon="FullScreen"
            ></el-button>
        </el-tooltip>

        <HostDialog ref="dialogRef" @on-conn-terminal="onConnTerminal" @load-host-tree="loadHostTree" />
    </div>
</template>

<script setup lang="ts">
import { ref, getCurrentInstance, watch, nextTick, computed, onMounted } from 'vue';
import Terminal from '@/components/terminal/index.vue';
import HostDialog from '@/views/host/terminal/terminal/host-create.vue';
import type Node from 'element-plus/es/components/tree/src/model/node';
import { ElTree } from 'element-plus';
import screenfull from 'screenfull';
import i18n from '@/lang';
import { Host } from '@/api/interface/host';
import { getCommandTree, getHostTree, testByID } from '@/api/modules/host';
import { GlobalStore } from '@/store';
import router from '@/routers';

const dialogRef = ref();
const ctx = getCurrentInstance() as any;
const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});

function toggleFullscreen() {
    if (screenfull.isEnabled) {
        screenfull.toggle();
    }
    globalStore.isFullScreen = !screenfull.isFullscreen;
}
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (globalStore.isFullScreen ? 'quitFullscreen' : 'fullscreen'));
};

const localHostID = ref();

let timer: NodeJS.Timer | null = null;
const terminalValue = ref();
const terminalTabs = ref([]) as any;
let tabIndex = 0;

const commandTree = ref();
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
    const res = await getHostTree({});
    hostTree.value = res.data;
    timer = setInterval(() => {
        syncTerminal();
    }, 1000 * 5);
    for (let gIndex = 0; gIndex < hostTree.value.length; gIndex++) {
        if (!hostTree.value[gIndex].children) {
            continue;
        }
        for (let i = 0; i < hostTree.value[gIndex].children.length; i++) {
            if (hostTree.value[gIndex].children[i].label.indexOf('@127.0.0.1:') !== -1) {
                localHostID.value = hostTree.value[gIndex].children[i].id;
                hostTree.value[gIndex].children.splice(i, 1);
                if (hostTree.value[gIndex].children.length === 0) {
                    hostTree.value.splice(gIndex, 1);
                }
                if (terminalTabs.value.length !== 0) {
                    return;
                }
                onNewLocal();
                return;
            }
        }
    }

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
    return globalStore.openMenuTabs ? '269px' : '229px';
};
const loadEmptyHeight = () => {
    return globalStore.openMenuTabs ? '240px' : '200px';
};
const loadFullScreenHeight = () => {
    return globalStore.openMenuTabs ? '66px' : '90px';
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
    const res = await getCommandTree();
    commandTree.value = res.data || [];
};

function quickInput(val: any) {
    if (val !== '' && ctx) {
        if (isBatch.value) {
            for (const tab of terminalTabs.value) {
                ctx.refs[`t-${tab.index}`] && ctx.refs[`t-${tab.index}`][0].sendMsg(val + '\n');
            }
            return;
        }
        ctx.refs[`t-${terminalValue.value}`] && ctx.refs[`t-${terminalValue.value}`][0].sendMsg(val + '\n');
        quickCmd.value = '';
    }
}

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
const onNewLocal = () => {
    onConnTerminal(i18n.global.t('terminal.localhost'), localHostID.value, false);
};

const onClickConn = (node: Node, data: Tree) => {
    if (node.level === 1) {
        return;
    }
    onConnTerminal(node.label, data.id, false);
};

const onReconnect = async (item: any) => {
    if (ctx) {
        ctx.refs[`t-${item.index}`] && ctx.refs[`t-${item.index}`][0].onClose();
    }
    item.Refresh = !item.Refresh;
    const res = await testByID(item.wsID);
    nextTick(() => {
        ctx.refs[`t-${item.index}`] &&
            ctx.refs[`t-${item.index}`][0].acceptParams({
                endpoint: '/api/v1/terminals',
                args: `id=${item.wsID}`,
                error: res.data ? '' : 'Failed to set up the connection. Please check the host information',
            });
    });
    syncTerminal();
};

const onConnTerminal = async (title: string, wsID: number, isLocal?: boolean) => {
    const res = await testByID(wsID);
    if (isLocal) {
        for (const tab of terminalTabs.value) {
            if (tab.title.indexOf('@127.0.0.1:') !== -1 || tab.title === i18n.global.t('terminal.localhost')) {
                onReconnect(tab);
            }
        }
        return;
    }
    terminalTabs.value.push({
        index: tabIndex,
        title: title,
        wsID: wsID,
        status: res.data ? 'online' : 'closed',
        latency: 0,
    });
    terminalValue.value = tabIndex;
    if (!res.data && title === i18n.global.t('terminal.localhost')) {
        dialogRef.value!.acceptParams({ isLocal: true });
    }
    nextTick(() => {
        ctx.refs[`t-${terminalValue.value}`] &&
            ctx.refs[`t-${terminalValue.value}`][0].acceptParams({
                endpoint: '/api/v1/terminals',
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

defineExpose({
    acceptParams,
    cleanTimer,
});

onMounted(() => {
    if (router.currentRoute.value.query.path) {
        const path = String(router.currentRoute.value.query.path);
        initCmd.value = `cd "${path}" \n`;
    }
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
.fullScreen {
    background-color: transparent;
    border: none;
    position: absolute;
    right: 50px;
    font-weight: 600;
    font-size: 14px;
}
.el-tabs--top.el-tabs--card > .el-tabs__header .el-tabs__item:last-child {
    padding-right: 0px;
}
.el-input__wrapper {
    border-radius: 50px;
}
.el-input-group__prepend {
    border-top-left-radius: 50px;
    border-bottom-left-radius: 50px;
}
</style>
