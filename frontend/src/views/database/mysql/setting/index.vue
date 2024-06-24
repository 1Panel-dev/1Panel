<template>
    <div v-loading="loading">
        <LayoutContent>
            <template #title>
                <back-button name="MySQL" :header="props.database + ' ' + $t('commons.button.set')">
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
                </back-button>
            </template>

            <template #app>
                <AppStatus :app-key="props.type" :app-name="props.database" v-model:loading="loading" />
            </template>

            <template #main>
                <div v-if="activeName === 'conf'">
                    <codemirror
                        :autofocus="true"
                        :placeholder="$t('commons.msg.noneData')"
                        :indent-with-tab="true"
                        :tabSize="4"
                        :style="{ height: `calc(100vh - ${loadHeight()})`, 'margin-top': '10px' }"
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

        <el-dialog
            v-model="upgradeVisible"
            :title="$t('app.checkTitle')"
            width="30%"
            :close-on-click-modal="false"
            :destroy-on-close="true"
        >
            <el-alert :closable="false" :title="$t('database.confNotFound')" type="info">
                <el-link icon="Position" @click="goUpgrade()" type="primary">
                    {{ $t('database.goUpgrade') }}
                </el-link>
            </el-alert>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="upgradeVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

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
import { onMounted, reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { loadDBFile, loadDBBaseInfo, loadMysqlVariables, updateDBFile } from '@/api/modules/database';
import { ChangePort, CheckAppInstalled, GetAppDefaultConfig } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import router from '@/routers';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const loading = ref(false);

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
const upgradeVisible = ref();

const useOld = ref(false);

const statusRef = ref();
const variablesRef = ref();
const slowLogRef = ref();

const mysqlName = ref();
const mysqlStatus = ref();
const mysqlVersion = ref();
const variables = ref();

interface DBProps {
    type: string;
    database: string;
}
const props = withDefaults(defineProps<DBProps>(), {
    type: '',
    database: '',
});

const dialogContainerLogRef = ref();
const jumpToConf = async () => {
    activeName.value = 'conf';
    loadMysqlConf();
};

const loadHeight = () => {
    return globalStore.openMenuTabs ? '405px' : '375px';
};

const jumpToSlowlog = async () => {
    activeName.value = 'slowLog';
    loadSlowLogs();
};

const onSubmitChangePort = async () => {
    let params = {
        key: props.type,
        name: props.database,
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
    await GetAppDefaultConfig(props.type, props.database)
        .then((res) => {
            mysqlConf.value = res.data;
            useOld.value = true;
            loading.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onSubmitChangeConf = async () => {
    let param = {
        type: props.type,
        database: props.database,
        file: mysqlConf.value,
    };
    loading.value = true;
    await updateDBFile(param)
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
    const res = await loadDBBaseInfo(props.type, props.database);
    mysqlName.value = res.data?.name;
    baseInfo.port = res.data?.port;
    baseInfo.containerID = res.data?.containerName;
    loadMysqlConf();
    loadContainerLog(baseInfo.containerID);
};

const changeLoading = (status: boolean) => {
    loading.value = status;
};

const loadVariables = async () => {
    const res = await loadMysqlVariables(props.type, props.database);
    variables.value = res.data;
    variablesRef.value!.acceptParams({
        type: props.type,
        database: props.database,
        version: mysqlVersion.value,
        variables: res.data,
    });
};

const loadSlowLogs = async () => {
    const res = await loadMysqlVariables(props.type, props.database);
    variables.value = res.data;

    let param = {
        type: props.type,
        database: props.database,
        variables: variables.value,
    };
    slowLogRef.value!.acceptParams(param);
};

const loadMysqlConf = async () => {
    useOld.value = false;
    await loadDBFile(props.type + '-conf', props.database)
        .then((res) => {
            loading.value = false;
            mysqlConf.value = res.data;
        })
        .catch(() => {
            upgradeVisible.value = true;
            loading.value = false;
        });
};

const goUpgrade = () => {
    router.push({ name: 'AppUpgrade' });
};

const onLoadInfo = async () => {
    await CheckAppInstalled(props.type, props.database).then((res) => {
        mysqlName.value = res.data.name;
        mysqlStatus.value = res.data.status;
        mysqlVersion.value = res.data.version;
        loadBaseInfo();
        if (mysqlStatus.value === 'Running') {
            loadVariables();
            loadSlowLogs();
            statusRef.value!.acceptParams({ type: props.type, database: props.database });
        }
    });
};

onMounted(() => {
    onLoadInfo();
});
</script>
