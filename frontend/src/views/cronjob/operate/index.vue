<template>
    <el-drawer v-model="drawerVisible" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader
                :header="title"
                :hideResource="dialogData.title === 'create'"
                :resource="dialogData.rowData?.name"
                :back="handleClose"
            />
        </template>
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('cronjob.taskType')" prop="type">
                        <el-select
                            v-if="dialogData.title === 'create'"
                            class="selectClass"
                            @change="changeType"
                            v-model="dialogData.rowData!.type"
                        >
                            <el-option value="shell" :label="$t('cronjob.shell')" />
                            <el-option value="app" :label="$t('cronjob.app')" />
                            <el-option value="website" :label="$t('cronjob.website')" />
                            <el-option value="database" :label="$t('cronjob.database')" />
                            <el-option value="directory" :label="$t('cronjob.directory')" />
                            <el-option value="log" :label="$t('cronjob.log')" />
                            <el-option value="curl" :label="$t('cronjob.curl')" />
                            <el-option value="cutWebsiteLog" :label="$t('cronjob.cutWebsiteLog')" />
                            <el-option value="clean" :label="$t('setting.diskClean')" />
                            <el-option value="snapshot" :label="$t('cronjob.snapshot')" />
                            <el-option value="ntp" :label="$t('cronjob.ntp')" />
                        </el-select>
                        <div v-else style="width: 100%">
                            <el-tag>{{ $t('cronjob.' + dialogData.rowData!.type) }}</el-tag>
                        </div>
                        <div v-if="dialogData.rowData!.type === 'log'" class="logText">
                            <span class="input-help">
                                {{ $t('cronjob.logHelper1') }}
                                <el-link class="link" icon="Position" @click="goRouter('/logs/system')" type="primary">
                                    {{ $t('firewall.quickJump') }}
                                </el-link>
                            </span>
                            <span class="input-help">
                                {{ $t('cronjob.logHelper2') }}
                                <el-link class="link" icon="Position" @click="goRouter('/logs/ssh')" type="primary">
                                    {{ $t('firewall.quickJump') }}
                                </el-link>
                            </span>
                            <span class="input-help">
                                {{ $t('cronjob.logHelper3') }}
                                <el-link class="link" icon="Position" @click="goRouter('/logs/website')" type="primary">
                                    {{ $t('firewall.quickJump') }}
                                </el-link>
                            </span>
                        </div>
                    </el-form-item>

                    <el-form-item :label="$t('cronjob.taskName')" prop="name">
                        <el-input
                            :disabled="dialogData.title === 'edit'"
                            clearable
                            v-model.trim="dialogData.rowData!.name"
                        />
                    </el-form-item>

                    <el-form-item :label="$t('cronjob.cronSpec')" prop="spec">
                        <el-select class="specTypeClass" v-model="dialogData.rowData!.specType">
                            <el-option
                                v-for="item in specOptions"
                                :key="item.label"
                                :value="item.value"
                                :label="item.label"
                            />
                        </el-select>
                        <el-select
                            v-if="dialogData.rowData!.specType === 'perWeek'"
                            class="specClass"
                            v-model="dialogData.rowData!.week"
                        >
                            <el-option
                                v-for="item in weekOptions"
                                :key="item.label"
                                :value="item.value"
                                :label="item.label"
                            />
                        </el-select>
                        <el-input v-if="hasDay()" class="specClass" v-model.number="dialogData.rowData!.day">
                            <template #append>{{ $t('cronjob.day') }}</template>
                        </el-input>
                        <el-input v-if="hasHour()" class="specClass" v-model.number="dialogData.rowData!.hour">
                            <template #append>{{ $t('commons.units.hour') }}</template>
                        </el-input>
                        <el-input
                            v-if="dialogData.rowData!.specType !== 'perNSecond'"
                            class="specClass"
                            v-model.number="dialogData.rowData!.minute"
                        >
                            <template #append>{{ $t('commons.units.minute') }}</template>
                        </el-input>
                        <el-input
                            v-if="dialogData.rowData!.specType === 'perNSecond'"
                            class="specClass"
                            v-model.number="dialogData.rowData!.second"
                        >
                            <template #append>{{ $t('commons.units.second') }}</template>
                        </el-input>
                    </el-form-item>

                    <el-form-item v-if="hasScript()">
                        <el-checkbox v-model="dialogData.rowData!.inContainer">
                            {{ $t('cronjob.containerCheckBox') }}
                        </el-checkbox>
                    </el-form-item>
                    <el-form-item
                        v-if="hasScript() && dialogData.rowData!.inContainer"
                        :label="$t('cronjob.containerName')"
                        prop="containerName"
                    >
                        <el-select class="selectClass" v-model="dialogData.rowData!.containerName">
                            <el-option v-for="item in containerOptions" :key="item" :value="item" :label="item" />
                        </el-select>
                    </el-form-item>

                    <el-form-item v-if="hasScript()" :label="$t('cronjob.shellContent')" prop="script">
                        <el-input
                            clearable
                            type="textarea"
                            :autosize="{ minRows: 3, maxRows: 6 }"
                            v-model="dialogData.rowData!.script"
                        />
                    </el-form-item>

                    <el-form-item
                        v-if="dialogData.rowData!.type === 'website' || dialogData.rowData!.type === 'cutWebsiteLog'"
                        :label="dialogData.rowData!.type === 'website' ? $t('cronjob.website'):$t('website.website')"
                        prop="website"
                    >
                        <el-select class="selectClass" v-model="dialogData.rowData!.website">
                            <el-option
                                :disabled="websiteOptions.length === 0"
                                :label="$t('commons.table.all')"
                                value="all"
                            />
                            <el-option v-for="item in websiteOptions" :key="item" :value="item" :label="item" />
                        </el-select>
                        <span class="input-help" v-if="dialogData.rowData!.type === 'cutWebsiteLog'">
                            {{ $t('cronjob.cutWebsiteLogHelper') }}
                        </span>
                    </el-form-item>

                    <div v-if="dialogData.rowData!.type === 'app'">
                        <el-form-item :label="$t('cronjob.app')" prop="appID">
                            <el-select class="selectClass" clearable v-model="dialogData.rowData!.appID">
                                <el-option
                                    :disabled="appOptions.length === 0"
                                    :label="$t('commons.table.all')"
                                    value="all"
                                />
                                <div v-for="item in appOptions" :key="item.id">
                                    <el-option :value="item.id + ''" :label="item.name">
                                        <span>{{ item.name }}</span>
                                        <el-tag class="tagClass">
                                            {{ item.key }}
                                        </el-tag>
                                    </el-option>
                                </div>
                            </el-select>
                        </el-form-item>
                    </div>

                    <div v-if="dialogData.rowData!.type === 'database'">
                        <el-form-item :label="$t('cronjob.database')">
                            <el-radio-group v-model="dialogData.rowData!.dbType" @change="loadDatabases">
                                <el-radio label="mysql">MySQL</el-radio>
                                <el-radio label="mariadb">Mariadb</el-radio>
                                <el-radio label="postgresql">PostgreSQL</el-radio>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item :label="$t('cronjob.database')" prop="dbName">
                            <el-select class="selectClass" clearable v-model="dialogData.rowData!.dbName">
                                <el-option
                                    :disabled="dbInfo.dbs.length === 0"
                                    :label="$t('commons.table.all')"
                                    value="all"
                                />
                                <el-option
                                    v-for="item in dbInfo.dbs"
                                    :key="item.id"
                                    :value="item.id + ''"
                                    :label="item.name"
                                >
                                    <span>{{ item.name }}</span>
                                    <el-tag class="tagClass">
                                        {{ item.from === 'local' ? $t('database.local') : $t('database.remote') }}
                                    </el-tag>
                                </el-option>
                            </el-select>
                        </el-form-item>
                    </div>

                    <el-form-item
                        v-if="dialogData.rowData!.type === 'directory'"
                        :label="$t('cronjob.sourceDir')"
                        prop="sourceDir"
                    >
                        <el-input v-model="dialogData.rowData!.sourceDir">
                            <template #prepend>
                                <FileList @choose="loadDir" :dir="true"></FileList>
                            </template>
                        </el-input>
                    </el-form-item>

                    <div v-if="isBackup()">
                        <el-form-item :label="$t('cronjob.target')" prop="targetDirID">
                            <el-select class="selectClass" v-model="dialogData.rowData!.targetDirID">
                                <div v-for="item in backupOptions" :key="item.label">
                                    <el-option
                                        v-if="item.label !== $t('setting.LOCAL') || (dialogData.rowData!.type !== 'snapshot' && dialogData.rowData!.type !== 'log')"
                                        :value="item.value"
                                        :label="item.label"
                                    />
                                </div>
                            </el-select>
                            <span class="input-help">
                                {{ $t('cronjob.targetHelper') }}
                                <el-link
                                    style="font-size: 12px"
                                    icon="Position"
                                    @click="goRouter('/settings/backupaccount')"
                                    type="primary"
                                >
                                    {{ $t('firewall.quickJump') }}
                                </el-link>
                            </span>
                        </el-form-item>
                        <el-form-item
                            v-if="dialogData.rowData!.targetDirID !== localDirID && dialogData.rowData!.type !== 'snapshot' && dialogData.rowData!.type !== 'log'"
                        >
                            <el-checkbox v-model="dialogData.rowData!.keepLocal">
                                {{ $t('cronjob.saveLocal') }}
                            </el-checkbox>
                        </el-form-item>
                    </div>

                    <el-form-item :label="$t('cronjob.retainCopies')" prop="retainCopies">
                        <el-input-number
                            :min="1"
                            :max="300"
                            step-strictly
                            :step="1"
                            v-model.number="dialogData.rowData!.retainCopies"
                        ></el-input-number>
                        <span class="input-help">{{ $t('cronjob.retainCopiesHelper') }}</span>
                    </el-form-item>

                    <el-form-item v-if="dialogData.rowData!.type === 'curl'" :label="$t('cronjob.url')" prop="url">
                        <el-input clearable v-model.trim="dialogData.rowData!.url" />
                    </el-form-item>

                    <el-form-item
                        v-if="dialogData.rowData!.type === 'directory'"
                        :label="$t('cronjob.exclusionRules')"
                        prop="exclusionRules"
                    >
                        <el-input
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
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { checkNumberRange, Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import { getBackupList } from '@/api/modules/setting';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Cronjob } from '@/api/interface/cronjob';
import { addCronjob, editCronjob } from '@/api/modules/cronjob';
import { listDbItems } from '@/api/modules/database';
import { GetWebsiteOptions } from '@/api/modules/website';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { useRouter } from 'vue-router';
import { listContainer } from '@/api/modules/container';
import { Database } from '@/api/interface/database';
import { ListAppInstalled } from '@/api/modules/app';
const router = useRouter();

interface DialogProps {
    title: string;
    rowData?: Cronjob.CronjobInfo;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    if (dialogData.value.title === 'create') {
        changeType();
        dialogData.value.rowData.dbType = 'mysql';
    }
    title.value = i18n.global.t('cronjob.' + dialogData.value.title);
    if (dialogData.value?.rowData?.exclusionRules) {
        dialogData.value.rowData.exclusionRules = dialogData.value.rowData.exclusionRules.replaceAll(',', '\n');
    }
    if (dialogData.value?.rowData?.containerName) {
        dialogData.value.rowData.inContainer = true;
    }
    drawerVisible.value = true;
    loadBackups();
    loadAppInstalls();
    loadWebsites();
    loadContainers();
    if (dialogData.value.rowData?.dbType) {
        loadDatabases(dialogData.value.rowData.dbType);
    } else {
        loadDatabases('mysql');
    }
};
const emit = defineEmits<{ (e: 'search'): void }>();

const goRouter = async (path: string) => {
    router.push({ path: path });
};

const handleClose = () => {
    drawerVisible.value = false;
};

const localDirID = ref();

const containerOptions = ref();
const websiteOptions = ref();
const backupOptions = ref();
const appOptions = ref();

const dbInfo = reactive({
    isExist: false,
    name: '',
    version: '',
    dbs: [] as Array<Database.DbItem>,
});

const verifySpec = (rule: any, value: any, callback: any) => {
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
        case 'perNSecond':
            if (!Number.isInteger(dialogData.value.rowData!.second)) {
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
    { label: i18n.global.t('cronjob.perNSecond'), value: 'perNSecond' },
];
const weekOptions = [
    { label: i18n.global.t('cronjob.monday'), value: 1 },
    { label: i18n.global.t('cronjob.tuesday'), value: 2 },
    { label: i18n.global.t('cronjob.wednesday'), value: 3 },
    { label: i18n.global.t('cronjob.thursday'), value: 4 },
    { label: i18n.global.t('cronjob.friday'), value: 5 },
    { label: i18n.global.t('cronjob.saturday'), value: 6 },
    { label: i18n.global.t('cronjob.sunday'), value: 0 },
];
const rules = reactive({
    name: [Rules.requiredInput],
    type: [Rules.requiredSelect],
    specType: [Rules.requiredSelect],
    spec: [
        { validator: verifySpec, trigger: 'blur', required: true },
        { validator: verifySpec, trigger: 'change', required: true },
    ],
    week: [Rules.requiredSelect, Rules.number],
    day: [Rules.number, checkNumberRange(1, 31)],
    hour: [Rules.number, checkNumberRange(1, 23)],
    minute: [Rules.number, checkNumberRange(1, 59)],

    script: [Rules.requiredInput],
    website: [Rules.requiredSelect],
    dbName: [Rules.requiredSelect],
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

const hasDay = () => {
    return dialogData.value.rowData!.specType === 'perMonth' || dialogData.value.rowData!.specType === 'perNDay';
};
const hasHour = () => {
    return (
        dialogData.value.rowData!.specType !== 'perHour' &&
        dialogData.value.rowData!.specType !== 'perNMinute' &&
        dialogData.value.rowData!.specType !== 'perNSecond'
    );
};

const loadDatabases = async (dbType: string) => {
    const data = await listDbItems(dbType);
    dbInfo.dbs = data.data || [];
};

const changeType = () => {
    switch (dialogData.value.rowData!.type) {
        case 'shell':
            dialogData.value.rowData.specType = 'perWeek';
            dialogData.value.rowData.week = 1;
            dialogData.value.rowData.hour = 1;
            dialogData.value.rowData.minute = 30;
            break;
        case 'app':
            dialogData.value.rowData.specType = 'perDay';
            dialogData.value.rowData.hour = 2;
            dialogData.value.rowData.minute = 30;
            break;
        case 'database':
            dialogData.value.rowData.specType = 'perDay';
            dialogData.value.rowData.hour = 2;
            dialogData.value.rowData.minute = 30;
            break;
        case 'clean':
        case 'website':
            dialogData.value.rowData.specType = 'perWeek';
            dialogData.value.rowData.week = 1;
            dialogData.value.rowData.hour = 1;
            dialogData.value.rowData.minute = 30;
            break;
        case 'log':
        case 'snapshot':
            dialogData.value.rowData.specType = 'perWeek';
            dialogData.value.rowData.week = 1;
            dialogData.value.rowData.hour = 1;
            dialogData.value.rowData.minute = 30;
            dialogData.value.rowData.keepLocal = false;
            dialogData.value.rowData.targetDirID = null;
            for (const item of backupOptions.value) {
                if (item.label !== i18n.global.t('setting.LOCAL')) {
                    dialogData.value.rowData.targetDirID = item.value;
                    break;
                }
            }
            break;
        case 'directory':
            dialogData.value.rowData.specType = 'perDay';
            dialogData.value.rowData.hour = 1;
            dialogData.value.rowData.minute = 30;
            break;
        case 'curl':
            dialogData.value.rowData.specType = 'perWeek';
            dialogData.value.rowData.week = 1;
            dialogData.value.rowData.hour = 1;
            dialogData.value.rowData.minute = 30;
            break;
    }
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
        backupOptions.value.push({ label: i18n.global.t('setting.' + item.type), value: item.id });
    }
};

const loadAppInstalls = async () => {
    const res = await ListAppInstalled();
    appOptions.value = res.data || [];
};

const loadWebsites = async () => {
    const res = await GetWebsiteOptions();
    websiteOptions.value = res.data || [];
};

const loadContainers = async () => {
    const res = await listContainer();
    containerOptions.value = res.data || [];
};

function isBackup() {
    return (
        dialogData.value.rowData!.type === 'app' ||
        dialogData.value.rowData!.type === 'website' ||
        dialogData.value.rowData!.type === 'database' ||
        dialogData.value.rowData!.type === 'directory' ||
        dialogData.value.rowData!.type === 'snapshot' ||
        dialogData.value.rowData!.type === 'log'
    );
}

function hasScript() {
    return dialogData.value.rowData!.type === 'shell';
}

function checkScript() {
    let row = dialogData.value.rowData;
    switch (row.specType) {
        case 'perMonth':
            return row.day > 0 && row.day < 32 && row.hour >= 0 && row.hour < 24 && row.minute >= 0 && row.minute < 60;
        case 'perWeek':
            return (
                row.week >= 0 && row.week < 7 && row.hour >= 0 && row.hour < 24 && row.minute >= 0 && row.minute < 60
            );
        case 'perDay':
            return row.hour >= 0 && row.hour < 24 && row.minute >= 0 && row.minute < 60;
        case 'perHour':
            return row.minute >= 0 && row.minute < 60;
        case 'perNDay':
            return row.day > 0 && row.day < 366 && row.hour >= 0 && row.hour < 24 && row.minute >= 0 && row.minute < 60;
        case 'perNHour':
            return row.hour > 0 && row.hour < 8784 && row.minute >= 0 && row.minute < 60;
        case 'perNMinute':
            return row.minute > 0 && row.minute < 527040;
        case 'perNSecond':
            return row.second > 0 && row.second < 31622400;
    }
}

const onSubmit = async (formEl: FormInstance | undefined) => {
    dialogData.value.rowData.week = Number(dialogData.value.rowData.week);
    dialogData.value.rowData.day = Number(dialogData.value.rowData.day);
    dialogData.value.rowData.hour = Number(dialogData.value.rowData.hour);
    dialogData.value.rowData.minute = Number(dialogData.value.rowData.minute);
    dialogData.value.rowData.second = Number(dialogData.value.rowData.second);
    if (!checkScript()) {
        MsgError(i18n.global.t('cronjob.cronSpecHelper'));
        return;
    }
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!dialogData.value.rowData.inContainer) {
            dialogData.value.rowData.containerName = '';
        }
        if (dialogData.value?.rowData?.exclusionRules) {
            dialogData.value.rowData.exclusionRules = dialogData.value.rowData.exclusionRules.replaceAll('\n', ',');
        }
        if (!dialogData.value.rowData) return;
        if (dialogData.value.title === 'create') {
            await addCronjob(dialogData.value.rowData);
        }
        if (dialogData.value.title === 'edit') {
            await editCronjob(dialogData.value.rowData);
        }

        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('search');
        drawerVisible.value = false;
    });
};

defineExpose({
    acceptParams,
});
</script>
<style scoped lang="scss">
.specClass {
    width: 22% !important;
    margin-left: 20px;
}
@media only screen and (max-width: 1000px) {
    .specClass {
        width: 100% !important;
        margin-top: 20px;
        margin-left: 0;
    }
}
.specTypeClass {
    width: 22% !important;
}
@media only screen and (max-width: 1000px) {
    .specTypeClass {
        width: 100% !important;
    }
}
.selectClass {
    width: 100%;
}
.tagClass {
    float: right;
    margin-right: 10px;
    font-size: 12px;
    margin-top: 5px;
}
.logText {
    line-height: 22px;
    font-size: 12px;
    .link {
        font-size: 12px;
        margin-top: -3px;
    }
}
</style>
