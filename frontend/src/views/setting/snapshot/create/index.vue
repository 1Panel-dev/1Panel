<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.snapshot')" @close="handleClose" size="large">
        <fu-steps
            v-loading="loading"
            class="steps"
            :space="50"
            ref="stepsRef"
            direction="vertical"
            :isLoading="stepLoading"
            :finishButtonText="$t('commons.button.create')"
            @change="changeStep"
            :beforeLeave="beforeLeave"
        >
            <fu-step id="baseData" :title="$t('setting.stepBaseData')">
                <el-form
                    v-loading="loading"
                    class="mt-5"
                    label-position="top"
                    ref="formRef"
                    :model="form"
                    :rules="rules"
                >
                    <el-form-item :label="$t('setting.backupAccount')" prop="fromAccounts">
                        <el-select multiple @change="changeAccount(false)" v-model="form.fromAccounts" clearable>
                            <div v-for="item in backupOptions" :key="item.id">
                                <el-option v-if="item.type !== $t('setting.LOCAL')" :value="item.id" :label="item.name">
                                    {{ item.name }}
                                    <el-tag class="tagClass" type="primary">{{ item.type }}</el-tag>
                                </el-option>
                                <el-option v-else :value="item.id" :label="item.type" />
                            </div>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('cronjob.default_download_path')" prop="downloadAccountID">
                        <el-select v-model="form.downloadAccountID" clearable>
                            <div v-for="item in accountOptions" :key="item.id">
                                <el-option v-if="item.type !== $t('setting.LOCAL')" :value="item.id" :label="item.name">
                                    {{ item.name }}
                                    <el-tag class="tagClass" type="primary">{{ item.type }}</el-tag>
                                </el-option>
                                <el-option v-else :value="item.id" :label="item.type" />
                            </div>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('setting.compressPassword')" prop="secret">
                        <el-input v-model="form.secret"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('cronjob.timeout')" prop="timeoutItem">
                        <el-input type="number" class="selectClass" v-model.number="form.timeoutItem">
                            <template #append>
                                <el-select v-model="form.timeoutUnit" style="width: 80px">
                                    <el-option :label="$t('commons.units.second')" value="s" />
                                    <el-option :label="$t('commons.units.minute')" value="m" />
                                    <el-option :label="$t('commons.units.hour')" value="h" />
                                </el-select>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.description')" prop="description">
                        <el-input type="textarea" clearable v-model="form.description" />
                    </el-form-item>
                </el-form>
            </fu-step>
            <fu-step id="appData" :title="$t('setting.stepAppData')">
                <div class="mt-5 mb-5" v-if="!form.appData || form.appData.length === 0">
                    <span class="input-help">{{ $t('setting.noAppData') }}</span>
                </div>
                <div v-else>
                    <el-checkbox
                        class="ml-6"
                        v-model="form.backupAllImage"
                        @change="selectAllImage"
                        :label="$t('setting.selectAllImage')"
                        size="large"
                    />
                    <el-tree
                        style="max-width: 600px"
                        ref="appRef"
                        node-key="id"
                        :data="form.appData"
                        :props="defaultProps"
                        @check-change="onChangeAppData"
                        show-checkbox
                    >
                        <template #default="{ data }">
                            <div class="float-left">
                                <span>{{ loadApp18n(data.label) }}</span>
                            </div>
                            <div class="ml-4 float-left">
                                <span v-if="data.size">{{ computeSize(data.size) }}</span>
                            </div>
                        </template>
                    </el-tree>
                </div>
            </fu-step>
            <fu-step id="panelData" :title="$t('setting.stepPanelData')">
                <el-tree
                    style="max-width: 600px"
                    ref="panelRef"
                    node-key="id"
                    :data="form.panelData"
                    :props="defaultProps"
                    show-checkbox
                >
                    <template #default="{ data }">
                        <div class="float-left">
                            <span>{{ data.label }}</span>
                        </div>
                        <div class="ml-4 float-left">
                            <span v-if="data.size">{{ computeSize(data.size) }}</span>
                        </div>
                    </template>
                </el-tree>
            </fu-step>
            <fu-step id="backupData" :title="$t('setting.stepBackupData')">
                <div class="mt-5 mb-5" v-if="!form.backupData || form.backupData.length === 0">
                    <span class="input-help">{{ $t('setting.noBackupData') }}</span>
                </div>
                <div v-else>
                    <el-tree
                        style="max-width: 600px"
                        ref="backupRef"
                        node-key="id"
                        :data="form.backupData"
                        :props="defaultProps"
                        show-checkbox
                    >
                        <template #default="{ node, data }">
                            <div class="float-left">
                                <span>{{ load18n(node, data.label) }}</span>
                            </div>
                            <div class="ml-4 float-left">
                                <span v-if="data.size">{{ computeSize(data.size) }}</span>
                            </div>
                        </template>
                    </el-tree>
                </div>
            </fu-step>
            <fu-step id="otherData" :title="$t('setting.stepOtherData')">
                <div class="ml-5">
                    <el-checkbox v-model="form.withDockerConf" :label="$t('setting.dockerConf')" size="large" />
                </div>
                <div class="ml-5">
                    <el-checkbox v-model="form.withOperationLog" :label="$t('logs.operation')" size="large" />
                </div>
                <div class="ml-5">
                    <el-checkbox v-model="form.withLoginLog" :label="$t('logs.login')" size="large" />
                </div>
                <div class="ml-5">
                    <el-checkbox v-model="form.withSystemLog" :label="$t('logs.system')" size="large" />
                </div>
                <div class="ml-5">
                    <el-checkbox v-model="form.withTaskLog" :label="$t('logs.task')" size="large" />
                </div>
                <div class="ml-5">
                    <el-checkbox v-model="form.withMonitorData" :label="$t('setting.monitorData')" size="large" />
                </div>
            </fu-step>
            <template #footer></template>
            <fu-step id="ignoreFiles" :title="$t('cronjob.exclusionRules')">
                <InputTag
                    class="w-full"
                    v-model:tags="form.ignoreFiles"
                    :withFile="true"
                    :baseDir="baseDir"
                    :egHelp="$t('cronjob.exclusionRulesHelper')"
                />
            </fu-step>
        </fu-steps>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button @click="prev" v-if="nowIndex !== 0">{{ $t('commons.button.prev') }}</el-button>
            <el-button type="primary" v-if="nowIndex === 5" :disabled="loading" @click="submitAddSnapshot">
                {{ $t('commons.button.create') }}
            </el-button>
            <el-button @click="next" v-else>{{ $t('commons.button.next') }}</el-button>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" width="70%" @close="handleClose" />
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { loadBaseDir, loadSnapshotInfo, snapshotCreate } from '@/api/modules/setting';
import { computeSize, newUUID, transferTimeToSecond } from '@/utils/util';
import i18n from '@/lang';
import TaskLog from '@/components/log/task/index.vue';
import InputTag from '@/components/input-tag/index.vue';
import { listBackupOptions } from '@/api/modules/backup';
import { Rules } from '@/global/form-rules';
import { ElForm } from 'element-plus';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const stepLoading = ref(false);
const stepsRef = ref();
const nowIndex = ref(0);

const appRef = ref();
const panelRef = ref();
const backupRef = ref();
const taskLogRef = ref();

const baseDir = ref();

const backupOptions = ref();
const accountOptions = ref();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const form = reactive({
    id: 0,
    taskID: '',
    downloadAccountID: '',
    fromAccounts: [],
    sourceAccountIDs: '',
    description: '',
    secret: '',

    timeout: 3600,
    timeoutItem: 3600,
    timeoutUnit: 's',

    backupAllImage: false,
    withDockerConf: true,
    withLoginLog: false,
    withOperationLog: false,
    withSystemLog: false,
    withTaskLog: false,
    withMonitorData: false,

    panelData: [],
    backupData: [],
    appData: [],
    ignoreFiles: [],
});
const rules = reactive({
    fromAccounts: [Rules.requiredSelect],
    downloadAccountID: [Rules.requiredSelect],

    timeoutItem: [Rules.number],
});

const defaultProps = {
    children: 'children',
    label: 'label',
    checked: 'isCheck',
    disabled: 'isDisable',
};
const drawerVisible = ref();

const emit = defineEmits(['search']);
const acceptParams = (): void => {
    form.downloadAccountID = '';
    form.fromAccounts = [];
    form.description = '';
    form.secret = '';
    nowIndex.value = 0;
    search();
    loadBackups();
    loadInstallDir();
    drawerVisible.value = true;
};

const handleClose = () => {
    drawerVisible.value = false;
    emit('search');
};

const submitForm = async (formEl: any) => {
    let bool;
    if (!formEl) return;
    await formEl.validate((valid: boolean) => {
        if (valid) {
            bool = true;
        } else {
            bool = false;
        }
    });
    return bool;
};
const beforeLeave = async (stepItem: any) => {
    switch (stepItem.id) {
        case 'baseData':
            if (await submitForm(formRef.value)) {
                stepsRef.value.next();
                return true;
            } else {
                return false;
            }
        case 'appData':
            if (form.appData && form.appData.length !== 0) {
                let appChecks = appRef.value.getCheckedNodes();
                loadCheckForSubmit(appChecks, form.appData);
            }
            return true;
        case 'panelData':
            let panelChecks = panelRef.value.getCheckedNodes();
            loadCheckForSubmit(panelChecks, form.panelData);
            return true;
        case 'backupData':
            if (!form.backupData || form.backupData.length === 0) {
                return true;
            }
            if (form.backupData && form.backupData.length !== 0) {
                let backupChecks = backupRef.value.getCheckedNodes();
                loadCheckForSubmit(backupChecks, form.backupData);
            }
            return true;
    }
    return true;
};

function next() {
    stepsRef.value.next();
}
function prev() {
    stepsRef.value.prev();
}

const loadApp18n = (label: string) => {
    switch (label) {
        case 'appData':
            return i18n.global.t('setting.appDataLabel');
        case 'appImage':
        case 'appBackup':
            return i18n.global.t('setting.' + label);
        default:
            return label;
    }
};

const loadBackups = async () => {
    const res = await listBackupOptions();
    backupOptions.value = [];
    for (const item of res.data) {
        if (item.id !== 0) {
            backupOptions.value.push({ id: item.id, type: i18n.global.t('setting.' + item.type), name: item.name });
        }
    }
    changeAccount(true);
};

const changeStep = (currentStep: any) => {
    nowIndex.value = currentStep.index;
    switch (currentStep.id) {
        case 'appData':
            if (appRef.value) {
                return;
            }
            nextTick(() => {
                setAppDefaultCheck(form.appData);
            });
            return;
        case 'panelData':
            if (panelRef.value) {
                return;
            }
            nextTick(() => {
                setPanelDefaultCheck(form.panelData);
                return;
            });
            return;
        case 'backupData':
            if (backupRef.value) {
                return;
            }
            nextTick(() => {
                setBackupDefaultCheck(form.backupData);
                return;
            });
            return;
    }
};

const changeAccount = async (isInit: boolean) => {
    accountOptions.value = [];
    let isInAccounts = false;
    for (const item of backupOptions.value) {
        let exist = false;
        for (const ac of form.fromAccounts) {
            if (item.id == ac) {
                exist = true;
                break;
            }
        }
        if (exist) {
            if (item.id === form.downloadAccountID) {
                isInAccounts = true;
            }
            accountOptions.value.push(item);
        }
    }
    if (!isInAccounts && !isInit) {
        form.downloadAccountID = form.fromAccounts.length === 0 ? undefined : form.fromAccounts[0];
    }
};

const load18n = (node: any, label: string) => {
    if (node.level === 1) {
        switch (label) {
            case 'log':
            case 'app':
            case 'database':
            case 'website':
            case 'directory':
                return i18n.global.t('setting.' + label + 'Label');
            case 'system_snapshot':
                return i18n.global.t('setting.snapshotLabel');
            case 'master':
                return i18n.global.t('xpack.node.masterBackup');
            default:
                return label;
        }
    }
    if (node.level === 2) {
        switch (label) {
            case 'App':
                return i18n.global.t('setting.appLabel');
            case 'AppStore':
                return i18n.global.t('menu.apps');
            case 'shell':
                return i18n.global.t('setting.shellLabel');
            default:
                return label;
        }
    }
    return label;
};

const submitAddSnapshot = async () => {
    loading.value = true;
    form.taskID = newUUID();
    form.sourceAccountIDs = form.fromAccounts.join(',');
    form.timeout = transferTimeToSecond(form.timeoutItem + form.timeoutUnit);
    await snapshotCreate(form)
        .then(() => {
            loading.value = false;
            drawerVisible.value = false;
            emit('search');
            openTaskLog(form.taskID);
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const loadCheckForSubmit = (checks: any, list: any) => {
    for (const item of list) {
        let isCheck = false;
        for (const check of checks) {
            if (item.id == check.id) {
                isCheck = true;
                break;
            }
        }
        item.isCheck = isCheck;
        if (item.children) {
            loadCheckForSubmit(checks, item.children);
        }
    }
};

const selectAllImage = () => {
    for (const item of form.appData) {
        for (const item2 of item.children) {
            if (item2.label === 'appImage') {
                appRef.value.setChecked(item2.id, form.backupAllImage, false);
            }
        }
    }
};

const loadInstallDir = async () => {
    const pathRes = await loadBaseDir();
    baseDir.value = pathRes.data;
};

const search = async () => {
    loading.value = true;
    await loadSnapshotInfo()
        .then((res) => {
            loading.value = false;
            form.panelData = res.data.panelData || [];
            form.backupData = res.data.backupData || [];
            form.appData = res.data.appData || [];
        })
        .catch(() => {
            loading.value = false;
        });
};

function onChangeAppData(data: any, isCheck: boolean) {
    if (data.label !== 'appData' || !data.relationItemID) {
        return;
    }
    data.isCheck = isCheck;
    let isDisable = false;
    for (const item of form.appData) {
        if (!item.children) {
            return;
        }
        for (const itemData of item.children) {
            if (itemData.label === 'appData' && itemData.relationItemID === data.relationItemID && itemData.isCheck) {
                isDisable = true;
                break;
            }
        }
    }
    for (const item of form.appData) {
        if (!item.children) {
            return;
        }
        for (const relationItem of item.children) {
            if (relationItem.id !== data.relationItemID) {
                continue;
            }
            relationItem.isDisable = isDisable;
            if (isDisable) {
                appRef.value.setChecked(relationItem.id, isDisable, isDisable);
            }
            break;
        }
    }
}
const setAppDefaultCheck = async (list: any) => {
    for (const item of list) {
        if (item.isCheck) {
            appRef.value.setChecked(item.id, true, true);
        }
        if (item.children) {
            setAppDefaultCheck(item.children);
        }
    }
};
const setPanelDefaultCheck = async (list: any) => {
    for (const item of list) {
        if (item.isCheck) {
            panelRef.value.setChecked(item.id, true, true);
            continue;
        }
        if (item.children) {
            setPanelDefaultCheck(item.children);
        }
    }
};
const setBackupDefaultCheck = async (list: any) => {
    if (!form.backupData || form.backupData.length === 0) {
        return;
    }
    for (const item of list) {
        if (item.isCheck) {
            backupRef.value.setChecked(item.id, true, true);
            continue;
        }
        if (item.children) {
            setBackupDefaultCheck(item.children);
        }
    }
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.steps {
    width: 100%;
    margin-top: 20px;
    :deep(.el-step) {
        .el-step__line {
            background-color: var(--el-color-primary-light-5);
        }
        .el-step__head.is-success {
            color: var(--el-color-primary-light-5);
            border-color: var(--el-color-primary-light-5);
        }
        .el-step__icon {
            color: var(--el-color-primary-light-2);
        }
        .el-step__icon.is-text {
            border-radius: 50%;
            border: 2px solid;
            border-color: var(--el-color-primary-light-2);
        }

        .el-step__title.is-finish {
            color: #717379;
            font-size: 13px;
            font-weight: bold;
        }

        .el-step__description.is-finish {
            color: #606266;
        }

        .el-step__title.is-success {
            font-weight: bold;
            color: var(--el-color-primary-light-2);
        }

        .el-step__title.is-process {
            font-weight: bold;
            color: var(--el-text-color-regular);
        }
    }
}
.tagClass {
    float: right;
    margin-right: 10px;
    font-size: 12px;
    margin-top: 5px;
}
</style>
