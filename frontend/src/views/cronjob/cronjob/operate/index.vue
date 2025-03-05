<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="title"
        :resource="dialogData.title === 'create' ? '' : dialogData.rowData?.name"
        @close="handleClose"
        size="large"
    >
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules">
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
                <div v-else class="w-full">
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
                <el-input :disabled="dialogData.title === 'edit'" clearable v-model.trim="dialogData.rowData!.name" />
            </el-form-item>
            <el-card class="mb-5">
                <el-form-item :label="$t('cronjob.cronSpec')" prop="specCustom">
                    <el-checkbox :label="$t('container.custom')" v-model="dialogData.rowData!.specCustom" />
                </el-form-item>
                <div v-if="!dialogData.rowData!.specCustom">
                    <el-form-item prop="spec">
                        <div v-for="(specObj, index) of dialogData.rowData.specObjs" :key="index" style="width: 100%">
                            <el-select class="specTypeClass" v-model="specObj.specType" @change="changeSpecType(index)">
                                <el-option
                                    v-for="item in specOptions"
                                    :key="item.label"
                                    :value="item.value"
                                    :label="item.label"
                                />
                            </el-select>
                            <el-select v-if="specObj.specType === 'perWeek'" class="specClass" v-model="specObj.week">
                                <el-option
                                    v-for="item in weekOptions"
                                    :key="item.label"
                                    :value="item.value"
                                    :label="item.label"
                                />
                            </el-select>
                            <el-input v-if="hasDay(specObj)" class="specClass" v-model.number="specObj.day">
                                <template #append>
                                    <div class="append">{{ $t('cronjob.day') }}</div>
                                </template>
                            </el-input>
                            <el-input v-if="hasHour(specObj)" class="specClass" v-model.number="specObj.hour">
                                <template #append>
                                    <div class="append">{{ $t('commons.units.hour') }}</div>
                                </template>
                            </el-input>
                            <el-input
                                v-if="specObj.specType !== 'perNSecond'"
                                class="specClass"
                                v-model.number="specObj.minute"
                            >
                                <template #append>
                                    <div class="append">{{ $t('commons.units.minute') }}</div>
                                </template>
                            </el-input>
                            <el-input
                                v-if="specObj.specType === 'perNSecond'"
                                class="specClass"
                                v-model.number="specObj.second"
                            >
                                <template #append>
                                    <div class="append">{{ $t('commons.units.second') }}</div>
                                </template>
                            </el-input>
                            <el-popover
                                placement="top-start"
                                :title="$t('cronjob.nextTime')"
                                width="200"
                                trigger="click"
                            >
                                <div v-for="(time, index_t) of nextTimes" :key="index_t">
                                    <el-tag class="mt-2">{{ time }}</el-tag>
                                </div>
                                <template #reference>
                                    <el-button class="ml-2.5" @click="loadNext(specObj)" link type="primary">
                                        {{ $t('commons.button.preview') }}
                                    </el-button>
                                </template>
                            </el-popover>
                            <el-button
                                class="ml-2.5"
                                link
                                type="primary"
                                @click="handleSpecDelete(index)"
                                v-if="dialogData.rowData.specObjs.length > 1"
                            >
                                {{ $t('commons.button.delete') }}
                            </el-button>
                            <el-divider v-if="dialogData.rowData.specObjs.length > 1" class="divider" />
                        </div>
                    </el-form-item>
                    <el-button class="mb-3" @click="handleSpecAdd()">
                        {{ $t('commons.button.add') }}
                    </el-button>
                </div>

                <div v-if="dialogData.rowData!.specCustom">
                    <el-form-item prop="spec">
                        <div v-for="(spec, index) of dialogData.rowData.specs" :key="index" class="w-full">
                            <el-input class="specCustom" v-model="dialogData.rowData.specs[index]" />
                            <el-popover
                                placement="top-start"
                                :title="$t('cronjob.nextTime')"
                                width="200"
                                trigger="click"
                            >
                                <div v-for="(time, index_t) of nextTimes" :key="index_t">
                                    <el-tag class="mt-2">{{ time }}</el-tag>
                                </div>
                                <template #reference>
                                    <el-button class="ml-2.5" @click="loadNext(spec)" link type="primary">
                                        {{ $t('commons.button.preview') }}
                                    </el-button>
                                </template>
                            </el-popover>
                            <el-button
                                class="ml-2.5"
                                link
                                type="primary"
                                @click="handleSpecCustomDelete(index)"
                                v-if="dialogData.rowData.specs.length > 1"
                            >
                                {{ $t('commons.button.delete') }}
                            </el-button>
                            <el-divider v-if="dialogData.rowData.specs.length > 1" class="divider" />
                        </div>
                    </el-form-item>
                    <el-button class="mb-3" @click="handleSpecCustomAdd()">
                        {{ $t('commons.button.add') }}
                    </el-button>
                </div>
            </el-card>

            <div v-if="hasScript()">
                <el-form-item class="mt-5">
                    <el-checkbox v-model="dialogData.rowData!.inContainer">
                        {{ $t('cronjob.containerCheckBox') }}
                    </el-checkbox>
                </el-form-item>
                <el-card v-if="dialogData.rowData!.inContainer">
                    <el-row :gutter="20">
                        <el-col :span="12">
                            <el-form-item :label="$t('cronjob.containerName')" prop="containerName">
                                <el-select class="selectClass" v-model="dialogData.rowData!.containerName">
                                    <el-option
                                        v-for="item in containerOptions"
                                        :key="item"
                                        :value="item"
                                        :label="item"
                                    />
                                </el-select>
                            </el-form-item>
                        </el-col>
                        <el-col :span="12">
                            <el-form-item :label="$t('container.command')" prop="command" :rules="Rules.requiredInput">
                                <el-checkbox border v-model="dialogData.rowData!.isCustom">
                                    {{ $t('container.custom') }}
                                </el-checkbox>
                                <el-select
                                    v-if="!dialogData.rowData!.isCustom"
                                    style="width: calc(100% - 100px)"
                                    filterable
                                    clearable
                                    v-model="dialogData.rowData!.command"
                                >
                                    <el-option value="ash" label="/bin/ash" />
                                    <el-option value="bash" label="/bin/bash" />
                                    <el-option value="sh" label="/bin/sh" />
                                </el-select>
                                <el-input
                                    clearable
                                    v-else
                                    style="width: calc(100% - 100px)"
                                    v-model="dialogData.rowData!.command"
                                />
                            </el-form-item>
                        </el-col>
                    </el-row>
                </el-card>
                <div v-if="!dialogData.rowData!.inContainer">
                    <el-card>
                        <el-row :gutter="20">
                            <el-col :span="12">
                                <el-form-item :label="$t('commons.table.user')" prop="user">
                                    <el-select filterable v-model="dialogData.rowData!.user">
                                        <div v-for="item in userOptions" :key="item">
                                            <el-option :value="item" :label="item" />
                                        </div>
                                    </el-select>
                                </el-form-item>
                            </el-col>
                            <el-col :span="12">
                                <el-form-item :label="$t('cronjob.executor')" prop="executor">
                                    <el-checkbox border v-model="dialogData.rowData!.isExecutorCustom">
                                        {{ $t('container.custom') }}
                                    </el-checkbox>
                                    <el-select
                                        v-if="!dialogData.rowData!.isExecutorCustom"
                                        style="width: calc(100% - 100px)"
                                        v-model="dialogData.rowData!.executor"
                                    >
                                        <el-option value="bash" label="bash" />
                                        <el-option value="python" label="python" />
                                        <el-option value="sh" label="sh" />
                                    </el-select>
                                    <el-input
                                        clearable
                                        v-else
                                        style="width: calc(100% - 100px)"
                                        v-model="dialogData.rowData!.executor"
                                    />
                                </el-form-item>
                            </el-col>
                        </el-row>
                    </el-card>

                    <el-form-item :label="$t('cronjob.shellContent')" prop="script" class="mt-5">
                        <el-radio-group
                            @change="dialogData.rowData!.script = ''"
                            v-model="dialogData.rowData!.scriptMode"
                        >
                            <el-radio value="input">{{ $t('commons.button.edit') }}</el-radio>
                            <el-radio value="select">{{ $t('container.pathSelect') }}</el-radio>
                        </el-radio-group>
                        <CodemirrorPro
                            v-if="dialogData.rowData!.scriptMode=== 'input'"
                            v-model="dialogData.rowData!.script"
                            placeholder="#Define or paste the content of your shell file here"
                            mode="javascript"
                            :heightDiff="400"
                        />
                        <el-input
                            v-if="dialogData.rowData!.scriptMode=== 'select'"
                            :placeholder="$t('commons.example') + '/tmp/test.sh'"
                            v-model="dialogData.rowData!.script"
                        >
                            <template #prepend>
                                <FileList @choose="loadScriptDir" :dir="false"></FileList>
                            </template>
                        </el-input>
                    </el-form-item>
                </div>
                <el-form-item :label="$t('cronjob.shellContent')" v-else prop="script" class="mt-5">
                    <CodemirrorPro
                        v-if="dialogData.rowData!.scriptMode=== 'input'"
                        v-model="dialogData.rowData!.script"
                        placeholder="#Define or paste the content of your shell file here"
                        mode="javascript"
                        :heightDiff="400"
                    />
                </el-form-item>
            </div>

            <el-form-item
                v-if="dialogData.rowData!.type === 'website' || dialogData.rowData!.type === 'cutWebsiteLog'"
                :label="dialogData.rowData!.type === 'website' ? $t('cronjob.website'):$t('menu.website')"
                prop="website"
            >
                <el-select class="selectClass" v-model="dialogData.rowData!.website">
                    <el-option :disabled="websiteOptions.length === 0" :label="$t('commons.table.all')" value="all" />
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
                <el-form-item :label="$t('cronjob.app')" prop="appID">
                    <el-select class="selectClass" clearable v-model="dialogData.rowData!.appID">
                        <el-option :disabled="appOptions.length === 0" :label="$t('commons.table.all')" value="all" />
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
                <el-form-item :label="$t('cronjob.database')" prop="dbName">
                    <el-select class="selectClass" clearable v-model="dialogData.rowData!.dbName">
                        <el-option :disabled="dbInfo.dbs.length === 0" :label="$t('commons.table.all')" value="all" />
                        <el-option v-for="item in dbInfo.dbs" :key="item.id" :value="item.id + ''" :label="item.name">
                            <span>{{ item.name }}</span>
                            <el-tag class="tagClass">
                                {{ item.from === 'local' ? $t('commons.table.local') : $t('database.remote') }}
                            </el-tag>
                            <el-tag class="tagClass">
                                {{ item.database }}
                            </el-tag>
                        </el-option>
                    </el-select>
                </el-form-item>
            </div>

            <el-form-item v-if="dialogData.rowData!.type === 'directory'" :label="$t('cronjob.backupContent')">
                <el-radio-group v-model="dialogData.rowData!.isDir">
                    <el-radio :value="true">{{ $t('file.dir') }}</el-radio>
                    <el-radio :value="false">{{ $t('menu.files') }}</el-radio>
                </el-radio-group>
            </el-form-item>

            <el-form-item v-if="dialogData.rowData!.type === 'directory' && dialogData.rowData!.isDir" prop="sourceDir">
                <el-input v-model="dialogData.rowData!.sourceDir">
                    <template #prepend>
                        <FileList @choose="loadDir" :dir="true" :path="dialogData.rowData!.sourceDir" />
                    </template>
                </el-input>
            </el-form-item>
            <div v-if="dialogData.rowData!.type === 'directory' && !dialogData.rowData!.isDir" class="mb-5">
                <el-input>
                    <template #prepend>
                        <FileList @choose="loadFile" :dir="false" />
                    </template>
                </el-input>
                <el-form-item prop="files">
                    <div style="width: 100%">
                        <ComplexTable
                            :show-header="false"
                            :data="dialogData.rowData.files"
                            v-if="dialogData.rowData.files"
                        >
                            <el-table-column prop="val" />
                            <el-table-column width="60">
                                <template #default="scope">
                                    <el-button link type="primary" @click="handleFileDelete(scope.$index)">
                                        {{ $t('commons.button.delete') }}
                                    </el-button>
                                </template>
                            </el-table-column>
                        </ComplexTable>
                    </div>
                </el-form-item>
            </div>

            <div v-if="isBackup()">
                <el-form-item :label="$t('setting.backupAccount')" prop="sourceAccountItems">
                    <el-select
                        multiple
                        class="selectClass"
                        v-model="dialogData.rowData!.sourceAccountItems"
                        @change="changeAccount"
                    >
                        <div v-for="item in backupOptions" :key="item.id">
                            <el-option
                                v-if="item.type !== $t('setting.LOCAL')"
                                :value="item.id"
                                :label="item.type + ' - ' + item.name"
                            />
                            <el-option v-else :value="item.id" :label="item.type" />
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
                <el-form-item :label="$t('cronjob.default_download_path')" prop="downloadAccountID">
                    <el-select class="selectClass" v-model="dialogData.rowData!.downloadAccountID">
                        <div v-for="item in accountOptions" :key="item.id">
                            <el-option
                                v-if="item.type !== $t('setting.LOCAL')"
                                :value="item.id"
                                :label="item.type + ' - ' + item.name"
                            />
                            <el-option v-else :value="item.id" :label="item.type" />
                        </div>
                    </el-select>
                </el-form-item>
            </div>

            <div v-if="!globalStore.isIntl">
                <el-form-item prop="hasAlert">
                    <el-checkbox v-model="dialogData.rowData!.hasAlert" :label="$t('xpack.alert.isAlert')" />
                    <span class="input-help">{{ $t('xpack.alert.cronJobHelper') }}</span>
                </el-form-item>
                <el-form-item
                    prop="alertCount"
                    v-if="dialogData.rowData!.hasAlert && isProductPro"
                    :label="$t('xpack.alert.alertCount')"
                >
                    <el-input-number
                        style="width: 200px"
                        :min="1"
                        step-strictly
                        :step="1"
                        v-model.number="dialogData.rowData!.alertCount"
                    ></el-input-number>
                    <span class="input-help">{{ $t('xpack.alert.alertCountHelper') }}</span>
                </el-form-item>
                <el-form-item v-if="dialogData.rowData!.hasAlert && !isProductPro">
                    <span>{{ $t('xpack.alert.licenseHelper') }}</span>
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

            <el-form-item v-if="hasExclusionRules()" :label="$t('cronjob.exclusionRules')" prop="exclusionRules">
                <el-input
                    type="textarea"
                    :placeholder="$t('cronjob.rulesHelper')"
                    :rows="3"
                    clearable
                    v-model="dialogData.rowData!.exclusionRules"
                />
                <span class="input-help">{{ $t('cronjob.exclusionRulesHelper') }}</span>
            </el-form-item>
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
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import { listBackupOptions } from '@/api/modules/backup';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Cronjob } from '@/api/interface/cronjob';
import { addCronjob, editCronjob, loadNextHandle } from '@/api/modules/cronjob';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import { listDbItems } from '@/api/modules/database';
import { getWebsiteOptions } from '@/api/modules/website';
import { MsgError, MsgSuccess } from '@/utils/message';
import { useRouter } from 'vue-router';
import { listContainer } from '@/api/modules/container';
import { Database } from '@/api/interface/database';
import { listAppInstalled } from '@/api/modules/app';
import {
    loadDefaultSpec,
    loadDefaultSpecCustom,
    specOptions,
    transObjToSpec,
    transSpecToObj,
    weekOptions,
} from './../helper';
import { loadUsers } from '@/api/modules/toolbox';
import { storeToRefs } from 'pinia';
import { GlobalStore } from '@/store';
import LicenseImport from '@/components/license-import/index.vue';
const router = useRouter();

const globalStore = GlobalStore();
const licenseRef = ref();
const { isProductPro } = storeToRefs(globalStore);

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
const nextTimes = ref([]);

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    if (!dialogData.value.rowData?.specCustom && dialogData.value.rowData?.spec) {
        let objs = [];
        for (const item of dialogData.value.rowData.spec.split(',')) {
            objs.push(transSpecToObj(item));
        }
        dialogData.value.rowData.specObjs = objs;
    }
    dialogData.value.rowData.specObjs = dialogData.value.rowData.specObjs || [];
    if (dialogData.value.rowData?.specCustom && dialogData.value.rowData?.spec) {
        dialogData.value.rowData.specs = dialogData.value.rowData.spec.split(',');
    }
    if (dialogData.value.rowData.sourceAccountIDs) {
        let list = [];
        dialogData.value.rowData.sourceAccountItems = [];
        list = dialogData.value.rowData.sourceAccountIDs.split(',');
        for (const item of list) {
            if (item) {
                dialogData.value.rowData.sourceAccountItems.push(Number(item));
            }
        }
    }
    dialogData.value.rowData.specs = dialogData.value.rowData.specs || [];
    dialogData.value.rowData.files = [];
    if (!dialogData.value.rowData.isDir) {
        let files = dialogData.value.rowData.sourceDir?.split(',') || [];
        for (const item of files) {
            dialogData.value.rowData.files.push({ val: item });
        }
    }
    if (dialogData.value.title === 'create') {
        changeType();
        dialogData.value.rowData.scriptMode = 'input';
        dialogData.value.rowData.dbType = 'mysql';
        dialogData.value.rowData.isDir = true;
    }
    dialogData.value.rowData.hasAlert = dialogData.value.rowData!.alertCount > 0;
    dialogData.value.rowData!.alertCount = dialogData.value.rowData!.alertCount || 3;
    dialogData.value.rowData!.command = dialogData.value.rowData!.command || 'sh';
    dialogData.value.rowData!.isCustom =
        dialogData.value.rowData!.command !== 'sh' &&
        dialogData.value.rowData!.command !== 'bash' &&
        dialogData.value.rowData!.command !== 'ash';

    dialogData.value.rowData!.executor = dialogData.value.rowData!.executor || 'bash';
    dialogData.value.rowData!.isCustom =
        dialogData.value.rowData!.executor !== 'sh' &&
        dialogData.value.rowData!.executor !== 'bash' &&
        dialogData.value.rowData!.executor !== 'python' &&
        dialogData.value.rowData!.executor !== 'python3';

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
    loadShellUsers();
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
const userOptions = ref([]);

const dbInfo = reactive({
    isExist: false,
    name: '',
    version: '',
    dbs: [] as Array<Database.DbItem>,
});

const verifyScript = (rule: any, value: any, callback: any) => {
    if (!dialogData.value.rowData!.script || dialogData.value.rowData!.script.length === 0) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
        return;
    }
    callback();
};

const verifySpec = (rule: any, value: any, callback: any) => {
    if (dialogData.value.rowData!.specCustom) {
        if (dialogData.value.rowData!.specs.length === 0) {
            callback(new Error(i18n.global.t('commons.rule.requiredInput')));
            return;
        }
        for (let i = 0; i < dialogData.value.rowData!.specs.length; i++) {
            if (dialogData.value.rowData!.specs[i]) {
                continue;
            }
            callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
            return;
        }
        callback();
        return;
    }
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

const verifyFiles = (rule: any, value: any, callback: any) => {
    if (!dialogData.value.rowData!.files || dialogData.value.rowData!.files.length === 0) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
        return;
    }
    callback();
};

const checkSendCount = (rule: any, value: any, callback: any) => {
    if (value === '') {
        callback();
    }
    const regex = /^(?:[1-9]|[12][0-9]|30)$/;
    if (!regex.test(value)) {
        return callback(new Error(i18n.global.t('commons.rule.numberRange', [1, 30])));
    }
    callback();
};

const rules = reactive({
    name: [Rules.requiredInput, Rules.noSpace],
    type: [Rules.requiredSelect],
    spec: [
        { validator: verifySpec, trigger: 'blur', required: true },
        { validator: verifySpec, trigger: 'change', required: true },
    ],

    script: [{ validator: verifyScript, trigger: 'blur', required: true }],
    containerName: [Rules.requiredSelect],
    appID: [Rules.requiredSelect],
    website: [Rules.requiredSelect],
    dbName: [Rules.requiredSelect],
    url: [Rules.requiredInput],
    files: [{ validator: verifyFiles, trigger: 'blur', required: true }],
    sourceDir: [Rules.requiredInput],
    sourceAccountItems: [Rules.requiredSelect],
    downloadAccountID: [Rules.requiredSelect],
    retainCopies: [Rules.number],
    alertCount: [Rules.integerNumber, { validator: checkSendCount, trigger: 'blur' }],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const loadDir = async (path: string) => {
    dialogData.value.rowData!.sourceDir = path;
};

const loadScriptDir = async (path: string) => {
    dialogData.value.rowData!.script = path;
};

const loadFile = async (path: string) => {
    for (const item of dialogData.value.rowData!.files) {
        if (item.val === path) {
            return;
        }
    }
    dialogData.value.rowData!.files.push({ val: path });
};

const hasDay = (item: any) => {
    return item.specType === 'perMonth' || item.specType === 'perNDay';
};
const hasHour = (item: any) => {
    return item.specType !== 'perHour' && item.specType !== 'perNMinute' && item.specType !== 'perNSecond';
};

const loadNext = async (spec: any) => {
    nextTimes.value = [];
    let specItem = '';
    if (!dialogData.value.rowData.specCustom) {
        specItem = transObjToSpec(spec.specType, spec.week, spec.day, spec.hour, spec.minute, spec.second);
    } else {
        specItem = spec;
    }
    if (!specItem) {
        MsgError(i18n.global.t('cronjob.cronSpecRule2'));
        return;
    }
    const data = await loadNextHandle(specItem);
    nextTimes.value = data.data || [];
};

const loadDatabases = async (dbType: string) => {
    const data = await listDbItems(dbType);
    dbInfo.dbs = data.data || [];
};

const changeType = () => {
    dialogData.value.rowData!.specObjs = [loadDefaultSpec(dialogData.value.rowData.type)];
    dialogData.value.rowData!.specs = [loadDefaultSpecCustom(dialogData.value.rowData.type)];
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

const handleSpecCustomAdd = () => {
    dialogData.value.rowData!.specs.push('');
};

const handleSpecDelete = (index: number) => {
    dialogData.value.rowData!.specObjs.splice(index, 1);
};

const handleSpecCustomDelete = (index: number) => {
    dialogData.value.rowData!.specs.splice(index, 1);
};

const handleFileDelete = (index: number) => {
    dialogData.value.rowData!.files.splice(index, 1);
};

const loadBackups = async () => {
    const res = await listBackupOptions();
    let options = res.data || [];
    backupOptions.value = [];
    let local = 0;
    for (const item of options) {
        if (item.id === 0) {
            continue;
        }
        if (item.type == 'LOCAL') {
            local = item.id;
        }
        backupOptions.value.push({ id: item.id, type: i18n.global.t('setting.' + item.type), name: item.name });
    }
    if (!dialogData.value.rowData!.sourceAccountItems) {
        dialogData.value.rowData!.sourceAccountItems = local === 0 ? [local] : [];
    }
    changeAccount();
};

const changeAccount = async () => {
    accountOptions.value = [];
    let isInAccounts = false;
    for (const item of backupOptions.value) {
        let exist = false;
        for (const ac of dialogData.value.rowData.sourceAccountItems) {
            if (item.id == ac) {
                exist = true;
                break;
            }
        }
        if (exist) {
            if (item.id === dialogData.value.rowData.downloadAccountID) {
                isInAccounts = true;
            }
            accountOptions.value.push(item);
        }
    }
    if (!isInAccounts) {
        dialogData.value.rowData.downloadAccountID = undefined;
    }
};

const loadShellUsers = async () => {
    const res = await loadUsers();
    userOptions.value = res.data || [];
};

const loadAppInstalls = async () => {
    const res = await listAppInstalled();
    appOptions.value = res.data || [];
};

const loadWebsites = async () => {
    const res = await getWebsiteOptions({});
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
        (dialogData.value.rowData!.type === 'directory' && dialogData.value.rowData!.isDir)
    );
}

function hasScript() {
    return dialogData.value.rowData!.type === 'shell';
}

const onSubmit = async (formEl: FormInstance | undefined) => {
    let specs = [];
    if (!dialogData.value.rowData.specCustom) {
        for (const item of dialogData.value.rowData.specObjs) {
            const itemSpec = transObjToSpec(item.specType, item.week, item.day, item.hour, item.minute, item.second);
            if (itemSpec === '') {
                MsgError(i18n.global.t('cronjob.cronSpecHelper'));
                return;
            }
            specs.push(itemSpec);
        }
    } else {
        specs = dialogData.value.rowData.specs;
    }
    if (!dialogData.value.rowData.isDir) {
        let files = [];
        for (const item of dialogData.value.rowData.files) {
            files.push(item.val);
        }
        dialogData.value.rowData.sourceDir = files.join(',');
    }
    dialogData.value.rowData.sourceAccountIDs = dialogData.value.rowData.sourceAccountItems.join(',');
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
            dialogData.value.rowData!.hasAlert && isProductPro.value ? dialogData.value.rowData.alertCount : 3;
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
.specClass {
    width: 17% !important;
    margin-left: 20px;
    .append {
        width: 20px;
    }
}
@media only screen and (max-width: 1000px) {
    .specClass {
        width: 100% !important;
        margin-top: 20px;
        margin-left: 0;
        .append {
            width: 43px;
        }
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
.specCustom {
    width: 80%;
}
@media only screen and (max-width: 1000px) {
    .specCustom {
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

.divider {
    display: block;
    height: 1px;
    width: 100%;
    margin: 3px 0;
    border-top: 1px var(--el-border-color) var(--el-border-style);
}
</style>
