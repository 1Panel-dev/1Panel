<template>
    <div v-show="onSetting" v-loading="loading">
        <LayoutContent :title="'MySQL ' + $t('commons.button.set')" :reload="true">
            <template #buttons>
                <el-button type="primary" :plain="activeName !== 'conf'" @click="jumpToConf">
                    {{ $t('database.confChange') }}
                </el-button>
                <el-button
                    type="primary"
                    :disabled="mysqlStatus !== 'Running'"
                    :plain="activeName !== 'status'"
                    @click="activeName = 'status'"
                >
                    {{ $t('database.currentStatus') }}
                </el-button>
                <el-button
                    type="primary"
                    :disabled="mysqlStatus !== 'Running'"
                    :plain="activeName !== 'tuning'"
                    @click="activeName = 'tuning'"
                >
                    {{ $t('database.performanceTuning') }}
                </el-button>
                <el-button type="primary" :plain="activeName !== 'port'" @click="activeName = 'port'">
                    {{ $t('commons.table.port') }}
                </el-button>
                <el-button
                    type="primary"
                    :disabled="mysqlStatus !== 'Running'"
                    :plain="activeName !== 'log'"
                    @click="activeName = 'log'"
                >
                    {{ $t('database.log') }}
                </el-button>
                <el-button
                    type="primary"
                    :disabled="mysqlStatus !== 'Running'"
                    @click="jumpToSlowlog"
                    :plain="activeName !== 'slowLog'"
                >
                    {{ $t('database.slowLog') }}
                </el-button>
            </template>
            <template #main>
                <div v-if="activeName === 'conf'">
                    <codemirror
                        :autofocus="true"
                        :placeholder="$t('commons.msg.noneData')"
                        :indent-with-tab="true"
                        :tabSize="4"
                        style="margin-top: 10px; height: calc(100vh - 375px)"
                        :lineWrapping="true"
                        :matchBrackets="true"
                        theme="cobalt"
                        :styleActiveLine="true"
                        :extensions="extensions"
                        v-model="mysqlConf"
                    />
                    <el-button style="margin-top: 10px" @click="getDefaultConfig()">
                        {{ $t('app.defaultConfig') }}
                    </el-button>
                    <el-button type="primary" style="margin-top: 10px" @click="onSaveConf">
                        {{ $t('commons.button.save') }}
                    </el-button>
                    <el-row>
                        <el-col :span="8">
                            <el-alert
                                v-if="useOld"
                                style="margin-top: 10px"
                                :title="$t('app.defaultConfigHelper')"
                                type="info"
                                :closable="false"
                            ></el-alert>
                        </el-col>
                    </el-row>
                </div>
                <Status v-show="activeName === 'status'" ref="statusRef" />
                <Variables @loading="changeLoading" v-show="activeName === 'tuning'" ref="variablesRef" />
                <div v-show="activeName === 'port'">
                    <el-form :model="baseInfo" ref="panelFormRef" label-position="top">
                        <el-row>
                            <el-col :span="1"><br /></el-col>
                            <el-col :span="10">
                                <el-form-item :label="$t('commons.table.port')" prop="port" :rules="Rules.port">
                                    <el-input clearable type="number" v-model.number="baseInfo.port" />
                                </el-form-item>
                                <el-form-item>
                                    <el-button type="primary" @click="onSavePort(panelFormRef)" icon="Collection">
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                </el-form-item>
                            </el-col>
                        </el-row>
                    </el-form>
                </div>
                <ContainerLog v-show="activeName === 'log'" ref="dialogContainerLogRef" />
                <SlowLog
                    @loading="changeLoading"
                    @refresh="loadBaseInfo"
                    v-show="activeName === 'slowLog'"
                    ref="slowLogRef"
                />
            </template>
        </LayoutContent>

        <ConfirmDialog ref="confirmPortRef" @confirm="onSubmitChangePort"></ConfirmDialog>
        <ConfirmDialog ref="confirmConfRef" @confirm="onSubmitChangeConf"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import { FormInstance } from 'element-plus';
import ContainerLog from '@/components/container-log/index.vue';
import Status from '@/views/database/mysql/setting/status/index.vue';
import Variables from '@/views/database/mysql/setting/variables/index.vue';
import SlowLog from '@/views/database/mysql/setting/slow-log/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { LoadFile } from '@/api/modules/files';
import { loadMysqlBaseInfo, loadMysqlVariables, updateMysqlConfByFile } from '@/api/modules/database';
import { ChangePort, GetAppDefaultConfig } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { loadBaseDir } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);

const baseDir = ref();
const extensions = [javascript(), oneDark];
const activeName = ref('conf');

const baseInfo = reactive({
    name: '',
    port: 3306,
    password: '',
    remoteConn: false,
    containerID: '',
});
const panelFormRef = ref<FormInstance>();
const mysqlConf = ref();

const useOld = ref(false);

const statusRef = ref();
const variablesRef = ref();
const slowLogRef = ref();

const onSetting = ref<boolean>(false);
const mysqlName = ref();
const mysqlStatus = ref();
const mysqlVersion = ref();
const variables = ref();

interface DialogProps {
    mysqlName: string;
    mysqlVersion: string;
    status: string;
}

const dialogContainerLogRef = ref();
const acceptParams = (props: DialogProps): void => {
    onSetting.value = true;
    mysqlStatus.value = props.status;
    mysqlVersion.value = props.mysqlVersion;
    loadBaseInfo();
    if (mysqlStatus.value === 'Running') {
        loadVariables();
        loadSlowLogs();
        statusRef.value!.acceptParams({ mysqlName: props.mysqlName });
    }
};
const onClose = (): void => {
    onSetting.value = false;
};

const jumpToConf = async () => {
    activeName.value = 'conf';
    const pathRes = await loadBaseDir();
    loadMysqlConf(`${pathRes.data}/apps/mysql/${mysqlName.value}/conf/my.cnf`);
};

const jumpToSlowlog = async () => {
    activeName.value = 'slowLog';
    loadSlowLogs();
};

const onSubmitChangePort = async () => {
    let params = {
        key: 'mysql',
        name: mysqlName.value,
        port: baseInfo.port,
    };
    loading.value = true;
    await ChangePort(params)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};
const confirmPortRef = ref();
const onSavePort = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    const result = await formEl.validateField('port', callback);
    if (!result) {
        return;
    }
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmPortRef.value!.acceptParams(params);
    return;
};
function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

const getDefaultConfig = async () => {
    loading.value = true;
    const res = await GetAppDefaultConfig('mysql');
    mysqlConf.value = res.data;
    useOld.value = true;
    loading.value = false;
};

const onSubmitChangeConf = async () => {
    let param = {
        mysqlName: mysqlName.value,
        file: mysqlConf.value,
    };
    loading.value = true;
    await updateMysqlConfByFile(param)
        .then(() => {
            useOld.value = false;
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};
const confirmConfRef = ref();
const onSaveConf = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmConfRef.value!.acceptParams(params);
    return;
};

const loadContainerLog = async (containerID: string) => {
    dialogContainerLogRef.value!.acceptParams({ containerID: containerID, container: containerID });
};

const loadBaseInfo = async () => {
    const res = await loadMysqlBaseInfo();
    mysqlName.value = res.data?.name;
    baseInfo.port = res.data?.port;
    baseInfo.containerID = res.data?.containerName;
    const pathRes = await loadBaseDir();
    baseDir.value = pathRes.data;
    loadMysqlConf(`${pathRes.data}/apps/mysql/${mysqlName.value}/conf/my.cnf`);
    loadContainerLog(baseInfo.containerID);
};

const changeLoading = (status: boolean) => {
    loading.value = status;
};

const loadVariables = async () => {
    const res = await loadMysqlVariables();
    variables.value = res.data;
    variablesRef.value!.acceptParams({
        mysqlName: mysqlName.value,
        mysqlVersion: mysqlVersion.value,
        variables: res.data,
    });
};

const loadSlowLogs = async () => {
    const res = await loadMysqlVariables();
    variables.value = res.data;

    let param = {
        mysqlName: mysqlName.value,
        variables: variables.value,
    };
    slowLogRef.value!.acceptParams(param);
};

const loadMysqlConf = async (path: string) => {
    useOld.value = false;
    const res = await LoadFile({ path: path });
    loading.value = false;
    mysqlConf.value = res.data;
};

defineExpose({
    acceptParams,
    onClose,
});
</script>
