<template>
    <DrawerPro
        v-model="visible"
        :header="dialogData.title === 'create' ? $t('xpack.alert.addTask') : $t('xpack.alert.editTask')"
        :resource="dialogData.title === 'create' ? dialogData.rowData?.title : ''"
        @close="handleClose"
        size="large"
    >
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('xpack.alert.taskType')" prop="type">
                        <el-select
                            class="selectClass"
                            @change="changeType"
                            v-model="dialogData.rowData!.type"
                            :disabled="dialogData.title === 'edit'"
                        >
                            <template v-for="item in allTaskOptions">
                                <el-option
                                    :key="item.value"
                                    v-if="item.show"
                                    :value="item.value"
                                    :label="$t(item.label)"
                                />
                            </template>
                        </el-select>
                        <span
                            class="input-help"
                            v-if="dialogData.rowData!.type === 'panelPwdEndTime' && expirationDays === 0"
                        >
                            {{ $t('xpack.alert.panelPwdEndTimeRulesHelper') }}
                            <el-link
                                style="font-size: 12px; margin-left: 5px"
                                icon="Position"
                                @click="quickJump('Safe')"
                                type="primary"
                            >
                                {{ $t('firewall.quickJump') }}
                            </el-link>
                        </span>
                    </el-form-item>

                    <el-form-item
                        v-if="dialogData.rowData!.type === 'cronJob'"
                        :label="$t('xpack.alert.cronJobType')"
                        prop="subType"
                    >
                        <el-select
                            class="selectClass"
                            @change="changeType"
                            v-model="dialogData.rowData!.subType"
                            :disabled="dialogData.title === 'edit'"
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
                    </el-form-item>

                    <!--                        网站/证书或磁盘-->
                    <el-form-item
                        v-if="dialogData.rowData!.type === 'ssl'"
                        :label="$t('xpack.alert.certificate')"
                        prop="project"
                    >
                        <el-select class="selectClass" v-model="dialogData.rowData!.project">
                            <el-option
                                :disabled="sslOptions.length === 0"
                                :label="$t('commons.table.all')"
                                value="all"
                            />
                            <el-option
                                v-for="(item, index) in sslOptions"
                                :key="index"
                                :value="item.id + ''"
                                :label="item.primaryDomain"
                            >
                                <span>{{ item.primaryDomain }}</span>
                                <el-tag class="tagClass" v-if="item.autoRenew">
                                    {{ $t('xpack.alert.autoRenew') }}
                                </el-tag>
                            </el-option>
                        </el-select>
                        <span class="input-help">
                            {{ $t('xpack.alert.autoRenewRulesHelper') }}
                        </span>
                    </el-form-item>

                    <el-form-item
                        v-if="dialogData.rowData!.type === 'siteEndTime'"
                        :label="$t('menu.website')"
                        prop="project"
                    >
                        <el-select class="selectClass" v-model="dialogData.rowData!.project">
                            <el-option
                                :disabled="websiteOptions && websiteOptions.length === 0"
                                :label="$t('commons.table.all')"
                                value="all"
                            />
                            <template v-for="(item, index) in websiteOptions" :key="item">
                                <el-option
                                    v-if="!isEver(item.expireDate)"
                                    :key="index"
                                    :value="item.id + ''"
                                    :label="item.primaryDomain"
                                >
                                    <span>{{ item.primaryDomain }}</span>
                                    <el-tag class="tagClass">
                                        {{ item.alias }}
                                    </el-tag>
                                </el-option>
                            </template>
                        </el-select>
                        <span class="input-help">
                            {{ $t('xpack.alert.siteEndTimeRulesHelper') }}
                        </span>
                    </el-form-item>

                    <el-form-item
                        v-if="diskTypes.includes(dialogData.rowData!.type)"
                        :label="$t('xpack.alert.diskInfo')"
                        prop="project"
                    >
                        <el-select class="selectClass" v-model="dialogData.rowData!.project">
                            <el-option
                                :disabled="diskOptions.length === 0"
                                :label="$t('commons.table.all')"
                                value="all"
                            />
                            <el-option
                                v-for="(item, index) in diskOptions"
                                :key="index"
                                :value="item.path"
                                :label="item.path"
                            >
                                <span>{{ item.path }}</span>
                            </el-option>
                        </el-select>
                    </el-form-item>

                    <el-form-item
                        v-if="dialogData.rowData!.type === 'clams'"
                        :label="$t('xpack.alert.taskName')"
                        prop="project"
                    >
                        <el-select class="selectClass" default-first-option v-model="dialogData.rowData!.project">
                            <el-option
                                v-for="(item, index) in clamsOptions"
                                :key="index"
                                :value="String(item.id)"
                                :label="item.name"
                            >
                                <span>{{ item.name }}</span>
                                <el-tag class="tagClass">
                                    {{ item.path }}
                                </el-tag>
                            </el-option>
                        </el-select>
                    </el-form-item>

                    <el-form-item
                        v-if="dialogData.rowData!.type === 'cronJob' && cronjobTypes.includes(dialogData.rowData!.subType)"
                        :label="$t('xpack.alert.taskName')"
                        prop="project"
                    >
                        <el-select class="selectClass" default-first-option v-model="dialogData.rowData!.project">
                            <el-option
                                v-for="(item, index) in cronJobOptions"
                                :key="index"
                                :value="String(item.id)"
                                :label="item.name"
                            >
                                <span>{{ item.name }}</span>
                            </el-option>
                        </el-select>
                    </el-form-item>
                    <span class="input-help" v-if="dialogData.rowData!.type === 'clams' && clamsOptions.length === 0">
                        {{ $t('xpack.alert.clamsRulesHelper') }}
                        <el-link
                            style="font-size: 12px; margin-left: 5px"
                            icon="Position"
                            @click="quickJump('Clam')"
                            type="primary"
                        >
                            {{ $t('firewall.quickJump') }}
                        </el-link>
                    </span>
                    <span
                        class="input-help"
                        v-if="cronjobTypes.includes(dialogData.rowData!.type) && cronJobOptions.length === 0"
                    >
                        {{ $t('xpack.alert.cronJobRulesHelper') }}
                        <el-link
                            style="font-size: 12px; margin-left: 5px"
                            icon="Position"
                            @click="quickJump('Cronjob')"
                            type="primary"
                        >
                            {{ $t('firewall.quickJump') }}
                        </el-link>
                    </span>
                    <!--                        网站或磁盘-->

                    <el-form-item
                        v-if="timeTypes.includes(dialogData.rowData!.type)"
                        :label="$t('xpack.alert.remainingDays')"
                        prop="cycle"
                    >
                        <el-input v-model.number="dialogData.rowData!.cycle" />
                    </el-form-item>

                    <el-form-item
                        v-if="diskTypes.includes(dialogData.rowData!.type)"
                        :label="$t('xpack.alert.monitoringType')"
                        prop="cycle"
                    >
                        <el-radio-group @change="changeCycle" v-model="dialogData.rowData!.cycle">
                            <el-radio-button :label="$t('xpack.alert.useDisk')" :value="1" />
                            <el-radio-button :label="$t('xpack.alert.usePercentage')" :value="2" />
                        </el-radio-group>
                    </el-form-item>

                    <el-form-item
                        v-if="avgTypes.includes(dialogData.rowData!.type)"
                        :label="$t('xpack.alert.specifiedTime')"
                        prop="cycle"
                    >
                        <el-select disabled class="selectClass" v-model.number="dialogData.rowData!.cycle">
                            <el-option :value="1" :label="1 + $t('commons.units.minute')" />
                            <el-option :value="5" :label="5 + $t('commons.units.minute')" />
                            <el-option :value="15" :label="15 + $t('commons.units.minute')" />
                        </el-select>
                    </el-form-item>

                    <el-form-item
                        v-if="diskTypes.includes(dialogData.rowData!.type)"
                        :label="$t('xpack.alert.useExceed')"
                        prop="count"
                    >
                        <el-input v-model.number="dialogData.rowData!.count">
                            <template #append>
                                {{ dialogData.rowData!.cycle === 1 ? 'GB' : ' % ' }}
                            </template>
                        </el-input>
                        <span class="input-help">{{ $t('xpack.alert.useExceedRulesHelper') }}</span>
                    </el-form-item>

                    <el-form-item
                        v-if="avgTypes.includes(dialogData.rowData!.type)"
                        :label="$t('xpack.alert.' + dialogData.rowData!.type + 'UseExceedAvg')"
                        prop="count"
                    >
                        <el-input v-model.number="dialogData.rowData!.count">
                            <template #append>%</template>
                        </el-input>
                        <span class="input-help">
                            {{ $t('xpack.alert.' + dialogData.rowData!.type + 'UseExceedAvgHelper') }}
                        </span>
                    </el-form-item>

                    <el-form-item
                        :label="$t('xpack.alert.triggerCondition')"
                        v-if="ipTypes.includes(dialogData.rowData!.type)"
                        prop="count"
                    >
                        <div
                            class="flex items-center flex-row md:flex-nowrap sm:flex-nowrap flex-wrap justify-between gap-2 w-full"
                        >
                            <el-form-item prop="cycle" class="md:flex-1 sm:flex-1">
                                <el-input v-model.number="dialogData.rowData!.cycle" :max="200">
                                    <template #append>{{ $t('commons.units.minute') }}</template>
                                </el-input>
                            </el-form-item>

                            <span class="whitespace-nowrap input-help w-[4.5rem]">
                                {{ $t('xpack.alert.loginFail') }}
                            </span>
                            <el-form-item prop="count" class="md:flex-1 sm:flex-1">
                                <el-input v-model.number="dialogData.rowData!.count">
                                    <template #append>{{ $t('commons.units.time') }}</template>
                                </el-input>
                            </el-form-item>
                        </div>
                    </el-form-item>

                    <el-form-item
                        :label="$t('setting.ipWhiteList')"
                        prop="advancedParams"
                        v-if="ipTypes.includes(dialogData.rowData!.type)"
                    >
                        <el-input
                            type="textarea"
                            :placeholder="$t('setting.ipWhiteListEgs')"
                            :rows="4"
                            v-model="dialogData.rowData!.advancedParams"
                        />
                        <span class="input-help">{{ $t('xpack.alert.ipWhiteListHelper') }}</span>
                    </el-form-item>

                    <el-form-item :label="$t('xpack.alert.sendCount')" prop="sendCount">
                        <el-input v-model.number="dialogData.rowData!.sendCount" />
                        <span class="input-help">
                            {{
                                timeTypes.includes(dialogData.rowData!.type)
                                    ? $t('xpack.alert.sendCountRulesHelper')
                                    : noParamTypes.includes(dialogData.rowData!.type)
                                    ? $t('xpack.alert.panelUpdateRulesHelper')
                                    : $t('xpack.alert.oneDaySendCountRulesHelper')
                            }}
                        </span>
                    </el-form-item>

                    <el-form-item :label="$t('xpack.alert.alertMethod')" prop="sendMethod">
                        <el-select class="selectClass" v-model="dialogData.rowData!.sendMethod" multiple cleanable>
                            <el-option value="mail" :label="$t('xpack.alert.mail')" />
                            <el-option
                                value="sms"
                                v-if="!globalStore.isIntl"
                                :disabled="!globalStore.isProductPro"
                                :label="$t('xpack.alert.sms')"
                            />
                        </el-select>
                    </el-form-item>
                    <span class="input-help">
                        {{
                            intervalTypes.includes(dialogData.rowData!.type)
                                ? $t('xpack.alert.resourceAlertRulesHelper')
                                : ''
                        }}
                    </span>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button
                    type="primary"
                    @click="onSubmit(formRef)"
                    :disabled="dialogData.rowData?.type === 'panelPwdEndTime' && expirationDays === 0 && loading"
                >
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { ElForm } from 'element-plus';
import { Alert } from '@/api/interface/alert';
import { listSSL, listWebsites } from '@/api/modules/website';
import { CreateAlert, ListDisks, UpdateAlert, ListClams, ListCronJob } from '@/api/modules/alert';
import { MsgSuccess } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { getSettingInfo } from '@/api/modules/setting';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';
import { routerToName } from '@/utils/router';
import { checkCidr, checkCidrV6, checkIpV4V6 } from '@/utils/util';

const globalStore = GlobalStore();
const { isMaster, isProductPro } = storeToRefs(globalStore);

interface DialogProps {
    title: string;
    rowData?: Alert.AlertInfo;
}
const dialogData = ref<DialogProps>({
    title: '',
});
const { t } = i18n.global;
const loading = ref(false);
const visible = ref(false);
const websiteOptions = ref();
const expirationDays = ref(0);
const sslOptions = ref([]);
const diskOptions = ref([]);
const clamsOptions = ref([]);
const cronJobOptions = ref([]);
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const timeTypes = ['ssl', 'siteEndTime', 'panelPwdEndTime'];
const avgTypes = ['cpu', 'memory', 'load'];
const ipTypes = ['sshLogin', 'panelLogin'];
const noParamTypes = ['panelUpdate'];
const intervalTypes = ['cpu', 'memory', 'load', 'disk', 'sshLogin', 'panelLogin', 'nodeException', 'licenseException'];

const diskTypes = ['disk'];
const cronjobTypes = [
    'shell',
    'app',
    'website',
    'database',
    'directory',
    'log',
    'snapshot',
    'curl',
    'cutWebsiteLog',
    'clean',
    'ntp',
];

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    dialogData.value.rowData.sendMethod = [];
    if (dialogData.value.rowData.method != '') {
        dialogData.value.rowData.sendMethod = dialogData.value.rowData.method.split(',');
    }
    if (cronjobTypes.includes(dialogData.value.rowData.type)) {
        dialogData.value.rowData.subType = dialogData.value.rowData.type;
        dialogData.value.rowData.type = 'cronJob';
    }
    initOptions(dialogData.value.rowData.type, dialogData.value.rowData.subType);
    visible.value = true;
};

const rules = reactive({
    type: [Rules.requiredSelect],
    subType: [Rules.requiredSelect],
    project: [Rules.requiredSelect],
    cycle: [Rules.requiredInput, Rules.integerNumber, { validator: checkCycle, trigger: 'blur' }],
    count: [Rules.requiredInput, Rules.integerNumber, { validator: checkCount, trigger: 'blur' }],
    sendCount: [Rules.requiredInput, Rules.integerNumber, { validator: checkSendCount, trigger: 'blur' }],
    sendMethod: [Rules.requiredSelect],
    advancedParams: [{ required: false, validator: checkIPs, trigger: 'blur' }],
});

const allTaskOptions = [
    { value: 'panelPwdEndTime', label: 'xpack.alert.panelPwdEndTime', show: isMaster.value },
    { value: 'panelLogin', label: 'xpack.alert.panelLogin', show: isMaster.value },
    { value: 'sshLogin', label: 'xpack.alert.sshLogin', show: true },
    { value: 'licenseException', label: 'xpack.alert.licenseException', show: isMaster.value && isProductPro.value },
    { value: 'ssl', label: 'xpack.alert.ssl', show: true },
    { value: 'siteEndTime', label: 'xpack.alert.siteEndTime', show: true },
    { value: 'nodeException', label: 'xpack.alert.nodeException', show: isMaster.value && isProductPro.value },
    { value: 'cpu', label: 'xpack.alert.cpu', show: true },
    { value: 'memory', label: 'xpack.alert.memory', show: true },
    { value: 'disk', label: 'xpack.alert.disk', show: true },
    { value: 'load', label: 'xpack.alert.load', show: true },
    { value: 'cronJob', label: 'xpack.alert.cronjob', show: true },
    { value: 'clams', label: 'xpack.alert.clams', show: true },
    { value: 'panelUpdate', label: 'xpack.alert.panelUpdate', show: isMaster.value },
];

function checkRange(value: any, min: number, max: number, callback: any) {
    const num = Number(value);
    if (isNaN(num)) {
        return callback(new Error(i18n.global.t('commons.rule.number')));
    }
    if (num < min || num > max) {
        return callback(new Error(i18n.global.t('commons.rule.numberRange', [min, max])));
    }
    callback();
}

function checkCycle(rule: any, value: any, callback: any) {
    if (value === '') return callback();

    const type = dialogData.value.rowData.type;

    if (type === 'ssl') {
        return checkRange(value, 1, 60, callback);
    } else if (ipTypes.includes(type)) {
        return checkRange(value, 1, 200, callback);
    } else {
        return checkRange(value, 1, 30, callback);
    }
}

function checkCount(rule: any, value: any, callback: any) {
    if (value === '') return callback();

    const type = dialogData.value.rowData.type;
    const cycle = dialogData.value.rowData.cycle;

    if (avgTypes.includes(type) || ipTypes.includes(type)) {
        return checkRange(value, 1, 100, callback);
    }
    if (type === 'disk' && cycle === 2) {
        return checkRange(value, 1, 100, callback);
    }
    callback();
}

function checkSendCount(rule: any, value: any, callback: any) {
    if (value === '') return callback();

    const type = dialogData.value.rowData.type;
    const cycle = dialogData.value.rowData.cycle;

    if (type === 'disk' || avgTypes.includes(type) || ipTypes.includes(type)) {
        return checkRange(value, 1, 50, callback);
    } else if (noParamTypes.includes(type)) {
        return checkRange(value, 1, 30, callback);
    } else {
        if (cycle > 0) {
            return checkRange(value, 1, cycle, callback);
        }
    }
    callback();
}

function checkIPs(rule: any, value: any, callback: any) {
    if (typeof value === 'string' && value.trim() !== '') {
        let addr = value.split('\n');
        for (const item of addr) {
            if (item === '') {
                continue;
            }
            if (item.indexOf('/') !== -1) {
                if (item.indexOf(':') !== -1) {
                    if (checkCidrV6(item)) {
                        return callback(new Error(i18n.global.t('firewall.addressFormatError')));
                    }
                } else if (checkCidr(item)) {
                    return callback(new Error(i18n.global.t('firewall.addressFormatError')));
                }
            } else if (checkIpV4V6(item)) {
                return callback(new Error(i18n.global.t('firewall.addressFormatError')));
            }
        }
    }
    callback();
}

const initOptions = (type: string, subType: string) => {
    if (type === 'ssl') {
        loadSSLs();
    }
    if (type === 'siteEndTime') {
        loadWebsites();
    }
    if (type === 'panelPwdEndTime') {
        loadSettings();
    }
    if (diskTypes.includes(type)) {
        loadDisks();
    }
    if (type === 'clams') {
        loadClams();
    }
    if (type === 'cronJob' && cronjobTypes.includes(subType)) {
        loadCronJob(subType);
    }
};

const handleClose = () => {
    visible.value = false;
    cronJobOptions.value = [];
};

const changeType = () => {
    const typeToCycleMap = {
        ssl: 15,
        siteEndTime: 15,
        panelPwdEndTime: 15,
        panelUpdate: 0,
        disk: 2,
        cpu: 5,
        load: 5,
        memory: 5,
        clams: 0,
        app: 0,
        website: 0,
        database: 0,
        directory: 0,
        log: 0,
        snapshot: 0,
        shell: 0,
        curl: 0,
        cutWebsiteLog: 0,
        clean: 0,
        ntp: 0,
        sshLogin: 30,
        panelLogin: 30,
        nodeException: 0,
        licenseException: 0,
    };

    const typeToCountMap = {
        ssl: 0,
        siteEndTime: 0,
        panelPwdEndTime: 0,
        panelUpdate: 0,
        disk: 80,
        cpu: 80,
        load: 80,
        memory: 80,
        clams: 0,
        app: 0,
        website: 0,
        database: 0,
        directory: 0,
        log: 0,
        snapshot: 0,
        shell: 0,
        curl: 0,
        cutWebsiteLog: 0,
        clean: 0,
        ntp: 0,
        sshLogin: 3,
        panelLogin: 3,
        nodeException: 0,
        licenseException: 0,
    };
    const typeToProjectMap = {
        ssl: 'all',
        siteEndTime: 'all',
        panelPwdEndTime: 'all',
        panelUpdate: 'all',
        disk: 'all',
        cpu: 'all',
        load: 'all',
        memory: 'all',
        clams: '',
        app: '',
        website: '',
        database: '',
        directory: '',
        log: '',
        snapshot: '',
        shell: '',
        curl: '',
        cutWebsiteLog: '',
        clean: '',
        ntp: '',
        sshLogin: 'all',
        panelLogin: 'all',
        nodeException: 'all',
        licenseException: 'all',
    };

    const rowData = dialogData.value.rowData;
    if (rowData) {
        let type = rowData.type;
        let subType = rowData.type;
        if (dialogData.value.rowData.type === 'cronJob') {
            subType = typeof rowData.subType === 'undefined' ? 'shell' : rowData.subType;
            type = subType;
            rowData.subType = subType;
        }
        rowData.project = typeof typeToProjectMap[type] !== 'undefined' ? typeToProjectMap[type] : rowData.project;
        rowData.cycle = typeof typeToCycleMap[type] !== 'undefined' ? typeToCycleMap[type] : rowData.cycle;
        rowData.count = typeof typeToCountMap[type] !== 'undefined' ? typeToCountMap[type] : rowData.count;

        rowData.sendCount = 3;
        formRef.value.validate();
        initOptions(rowData.type, subType);
    }
};

const changeCycle = () => {
    if (diskTypes.includes(dialogData.value.rowData.type)) {
        dialogData.value.rowData.count = dialogData.value.rowData.cycle == 1 ? 30 : 80;
        formRef.value.validate();
    }
};

const loadWebsites = async () => {
    const res = await listWebsites();
    websiteOptions.value = res.data || [];
};

const loadSSLs = async () => {
    const res = await listSSL({});
    sslOptions.value = res.data || [];
};

const loadDisks = async () => {
    const res = await ListDisks();
    diskOptions.value = res.data || [];
};

const loadClams = async () => {
    const res = await ListClams();
    clamsOptions.value = res.data || [];
    dialogData.value.rowData.project = dialogData.value.rowData.project || String(clamsOptions.value[0].id);
};

const loadCronJob = async (jobType: string) => {
    const res = await ListCronJob({
        type: jobType,
        name: '',
        status: '',
    });
    cronJobOptions.value = res.data || [];
    dialogData.value.rowData.project = dialogData.value.rowData.project || String(cronJobOptions.value[0].id);
};

const loadSettings = async () => {
    const res = await getSettingInfo();
    expirationDays.value = Number(res.data.expirationDays);
};

const formatTitle = (row: Alert.AlertInfo) => {
    if (row.type === 'cronJob') {
        row.type = row.subType;
    }
    const titleTemplates = {
        ssl: () => {
            return row.project === 'all'
                ? t('xpack.alert.allSslTitle')
                : t('xpack.alert.sslTitle', [formatSSLName(Number(row.project))]);
        },
        siteEndTime: () => {
            return row.project === 'all'
                ? t('xpack.alert.allSiteEndTimeTitle')
                : t('xpack.alert.siteEndTimeTitle', [formatWebsiteName(Number(row.project))]);
        },
        panelPwdEndTime: () => t('xpack.alert.panelPwdEndTimeTitle'),
        panelUpdate: () => t('xpack.alert.panelUpdateTitle'),
        cpu: () => t('xpack.alert.cpuTitle'),
        memory: () => t('xpack.alert.memoryTitle'),
        load: () => t('xpack.alert.loadTitle'),
        disk: () => {
            return row.project === 'all' ? t('xpack.alert.allDiskTitle') : t('xpack.alert.diskTitle', [row.project]);
        },
        clams: () => t('xpack.alert.clamsTitle', [formatClamName(Number(row.project))]),
        app: () => t('xpack.alert.cronJobAppTitle', [formatCronJobName(Number(row.project))]),
        website: () => t('xpack.alert.cronJobWebsiteTitle', [formatCronJobName(Number(row.project))]),
        database: () => t('xpack.alert.cronJobDatabaseTitle', [formatCronJobName(Number(row.project))]),
        directory: () => t('xpack.alert.cronJobDirectoryTitle', [formatCronJobName(Number(row.project))]),
        log: () => t('xpack.alert.cronJobLogTitle', [formatCronJobName(Number(row.project))]),
        snapshot: () => t('xpack.alert.cronJobSnapshotTitle', [formatCronJobName(Number(row.project))]),
        shell: () => t('xpack.alert.cronJobShellTitle', [formatCronJobName(Number(row.project))]),
        curl: () => t('xpack.alert.cronJobCurlTitle', [formatCronJobName(Number(row.project))]),
        cutWebsiteLog: () => t('xpack.alert.cronJobCutWebsiteLogTitle', [formatCronJobName(Number(row.project))]),
        clean: () => t('xpack.alert.cronJobCleanTitle', [formatCronJobName(Number(row.project))]),
        ntp: () => t('xpack.alert.cronJobNtpTitle', [formatCronJobName(Number(row.project))]),
        nodeException: () => t('xpack.alert.nodeException'),
        licenseException: () => t('xpack.alert.licenseException'),
        panelLogin: () => t('xpack.alert.panelLogin'),
        sshLogin: () => t('xpack.alert.sshLogin'),
    };

    return titleTemplates[row.type] ? titleTemplates[row.type]() : '';
};

const formatSSLName = (id: number) => {
    const sslObject = sslOptions.value.find((item) => item.id === id);
    return sslObject ? sslObject.primaryDomain : undefined;
};

const formatWebsiteName = (id: number) => {
    const websiteOption = websiteOptions.value.find((item: { id: number }) => item.id === id);
    return websiteOption ? websiteOption.primaryDomain : undefined;
};

const formatClamName = (id: number) => {
    const clamObject = clamsOptions.value.find((item) => item.id === id);
    return clamObject ? clamObject.name : undefined;
};

const formatCronJobName = (id: number) => {
    const cronJobOption = cronJobOptions.value.find((item: { id: number }) => item.id === id);
    return cronJobOption ? cronJobOption.name : undefined;
};

const emit = defineEmits<{ (e: 'search'): void }>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (loading.value) return;
    loading.value = true;
    if (!formEl) return;
    await formEl.validate(async (valid) => {
        if (!valid) return;
        if (!dialogData.value.rowData) return;
        dialogData.value.rowData.method = dialogData.value.rowData.sendMethod.join(',');
        dialogData.value.rowData.title = formatTitle(dialogData.value.rowData);
        if (dialogData.value.rowData.type === 'cronJob') {
            dialogData.value.rowData.type = dialogData.value.rowData.subType;
        }
        try {
            if (dialogData.value.title === 'create') {
                await CreateAlert(dialogData.value.rowData);
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
            }
            if (dialogData.value.title === 'edit') {
                await UpdateAlert(dialogData.value.rowData);
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            }
            emit('search');
            visible.value = false;
        } finally {
            loading.value = false;
        }
    });
};

const isEver = (time: string) => {
    const expireDate = new Date(time);
    return expireDate < new Date('1970-01-02');
};

const quickJump = (name: string) => {
    handleClose();
    routerToName(name);
};

defineExpose({
    acceptParams,
});
</script>
<style scoped lang="scss">
.tagClass {
    float: right;
    margin-right: 10px;
    font-size: 12px;
    margin-top: 5px;
}
</style>
