<template>
    <div v-loading="loading">
        <LayoutContent backName="PostgreSQL">
            <template #leftToolBar>
                <el-text class="mx-1">
                    {{ props.database }}
                </el-text>
                <el-divider direction="vertical" />
                <el-button type="primary" :plain="activeName !== 'conf'" @click="jumpToConf">
                    {{ $t('database.confChange') }}
                </el-button>
                <el-button type="primary" :plain="activeName !== 'port'" @click="activeName = 'port'">
                    {{ $t('commons.table.port') }}
                </el-button>
                <el-button
                    type="primary"
                    :disabled="postgresqlStatus !== 'Running'"
                    :plain="activeName !== 'log'"
                    @click="activeName = 'log'"
                >
                    {{ $t('commons.button.log') }}
                </el-button>
            </template>

            <template #app>
                <AppStatus :app-key="props.type" :app-name="props.database" v-model:loading="loading" />
            </template>

            <template #main>
                <div v-if="activeName === 'conf'">
                    <CodemirrorPro :heightDiff="320" v-model="postgresqlConf"></CodemirrorPro>
                    <el-button type="primary" class="mt-5" @click="onSaveConf">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </div>
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
                <ContainerLog v-if="activeName === 'log'" :container="baseInfo.containerID" :highlightDiff="350" />
            </template>
        </LayoutContent>

        <DialogPro v-model="open" :title="$t('app.checkTitle')" size="small">
            <el-alert :closable="false" :title="$t('database.confNotFound')" type="info">
                <el-link icon="Position" @click="goUpgrade()" type="primary">
                    {{ $t('database.goUpgrade') }}
                </el-link>
            </el-alert>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </DialogPro>

        <ConfirmDialog ref="confirmPortRef" @confirm="onSubmitChangePort"></ConfirmDialog>
        <ConfirmDialog ref="confirmConfRef" @confirm="onSubmitChangeConf"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import { FormInstance } from 'element-plus';
import ContainerLog from '@/components/log/container/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { loadDBFile, loadDBBaseInfo, updateDBFile } from '@/api/modules/database';
import { changePort, checkAppInstalled } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import { routerToName } from '@/utils/router';

const loading = ref(false);

const activeName = ref('conf');

const baseInfo = reactive({
    name: '',
    port: 5432,
    password: '',
    remoteConn: false,
    containerID: '',
});
const panelFormRef = ref<FormInstance>();
const postgresqlConf = ref();
const open = ref();

const postgresqlName = ref();
const postgresqlStatus = ref();
const postgresqlVersion = ref();

interface DBProps {
    type: string;
    database: string;
}
const props = withDefaults(defineProps<DBProps>(), {
    type: '',
    database: '',
});

const jumpToConf = async () => {
    activeName.value = 'conf';
    loadPostgresqlConf();
};

const onSubmitChangePort = async () => {
    let params = {
        key: props.type,
        name: props.database,
        port: baseInfo.port,
    };
    loading.value = true;
    await changePort(params)
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

const onSubmitChangeConf = async () => {
    let param = {
        type: props.type,
        database: props.database,
        file: postgresqlConf.value,
    };
    loading.value = true;
    await updateDBFile(param)
        .then(() => {
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

const loadBaseInfo = async () => {
    const res = await loadDBBaseInfo(props.type, props.database);
    postgresqlName.value = res.data?.name;
    baseInfo.port = res.data?.port;
    baseInfo.containerID = res.data?.containerName;
    loadPostgresqlConf();
};

const loadPostgresqlConf = async () => {
    await loadDBFile(props.type + '-conf', props.database)
        .then((res) => {
            loading.value = false;
            postgresqlConf.value = res.data;
        })
        .catch(() => {
            open.value = true;
            loading.value = false;
        });
};

const goUpgrade = () => {
    routerToName('AppUpgrade');
};

const onLoadInfo = async () => {
    await checkAppInstalled(props.type, props.database).then((res) => {
        postgresqlName.value = res.data.name;
        postgresqlStatus.value = res.data.status;
        postgresqlVersion.value = res.data.version;
        loadBaseInfo();
    });
};

onMounted(() => {
    onLoadInfo();
});
</script>
