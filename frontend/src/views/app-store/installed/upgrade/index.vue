<template>
    <DrawerPro
        v-model="open"
        :header="$t('commons.button.' + operateReq.operate)"
        :resource="resourceName"
        @close="handleClose"
    >
        <div v-loading="loading">
            <el-descriptions direction="vertical">
                <el-descriptions-item>
                    <el-link @click="toLink(app.website)">
                        <el-icon><OfficeBuilding /></el-icon>
                        <span>{{ $t('app.appOfficeWebsite') }}</span>
                    </el-link>
                </el-descriptions-item>
                <el-descriptions-item>
                    <el-link @click="toLink(app.document)">
                        <el-icon><Document /></el-icon>
                        <span>{{ $t('app.document') }}</span>
                    </el-link>
                </el-descriptions-item>
                <el-descriptions-item>
                    <el-link @click="toLink(app.github)">
                        <el-icon><Link /></el-icon>
                        <span>{{ $t('app.github') }}</span>
                    </el-link>
                </el-descriptions-item>
            </el-descriptions>
            <el-form @submit.prevent ref="updateRef" :rules="rules" label-position="top" :model="operateReq">
                <el-form-item :label="$t('app.versionSelect')" prop="detailId">
                    <el-select v-model="operateReq.version" @change="getVersions(operateReq.version)">
                        <el-option
                            v-if="operateReq.operate == 'ignore'"
                            :value="'all'"
                            :label="$t('commons.table.all') + $t('app.version')"
                        ></el-option>
                        <el-option
                            v-for="(version, index) in versions"
                            :key="index"
                            :value="version.version"
                            :label="version.version"
                        ></el-option>
                    </el-select>
                </el-form-item>
                <el-form-item prop="backup" v-if="operateReq.operate === 'upgrade'">
                    <el-checkbox v-model="operateReq.backup" :label="$t('app.backupApp')" />
                    <span class="input-help">
                        <el-text type="warning">{{ $t('app.backupAppHelper') }}</el-text>
                    </span>
                </el-form-item>
                <el-form-item prop="pullImage" v-if="operateReq.operate === 'upgrade'">
                    <el-checkbox v-model="operateReq.pullImage" :label="$t('app.pullImage')" size="large" />
                    <span class="input-help">{{ $t('app.pullImageHelper') }}</span>
                </el-form-item>
            </el-form>

            <div v-if="operateReq.operate === 'upgrade'">
                <el-text type="warning">{{ $t('app.upgradeWarn') }}</el-text>
                <br />
                <el-button @click="openDiff()" type="primary" class="mt-2">
                    {{ $t('app.showDiff') }}
                </el-button>
                <div>
                    <el-checkbox v-model="useNewCompose" :label="$t('app.useCustom')" size="large" />
                </div>
                <div v-if="useNewCompose">
                    <el-text type="danger">{{ $t('app.useCustomHelper') }}</el-text>
                </div>
                <CodemirrorPro v-if="useNewCompose" v-model="newCompose" mode="yaml"></CodemirrorPro>
            </div>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onOperate" :disabled="versions == null || loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
        <Diff ref="composeDiffRef" @confirm="getNewCompose" />
    </DrawerPro>
    <TaskLog ref="taskLogRef" />
</template>
<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { getAppUpdateVersions, ignoreUpgrade, installedOp } from '@/api/modules/app';
import { getAppStoreConfig } from '@/api/modules/setting';
import i18n from '@/lang';
import { ElMessageBox, FormInstance } from 'element-plus';
import { reactive, ref, onBeforeUnmount } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import Diff from './diff/index.vue';
import bus from '@/global/bus';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import { v4 as uuidv4 } from 'uuid';

const composeDiffRef = ref();
const updateRef = ref<FormInstance>();
const open = ref(false);
const loading = ref(false);
const versions = ref<App.VersionDetail[]>();
const operateReq = reactive({
    detailId: 0,
    operate: 'upgrade',
    installId: 0,
    backup: true,
    pullImage: true,
    version: '',
    dockerCompose: '',
    taskID: '',
});
const resourceName = ref('');
const rules = ref<any>({
    detailId: [Rules.requiredSelect],
});
const app = ref();
const oldContent = ref('');
const newContent = ref('');
const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    em('close', open);
};

const newCompose = ref('');
const useNewCompose = ref(false);
const appInstallID = ref(0);
const taskLogRef = ref();
const ignoreAppReq = reactive({
    appID: 0,
    appDetailID: 0,
    scope: 'app',
});

const toLink = (link: string) => {
    window.open(link, '_blank');
};

const openDiff = async () => {
    if (newContent.value === '') {
        await getVersions(operateReq.version);
    }
    composeDiffRef.value.acceptParams(oldContent.value, newContent.value);
};

const getNewCompose = (compose: string) => {
    if (compose !== '') {
        newCompose.value = compose;
        useNewCompose.value = true;
    } else {
        newCompose.value = newContent.value;
        useNewCompose.value = false;
    }
};

const initData = async () => {
    const config = await getAppStoreConfig();
    newCompose.value = '';
    useNewCompose.value = false;
    operateReq.backup = config.data.upgradeBackup == 'Enable';
    operateReq.pullImage = true;
    operateReq.dockerCompose = '';
};

const acceptParams = (id: number, name: string, dockerCompose: string, op: string, appDetail: App.AppDetail) => {
    initData();
    operateReq.installId = id;
    operateReq.operate = op;
    resourceName.value = name;
    app.value = appDetail;
    oldContent.value = dockerCompose;
    appInstallID.value = id;
    getVersions('');
    ignoreAppReq.appID = appDetail.appId;
    if (op === 'ignore') {
        operateReq.detailId = -1;
    }
    open.value = true;
};

const getVersions = async (version: string) => {
    const req = {
        appInstallID: appInstallID.value,
    };
    if (version !== '') {
        req['updateVersion'] = version;
    }
    try {
        const res = await getAppUpdateVersions(req);
        versions.value = res.data || [];
        if (res.data != null && res.data.length > 0) {
            let item = res.data[0];
            if (version != '') {
                item = res.data.find((v) => v.version === version);
            }
            operateReq.detailId = item.detailId;
            operateReq.version = item.version;
            newContent.value = item.dockerCompose;
            newCompose.value = item.dockerCompose;
            useNewCompose.value = false;
        }
    } catch (error) {}
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const operate = async () => {
    loading.value = true;
    if (operateReq.operate === 'upgrade') {
        if (useNewCompose.value) {
            operateReq.dockerCompose = newCompose.value;
        }
        const taskID = uuidv4();
        operateReq.taskID = taskID;
        await installedOp(operateReq)
            .then(() => {
                bus.emit('upgrade', true);
                handleClose();
                openTaskLog(taskID);
            })
            .finally(() => {
                loading.value = false;
            });
    } else {
        await ignoreUpgrade(ignoreAppReq)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                bus.emit('upgrade', true);
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    }
};

const onOperate = async () => {
    ElMessageBox.confirm(
        i18n.global.t('app.operatorHelper', [i18n.global.t('commons.button.' + operateReq.operate)]),
        i18n.global.t('commons.button.upgrade'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(() => {
        operate();
    });
};

onBeforeUnmount(() => {
    bus.off('upgrade');
});

defineExpose({
    acceptParams,
});
</script>
