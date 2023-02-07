<template>
    <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader :header="$t('cronjob.cronTask')" :back="handleClose" />
        </template>
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules" label-width="120px">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('cronjob.taskType')" prop="type">
                        <el-select style="width: 100%" v-model="dialogData.rowData!.type">
                            <el-option value="shell" :label="$t('cronjob.shell')" />
                            <el-option value="website" :label="$t('cronjob.website')" />
                            <el-option value="database" :label="$t('cronjob.database')" />
                            <el-option value="directory" :label="$t('cronjob.directory')" />
                            <el-option value="curl" :label="$t('cronjob.curl')" />
                        </el-select>
                    </el-form-item>

                    <el-form-item :label="$t('cronjob.taskName')" prop="name">
                        <el-input style="width: 100%" clearable v-model.trim="dialogData.rowData!.name" />
                    </el-form-item>

                    <el-form-item :label="$t('cronjob.cronSpec')" prop="spec">
                        <el-select style="width: 20%" v-model="dialogData.rowData!.specType">
                            <el-option
                                v-for="item in specOptions"
                                :key="item.label"
                                :value="item.value"
                                :label="item.label"
                            />
                        </el-select>
                        <el-select
                            v-if="dialogData.rowData!.specType === 'perWeek'"
                            style="width: 12%; margin-left: 20px"
                            v-model="dialogData.rowData!.week"
                        >
                            <el-option
                                v-for="item in weekOptions"
                                :key="item.label"
                                :value="item.value"
                                :label="item.label"
                            />
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
                        <el-input
                            style="width: 100%"
                            clearable
                            type="textarea"
                            :autosize="{ minRows: 3, maxRows: 6 }"
                            v-model="dialogData.rowData!.script"
                        />
                    </el-form-item>

                    <el-form-item
                        v-if="dialogData.rowData!.type === 'website'"
                        :label="$t('cronjob.website')"
                        prop="website"
                    >
                        <el-select style="width: 100%" v-model="dialogData.rowData!.website">
                            <el-option v-for="item in websiteOptions" :key="item" :value="item" :label="item" />
                        </el-select>
                    </el-form-item>

                    <div v-if="dialogData.rowData!.type === 'database'">
                        <el-form-item :label="$t('cronjob.database')" prop="dbName">
                            <el-select style="width: 100%" clearable v-model="dialogData.rowData!.dbName">
                                <el-option v-for="item in mysqlInfo.dbNames" :key="item" :label="item" :value="item" />
                            </el-select>
                        </el-form-item>
                    </div>

                    <el-form-item
                        v-if="dialogData.rowData!.type === 'directory'"
                        :label="$t('cronjob.sourceDir')"
                        prop="sourceDir"
                    >
                        <el-input style="width: 100%" disabled v-model="dialogData.rowData!.sourceDir">
                            <template #append>
                                <FileList @choose="loadDir" :dir="true"></FileList>
                            </template>
                        </el-input>
                    </el-form-item>

                    <div v-if="isBackup()">
                        <el-form-item :label="$t('cronjob.target')" prop="targetDirID">
                            <el-select style="width: 100%" v-model="dialogData.rowData!.targetDirID">
                                <el-option
                                    v-for="item in backupOptions"
                                    :key="item.label"
                                    :value="item.value"
                                    :label="item.label"
                                />
                            </el-select>
                        </el-form-item>
                        <el-form-item v-if="dialogData.rowData!.targetDirID !== localDirID">
                            <el-checkbox v-model="dialogData.rowData!.keepLocal">
                                {{ $t('cronjob.saveLocal') }}
                            </el-checkbox>
                        </el-form-item>
                        <el-form-item :label="$t('cronjob.retainCopies')" prop="retainCopies">
                            <el-input-number
                                :min="1"
                                :max="30"
                                v-model.number="dialogData.rowData!.retainCopies"
                            ></el-input-number>
                        </el-form-item>
                    </div>

                    <el-form-item v-if="dialogData.rowData!.type === 'curl'" :label="$t('cronjob.url')" prop="url">
                        <el-input style="width: 100%" clearable v-model.trim="dialogData.rowData!.url" />
                    </el-form-item>

                    <el-form-item
                        v-if="dialogData.rowData!.type === 'website' || dialogData.rowData!.type === 'directory'"
                        :label="$t('cronjob.exclusionRules')"
                        prop="exclusionRules"
                    >
                        <el-input
                            style="width: 100%"
                            type="textarea"
                            :placeholder="$t('cronjob.rulesHelper')"
                            :autosize="{ minRows: 3, maxRows: 6 }"
                            clearable
                            v-model="dialogData.rowData!.exclusionRules"
                        />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import { loadBackupName } from '@/views/setting/helper';
import FileList from '@/components/file-list/index.vue';
import { getBackupList } from '@/api/modules/backup';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { Cronjob } from '@/api/interface/cronjob';
import { addCronjob, editCronjob } from '@/api/modules/cronjob';
import { loadDBNames } from '@/api/modules/database';
import { CheckAppInstalled } from '@/api/modules/app';
import { GetWebsiteOptions } from '@/api/modules/website';
import DrawerHeader from '@/components/drawer-header/index.vue';

interface DialogProps {
    title: string;
    rowData?: Cronjob.CronjobInfo;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisiable = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    drawerVisiable.value = true;
    checkMysqlInstalled();
    loadBackups();
    loadWebsites();
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisiable.value = false;
};

const localDirID = ref();

const websiteOptions = ref();
const backupOptions = ref();

const mysqlInfo = reactive({
    isExist: false,
    name: '',
    version: '',
    dbNames: [] as Array<string>,
});

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
        case 'perDay':
            if (
                !(
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

const specOptions = [
    { label: i18n.global.t('cronjob.perMonth'), value: 'perMonth' },
    { label: i18n.global.t('cronjob.perWeek'), value: 'perWeek' },
    { label: i18n.global.t('cronjob.perDay'), value: 'perDay' },
    { label: i18n.global.t('cronjob.perHour'), value: 'perHour' },
    { label: i18n.global.t('cronjob.perNDay'), value: 'perNDay' },
    { label: i18n.global.t('cronjob.perNHour'), value: 'perNHour' },
    { label: i18n.global.t('cronjob.perNMinute'), value: 'perNMinute' },
];
const weekOptions = [
    { label: i18n.global.t('cronjob.monday'), value: 1 },
    { label: i18n.global.t('cronjob.tuesday'), value: 2 },
    { label: i18n.global.t('cronjob.wednesday'), value: 3 },
    { label: i18n.global.t('cronjob.thursday'), value: 4 },
    { label: i18n.global.t('cronjob.friday'), value: 5 },
    { label: i18n.global.t('cronjob.saturday'), value: 6 },
    { label: i18n.global.t('cronjob.sunday'), value: 7 },
];
const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
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
    dbName: [Rules.requiredSelect],
    url: [Rules.requiredInput],
    sourceDir: [Rules.requiredSelect],
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
        if (item.id === 0) {
            continue;
        }
        if (item.type === 'LOCAL') {
            localDirID.value = item.id;
            if (!dialogData.value.rowData!.targetDirID) {
                dialogData.value.rowData!.targetDirID = item.id;
            }
        }
        backupOptions.value.push({ label: loadBackupName(item.type), value: item.id });
    }
};

const loadWebsites = async () => {
    const res = await GetWebsiteOptions();
    websiteOptions.value = res.data;
};

const checkMysqlInstalled = async () => {
    const res = await CheckAppInstalled('mysql');
    mysqlInfo.isExist = res.data.isExist;
    mysqlInfo.name = res.data.name;
    mysqlInfo.version = res.data.version;
    if (mysqlInfo.isExist) {
        const data = await loadDBNames();
        mysqlInfo.dbNames = data.data;
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
    return dialogData.value.rowData!.type === 'shell';
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
        emit('search');
        drawerVisiable.value = false;
    });
};

defineExpose({
    acceptParams,
});
</script>
