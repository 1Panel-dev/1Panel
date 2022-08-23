<template>
    <LayoutContent :header="$t('menu.terminal')">
        <el-button class="drawer-container" icon="arrowLeftBold" @click="hostDrawer = true">
            {{ $t('terminal.connHistory') }}
        </el-button>

        <div>
            <el-tabs
                type="card"
                editable
                style="background-color: #efefef"
                v-model="terminalValue"
                @edit="handleTabsEdit"
            >
                <el-tab-pane :key="item.key" v-for="item in terminalTabs" :label="item.title" :name="item.key">
                    <template #label>
                        <span class="custom-tabs-label">
                            <el-icon color="#67C23A" v-if="item.status === 'online'"><circleCheck /></el-icon>
                            <el-icon color="#F56C6C" v-if="item.status === 'closed'"><circleClose /></el-icon>
                            <span> &nbsp;{{ item.title }}&nbsp;&nbsp;</span>
                        </span>
                    </template>
                    <Terminal
                        style="height: calc(100vh - 265px); background-color: #000"
                        :ref="'Ref' + item.key"
                        :wsID="item.wsID"
                        :terminalID="item.key"
                    ></Terminal>
                </el-tab-pane>
                <div v-if="terminalTabs.length === 0">
                    <el-empty
                        style="background-color: #000; height: calc(100vh - 265px)"
                        :description="$t('terminal.emptyTerminal')"
                    ></el-empty>
                </div>
            </el-tabs>
        </div>

        <el-drawer :size="320" v-model="hostDrawer" :title="$t('terminal.hostHistory')" direction="rtl">
            <el-button @click="onAddHost">{{ $t('terminal.addHost') }}</el-button>
            <div v-infinite-scroll="nextPage" style="overflow: auto">
                <el-card
                    @click="onConnLocal()"
                    style="margin-top: 5px; cursor: pointer"
                    :title="$t('terminal.localhost')"
                    shadow="hover"
                >
                    <div :inline="true">
                        <div>
                            <span>{{ $t('terminal.localhost') }}</span>
                        </div>
                        <span style="font-size: 14px; line-height: 25px"> [ 127.0.0.1 ]</span>
                    </div>
                </el-card>
                <div v-for="(item, index) in data" :key="item.id" @mouseover="hover = index" @mouseleave="hover = null">
                    <el-card @click="onConn(item)" style="margin-top: 5px; cursor: pointer" shadow="hover">
                        <div :inline="true">
                            <div>
                                <span>{{ item.name }}</span>
                            </div>
                            <span style="font-size: 14px; line-height: 25px">
                                [ {{ item.addr + ':' + item.port }} ]
                                <el-button
                                    style="float: right; margin-left: 5px"
                                    size="small"
                                    circle
                                    @click="onDeleteHost(item)"
                                    v-if="hover === index"
                                    icon="delete"
                                ></el-button>
                                <el-button
                                    style="float: right; margin-left: 5px"
                                    size="small"
                                    circle
                                    @click="onEditHost(item)"
                                    v-if="hover === index"
                                    icon="edit"
                                ></el-button>
                                <div v-if="item.description && hover === index">
                                    <span style="font-size: 12px">{{ item.description }}</span>
                                </div>
                            </span>
                        </div>
                    </el-card>
                </div>
            </div>
        </el-drawer>

        <el-dialog v-model="connVisiable" :title="$t('terminal.addHost')" width="30%">
            <el-form ref="hostInfoRef" label-width="80px" :model="hostInfo" :rules="rules">
                <el-form-item :label="$t('commons.table.name')" prop="name">
                    <el-input v-model="hostInfo.name" style="width: 80%" />
                </el-form-item>
                <el-form-item label="IP" prop="addr">
                    <el-input v-model="hostInfo.addr" style="width: 80%" />
                </el-form-item>
                <el-form-item :label="$t('terminal.port')" prop="port">
                    <el-input v-model.number="hostInfo.port" style="width: 80%" />
                </el-form-item>
                <el-form-item :label="$t('terminal.user')" prop="user">
                    <el-input v-model="hostInfo.user" style="width: 80%" />
                </el-form-item>
                <el-form-item :label="$t('terminal.authMode')" prop="authMode">
                    <el-radio-group v-model="hostInfo.authMode">
                        <el-radio label="password">{{ $t('terminal.passwordMode') }}</el-radio>
                        <el-radio label="key">{{ $t('terminal.keyMode') }}</el-radio>
                    </el-radio-group>
                </el-form-item>
                <el-form-item :label="$t('terminal.password')" v-if="hostInfo.authMode === 'password'" prop="password">
                    <el-input show-password type="password" v-model="hostInfo.password" style="width: 80%" />
                </el-form-item>
                <el-form-item :label="$t('terminal.key')" v-if="hostInfo.authMode === 'key'" prop="privateKey">
                    <el-input type="textarea" v-model="hostInfo.privateKey" style="width: 80%" />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="connVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button v-if="operation === 'conn'" type="primary" @click="submitAddHost(hostInfoRef)">
                        {{ $t('commons.button.conn') }}
                    </el-button>
                    <el-button v-else type="primary" @click="submitAddHost(hostInfoRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </LayoutContent>
</template>

<script setup lang="ts">
import { onMounted, onBeforeMount, ref, nextTick, reactive, getCurrentInstance } from 'vue';
import { Rules } from '@/global/form-rues';
import { getHostList, addHost, editHost, deleteHost } from '@/api/modules/host';
import { useDeleteData } from '@/hooks/use-delete-data';
import LayoutContent from '@/layout/layout-content.vue';
import i18n from '@/lang';
import type { ElForm } from 'element-plus';
import { Host } from '@/api/interface/host';
import { ElMessage } from 'element-plus';
import Terminal from '@/views/terminal/terminal.vue';

let timer: NodeJS.Timer | null = null;

const terminalValue = ref();
const terminalTabs = ref([]) as any;
let tabIndex = 0;
const data = ref();
const hostDrawer = ref(false);

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 8,
    total: 0,
});

const connVisiable = ref<boolean>(false);
const operation = ref();
const hover = ref();
type FormInstance = InstanceType<typeof ElForm>;
const hostInfoRef = ref<FormInstance>();
const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
    addr: [Rules.requiredInput, Rules.ip],
    port: [Rules.requiredInput, Rules.port],
    user: [Rules.requiredInput],
    authMode: [Rules.requiredSelect],
    password: [Rules.requiredInput],
    privateKey: [Rules.requiredInput],
});

let hostInfo = reactive<Host.HostOperate>({
    id: 0,
    name: '',
    addr: '',
    port: 22,
    user: '',
    authMode: 'password',
    password: '',
    privateKey: '',
    description: '',
});

const ctx = getCurrentInstance() as any;

const handleTabsEdit = (targetName: string, action: 'remove' | 'add') => {
    if (action === 'add') {
        connVisiable.value = true;
        operation.value = 'conn';
        if (hostInfoRef.value) {
            hostInfoRef.value.resetFields();
        }
    } else if (action === 'remove') {
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
    }
};

const loadHost = async () => {
    const res = await getHostList({ page: paginationConfig.currentPage, pageSize: paginationConfig.pageSize });
    data.value = res.data.items;
    paginationConfig.total = res.data.total;
};

const nextPage = () => {
    if (paginationConfig.pageSize >= paginationConfig.total) {
        return;
    }
    paginationConfig.pageSize = paginationConfig.pageSize + 3;
    loadHost();
};

function onAddHost() {
    connVisiable.value = true;
    operation.value = 'create';
    if (hostInfoRef.value) {
        hostInfoRef.value.resetFields();
    }
}

function onEditHost(row: Host.Host) {
    hostInfo.id = row.id;
    hostInfo.name = row.name;
    hostInfo.addr = row.addr;
    hostInfo.port = row.port;
    hostInfo.user = row.user;
    hostInfo.authMode = row.authMode;
    hostInfo.password = '';
    hostInfo.privateKey = '';
    operation.value = 'update';
    connVisiable.value = true;
}

const submitAddHost = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        try {
            switch (operation.value) {
                case 'create':
                    await addHost(hostInfo);
                    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
                    break;
                case 'update':
                    await editHost(hostInfo);
                    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
                    break;
                case 'conn':
                    const res = await addHost(hostInfo);
                    terminalTabs.value.push({
                        key: `${res.data.addr}-${++tabIndex}`,
                        title: res.data.addr,
                        wsID: res.data.id,
                        status: 'online',
                    });
                    terminalValue.value = `${res.data.addr}-${tabIndex}`;
            }
            connVisiable.value = false;
            loadHost();
        } catch (error) {
            ElMessage.success(i18n.global.t('commons.msg.loginSuccess') + ':' + error);
        }
    });
};

const onConn = (row: Host.Host) => {
    terminalTabs.value.push({
        key: `${row.addr}-${++tabIndex}`,
        title: row.addr,
        wsID: row.id,
        status: 'online',
    });
    terminalValue.value = `${row.addr}-${tabIndex}`;
    hostDrawer.value = false;
};

const onConnLocal = () => {
    terminalTabs.value.push({
        key: `127.0.0.1-${++tabIndex}`,
        title: '127.0.0.1',
        wsID: 0,
        status: 'online',
    });
    terminalValue.value = `127.0.0.1-${tabIndex}`;
    hostDrawer.value = false;
};

const onDeleteHost = async (row: Host.Host) => {
    let ids: Array<number> = [row.id];
    await useDeleteData(deleteHost, { ids: ids }, 'commons.msg.delete');
    loadHost();
};

function changeFrameHeight() {
    let ifm = document.getElementById('iframeTerminal') as HTMLInputElement | null;
    if (ifm) {
        ifm.style.height = document.documentElement.clientHeight - 300 + 'px';
    }
}

function syncTerminal() {
    for (const terminal of terminalTabs.value) {
        if (ctx && ctx.refs[`Ref${terminal.key}`]) {
            terminal.status = ctx.refs[`Ref${terminal.key}`][0].isWsOpen() ? 'online' : 'closed';
        }
    }
}

onMounted(() => {
    onConnLocal();
    nextTick(() => {
        changeFrameHeight();
        window.addEventListener('resize', changeFrameHeight);
    });
    loadHost();
    timer = setInterval(() => {
        syncTerminal();
    }, 1000 * 8);
});
onBeforeMount(() => {
    clearInterval(Number(timer));
});
</script>
<style lang="scss" scoped>
.drawer-container {
    transition: all 0.2s;
    &:hover {
        right: 0;
    }
    position: fixed;
    right: -90px;
    top: 15%;
    height: 40px;
    width: 130px;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 999;
    border-radius: 4px 0 0 4px;
    cursor: pointer;
}
.el-tabs {
    :deep .el-tabs__header {
        padding: 0;
        position: relative;
        margin: 0 0 3px 0;
    }
    :deep .el-tabs__nav {
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
    :deep .el-tabs__new-tab {
        display: flex;
        align-items: center;
        justify-content: center;
        float: right;
        border: 1pxsolidvar (--el-border-color);
        height: 20px;
        width: 20px;
        line-height: 20px;
        margin: 10px 30px 10px 10px;
        border-radius: 3px;
        text-align: center;
        font-size: 24px;
        color: var(--el-text-color-primary);
        cursor: pointer;
        transition: all 0.15s;
    }
}
</style>
