<template>
    <div>
        <ComplexTable
            :pagination-config="paginationConfig"
            v-model:selects="selects"
            @search="search"
            style="margin-top: 20px"
            :data="data"
        >
            <template #toolbar>
                <el-button type="primary" @click="onCreate()">{{ $t('commons.button.create') }}</el-button>
                <el-button type="danger" plain :disabled="selects.length === 0" @click="onBatchDelete()">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <el-table-column type="expand">
                <template #default="{ row }">
                    <ul>
                        <li>{{ row.name }} {{ $t('cronjob.handle') }}记录 1</li>
                        <li>{{ row.name }} {{ $t('cronjob.handle') }}记录 2</li>
                        <li>{{ row.name }} {{ $t('cronjob.handle') }}记录 3</li>
                        <li>{{ row.name }} {{ $t('cronjob.handle') }}记录 4</li>
                    </ul>
                </template>
            </el-table-column>

            <el-table-column type="selection" fix />
            <el-table-column :label="$t('cronjob.taskName')" prop="name" />
            <el-table-column :label="$t('commons.table.status')" prop="status">
                <template #default="{ row }">
                    <el-switch
                        v-model="row.status"
                        active-text="running"
                        inactive-text="stoped"
                        active-value="running"
                        inactive-value="stoped"
                    ></el-switch>
                </template>
            </el-table-column>
            <el-table-column :label="$t('cronjob.cronSpec')">
                <template #default="{ row }">
                    <span v-if="row.specType.indexOf('N') === -1 || row.specType === 'perWeek'">
                        {{ $t('cronjob.' + row.specType) }}
                    </span>
                    <span v-else>{{ $t('cronjob.per') }}</span>
                    <span v-if="row.specType === 'perMonth'">
                        {{ row.day }}{{ $t('cronjob.day') }} {{ loadZero(row.hour) }} :
                        {{ loadZero(row.minute) }}
                    </span>
                    <span v-if="row.specType === 'perWeek'">
                        {{ loadWeek(row.week) }} {{ loadZero(row.hour) }} : {{ loadZero(row.minute) }}
                    </span>
                    <span v-if="row.specType === 'perNDay'">
                        {{ row.day }}{{ $t('cronjob.day1') }}, {{ loadZero(row.hour) }} : {{ loadZero(row.minute) }}
                    </span>
                    <span v-if="row.specType === 'perNHour'">
                        {{ row.hour }}{{ $t('cronjob.hour') }}, {{ loadZero(row.minute) }}
                    </span>
                    <span v-if="row.specType === 'perHour'">{{ loadZero(row.minute) }}</span>
                    <span v-if="row.specType === 'perNMinute'">{{ row.minute }}{{ $t('cronjob.minute') }}</span>
                    {{ $t('cronjob.handle') }}
                </template>
            </el-table-column>
            <el-table-column :label="$t('cronjob.retainCopies')" prop="retainCopies" />
            <el-table-column :label="$t('cronjob.target')" prop="targetDir" />
            <fu-table-operations type="icon" :buttons="buttons" :label="$t('commons.table.operate')" fix />
        </ComplexTable>

        <el-dialog @close="search" v-model="cronjobVisiable" width="50%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('cronjob.createCronTask') }}</span>
                </div>
            </template>
            <el-form :model="form" ref="formRef" label-position="left" :rules="rules" label-width="120px">
                <el-form-item :label="$t('cronjob.taskType')" prop="type">
                    <el-select
                        @change="changeName(true, form.type, form.website)"
                        style="width: 100%"
                        v-model="form.type"
                    >
                        <el-option
                            v-for="item in typeOptions"
                            :key="item.label"
                            :value="item.value"
                            :label="item.label"
                        />
                    </el-select>
                </el-form-item>

                <el-form-item :label="$t('cronjob.taskName')" prop="name">
                    <el-input
                        :disabled="form.type === 'website' || form.type === 'database'"
                        style="width: 100%"
                        clearable
                        v-model="form.name"
                    />
                </el-form-item>

                <el-form-item :label="$t('cronjob.cronSpec')" prop="spec">
                    <el-select style="width: 15%" v-model="form.specType">
                        <el-option
                            v-for="item in specOptions"
                            :key="item.label"
                            :value="item.value"
                            :label="item.label"
                        />
                    </el-select>
                    <el-select
                        v-if="form.specType === 'perWeek'"
                        style="width: 12%; margin-left: 20px"
                        v-model="form.week"
                    >
                        <el-option
                            v-for="item in weekOptions"
                            :key="item.label"
                            :value="item.value"
                            :label="item.label"
                        />
                    </el-select>
                    <el-input
                        v-if="form.specType === 'perMonth' || form.specType === 'perNDay'"
                        style="width: 20%; margin-left: 20px"
                        v-model.number="form.day"
                    >
                        <template #append>{{ $t('cronjob.day') }}</template>
                    </el-input>
                    <el-input
                        v-if="form.specType !== 'perHour' && form.specType !== 'perNMinute'"
                        style="width: 20%; margin-left: 20px"
                        v-model.number="form.hour"
                    >
                        <template #append>{{ $t('cronjob.hour') }}</template>
                    </el-input>
                    <el-input style="width: 20%; margin-left: 20px" v-model.number="form.minute">
                        <template #append>{{ $t('cronjob.minute') }}</template>
                    </el-input>
                </el-form-item>

                <el-form-item v-if="hasScript()" :label="$t('cronjob.shellContent')" prop="script">
                    <el-input style="width: 100%" clearable type="textarea" v-model="form.script" />
                </el-form-item>

                <el-form-item v-if="form.type === 'website'" :label="$t('cronjob.website')" prop="website">
                    <el-select
                        @change="changeName(false, form.type, form.website)"
                        style="width: 100%"
                        v-model="form.website"
                    >
                        <el-option
                            v-for="item in websiteOptions"
                            :key="item.label"
                            :value="item.value"
                            :label="item.label"
                        />
                    </el-select>
                </el-form-item>
                <el-form-item v-if="form.type === 'database'" :label="$t('cronjob.database')" prop="database">
                    <el-input style="width: 100%" clearable v-model="form.database" />
                </el-form-item>
                <el-form-item v-if="form.type === 'directory'" :label="$t('cronjob.sourceDir')" prop="sourceDir">
                    <el-input
                        @input="changeName(false, form.type, form.website)"
                        style="width: 100%"
                        clearable
                        v-model="form.sourceDir"
                    >
                        <template #append>
                            <FileList @choose="loadDir" :dir="true"></FileList>
                        </template>
                    </el-input>
                </el-form-item>

                <el-form-item v-if="isBackup()" :label="$t('cronjob.target')" prop="targetDirID">
                    <el-select style="width: 100%" v-model="form.targetDirID">
                        <el-option
                            v-for="item in backupOptions"
                            :key="item.label"
                            :value="item.value"
                            :label="item.label"
                        />
                    </el-select>
                </el-form-item>
                <el-form-item v-if="isBackup()" :label="$t('cronjob.retainCopies')" prop="retainCopies">
                    <el-input-number :min="1" v-model.number="form.retainCopies"></el-input-number>
                </el-form-item>

                <el-form-item v-if="form.type === 'curl'" :label="$t('cronjob.url') + 'URL'" prop="url">
                    <el-input style="width: 100%" clearable v-model="form.url" />
                </el-form-item>

                <el-form-item
                    v-if="form.type === 'website' || form.type === 'directory'"
                    :label="$t('cronjob.exclusionRules')"
                    prop="exclusionRules"
                >
                    <el-input style="width: 100%" type="textarea" clearable v-model="form.exclusionRules" />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="cronjobVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="onSubmit(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { getBackupList } from '@/api/modules/backup';
import { getCronjobPage, addCronjob, editCronjob } from '@/api/modules/cronjob';
import { ElForm, ElMessage } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { typeOptions, specOptions, weekOptions, loadWeek } from './options';
import FileList from '@/components/file-list/index.vue';
import i18n from '@/lang';

const cronjobVisiable = ref<boolean>(false);
const operation = ref<string>('create');
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const selects = ref<any>([]);

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 5,
    total: 0,
});

const logSearch = reactive({
    page: 1,
    pageSize: 5,
});

const varifySpec = (rule: any, value: any, callback: any) => {
    switch (form.specType) {
        case 'perMonth':
        case 'perNDay':
            if (!(Number.isInteger(form.day) && Number.isInteger(form.hour) && Number.isInteger(form.minute))) {
                callback(new Error(i18n.global.t('cronjob.cronSpecRule')));
            }
            break;
        case 'perWeek':
            if (!(Number.isInteger(form.week) && Number.isInteger(form.hour) && Number.isInteger(form.minute))) {
                callback(new Error(i18n.global.t('cronjob.cronSpecRule')));
            }
            break;
        case 'perNHour':
            if (!(Number.isInteger(form.hour) && Number.isInteger(form.minute))) {
                callback(new Error(i18n.global.t('cronjob.cronSpecRule')));
            }
            break;
        case 'perHour':
        case 'perNMinute':
            if (!Number.isInteger(form.minute)) {
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

const form = reactive({
    id: 0,
    name: '',
    type: '',
    specType: 'perMonth',
    spec: '',
    week: 1,
    day: 1,
    hour: 2,
    minute: 3,

    script: '',
    website: '',
    exclusionRules: '',
    database: '',
    url: '',
    sourceDir: '',
    targetDirID: 0,
    retainCopies: 3,
    status: '',
});

const websiteOptions = ref([
    { label: '所有', value: 'all' },
    { label: '网站1', value: 'web1' },
    { label: '网站2', value: 'web2' },
]);

const backupOptions = ref();

const search = async () => {
    logSearch.page = paginationConfig.currentPage;
    logSearch.pageSize = paginationConfig.pageSize;
    const res = await getCronjobPage(logSearch);
    data.value = res.data.items;
    for (const item of data.value) {
        if (item.targetDir !== '-') {
            item.targetDir = loadBackupName(item.targetDir);
        }
    }
    paginationConfig.total = res.data.total;
};

const onCreate = async () => {
    operation.value = 'create';
    form.id = 0;
    form.name = '';
    form.type = '';
    form.specType = 'perMonth';
    form.spec = '';
    form.week = 1;
    form.day = 1;
    form.hour = 2;
    form.minute = 3;
    form.script = '';
    form.website = '';
    form.exclusionRules = '';
    form.database = '';
    form.url = '';
    form.sourceDir = '';
    form.targetDirID = backupOptions.value.length === 0 ? 0 : backupOptions.value[0].value;
    form.retainCopies = 3;
    cronjobVisiable.value = true;
};
const onEdit = async () => {
    cronjobVisiable.value = true;
};
const onBatchDelete = async () => {};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (form.id !== 0 && operation.value === 'edit') {
            await editCronjob(form);
        } else if (form.id === 0 && operation.value === 'create') {
            await addCronjob(form);
        } else {
            ElMessage.success(i18n.global.t('commons.msg.notSupportOperation'));
            return;
        }
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        restForm();
        search();
        cronjobVisiable.value = false;
    });
};
function restForm() {
    if (formRef.value) {
        formRef.value.resetFields();
    }
}

const loadBackups = async () => {
    const res = await getBackupList();
    backupOptions.value = [];
    for (const item of res.data) {
        backupOptions.value.push({ label: loadBackupName(item.type), value: item.id });
    }
};
const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        icon: 'Edit',
        click: onEdit,
    },
    {
        label: i18n.global.t('commons.button.delete'),
        icon: 'Delete',
        click: onBatchDelete,
    },
];
const loadDir = async (path: string) => {
    form.sourceDir = path;
};

function isBackup() {
    return form.type === 'website' || form.type === 'database' || form.type === 'directory';
}
function hasScript() {
    return form.type === 'shell' || form.type === 'sync';
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
            form.name = `${i18n.global.t('cronjob.website')} [ ${targetName} ]`;
            break;
        case 'database':
            targetName = targetName ? targetName : i18n.global.t('cronjob.all');
            form.name = `${i18n.global.t('cronjob.database')} [ ${targetName} ]`;
            break;
        case 'directory':
            targetName = targetName ? targetName : '/etc/1panel';
            form.name = `${i18n.global.t('cronjob.directory')} [ ${targetName} ]`;
            break;
        case 'sync':
            form.name = i18n.global.t('cronjob.syncDateName');
            break;
        case 'release':
            form.name = i18n.global.t('cronjob.releaseMemory');
            break;
        case 'curl':
            form.name = i18n.global.t('cronjob.curl');
            form.url = 'http://';
            break;
        default:
            form.name = '';
            break;
    }
}
function loadZero(i: number) {
    return i < 10 ? '0' + i : '' + i;
}
const loadBackupName = (type: string) => {
    switch (type) {
        case 'OSS':
            return i18n.global.t('setting.OSS');
            break;
        case 'S3':
            return i18n.global.t('setting.S3');
            break;
        case 'LOCAL':
            return i18n.global.t('setting.serverDisk');
            break;
        default:
            return type;
    }
};
onMounted(() => {
    search();
    loadBackups();
});
</script>
