<template>
    <LayoutContent :header="$t('menu.terminal')">
        <div>
            <el-tabs editable type="card" v-model="terminalValue" @edit="handleTabsEdit">
                <el-tab-pane :key="item.name" v-for="item in terminalTabs" :label="item.title" :name="item.name">
                    <iframe
                        v-if="item.type === 'local'"
                        id="iframeTerminal"
                        name="iframeTerminal"
                        width="100%"
                        frameborder="0"
                        :src="item.src"
                    />
                    <Terminal v-else :ref="'Ref' + item.name" :id="item.wsID"></Terminal>
                </el-tab-pane>
            </el-tabs>
        </div>
        <el-button class="term-tool-button" icon="arrowLeftBold" @click="hostDrawer = true"></el-button>

        <el-drawer :size="320" v-model="hostDrawer" title="历史主机信息" direction="rtl">
            <el-button @click="onAddHost">添加主机</el-button>
            <div v-infinite-scroll="nextPage" style="overflow: auto">
                <div v-for="(item, index) in data" :key="item.id" @mouseover="hover = index" @mouseleave="hover = null">
                    <el-card @click="onConn(item)" style="margin-top: 5px" :title="item.name" shadow="hover">
                        <div :inline="true">
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
                            </span>
                        </div>
                    </el-card>
                </div>
            </div>
        </el-drawer>

        <el-dialog v-model="connVisiable" title="添加主机信息" width="30%">
            <el-form ref="hostInfoRef" label-width="80px" :model="hostInfo" :rules="rules">
                <el-form-item label="名称" prop="name">
                    <el-input v-model="hostInfo.name" style="width: 80%" />
                </el-form-item>
                <el-form-item label="addr" prop="addr">
                    <el-input v-model="hostInfo.addr" style="width: 80%" />
                </el-form-item>
                <el-form-item label="端口" prop="port">
                    <el-input v-model="hostInfo.port" style="width: 80%" />
                </el-form-item>
                <el-form-item label="用户" prop="user">
                    <el-input v-model="hostInfo.user" style="width: 80%" />
                </el-form-item>
                <el-form-item label="认证方式" prop="authMode">
                    <el-radio-group v-model="hostInfo.authMode">
                        <el-radio label="password">密码输入</el-radio>
                        <el-radio label="key">密钥输入</el-radio>
                    </el-radio-group>
                </el-form-item>
                <el-form-item label="密码" show-password v-if="hostInfo.authMode === 'password'" prop="password">
                    <el-input type="password" v-model="hostInfo.password" style="width: 80%" />
                </el-form-item>
                <el-form-item label="密钥" v-if="hostInfo.authMode === 'key'" prop="password">
                    <el-input type="textarea" v-model="hostInfo.privateKey" style="width: 80%" />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="connVisiable = false">取消</el-button>
                    <el-button v-if="operation === 'conn'" type="primary" @click="submitAddHost(hostInfoRef)">
                        连 接
                    </el-button>
                    <el-button v-else type="primary" @click="submitAddHost(hostInfoRef)"> 提 交 </el-button>
                </span>
            </template>
        </el-dialog>
    </LayoutContent>
</template>

<script setup lang="ts">
import { onMounted, ref, nextTick, reactive, getCurrentInstance } from 'vue';
import { Rules } from '@/global/form-rues';
import { getHostList, addHost, editHost, deleteHost } from '@/api/modules/host';
import { useDeleteData } from '@/hooks/use-delete-data';
import LayoutContent from '@/layout/layout-content.vue';
import i18n from '@/lang';
import type { ElForm } from 'element-plus';
import { Host } from '@/api/interface/host';
import { ElMessage } from 'element-plus';
import Terminal from '@/views/terminal/terminal.vue';

const terminalValue = ref();
const terminalTabs = ref([]) as any;
const hostDrawer = ref(false);
const data = ref();

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
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
    } else if (action === 'remove') {
        if (ctx) {
            ctx.refs[`Ref${targetName}`] && ctx.refs[`Ref${targetName}`][0].onClose();
        }
        const tabs = terminalTabs.value;
        let activeName = terminalValue.value;
        if (activeName === targetName) {
            tabs.forEach((tab: any, index: any) => {
                if (tab.name === targetName) {
                    const nextTab = tabs[index + 1] || tabs[index - 1];
                    if (nextTab) {
                        activeName = nextTab.name;
                    }
                }
            });
        }
        terminalValue.value = activeName;
        terminalTabs.value = tabs.filter((tab: any) => tab.name !== targetName);
    }
};

const loadHost = async () => {
    const res = await getHostList({ page: paginationConfig.currentPage, pageSize: paginationConfig.pageSize });
    data.value = res.data.items;
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
                        name: res.data.addr,
                        title: res.data.addr,
                        wsID: res.data.id,
                        type: 'remote',
                    });
                    terminalValue.value = res.data.addr;
            }
            connVisiable.value = false;
            loadHost();
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        } catch (error) {
            ElMessage.success(i18n.global.t('commons.msg.loginSuccess') + ':' + error);
        }
    });
};

const onConn = (row: Host.Host) => {
    terminalTabs.value.push({
        name: row.addr,
        title: row.addr,
        wsID: row.id,
        type: 'remote',
    });
    terminalValue.value = row.addr;
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
        ifm.style.height = document.documentElement.clientHeight - 280 + 'px';
    }
}

onMounted(() => {
    terminalTabs.value.push({ name: '本地服务器', title: '本地服务器', src: 'http://localhost:8080', type: 'local' });
    terminalValue.value = '本地服务器';
    nextTick(() => {
        changeFrameHeight();
        window.addEventListener('resize', changeFrameHeight);
    });
    loadHost();
});
</script>
<style lang="scss" scoped>
.term-tool-button {
    position: absolute;
    right: -7px;
    top: 50%;
    width: 28px;
    height: 60px;
    background-color: #565656;
    border-top-left-radius: 30px;
    border-bottom-left-radius: 30px;
    cursor: pointer;
    z-index: 999;
    margin-top: -30px;
}
</style>
