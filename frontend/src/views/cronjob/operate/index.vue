<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
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
                        <div v-if="dialogData.rowData!.type === 'ntp'">
                            <span class="input-help">
                                {{ $t('cronjob.ntp_helper') }}
                                <el-link
                                    style="font-size: 12px"
                                    icon="Position"
                                    @click="goRouter('/toolbox/device')"
                                    type="primary"
                                >
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
                        <div v-for="(specObj, index) of dialogData.rowData.specObjs" :key="index" style="width: 100%">
                            <div class="grid sm:grid-cols-9 gap-4 grid-cols-1">
                                <el-select
                                    v-model="specObj.specType"
                                    class="col-span-2"
                                    @change="changeSpecType(index)"
                                >
                                    <el-option
                                        v-for="item in specOptions"
                                        :key="item.label"
                                        :value="item.value"
                                        :label="item.label"
                                    />
                                </el-select>
                                <el-select
                                    v-if="specObj.specType === 'perWeek'"
                                    class="col-span-2"
                                    v-model="specObj.week"
                                >
                                    <el-option
                                        v-for="item in weekOptions"
                                        :key="item.label"
                                        :value="item.value"
                                        :label="item.label"
                                    />
                                </el-select>
                                <el-input v-if="hasDay(specObj)" class="col-span-2" v-model.number="specObj.day">
                                    <template #append>
                                        <div class="sm:min-w-8 min-w-14 text-center">
                                            <el-tooltip :content="$t('cronjob.day')" placement="top">
                                                {{ $t('cronjob.dayUnit') }}
                                            </el-tooltip>
                                        </div>
                                    </template>
                                </el-input>
                                <el-input v-if="hasHour(specObj)" class="col-span-2" v-model.number="specObj.hour">
                                    <template #append>
                                        <div class="sm:min-w-8 min-w-14 text-center">
                                            <el-tooltip :content="$t('commons.units.hour')" placement="top">
                                                {{ $t('commons.units.hourUnit') }}
                                            </el-tooltip>
                                        </div>
                                    </template>
                                </el-input>
                                <el-input
                                    v-if="specObj.specType !== 'perNSecond'"
                                    class="col-span-2"
                                    v-model.number="specObj.minute"
                                >
                                    <template #append>
                                        <div class="sm:min-w-8 min-w-14 text-center">
                                            <el-tooltip :content="$t('commons.units.minute')" placement="top">
                                                {{ $t('commons.units.minuteUnit') }}
                                            </el-tooltip>
                                        </div>
                                    </template>
                                </el-input>
                                <el-input
                                    v-if="specObj.specType === 'perNSecond'"
                                    class="col-span-2"
                                    v-model.number="specObj.second"
                                >
                                    <template #append>
                                        <div class="sm:min-w-8 min-w-14 text-center">
                                            <el-tooltip :content="$t('commons.units.second')" placement="top">
                                                {{ $t('commons.units.secondUnit') }}
                                            </el-tooltip>
                                        </div>
                                    </template>
                                </el-input>
                                <el-button
                                    link
                                    type="primary"
                                    class="col-span-1"
                                    style="float: right; margin-top: 5px"
                                    @click="handleSpecDelete(index)"
                                    v-if="dialogData.rowData.specObjs.length > 1"
                                >
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                            <el-divider v-if="dialogData.rowData.specObjs.length > 1" class="divider" />
                        </div>
                    </el-form-item>
                    <el-button class="mb-3" @click="handleSpecAdd()">
                        {{ $t('commons.button.add') }}
                    </el-button>

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
                        <el-select class="selectClass" filterable v-model="dialogData.rowData!.containerName">
                            <el-option v-for="item in containerOptions" :key="item" :value="item" :label="item" />
                        </el-select>
                    </el-form-item>
                    <el-form-item
                        v-if="hasScript() && dialogData.rowData!.inContainer"
                        :label="$t('container.command')"
                        prop="command"
                        :rules="Rules.requiredInput"
                    >
                        <el-checkbox style="width: 100px" border v-model="dialogData.rowData!.isCustom">
                            {{ $t('container.custom') }}
                        </el-checkbox>
                        <el-select
                            style="width: calc(100% - 100px)"
                            filterable
                            clearable
                            v-model="dialogData.rowData!.command"
                        >
                            <el-option value="ash" label="/bin/ash" />
                            <el-option value="bash" label="/bin/bash" />
                            <el-option value="sh" label="/bin/sh" />
                        </el-select>
                    </el-form-item>

                    <el-form-item v-if="hasScript()" :label="$t('cronjob.shellContent')" prop="script">
                        <codemirror
                            :autofocus="true"
                            :indent-with-tab="true"
                            :tabSize="4"
                            :lineWrapping="true"
                            style="width: 100%; margin-top: 5px; min-height: 100px"
                            theme="cobalt"
                            :styleActiveLine="true"
                            :matchBrackets="true"
                            :extensions="extensions"
                            v-model="dialogData.rowData!.script"
                        />
                    </el-form-item>

                    <el-form-item
                        v-if="dialogData.rowData!.type === 'website' || dialogData.rowData!.type === 'cutWebsiteLog'"
                        :label="dialogData.rowData!.type === 'website' ? $t('cronjob.website'):$t('website.website')"
                        prop="websiteList"
                    >
                        <el-select class="selectClass" multiple v-model="dialogData.rowData!.websiteList">
                            <el-option
                                :disabled="websiteOptions.length === 0"
                                :label="$t('commons.table.all')"
                                value="all"
                            />
                            <el-option
                                v-for="(item, index) in websiteOptions"
                                :key="index"
                                :value="item.id + ''"
                                :label="item.primaryDomain"
                            >
                                <span>{{ item.primaryDomain }}</span>
                                <el-tag class="tagClass">
                                    {{ item.alias }}
                                </el-tag>
                            </el-option>
                        </el-select>
                        <span class="input-help" v-if="dialogData.rowData!.type === 'cutWebsiteLog'">
                            {{ $t('cronjob.cutWebsiteLogHelper') }}
                        </span>
                    </el-form-item>

                    <div v-if="dialogData.rowData!.type === 'app'">
                        <el-form-item :label="$t('cronjob.app')" prop="appIdList">
                            <el-select class="selectClass" multiple clearable v-model="dialogData.rowData!.appIdList">
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
                                <el-radio value="mysql">MySQL</el-radio>
                                <el-radio value="mariadb">Mariadb</el-radio>
                                <el-radio value="postgresql">PostgreSQL</el-radio>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item :label="$t('cronjob.database')" prop="dbNameList">
                            <el-select class="selectClass" multiple clearable v-model="dialogData.rowData!.dbNameList">
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
                                    <el-tag class="tagClass">
                                        {{ item.database }}
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
                                <FileList
                                    @choose="loadDir"
                                    :dir="true"
                                    :path="dialogData.rowData!.sourceDir"
                                ></FileList>
                            </template>
                        </el-input>
                    </el-form-item>

                    <div v-if="isBackup()">
                        <el-form-item :label="$t('setting.backupAccount')" prop="backupAccountList">
                            <el-select
                                multiple
                                class="selectClass"
                                v-model="dialogData.rowData!.backupAccountList"
                                @change="changeAccount"
                            >
                                <div v-for="item in backupOptions" :key="item.label">
                                    <el-option :value="item.value" :label="item.label" />
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
                            :label="$t('setting.compressPassword')"
                            prop="secret"
                            v-if="isBackup() && dialogData.rowData!.type !== 'database'"
                        >
                            <el-input v-model="dialogData.rowData!.secret" />
                        </el-form-item>
                        <el-form-item :label="$t('cronjob.default_download_path')" prop="defaultDownload">
                            <el-select class="selectClass" v-model="dialogData.rowData!.defaultDownload">
                                <div v-for="item in accountOptions" :key="item.label">
                                    <el-option :value="item.value" :label="item.label" />
                                </div>
                            </el-select>
                        </el-form-item>
                    </div>

                    <div v-if="!globalStore.isIntl">
                        <el-form-item prop="hasAlert">
                            <el-checkbox v-model="dialogData.rowData!.hasAlert" :label="$t('alert.isAlert')" />
                            <span class="input-help">{{ $t('alert.cronJobHelper') }}</span>
                        </el-form-item>
                        <el-form-item
                            prop="alertCount"
                            v-if="dialogData.rowData!.hasAlert && isProductPro"
                            :label="$t('alert.alertCount')"
                        >
                            <el-input-number
                                style="width: 200px"
                                :min="1"
                                step-strictly
                                :step="1"
                                v-model.number="dialogData.rowData!.alertCount"
                            ></el-input-number>
                            <span class="input-help">{{ $t('alert.alertCountHelper') }}</span>
                        </el-form-item>
                        <el-form-item v-if="dialogData.rowData!.hasAlert && !isProductPro">
                            <span>{{ $t('alert.licenseHelper') }}</span>
                            <el-button link type="primary" @click="toUpload">
                                {{ $t('license.levelUpPro') }}
                            </el-button>
                        </el-form-item>
                    </div>

                    <el-form-item :label="$t('cronjob.retainCopies')" prop="retainCopies">
                        <el-input-number
                            style="width: 200px"
                            :min="1"
                            step-strictly
                            :step="1"
                            v-model.number="dialogData.rowData!.retainCopies"
                        ></el-input-number>
                        <span v-if="isBackup()" class="input-help">{{ $t('cronjob.retainCopiesHelper1') }}</span>
                        <span v-else class="input-help">{{ $t('cronjob.retainCopiesHelper') }}</span>
                    </el-form-item>

                    <el-form-item v-if="dialogData.rowData!.type === 'curl'" :label="$t('cronjob.url')" prop="url">
                        <el-input clearable v-model.trim="dialogData.rowData!.url" />
                    </el-form-item>

                    <el-form-item
                        v-if="hasExclusionRules()"
                        :label="$t('cronjob.exclusionRules')"
                        prop="exclusionRules"
                    >
                        <el-input
                            type="textarea"
                            :placeholder="$t('cronjob.rulesHelper')"
                            :rows="3"
                            clearable
                            v-model="dialogData.rowData!.exclusionRules"
                        />
                        <span class="input-help">{{ $t('cronjob.exclusionRulesHelper') }}</span>
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
        <LicenseImport ref="licenseRef" />
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
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
import { loadDefaultSpec, specOptions, transObjToSpec, transSpecToObj, weekOptions } from './../helper';
import { storeToRefs } from 'pinia';
import { GlobalStore } from '@/store';
import LicenseImport from '@/components/license-import/index.vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';

const extensions = [javascript(), oneDark];

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

const globalStore = GlobalStore();
const licenseRef = ref();
const { isProductPro } = storeToRefs(globalStore);

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    if (dialogData.value.rowData?.spec) {
        let objs = [];
        for (const item of dialogData.value.rowData.spec.split(',')) {
            objs.push(transSpecToObj(item));
        }
        dialogData.value.rowData.specObjs = objs;
    }
    if (dialogData.value.title === 'create') {
        changeType();
        dialogData.value.rowData.dbType = 'mysql';
        dialogData.value.rowData.defaultDownload = 'LOCAL';
    }
    if (dialogData.value.rowData.backupAccounts) {
        dialogData.value.rowData.backupAccountList = dialogData.value.rowData.backupAccounts.split(',');
    }
    if (dialogData.value.rowData.appID) {
        dialogData.value.rowData.appIdList = dialogData.value.rowData.appID.split(',');
    }
    if (dialogData.value.rowData.website) {
        dialogData.value.rowData.websiteList = dialogData.value.rowData.website.split(',');
    }
    if (dialogData.value.rowData.dbName) {
        dialogData.value.rowData.dbNameList = dialogData.value.rowData.dbName.split(',');
    }
    dialogData.value.rowData.hasAlert = dialogData.value.rowData!.alertCount > 0;
    dialogData.value.rowData!.alertCount = dialogData.value.rowData!.alertCount || 3;
    dialogData.value.rowData!.command = dialogData.value.rowData!.command || 'sh';
    dialogData.value.rowData!.isCustom =
        dialogData.value.rowData!.command !== 'sh' &&
        dialogData.value.rowData!.command !== 'bash' &&
        dialogData.value.rowData!.command === 'ash';

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

const containerOptions = ref([]);
const websiteOptions = ref([]);
const backupOptions = ref([]);
const accountOptions = ref([]);
const appOptions = ref([]);

const dbInfo = reactive({
    isExist: false,
    name: '',
    version: '',
    dbs: [] as Array<Database.DbItem>,
});

const verifySpec = (rule: any, value: any, callback: any) => {
    if (dialogData.value.rowData!.specObjs.length === 0) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
    }
    for (let i = 0; i < dialogData.value.rowData!.specObjs.length; i++) {
        let item = dialogData.value.rowData!.specObjs[i];
        if (
            !Number.isInteger(item.day) ||
            !Number.isInteger(item.hour) ||
            !Number.isInteger(item.minute) ||
            !Number.isInteger(item.second) ||
            !Number.isInteger(item.week)
        ) {
            callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
            return;
        }
        switch (item.specType) {
            case 'perMonth':
                if (
                    item.day < 0 ||
                    item.day > 31 ||
                    item.hour < 0 ||
                    item.hour > 23 ||
                    item.minute < 0 ||
                    item.minute > 59
                ) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
            case 'perNDay':
                if (
                    item.day < 0 ||
                    item.day > 366 ||
                    item.hour < 0 ||
                    item.hour > 23 ||
                    item.minute < 0 ||
                    item.minute > 59
                ) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
            case 'perWeek':
                if (
                    item.week < 0 ||
                    item.week > 6 ||
                    item.hour < 0 ||
                    item.hour > 23 ||
                    item.minute < 0 ||
                    item.minute > 59
                ) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
            case 'perDay':
                if (item.hour < 0 || item.hour > 23 || item.minute < 0 || item.minute > 59) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
            case 'perNHour':
                if (item.hour < 0 || item.hour > 8784 || item.minute < 0 || item.minute > 59) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
            case 'perHour':
                if (item.minute < 0 || item.minute > 59) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
            case 'perNMinute':
                if (item.minute < 0 || item.minute > 527040) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
            case 'perNSecond':
                if (item.second < 0 || item.second > 31622400) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
        }
    }
    callback();
};

function checkSendCount(rule: any, value: any, callback: any) {
    if (value === '') {
        callback();
    }
    const regex = /^(?:[1-9]|[12][0-9]|30)$/;
    if (!regex.test(value)) {
        return callback(new Error(i18n.global.t('commons.rule.numberRange', [1, 30])));
    }

    callback();
}

const rules = reactive({
    name: [Rules.requiredInput, Rules.noSpace],
    type: [Rules.requiredSelect],
    spec: [
        { validator: verifySpec, trigger: 'blur', required: true },
        { validator: verifySpec, trigger: 'change', required: true },
    ],

    script: [Rules.requiredInput],
    websiteList: [Rules.requiredSelect],
    appIdList: [Rules.requiredSelect],
    dbNameList: [Rules.requiredSelect],
    backupAccountList: [Rules.requiredSelect],
    url: [Rules.requiredInput],
    sourceDir: [Rules.requiredInput],
    backupAccounts: [Rules.requiredSelect],
    defaultDownload: [Rules.requiredSelect],
    retainCopies: [Rules.number],
    alertCount: [Rules.integerNumber, { validator: checkSendCount, trigger: 'blur' }],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const loadDir = async (path: string) => {
    dialogData.value.rowData!.sourceDir = path;
};

const hasDay = (item: any) => {
    return item.specType === 'perMonth' || item.specType === 'perNDay';
};
const hasHour = (item: any) => {
    return item.specType !== 'perHour' && item.specType !== 'perNMinute' && item.specType !== 'perNSecond';
};

const loadDatabases = async (dbType: string) => {
    const data = await listDbItems(dbType);
    dbInfo.dbs = data.data || [];
};

const changeType = () => {
    dialogData.value.rowData!.specObjs = [loadDefaultSpec(dialogData.value.rowData.type)];
};

const changeSpecType = (index: number) => {
    let item = dialogData.value.rowData!.specObjs[index];
    switch (item.specType) {
        case 'perMonth':
        case 'perNDay':
            item.day = 3;
            item.hour = 1;
            item.minute = 30;
            break;
        case 'perWeek':
            item.week = 1;
            item.hour = 1;
            item.minute = 30;
            break;
        case 'perDay':
        case 'perNHour':
            item.hour = 2;
            item.minute = 30;
            break;
        case 'perHour':
        case 'perNMinute':
            item.minute = 30;
            break;
        case 'perNSecond':
            item.second = 30;
            break;
    }
};

const handleSpecAdd = () => {
    let item = {
        specType: 'perWeek',
        week: 1,
        day: 0,
        hour: 1,
        minute: 30,
        second: 0,
    };
    dialogData.value.rowData!.specObjs.push(item);
};

const handleSpecDelete = (index: number) => {
    dialogData.value.rowData!.specObjs.splice(index, 1);
};

const loadBackups = async () => {
    const res = await getBackupList();
    backupOptions.value = [];
    if (!dialogData.value.rowData!.backupAccountList) {
        dialogData.value.rowData!.backupAccountList = ['LOCAL'];
    }
    for (const item of res.data) {
        if (item.id === 0) {
            continue;
        }
        backupOptions.value.push({ label: i18n.global.t('setting.' + item.type), value: item.type });
    }
    changeAccount();
};

const changeAccount = async () => {
    accountOptions.value = [];
    let isInAccounts = false;
    for (const item of backupOptions.value) {
        let exist = false;
        for (const ac of dialogData.value.rowData.backupAccountList) {
            if (item.value == ac) {
                exist = true;
                break;
            }
        }
        if (exist) {
            if (item.value === dialogData.value.rowData.defaultDownload) {
                isInAccounts = true;
            }
            accountOptions.value.push(item);
        }
    }
    if (!isInAccounts) {
        dialogData.value.rowData.defaultDownload = '';
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

function hasExclusionRules() {
    return (
        dialogData.value.rowData!.type === 'app' ||
        dialogData.value.rowData!.type === 'website' ||
        dialogData.value.rowData!.type === 'directory'
    );
}

function hasScript() {
    return dialogData.value.rowData!.type === 'shell';
}

const onSubmit = async (formEl: FormInstance | undefined) => {
    const specs = [];
    for (const item of dialogData.value.rowData.specObjs) {
        const itemSpec = transObjToSpec(item.specType, item.week, item.day, item.hour, item.minute, item.second);
        if (itemSpec === '') {
            MsgError(i18n.global.t('cronjob.cronSpecHelper'));
            return;
        }
        specs.push(itemSpec);
    }
    dialogData.value.rowData.backupAccounts = dialogData.value.rowData.backupAccountList.join(',');

    if (dialogData.value.rowData.appIdList) {
        dialogData.value.rowData.appID = dialogData.value.rowData.appIdList.join(',');
    }
    if (dialogData.value.rowData.websiteList) {
        dialogData.value.rowData.website = dialogData.value.rowData.websiteList.join(',');
    }
    if (dialogData.value.rowData.dbNameList) {
        dialogData.value.rowData.dbName = dialogData.value.rowData.dbNameList.join(',');
    }

    dialogData.value.rowData.spec = specs.join(',');
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!dialogData.value.rowData.inContainer) {
            dialogData.value.rowData.containerName = '';
        }
        if (dialogData.value?.rowData?.exclusionRules) {
            dialogData.value.rowData.exclusionRules = dialogData.value.rowData.exclusionRules.replaceAll('\n', ',');
        }
        dialogData.value.rowData.alertCount =
            dialogData.value.rowData!.hasAlert && isProductPro.value ? dialogData.value.rowData.alertCount : 0;
        dialogData.value.rowData.alertTitle =
            dialogData.value.rowData!.hasAlert && isProductPro.value
                ? i18n.global.t('cronjob.alertTitle', [
                      i18n.global.t('cronjob.' + dialogData.value.rowData.type),
                      dialogData.value.rowData.name,
                  ])
                : '';
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

const toUpload = () => {
    licenseRef.value.acceptParams();
};

defineExpose({
    acceptParams,
});
</script>
<style scoped lang="scss">
:deep(.el-input-group__append) {
    padding: 0 10px;
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

.divider {
    display: block;
    height: 1px;
    width: 100%;
    margin: 3px 0;
    border-top: 1px var(--el-border-color) var(--el-border-style);
}
</style>
