<template>
    <div>
        <el-card class="topCard">
            <el-radio-group @change="handleChange" v-model="activeNames">
                <el-radio-button class="topButton" size="default" label="terminal">
                    {{ $t('menu.terminal') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="default" label="host">
                    {{ $t('menu.host') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="default" label="command">
                    {{ $t('terminal.quickCommand') }}
                </el-radio-button>
            </el-radio-group>
        </el-card>
        <div v-show="activeNames === 'terminal'">
            <el-tabs
                type="card"
                class="terminal-tabs"
                style="background-color: #efefef; margin-top: 20px"
                v-model="terminalValue"
                :before-leave="beforeLeave"
                @edit="handleTabsRemove"
            >
                <el-tab-pane
                    :key="item.key"
                    v-for="item in terminalTabs"
                    :closable="true"
                    :label="item.title"
                    :name="item.key"
                >
                    <template #label>
                        <span class="custom-tabs-label">
                            <el-icon style="margin-top: 1px" color="#67C23A" v-if="item.status === 'online'">
                                <circleCheck />
                            </el-icon>
                            <el-button
                                v-if="item.status === 'closed'"
                                icon="Refresh"
                                style="color: white"
                                size="default"
                                link
                                @click="onReconnect(item)"
                            />
                            <span>&nbsp;{{ item.title }}&nbsp;&nbsp;</span>
                        </span>
                    </template>
                    <Terminal
                        style="height: calc(100vh - 178px); background-color: #000"
                        :ref="'Ref' + item.key"
                        :key="item.Refresh"
                        :wsID="item.wsID"
                        :terminalID="item.key"
                    ></Terminal>
                    <div>
                        <el-select
                            v-model="quickCmd"
                            clearable
                            filterable
                            @change="quickInput"
                            style="width: 25%"
                            :placeholder="$t('terminal.quickCommand')"
                        >
                            <el-option
                                v-for="cmd in commandList"
                                :key="cmd.id"
                                :label="cmd.name + ' [ ' + cmd.command + ' ] '"
                                :value="cmd.command"
                            />
                        </el-select>
                        <el-input
                            :placeholder="$t('terminal.batchInput')"
                            v-model="batchVal"
                            @keyup.enter="batchInput"
                            style="width: 75%"
                        >
                            <template #append>
                                <el-switch v-model="isBatch" class="ml-2" />
                            </template>
                        </el-input>
                    </div>
                </el-tab-pane>
                <el-tab-pane :closable="false" name="newTabs">
                    <template #label>
                        <el-button
                            v-popover="popoverRef"
                            style="background-color: #ededed; border: 0"
                            icon="Plus"
                        ></el-button>
                        <el-popover ref="popoverRef" width="250px" trigger="hover" virtual-triggering persistent>
                            <el-button-group style="width: 100%">
                                <el-button @click="onNewSsh">New ssh</el-button>
                                <el-button @click="onConnLocal">New tab</el-button>
                            </el-button-group>
                            <el-input clearable style="margin-top: 5px" v-model="hostfilterInfo">
                                <template #append><el-button icon="search" /></template>
                            </el-input>
                            <el-tree
                                ref="treeRef"
                                :expand-on-click-node="false"
                                node-key="id"
                                :default-expand-all="true"
                                :data="hostTree"
                                :props="defaultProps"
                                :filter-node-method="filterHost"
                            >
                                <template #default="{ node, data }">
                                    <span class="custom-tree-node">
                                        <span>
                                            <a @click="onConn(node, data)">{{ node.label }}</a>
                                        </span>
                                    </span>
                                </template>
                            </el-tree>
                        </el-popover>
                    </template>
                </el-tab-pane>
                <div v-if="terminalTabs.length === 0">
                    <el-empty
                        style="background-color: #000; height: calc(100vh - 150px)"
                        :description="$t('terminal.emptyTerminal')"
                    ></el-empty>
                </div>
            </el-tabs>
            <el-button @click="toggleFullscreen" class="fullScreen" icon="FullScreen"></el-button>
        </div>
        <div v-if="activeNames === 'host'"><HostTab ref="hostTabRef" /></div>
        <div v-if="activeNames === 'command'"><CommandTab ref="commandTabRef" /></div>

        <el-dialog v-model="connVisiable" :title="$t('terminal.addHost')" width="30%">
            <el-form ref="hostInfoRef" label-width="100px" label-position="left" :model="hostInfo" :rules="rules">
                <el-form-item :label="$t('commons.table.name')" prop="name">
                    <el-input clearable v-model="hostInfo.name" />
                </el-form-item>
                <el-form-item label="IP" prop="addr">
                    <el-input clearable v-model="hostInfo.addr" />
                </el-form-item>
                <el-form-item :label="$t('terminal.port')" prop="port">
                    <el-input clearable v-model.number="hostInfo.port" />
                </el-form-item>
                <el-form-item :label="$t('terminal.user')" prop="user">
                    <el-input clearable v-model="hostInfo.user" />
                </el-form-item>
                <el-form-item :label="$t('terminal.authMode')" prop="authMode">
                    <el-radio-group v-model="hostInfo.authMode">
                        <el-radio label="password">{{ $t('terminal.passwordMode') }}</el-radio>
                        <el-radio label="key">{{ $t('terminal.keyMode') }}</el-radio>
                    </el-radio-group>
                </el-form-item>
                <el-form-item :label="$t('terminal.password')" v-if="hostInfo.authMode === 'password'" prop="password">
                    <el-input clearable show-password type="password" v-model="hostInfo.password" />
                </el-form-item>
                <el-form-item :label="$t('terminal.key')" v-if="hostInfo.authMode === 'key'" prop="privateKey">
                    <el-input clearable type="textarea" v-model="hostInfo.privateKey" />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="connVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button @click="submitAddHost(hostInfoRef, 'testConn')">
                        {{ $t('terminal.testConn') }}
                    </el-button>
                    <el-button type="primary" @click="submitAddHost(hostInfoRef, 'saveAndConn')">
                        {{ $t('terminal.saveAndConn') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeMount, ref, watch, reactive, getCurrentInstance } from 'vue';
import { Rules } from '@/global/form-rules';
import { testConn, getHostTree, addHost } from '@/api/modules/host';
import { getCommandList } from '@/api/modules/command';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Host } from '@/api/interface/host';
import { ElMessage } from 'element-plus';
import Terminal from '@/views/host/terminal/terminal/index.vue';
import HostTab from '@/views/host/terminal/host/index.vue';
import CommandTab from '@/views/host/terminal/command/index.vue';
import type Node from 'element-plus/es/components/tree/src/model/node';
import { ElTree } from 'element-plus';
import screenfull from 'screenfull';

let timer: NodeJS.Timer | null = null;

const activeNames = ref<string>('terminal');
const hostTabRef = ref();
const commandTabRef = ref();

const terminalValue = ref();
const terminalTabs = ref([]) as any;
let tabIndex = 0;
const commandList = ref();
let quickCmd = ref();
let batchVal = ref();
let isBatch = ref<boolean>(false);
const popoverRef = ref();

const connVisiable = ref<boolean>(false);
type FormInstance = InstanceType<typeof ElForm>;

const hostInfoRef = ref<FormInstance>();
const hostTree = ref<Array<Host.HostTree>>();
const treeRef = ref<InstanceType<typeof ElTree>>();
const defaultProps = {
    label: 'label',
    children: 'children',
};
const hostfilterInfo = ref('');
interface Tree {
    id: number;
    label: string;
    children?: Tree[];
}
const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
    addr: [Rules.requiredInput, Rules.ip],
    port: [Rules.requiredInput, Rules.port],
    user: [Rules.requiredInput],
    authMode: [Rules.requiredSelect],
    password: [Rules.requiredInput],
    privateKey: [Rules.requiredInput],
});

const localHostID = ref();

let hostInfo = reactive<Host.HostOperate>({
    id: 0,
    name: '',
    groupBelong: '',
    addr: '',
    port: 22,
    user: '',
    authMode: 'password',
    password: '',
    privateKey: '',
    description: '',
});

const ctx = getCurrentInstance() as any;

function toggleFullscreen() {
    if (screenfull.isEnabled) {
        screenfull.toggle();
    }
}
const handleChange = (tab: any) => {
    if (tab === 'host') {
        hostTabRef.value!.onInit();
    }
    if (tab === 'command') {
        commandTabRef.value!.onInit();
    }
    if (tab === 'terminal') {
        loadCommand();
        loadHost();
    }
};

const handleTabsRemove = (targetName: string, action: 'remove' | 'add') => {
    if (action !== 'remove') {
        return;
    }
    if (ctx) {
        ctx.refs[`Ref${targetName}`] && ctx.refs[`Ref${targetName}`][0].onClose();
    }
    const tabs = terminalTabs.value;
    let activeName = terminalValue.value;
    if (activeName === targetName) {
        tabs.forEach((tab: any, index: any) => {
            if (tab.key === targetName) {
                const nextTab = tabs[index + 1] || tabs[index - 1];
                if (nextTab) {
                    activeName = nextTab.key;
                }
            }
        });
    }
    terminalValue.value = activeName;
    terminalTabs.value = tabs.filter((tab: any) => tab.key !== targetName);
};

const loadHost = async () => {
    const res = await getHostTree({});
    hostTree.value = res.data;
    for (const item of hostTree.value) {
        if (!item.children) {
            continue;
        }
        for (const host of item.children) {
            if (host.label.indexOf('127.0.0.1') !== -1) {
                localHostID.value = host.id;
                if (terminalTabs.value.length === 0) {
                    onConnLocal();
                }
                return;
            }
        }
    }
    hostInfo.name = 'localhost';
    hostInfo.addr = '127.0.0.1';
    connVisiable.value = true;
};
watch(hostfilterInfo, (val: any) => {
    treeRef.value!.filter(val);
});
const filterHost = (value: string, data: any) => {
    if (!value) return true;
    return data.label.includes(value);
};
const loadCommand = async () => {
    const res = await getCommandList();
    commandList.value = res.data;
};

function quickInput(val: any) {
    if (val !== '') {
        if (ctx) {
            ctx.refs[`Ref${terminalValue.value}`] && ctx.refs[`Ref${terminalValue.value}`][0].onSendMsg(val + '\n');
            quickCmd.value = '';
        }
    }
}

function batchInput() {
    if (batchVal.value === '' || !ctx) {
        return;
    }
    if (isBatch.value) {
        for (const tab of terminalTabs.value) {
            ctx.refs[`Ref${tab.key}`] && ctx.refs[`Ref${tab.key}`][0].onSendMsg(batchVal.value + '\n');
        }
        batchVal.value = '';
        return;
    }
    ctx.refs[`Ref${terminalValue.value}`] && ctx.refs[`Ref${terminalValue.value}`][0].onSendMsg(batchVal.value + '\n');
    batchVal.value = '';
}

function beforeLeave(activeName: string) {
    if (activeName === 'newTabs') {
        return false;
    }
}

const onNewSsh = () => {
    connVisiable.value = true;
    if (hostInfoRef.value) {
        hostInfoRef.value.resetFields();
    }
};

const onConn = (node: Node, data: Tree) => {
    if (node.level === 1) {
        return;
    }
    let addr = data.label.split('@')[1].split(':')[0];
    terminalTabs.value.push({
        key: `${addr}-${++tabIndex}`,
        title: addr,
        wsID: data.id,
        status: 'online',
    });
    terminalValue.value = `${addr}-${tabIndex}`;
};

const onReconnect = async (item: any) => {
    if (ctx) {
        ctx.refs[`Ref${item.key}`] && ctx.refs[`Ref${item.key}`][0].onClose();
    }
    item.Refresh = !item.Refresh;
    ctx.refs[`Ref${item.key}`];
    syncTerminal();
};

const submitAddHost = (formEl: FormInstance | undefined, ops: string) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        hostInfo.groupBelong = 'default';
        switch (ops) {
            case 'testConn':
                await testConn(hostInfo);
                ElMessage.success(i18n.global.t('terminal.connTestOk'));
                break;
            case 'saveAndConn':
                const res = await addHost(hostInfo);
                terminalTabs.value.push({
                    key: `${res.data.addr}-${++tabIndex}`,
                    title: res.data.addr,
                    wsID: res.data.id,
                    status: 'online',
                });
                terminalValue.value = `${res.data.addr}-${tabIndex}`;
                connVisiable.value = false;
                loadHost();
        }
    });
};

const onConnLocal = () => {
    terminalTabs.value.push({
        key: `127.0.0.1-${++tabIndex}`,
        title: i18n.global.t('terminal.localhost'),
        wsID: localHostID.value,
        status: 'online',
    });
    terminalValue.value = `127.0.0.1-${tabIndex}`;
};

function syncTerminal() {
    for (const terminal of terminalTabs.value) {
        if (ctx && ctx.refs[`Ref${terminal.key}`][0]) {
            terminal.status = ctx.refs[`Ref${terminal.key}`][0].isWsOpen() ? 'online' : 'closed';
        }
    }
}

onMounted(() => {
    loadHost();
    loadCommand();
    timer = setInterval(() => {
        syncTerminal();
    }, 1000 * 5);
});
onBeforeMount(() => {
    clearInterval(Number(timer));
    timer = null;
});
</script>
<style lang="scss" scoped>
.terminal-tabs {
    :deep .el-tabs__header {
        padding: 0;
        position: relative;
        margin: 0 0 3px 0;
    }
    ::deep .el-tabs__nav {
        white-space: nowrap;
        position: relative;
        transition: transform var(--el-transition-duration);
        float: left;
        z-index: calc(var(--el-index-normal) + 1);
    }
    :deep .el-tabs__item {
        color: #575758;
        padding: 0 0px;
    }
    :deep .el-tabs__item.is-active {
        color: #ebeef5;
        background-color: #575758;
    }
}

.vertical-tabs > .el-tabs__content {
    padding: 32px;
    color: #6b778c;
    font-size: 32px;
    font-weight: 600;
}
.fullScreen {
    position: absolute;
    right: 50px;
    top: 86px;
    font-weight: 600;
    font-size: 14px;
}
.el-tabs--top.el-tabs--card > .el-tabs__header .el-tabs__item:last-child {
    padding-right: 0px;
}
.topCard {
    --el-card-border-color: var(--el-border-color-light);
    --el-card-border-radius: 4px;
    --el-card-padding: 0px;
    --el-card-bg-color: var(--el-fill-color-blank);
}
.topButton .el-radio-button__inner {
    display: inline-block;
    line-height: 1;
    white-space: nowrap;
    vertical-align: middle;
    background: var(--el-button-bg-color, var(--el-fill-color-blank));
    border: 0;
    font-weight: 350;
    border-left: 0;
    color: var(--el-button-text-color, var(--el-text-color-regular));
    text-align: center;
    box-sizing: border-box;
    outline: 0;
    margin: 0;
    position: relative;
    cursor: pointer;
    transition: var(--el-transition-all);
    -webkit-user-select: none;
    user-select: none;
    padding: 8px 15px;
    font-size: var(--el-font-size-base);
    border-radius: 0;
}
</style>
