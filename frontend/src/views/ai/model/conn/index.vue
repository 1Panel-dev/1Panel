<template>
    <DrawerPro v-model="drawerVisible" :header="$t('database.databaseConnInfo')" @close="handleClose" size="small">
        <el-form @submit.prevent v-loading="loading" :model="form" label-position="top">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('database.containerConn')">
                        <el-card class="mini-border-card">
                            <el-descriptions :column="1">
                                <el-descriptions-item :label="$t('database.connAddress')">
                                    {{ form.containerName }}
                                    <CopyButton :content="form.containerName" type="icon" />
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('commons.table.port')">
                                    11434
                                    <CopyButton content="11434" type="icon" />
                                </el-descriptions-item>
                            </el-descriptions>
                        </el-card>
                        <span class="input-help">
                            {{ $t('aiTools.model.container_conn_helper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('setting.proxyUrl')" v-if="bindDomain.connUrl != ''">
                        <el-card class="mini-border-card">
                            <el-descriptions :column="1">
                                <el-descriptions-item :label="$t('database.connAddress')">
                                    {{ bindDomain.connUrl }}
                                    <CopyButton :content="bindDomain.connUrl" type="icon" />
                                </el-descriptions-item>
                            </el-descriptions>
                        </el-card>
                        <span class="input-help">
                            {{ $t('database.remoteConnHelper2') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('database.remoteConn')" v-else>
                        <el-card class="mini-border-card">
                            <el-descriptions :column="1">
                                <el-descriptions-item :label="$t('database.connAddress')">
                                    {{ form.systemIP }}
                                    <CopyButton :content="form.systemIP" type="icon" />
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('commons.table.port')">
                                    {{ form.port }}
                                    <CopyButton :content="form.port + ''" type="icon" />
                                </el-descriptions-item>
                            </el-descriptions>
                        </el-card>
                        <span class="input-help">
                            {{ $t('database.remoteConnHelper2') }}
                        </span>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>

        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { getSettingInfo } from '@/api/modules/setting';
import { getBindDomain } from '@/api/modules/ai';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const loading = ref(false);

const drawerVisible = ref(false);
const form = reactive({
    systemIP: '',
    containerName: '',
    port: 0,

    remoteIP: '',
});
const bindDomain = ref({
    connUrl: '',
});

interface DialogProps {
    port: number;
    containerName: string;
    appinstallID: number;
}

const acceptParams = (param: DialogProps): void => {
    form.containerName = param.containerName;
    form.port = param.port;
    loadSystemIP();
    loadBindDomain(param.appinstallID);
    drawerVisible.value = true;
};

const handleClose = () => {
    drawerVisible.value = false;
};

const loadSystemIP = async () => {
    if (globalStore.currentNode !== 'local') {
        form.systemIP = globalStore.currentNode || i18n.global.t('database.localIP');
        return;
    }
    const res = await getSettingInfo();
    form.systemIP = res.data.systemIP || i18n.global.t('database.localIP');
};

const loadBindDomain = async (appInstallID: number) => {
    if (appInstallID == undefined || appInstallID <= 0) {
        return;
    }
    try {
        const res = await getBindDomain({
            appInstallID: appInstallID,
        });
        if (res.data.websiteID > 0) {
            bindDomain.value.connUrl = res.data.connUrl;
        }
    } catch (e) {}
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.copy_button {
    border-radius: 0px;
    border-left-width: 0px;
}
:deep(.el-input__wrapper) {
    border-top-right-radius: 0px;
    border-bottom-right-radius: 0px;
}
</style>
