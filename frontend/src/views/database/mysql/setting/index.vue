<template>
    <div class="demo-collapse" v-show="onSetting">
        <el-card style="margin-top: 5px" v-loading="loading">
            <LayoutContent :header="'Mysql ' + $t('database.setting')" :back-name="'Mysql'" :reload="true">
                <el-collapse v-model="activeName" accordion>
                    <el-collapse-item :title="$t('database.confChange')" name="1">
                        <codemirror
                            :autofocus="true"
                            placeholder="None data"
                            :indent-with-tab="true"
                            :tabSize="4"
                            style="margin-top: 10px; max-height: 500px"
                            :lineWrapping="true"
                            :matchBrackets="true"
                            theme="cobalt"
                            :styleActiveLine="true"
                            :extensions="extensions"
                            v-model="mysqlConf"
                            :readOnly="true"
                        />
                        <el-button type="primary" style="margin-top: 10px" @click="onSaveConf">
                            {{ $t('commons.button.save') }}
                        </el-button>
                    </el-collapse-item>
                    <el-collapse-item :title="$t('database.currentStatus')" name="2">
                        <Status ref="statusRef" />
                    </el-collapse-item>
                    <el-collapse-item :title="$t('database.performanceTuning')" name="3">
                        <Variables ref="variablesRef" />
                    </el-collapse-item>
                    <el-collapse-item :title="$t('database.portSetting')" name="4">
                        <el-form :model="baseInfo" ref="panelFormRef" label-width="120px">
                            <el-row>
                                <el-col :span="1"><br /></el-col>
                                <el-col :span="10">
                                    <el-form-item :label="$t('setting.port')" prop="port" :rules="Rules.port">
                                        <el-input clearable type="number" v-model.number="baseInfo.port">
                                            <template #append>
                                                <el-button @click="onSavePort(panelFormRef)" icon="Collection">
                                                    {{ $t('commons.button.save') }}
                                                </el-button>
                                            </template>
                                        </el-input>
                                    </el-form-item>
                                </el-col>
                            </el-row>
                        </el-form>
                    </el-collapse-item>
                    <el-collapse-item :title="$t('database.log')" name="5">
                        <ContainerLog ref="dialogContainerLogRef" />
                    </el-collapse-item>

                    <el-collapse-item :title="$t('database.slowLog')" name="6">
                        <SlowLog ref="slowLogRef" />
                    </el-collapse-item>
                </el-collapse>
            </LayoutContent>
        </el-card>

        <ConfirmDialog ref="confirmPortRef" @confirm="onSubmitChangePort"></ConfirmDialog>
        <ConfirmDialog ref="confirmConfRef" @confirm="onSubmitChangeConf"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import { ElMessage, FormInstance } from 'element-plus';
import LayoutContent from '@/layout/layout-content.vue';
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
import { ChangePort } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';

const loading = ref(false);

const extensions = [javascript(), oneDark];
const activeName = ref('1');

const baseInfo = reactive({
    name: '',
    port: 3306,
    password: '',
    remoteConn: false,
    containerID: '',
});
const panelFormRef = ref<FormInstance>();
const mysqlConf = ref();

const statusRef = ref();
const variablesRef = ref();
const slowLogRef = ref();

const onSetting = ref<boolean>(false);
const mysqlName = ref();
const variables = ref();

interface DialogProps {
    mysqlName: string;
}

const dialogContainerLogRef = ref();
const acceptParams = (params: DialogProps): void => {
    onSetting.value = true;
    loadBaseInfo();
    loadVariables();
    loadSlowLogs();
    statusRef.value!.acceptParams({ mysqlName: params.mysqlName });
};
const onClose = (): void => {
    onSetting.value = false;
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
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
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

const onSubmitChangeConf = async () => {
    let param = {
        mysqlName: mysqlName.value,
        file: mysqlConf.value,
    };
    loading.value = true;
    await updateMysqlConfByFile(param)
        .then(() => {
            loading.value = false;
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
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
    dialogContainerLogRef.value!.acceptParams({ containerID: containerID, container: mysqlName.value });
};

const loadBaseInfo = async () => {
    const res = await loadMysqlBaseInfo();
    mysqlName.value = res.data?.name;
    baseInfo.name = res.data?.name;
    baseInfo.port = res.data?.port;
    baseInfo.password = res.data?.password;
    baseInfo.remoteConn = res.data?.remoteConn;
    baseInfo.containerID = res.data?.containerName;
    loadMysqlConf(`/opt/1Panel/data/apps/mysql/${baseInfo.name}/conf/my.cnf`);
    loadContainerLog(baseInfo.containerID);
};

const loadVariables = async () => {
    const res = await loadMysqlVariables();
    variables.value = res.data;
    variablesRef.value!.acceptParams({ mysqlName: mysqlName.value, variables: res.data });
};

const loadSlowLogs = async () => {
    await Promise.all([loadBaseInfo(), loadVariables()]);
    let param = {
        mysqlName: mysqlName.value,
        variables: variables.value,
    };
    slowLogRef.value!.acceptParams(param);
};

const loadMysqlConf = async (path: string) => {
    const res = await LoadFile({ path: path });
    loading.value = false;
    mysqlConf.value = res.data;
};

defineExpose({
    acceptParams,
    onClose,
});
</script>
