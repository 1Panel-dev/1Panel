<template>
    <el-dialog v-model="cronjobVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="50%">
        <template #header>
            <div class="card-header">
                <span>{{ title }}{{ $t('cronjob.cronTask') }}</span>
            </div>
        </template>
        <el-form
            :disabled="dialogData.isView"
            ref="formRef"
            :model="dialogData.rowData"
            label-position="left"
            :rules="rules"
            label-width="120px"
            :hide-required-asterisk="dialogData.isView"
        >
            <el-form-item :label="$t('cronjob.taskType')" prop="type">
                <el-select
                    @change="changeName(true, dialogData.rowData!.type, dialogData.rowData!.website)"
                    style="width: 100%"
                    v-model="dialogData.rowData!.type"
                >
                    <el-option v-for="item in typeOptions" :key="item.label" :value="item.value" :label="item.label" />
                </el-select>
            </el-form-item>

            <el-form-item :label="$t('cronjob.taskName')" prop="name">
                <el-input
                    :disabled="dialogData.rowData!.type === 'website' || dialogData.rowData!.type === 'database'"
                    style="width: 100%"
                    clearable
                    v-model="dialogData.rowData!.name"
                />
            </el-form-item>

            <el-form-item :label="$t('cronjob.cronSpec')" prop="spec">
                <el-select style="width: 15%" v-model="dialogData.rowData!.specType">
                    <el-option v-for="item in specOptions" :key="item.label" :value="item.value" :label="item.label" />
                </el-select>
                <el-select
                    v-if="dialogData.rowData!.specType === 'perWeek'"
                    style="width: 12%; margin-left: 20px"
                    v-model="dialogData.rowData!.week"
                >
                    <el-option v-for="item in weekOptions" :key="item.label" :value="item.value" :label="item.label" />
                </el-select>
                <el-input
                    v-if="dialogData.rowData!.specType === 'perMonth' || dialogData.rowData!.specType === 'perNDay'"
                    style="width: 20%; margin-left: 20px"
                    v-model.number="dialogData.rowData!.day"
                >
                    <template #append>{{ $t('cronjob.day') }}</template>
                </el-input>
                <el-input
                    v-if="dialogData.rowData!.specType !== 'perHour' && dialogData.rowData!.specType !== 'perNMinute'"
                    style="width: 20%; margin-left: 20px"
                    v-model.number="dialogData.rowData!.hour"
                >
                    <template #append>{{ $t('cronjob.hour') }}</template>
                </el-input>
                <el-input style="width: 20%; margin-left: 20px" v-model.number="dialogData.rowData!.minute">
                    <template #append>{{ $t('cronjob.minute') }}</template>
                </el-input>
            </el-form-item>

            <el-form-item v-if="hasScript()" :label="$t('cronjob.shellContent')" prop="script">
                <el-input style="width: 100%" clearable type="textarea" v-model="dialogData.rowData!.script" />
            </el-form-item>

            <el-form-item v-if="dialogData.rowData!.type === 'website'" :label="$t('cronjob.website')" prop="website">
                <el-select
                    @change="changeName(false, dialogData.rowData!.type, dialogData.rowData!.website)"
                    style="width: 100%"
                    v-model="dialogData.rowData!.website"
                >
                    <el-option
                        v-for="item in websiteOptions"
                        :key="item.label"
                        :value="item.value"
                        :label="item.label"
                    />
                </el-select>
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'database'"
                :label="$t('cronjob.database')"
                prop="database"
            >
                <el-input style="width: 100%" clearable v-model="dialogData.rowData!.database" />
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'directory'"
                :label="$t('cronjob.sourceDir')"
                prop="sourceDir"
            >
                <el-input
                    @input="changeName(false, dialogData.rowData!.type, dialogData.rowData!.website)"
                    style="width: 100%"
                    clearable
                    v-model="dialogData.rowData!.sourceDir"
                >
                    <template #append>
                        <FileList @choose="loadDir" :dir="true"></FileList>
                    </template>
                </el-input>
            </el-form-item>

            <el-form-item v-if="isBackup()" :label="$t('cronjob.target')" prop="targetDirID">
                <el-select style="width: 100%" v-model="dialogData.rowData!.targetDirID">
                    <el-option
                        v-for="item in backupOptions"
                        :key="item.label"
                        :value="item.value"
                        :label="item.label"
                    />
                </el-select>
            </el-form-item>
            <el-form-item v-if="isBackup()" :label="$t('cronjob.retainCopies')" prop="retainCopies">
                <el-input-number :min="1" v-model.number="dialogData.rowData!.retainCopies"></el-input-number>
            </el-form-item>

            <el-form-item v-if="dialogData.rowData!.type === 'curl'" :label="$t('cronjob.url') + 'URL'" prop="url">
                <el-input style="width: 100%" clearable v-model="dialogData.rowData!.url" />
            </el-form-item>

            <el-form-item
                v-if="dialogData.rowData!.type === 'website' || dialogData.rowData!.type === 'directory'"
                :label="$t('cronjob.exclusionRules')"
                prop="exclusionRules"
            >
                <el-input style="width: 100%" type="textarea" clearable v-model="dialogData.rowData!.exclusionRules" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="cronjobVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button v-show="!dialogData.isView" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import { loadBackupName } from '@/views/setting/helper';
import { typeOptions, specOptions, weekOptions } from '../options';
import FileList from '@/components/file-list/index.vue';
import { getBackupList } from '@/api/modules/backup';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { Cronjob } from '@/api/interface/cronjob';
import { addCronjob, editCronjob } from '@/api/modules/cronjob';

interface DialogProps {
    title: string;
    isView: boolean;
    rowData?: Cronjob.CronjobInfo;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const cronjobVisiable = ref(false);
const dialogData = ref<DialogProps>({
    isView: false,
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    cronjobVisiable.value = true;
};

const websiteOptions = ref([
    { label: '所有', value: 'all' },
    { label: '网站1', value: 'web1' },
    { label: '网站2', value: 'web2' },
]);
const backupOptions = ref();

const emit = defineEmits<{ (e: 'search'): void }>();

const varifySpec = (rule: any, value: any, callback: any) => {
    switch (dialogData.value.rowData!.specType) {
        case 'perMonth':
        case 'perNDay':
            if (
                !(
                    Number.isInteger(dialogData.value.rowData!.day) &&
                    Number.isInteger(dialogData.value.rowData!.hour) &&
                    Number.isInteger(dialogData.value.rowData!.minute)
                )
            ) {
                callback(new Error(i18n.global.t('cronjob.cronSpecRule')));
            }
            break;
        case 'perWeek':
            if (
                !(
                    Number.isInteger(dialogData.value.rowData!.week) &&
                    Number.isInteger(dialogData.value.rowData!.hour) &&
                    Number.isInteger(dialogData.value.rowData!.minute)
                )
            ) {
                callback(new Error(i18n.global.t('cronjob.cronSpecRule')));
            }
            break;
        case 'perNHour':
            if (
                !(
                    Number.isInteger(dialogData.value.rowData!.hour) &&
                    Number.isInteger(dialogData.value.rowData!.minute)
                )
            ) {
                callback(new Error(i18n.global.t('cronjob.cronSpecRule')));
            }
            break;
        case 'perHour':
        case 'perNMinute':
            if (!Number.isInteger(dialogData.value.rowData!.minute)) {
                callback(new Error(i18n.global.t('cronjob.cronSpecRule')));
            }
            break;
    }
    callback();
};
const rules = reactive({
    name: [Rules.requiredInput],
    type: [Rules.requiredSelect],
    specType: [Rules.requiredSelect],
    spec: [
        { validator: varifySpec, trigger: 'blur', required: true },
        { validator: varifySpec, trigger: 'change', required: true },
    ],
    week: [Rules.requiredSelect, Rules.number],
    day: [Rules.number, { max: 31, min: 1 }],
    hour: [Rules.number, { max: 23, min: 0 }],
    minute: [Rules.number, { max: 60, min: 1 }],

    script: [Rules.requiredInput],
    website: [Rules.requiredSelect],
    database: [Rules.requiredSelect],
    url: [Rules.requiredInput],
    sourceDir: [Rules.requiredInput],
    targetDirID: [Rules.requiredSelect, Rules.number],
    retainCopies: [Rules.number],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const loadDir = async (path: string) => {
    dialogData.value.rowData!.sourceDir = path;
};

const loadBackups = async () => {
    const res = await getBackupList();
    backupOptions.value = [];
    for (const item of res.data) {
        backupOptions.value.push({ label: loadBackupName(item.type), value: item.id });
    }
};
function isBackup() {
    return (
        dialogData.value.rowData!.type === 'website' ||
        dialogData.value.rowData!.type === 'database' ||
        dialogData.value.rowData!.type === 'directory'
    );
}
function hasScript() {
    return dialogData.value.rowData!.type === 'shell' || dialogData.value.rowData!.type === 'sync';
}
function changeName(isChangeType: boolean, type: string, targetName: string) {
    if (isChangeType) {
        targetName = '';
        if (isBackup()) {
            if (backupOptions.value.length === 0) {
                ElMessage.error(i18n.global.t('cronjob.missBackupAccount'));
            }
        }
    }
    switch (type) {
        case 'website':
            targetName = targetName ? targetName : i18n.global.t('cronjob.all');
            dialogData.value.rowData!.name = `${i18n.global.t('cronjob.website')} [ ${targetName} ]`;
            break;
        case 'database':
            targetName = targetName ? targetName : i18n.global.t('cronjob.all');
            dialogData.value.rowData!.name = `${i18n.global.t('cronjob.database')} [ ${targetName} ]`;
            break;
        case 'directory':
            targetName = targetName ? targetName : '/etc/1panel';
            dialogData.value.rowData!.name = `${i18n.global.t('cronjob.directory')} [ ${targetName} ]`;
            break;
        case 'sync':
            dialogData.value.rowData!.name = i18n.global.t('cronjob.syncDateName');
            break;
        case 'release':
            dialogData.value.rowData!.name = i18n.global.t('cronjob.releaseMemory');
            break;
        case 'curl':
            dialogData.value.rowData!.name = i18n.global.t('cronjob.curl');
            dialogData.value.rowData!.url = 'http://';
            break;
        default:
            dialogData.value.rowData!.name = '';
            break;
    }
}
function restForm() {
    if (formRef.value) {
        formRef.value.resetFields();
    }
}
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!dialogData.value.rowData) return;
        if (dialogData.value.title === 'create') {
            await addCronjob(dialogData.value.rowData);
        }
        if (dialogData.value.title === 'edit') {
            await editCronjob(dialogData.value.rowData);
        }

        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        restForm();
        emit('search');
        cronjobVisiable.value = false;
    });
};

onMounted(() => {
    loadBackups();
});
defineExpose({
    acceptParams,
});
</script>
