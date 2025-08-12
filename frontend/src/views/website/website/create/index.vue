<template>
    <DrawerPro v-model="open" :header="$t('website.create')" size="60%" @close="handleClose">
        <template #buttons>
            <span class="drawer-header-button">
                <template v-for="item in WebsiteTypes" :key="item.value">
                    <el-button
                        :class="website.type === item.value ? 'active-button' : ''"
                        @click="changeType(item.value)"
                        :type="website.type === item.value ? '' : 'info'"
                        :plain="website.type === item.value"
                        :text="website.type !== item.value"
                        :bg="website.type !== item.value"
                    >
                        {{ item.label }}
                    </el-button>
                </template>
            </span>
        </template>
        <div v-loading="loading" :class="{ mask: !versionExist }">
            <el-alert
                v-if="website.type == 'deployment'"
                :title="$t('website.websiteDeploymentHelper')"
                type="info"
                :closable="false"
            />
            <el-alert
                v-if="website.type == 'static'"
                :title="$t('website.websiteStatictHelper')"
                type="info"
                :closable="false"
            />
            <el-alert
                v-if="website.type == 'proxy'"
                :title="$t('website.websiteProxyHelper')"
                type="info"
                :closable="false"
            />
            <el-alert
                v-if="website.type == 'runtime'"
                :title="$t('website.runtimeProxyHelper')"
                type="info"
                :closable="false"
            />
            <el-alert
                v-if="website.type == 'subsite'"
                :title="$t('website.subsiteHelper')"
                type="info"
                :closable="false"
            />
            <br />
            <el-form
                ref="websiteForm"
                label-position="top"
                :model="website"
                label-width="125px"
                :rules="rules"
                :validate-on-rule-change="false"
                v-loading="loading"
            >
                <el-form-item :label="$t('commons.table.group')" prop="webSiteGroupId">
                    <el-select v-model="website.webSiteGroupId">
                        <el-option
                            v-for="(group, index) in groups"
                            :key="index"
                            :label="group.name == 'Default' ? $t('commons.table.default') : group.name"
                            :value="group.id"
                        ></el-option>
                    </el-select>
                </el-form-item>
                <div v-if="website.type === 'deployment'">
                    <el-form-item prop="appType">
                        <el-radio-group v-model="website.appType" @change="changeAppType(website.appType)">
                            <el-radio :label="'installed'" :value="'installed'">
                                {{ $t('website.appInstalled') }}
                            </el-radio>
                            <el-radio :label="'new'" :value="'new'">
                                {{ $t('website.appNew') }}
                            </el-radio>
                        </el-radio-group>
                    </el-form-item>
                    <el-form-item
                        v-if="website.appType == 'installed'"
                        :label="$t('website.appInstalled')"
                        prop="appInstallId"
                    >
                        <el-select v-model="website.appInstallId" class="p-w-300">
                            <el-option
                                v-for="(appInstall, index) in appInstalls"
                                :key="index"
                                :label="appInstall.name"
                                :value="appInstall.id"
                                :disabled="appInstall.status !== 'Running'"
                            >
                                <div class="flex justify-between items-center w-full">
                                    <span>{{ appInstall.name }}</span>
                                    <span><Status :key="appInstall.status" :status="appInstall.status"></Status></span>
                                </div>
                            </el-option>
                        </el-select>
                    </el-form-item>
                    <div v-if="website.appType == 'new'">
                        <el-form-item :label="$t('app.app')" prop="appinstall.appId">
                            <el-select
                                v-model="website.appinstall.appId"
                                @change="changeApp()"
                                class="p-w-300"
                                filterable
                            >
                                <el-option
                                    v-for="(app, index) in apps"
                                    :key="index"
                                    :label="app.name"
                                    :value="app.id"
                                ></el-option>
                            </el-select>
                        </el-form-item>
                        <AppInstallForm ref="installFormRef" v-model="website.appinstall" :loading="loading" />
                    </div>
                </div>
                <div v-if="website.type === 'subsite'">
                    <el-form-item :label="$t('website.parentWbeiste')" prop="parentWebsiteID">
                        <el-select v-model="website.parentWebsiteID" @change="getDir(website.parentWebsiteID)">
                            <el-option
                                v-for="(site, index) in parentWebsites"
                                :key="index"
                                :label="site.primaryDomain"
                                :value="site.id"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('website.runDir')" prop="siteDir">
                        <el-select v-model="website.siteDir" filterable class="p-w-200">
                            <el-option
                                v-for="(item, index) in dirs"
                                :label="item"
                                :value="item"
                                :key="index"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                </div>
                <div v-if="website.type === 'runtime'">
                    <el-row :gutter="20">
                        <el-col :span="8">
                            <el-form-item :label="$t('commons.table.type')" prop="runtimeType">
                                <el-select v-model="website.runtimeType" @change="changeRuntimeType()">
                                    <el-option label="PHP" value="php"></el-option>
                                    <el-option label="Node.js" value="node"></el-option>
                                    <el-option label="Java" value="java"></el-option>
                                    <el-option label="Go" value="go"></el-option>
                                    <el-option label="Python" value="python"></el-option>
                                    <el-option label=".NET" value="dotnet"></el-option>
                                </el-select>
                            </el-form-item>
                        </el-col>
                        <el-col :span="16">
                            <el-form-item :label="$t('runtime.runtime')" prop="runtimeID">
                                <el-select
                                    v-model="website.runtimeID"
                                    @change="changeRuntime(website.runtimeID)"
                                    filterable
                                >
                                    <el-option
                                        v-for="run in runtimes"
                                        :key="run.name"
                                        :label="run.name + ' [' + getRuntimeLabel(run.resource) + ']'"
                                        :value="run.id"
                                    >
                                        <el-row>
                                            <el-col :span="14">
                                                <span class="runtimeName">
                                                    {{ run.name }}
                                                </span>
                                            </el-col>
                                            <el-col :span="10">
                                                {{ ' [' + getRuntimeLabel(run.resource) + ']' }}
                                            </el-col>
                                        </el-row>
                                    </el-option>
                                </el-select>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <div v-if="website.runtimeType === 'php' && runtimeResource === 'local'">
                        <el-form-item :label="$t('website.proxyType')" prop="proxyType">
                            <el-select v-model="website.proxyType">
                                <el-option :label="$t('website.tcp')" :value="'tcp'"></el-option>
                                <el-option :label="$t('website.unix')" :value="'unix'"></el-option>
                            </el-select>
                        </el-form-item>
                        <el-form-item v-if="website.proxyType === 'tcp'" :label="$t('commons.table.port')" prop="port">
                            <el-input v-model.number="website.port"></el-input>
                        </el-form-item>
                    </div>
                    <el-form-item
                        :label="$t('setting.proxyPort')"
                        prop="port"
                        v-if="website.runtimeType !== 'php' && runtimePorts.length > 1"
                    >
                        <el-select v-model="website.port">
                            <el-option
                                v-for="(port, index) in runtimePorts"
                                :key="index"
                                :label="port"
                                :value="port"
                            ></el-option>
                        </el-select>
                        <span class="input-help">{{ $t('website.runtimePortHelper') }}</span>
                    </el-form-item>
                    <el-text
                        v-if="
                            runtimes.length > 0 &&
                            website.type === 'runtime' &&
                            website.runtimeType !== 'php' &&
                            website.port == 0
                        "
                        type="danger"
                    >
                        {{ $t('website.runtimePortWarn') }}
                    </el-text>
                </div>
                <DomainCreate v-model:form="website" @gengerate="websiteForm.clearValidate()"></DomainCreate>
                <el-form-item prop="IPV6">
                    <el-checkbox v-model="website.IPV6" :label="$t('website.ipv6')" size="large" />
                </el-form-item>
                <el-form-item :label="$t('website.alias')" prop="alias">
                    <el-input v-model.trim="website.alias" :placeholder="$t('website.aliasHelper')"></el-input>
                    <div>
                        <span class="input-help">
                            <span>{{ $t('website.staticPath') + staticPath + website.alias }}</span>
                            <span v-if="website.type === 'static' && website.alias != ''">{{ '/index' }}</span>
                        </span>
                    </div>
                </el-form-item>

                <el-form-item prop="enableFtp" v-if="website.type === 'static' || website.type === 'runtime'">
                    <el-checkbox
                        @change="random"
                        v-model="website.enableFtp"
                        :label="$t('commons.button.create') + ' FTP'"
                        size="large"
                    />
                    <span class="input-help">{{ $t('website.ftpHelper') }}</span>
                </el-form-item>
                <el-row :gutter="20" v-if="website.enableFtp">
                    <el-col :span="12">
                        <el-form-item prop="ftpUser" :label="$t('website.ftpUser')">
                            <el-input v-model="website.ftpUser" />
                        </el-form-item>
                    </el-col>
                    <el-col :span="12">
                        <el-form-item prop="ftpPassword" :label="$t('website.ftpPassword')">
                            <el-input type="password" clearable v-model="website.ftpPassword" show-password>
                                <template #append>
                                    <el-button @click="random">{{ $t('commons.button.random') }}</el-button>
                                </template>
                            </el-input>
                        </el-form-item>
                    </el-col>
                </el-row>

                <el-row :gutter="20" v-if="website.type === 'proxy'">
                    <el-col :span="12">
                        <el-form-item :label="$t('website.proxyAddress')" prop="proxyAddress">
                            <el-input v-model="website.proxyAddress" :placeholder="$t('website.proxyHelper')">
                                <template #prepend>
                                    <el-select v-model="website.proxyProtocol" class="pre-select">
                                        <el-option label="http" value="http://" />
                                        <el-option label="https" value="https://" />
                                        <el-option :label="$t('website.other')" value="" />
                                    </el-select>
                                </template>
                            </el-input>
                        </el-form-item>
                    </el-col>
                    <el-col :span="12">
                        <el-form-item :label="$t('app.app')">
                            <el-select
                                v-model="website.appInstallId"
                                class="p-w-200"
                                filterable
                                @change="changeInstall"
                            >
                                <el-option
                                    v-for="(appInstall, index) in appInstalls"
                                    :key="index"
                                    :label="appInstall.name"
                                    :value="appInstall.id"
                                ></el-option>
                            </el-select>
                        </el-form-item>
                    </el-col>
                </el-row>

                <el-form-item prop="createDb" v-if="website.type === 'runtime'">
                    <el-checkbox
                        @change="randomDbPassword"
                        v-model="website.createDb"
                        :label="$t('website.createDb')"
                        size="large"
                    />
                </el-form-item>
                <el-row :gutter="20" v-if="website.type === 'runtime' && website.createDb">
                    <el-col :span="24">
                        <el-form-item :label="$t('menu.database')" prop="dbHost">
                            <el-row :gutter="20">
                                <el-col :span="12">
                                    <el-select
                                        v-model="website.dbType"
                                        class="p-w-200"
                                        @change="getAppByService(website.dbType)"
                                    >
                                        <el-option label="MySQL" value="mysql" />
                                        <el-option label="MariaDB" value="mariadb" />
                                        <el-option label="PostgreSQL" value="postgresql" />
                                    </el-select>
                                </el-col>
                                <el-col :span="12">
                                    <el-select v-model="website.dbHost" class="p-w-200">
                                        <el-option
                                            v-for="(service, index) in dbServices"
                                            :key="index"
                                            :label="service.label"
                                            :value="service.value"
                                        ></el-option>
                                    </el-select>
                                </el-col>
                            </el-row>
                        </el-form-item>
                    </el-col>
                    <el-col :span="24">
                        <el-form-item :label="$t('commons.table.name')" prop="dbName">
                            <el-input clearable v-model.trim="website.dbName" @input="website.dbUser = website.dbName">
                                <template #append>
                                    <el-select v-model="website.dbFormat" class="p-w-100">
                                        <el-option label="utf8mb4" value="utf8mb4" />
                                        <el-option label="utf-8" value="utf8" />
                                        <el-option label="gbk" value="gbk" />
                                        <el-option label="big5" value="big5" />
                                    </el-select>
                                </template>
                            </el-input>
                        </el-form-item>
                    </el-col>
                    <el-col :span="12">
                        <el-form-item :label="$t('commons.login.username')" prop="dbUser">
                            <el-input clearable v-model.trim="website.dbUser" />
                        </el-form-item>
                    </el-col>
                    <el-col :span="12">
                        <el-form-item :label="$t('commons.login.password')" prop="dbPassword">
                            <el-input type="dbPassword" clearable show-password v-model.trim="website.dbPassword">
                                <template #append>
                                    <el-button @click="randomDbPassword">{{ $t('commons.button.random') }}</el-button>
                                </template>
                            </el-input>
                        </el-form-item>
                    </el-col>
                </el-row>

                <el-form-item prop="enableSSL">
                    <el-checkbox v-model="website.enableSSL" :label="$t('website.enableHTTPS')" size="large" />
                    <span class="input-help">{{ $t('website.enableSSLHelper') }}</span>
                </el-form-item>

                <div v-if="website.enableSSL">
                    <el-form-item :label="$t('website.acmeAccountManage')" prop="acmeAccountID">
                        <el-select
                            v-model="website.acmeAccountID"
                            :placeholder="$t('website.selectAcme')"
                            @change="listSSLs"
                        >
                            <el-option :key="0" :label="$t('website.imported')" :value="0"></el-option>
                            <el-option
                                v-for="(acme, index) in acmeAccounts"
                                :key="index"
                                :label="acme.email"
                                :value="acme.id"
                            >
                                <span>
                                    {{ acme.email }}
                                    <el-tag class="ml-5">{{ getAccountName(acme.type) }}</el-tag>
                                </span>
                            </el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('website.ssl')" prop="websiteSSLID" :hide-required-asterisk="true">
                        <el-select
                            v-model="website.websiteSSLID"
                            :placeholder="$t('website.selectSSL')"
                            @change="changeSSl(website.websiteSSLID)"
                        >
                            <el-option
                                v-for="(ssl, index) in ssls"
                                :key="index"
                                :label="ssl.primaryDomain"
                                :value="ssl.id"
                                :disabled="ssl.pem == ''"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="' '" v-if="websiteSSL && websiteSSL.id > 0">
                        <el-descriptions :column="7" border direction="vertical">
                            <el-descriptions-item :label="$t('website.primaryDomain')">
                                {{ websiteSSL.primaryDomain }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('website.otherDomains')">
                                {{ websiteSSL.domains }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('website.brand')">
                                {{ websiteSSL.organization }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('ssl.provider')">
                                {{ getProvider(websiteSSL.provider) }}
                            </el-descriptions-item>
                            <el-descriptions-item
                                :label="$t('ssl.acmeAccount')"
                                v-if="websiteSSL.acmeAccount && websiteSSL.provider !== 'manual'"
                            >
                                {{ websiteSSL.acmeAccount.email }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('website.expireDate')">
                                {{ dateFormatSimple(websiteSSL.expireDate) }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('website.remark')">
                                {{ websiteSSL.description }}
                            </el-descriptions-item>
                        </el-descriptions>
                    </el-form-item>
                </div>
                <el-form-item :label="$t('website.remark')" prop="remark">
                    <el-input type="textarea" :rows="3" clearable v-model="website.remark" />
                </el-form-item>
            </el-form>
        </div>
        <template #footer>
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="submit(websiteForm)" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
        <Check ref="preCheckRef"></Check>
        <el-card width="30%" v-if="!versionExist" class="mask-prompt">
            <span>
                {{ $t('runtime.openrestyWarn') }}
            </span>
        </el-card>
        <TaskLog ref="taskLog" />
    </DrawerPro>
</template>

<script lang="ts" setup name="CreateWebSite">
import { App } from '@/api/interface/app';
import { searchApp, getAppInstalled } from '@/api/modules/app';
import {
    createWebsite,
    getWebsiteOptions,
    listSSL,
    preCheck,
    searchAcmeAccount,
    getDirConfig,
} from '@/api/modules/website';
import { Rules, checkNumberRange } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import Check from '../check/index.vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { getAgentGroupList } from '@/api/modules/group';
import { Group } from '@/api/interface/group';
import { SearchRuntimes } from '@/api/modules/runtime';
import { Runtime } from '@/api/interface/runtime';
import { getRandomStr, getRuntimeLabel } from '@/utils/util';
import TaskLog from '@/components/log/task/index.vue';
import { getAppService } from '@/api/modules/app';
import { v4 as uuidv4 } from 'uuid';
import { dateFormatSimple, getProvider, getAccountName } from '@/utils/util';
import { Website } from '@/api/interface/website';
import DomainCreate from '@/views/website/website/domain-create/index.vue';
import { getPathByType } from '@/api/modules/files';
import { getWebsiteTypes } from '@/global/mimetype';
import AppInstallForm from '@/views/app-store/detail/form/index.vue';

const websiteForm = ref<FormInstance>();

const initData = () => ({
    primaryDomain: '',
    type: 'deployment',
    alias: '',
    remark: '',
    appType: 'installed',
    appInstallId: undefined,
    webSiteGroupId: 1,
    otherDomains: '',
    proxy: '',
    runtimeID: undefined,
    appinstall: {
        appId: 0,
        name: '',
        appDetailId: 0,
        params: {},
        version: '',
        appkey: '',

        advanced: false,
        cpuQuota: 0,
        memoryLimit: 0,
        memoryUnit: 'MB',
        containerName: '',
        allowPort: false,
    },
    IPV6: false,
    enableFtp: false,
    ftpUser: '',
    ftpPassword: '',
    proxyType: 'tcp',
    port: 9000,
    proxyProtocol: 'http://',
    proxyAddress: '',
    runtimeType: 'php',
    taskID: '',
    createDb: false,
    dbName: '',
    dbPassword: '',
    dbFormat: 'utf8mb4',
    dbUser: '',
    dbType: 'mysql',
    dbHost: '',
    enableSSL: false,
    websiteSSLID: undefined,
    acmeAccountID: undefined,
    domains: [],
    parentWebsiteID: undefined,
    siteDir: '',
});
const website = ref(initData());
const rules = ref<any>({
    primaryDomain: [Rules.domainWithPort],
    alias: [Rules.alias],
    type: [Rules.requiredInput],
    webSiteGroupId: [Rules.requiredSelectBusiness],
    appInstallId: [Rules.requiredSelectBusiness],
    appType: [Rules.requiredInput],
    proxyAddress: [Rules.requiredInput],
    runtimeID: [Rules.requiredSelectBusiness],
    appinstall: {
        name: [Rules.appName],
        appId: [Rules.requiredSelectBusiness],
        params: {},
        cpuQuota: [Rules.requiredInput, checkNumberRange(0, 99999)],
        memoryLimit: [Rules.requiredInput, checkNumberRange(0, 9999999999)],
        containerName: [Rules.containerName],
    },
    ftpUser: [Rules.simpleName],
    ftpPassword: [Rules.simplePassword],
    proxyType: [Rules.requiredSelect],
    port: [Rules.port],
    runtimeType: [Rules.requiredInput],
    dbName: [Rules.requiredInput, Rules.dbName],
    dbUser: [Rules.requiredInput, Rules.name],
    dbPassword: [Rules.requiredInput, Rules.paramComplexity],
    dbHost: [Rules.requiredSelect],
    websiteSSLID: [Rules.requiredSelect],
    parentWebsiteID: [Rules.requiredSelect],
    siteDir: [Rules.requiredSelect],
});

const open = ref(false);
const loading = ref(false);
const groups = ref<Group.GroupInfo[]>([]);
const acmeAccounts = ref();
const appInstalls = ref<App.AppInstalled[]>([]);
const appReq = reactive({
    type: 'website',
    page: 1,
    pageSize: 100,
});
const apps = ref<App.App[]>([]);
const preCheckRef = ref();
const staticPath = ref('');
const runtimeResource = ref('appstore');
const initRuntimeReq = () => ({
    page: 1,
    pageSize: 100,
    status: 'Running',
    type: 'php',
});
const runtimeReq = ref<Runtime.RuntimeReq>(initRuntimeReq());
const runtimes = ref<Runtime.RuntimeDTO[]>([]);
const versionExist = ref(true);
const em = defineEmits(['close']);
const taskLog = ref();
const dbServices = ref();
const ssls = ref();
const websiteSSL = ref();
const parentWebsites = ref();
const dirs = ref([]);
const runtimePorts = ref([]);
const WebsiteTypes = getWebsiteTypes();
const installFormRef = ref();

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const random = async () => {
    website.value.ftpPassword = getRandomStr(16);
};

const randomDbPassword = async () => {
    website.value.dbPassword = getRandomStr(16);
};

const changeType = (type: string) => {
    switch (type) {
        case 'deployment':
            website.value.appType = 'installed';
            if (appInstalls.value && appInstalls.value.length > 0 && appInstalls.value[0].status === 'Running') {
                website.value.appInstallId = appInstalls.value[0].id;
            }
            break;
        case 'runtime':
            getRuntimes();
            getAppByService(website.value.dbType);
            break;
        case 'proxy':
            website.value.proxyAddress = '';
            searchAppInstalled('proxy');
            break;
        case 'subsite':
            listWebsites();
            break;
        default:
            website.value.appInstallId = undefined;
            break;
    }
    website.value.type = type;
    versionExist.value = true;
};

const searchAppInstalled = (appType: string) => {
    getAppInstalled({ type: appType, unused: true, all: true, page: 1, pageSize: 100 }).then((res) => {
        appInstalls.value = res.data.items;
        website.value.appInstallId = undefined;
        if (
            appType == 'website' &&
            res.data.items &&
            res.data.items.length > 0 &&
            res.data.items[0].status === 'Running'
        ) {
            website.value.appInstallId = res.data.items[0].id;
        }
    });
};

const getAppByService = async (key: string) => {
    const res = await getAppService(key);
    dbServices.value = res.data;
};

const changeInstall = () => {
    appInstalls.value.forEach((app) => {
        if (app.id === website.value.appInstallId) {
            website.value.proxyProtocol = 'http://';
            website.value.proxyAddress = '127.0.0.1:' + app.httpPort;
        }
    });
};

const searchAppList = () => {
    searchApp(appReq).then((res) => {
        apps.value = res.data.items;
        if (res.data.items.length > 0) {
            website.value.appinstall.appId = res.data.items[0].id;
            website.value.appinstall.appkey = res.data.items[0].key;
            changeApp();
        }
    });
};

const changeApp = () => {
    apps.value.forEach((app) => {
        if (app.id === website.value.appinstall.appId) {
            website.value.appinstall.appkey = app.key;
            installFormRef.value.initForm(app.key);
        }
    });
};

const changeRuntimeType = () => {
    runtimeReq.value.type = website.value.runtimeType;
    website.value.appinstall.advanced = false;
    website.value.runtimeID = undefined;
    getRuntimes();
};

const changeRuntime = (runID: number) => {
    website.value.port = 0;
    runtimes.value.forEach((item) => {
        if (item.id === runID) {
            runtimeResource.value = item.resource;
            if (runtimeResource.value == 'local') {
                website.value.port = 9000;
            } else {
                runtimePorts.value = item.port.split(',').map((port: string) => parseInt(port.trim(), 10));
                if (runtimePorts.value.length > 0) {
                    website.value.port = runtimePorts.value[0];
                }
            }
        }
    });
};

const getRuntimes = async () => {
    website.value.port = 0;
    try {
        const res = await SearchRuntimes(runtimeReq.value);
        runtimes.value = res.data.items || [];
        if (runtimes.value.length > 0) {
            const first = runtimes.value[0];
            website.value.runtimeID = first.id;
            runtimeResource.value = first.resource;
            runtimePorts.value = first.port.split(',').map((port: string) => parseInt(port.trim(), 10));
            if (runtimePorts.value.length > 0) {
                website.value.port = runtimePorts.value[0];
            }
        }
    } catch (error) {}
};

const acceptParams = async () => {
    website.value = initData();
    if (websiteForm.value) {
        websiteForm.value.resetFields();
    }
    getPathByType('websiteDir').then((res) => {
        staticPath.value = res.data + '/sites/';
    });

    const res = await getAgentGroupList('website');
    groups.value = res.data;
    website.value.webSiteGroupId = res.data[0].id;
    website.value.type = 'deployment';
    runtimeResource.value = 'appstore';
    runtimeReq.value = initRuntimeReq();

    searchAppInstalled('website');
    listAcmeAccount();

    open.value = true;
};

const changeAppType = (type: string) => {
    if (type === 'installed') {
        searchAppInstalled('website');
    } else {
        searchAppList();
    }
};

const openTaskLog = (taskID: string) => {
    taskLog.value.acceptParams(taskID);
};

const listAcmeAccount = () => {
    searchAcmeAccount({ page: 1, pageSize: 100 }).then((res) => {
        acmeAccounts.value = res.data.items || [];
    });
};

const listSSLs = () => {
    listSSL({
        acmeAccountID: String(website.value.acmeAccountID),
    }).then((res) => {
        ssls.value = res.data || [];
        website.value.websiteSSLID = undefined;
        websiteSSL.value = {};
        if (ssls.value.length > 0) {
            for (const ssl of ssls.value) {
                if (ssl.pem != '') {
                    website.value.websiteSSLID = ssl.id;
                    changeSSl(website.value.websiteSSLID);
                    break;
                }
            }
        }
    });
};

const changeSSl = (sslid: number) => {
    const res = ssls.value.filter((element: Website.SSL) => {
        return element.id == sslid;
    });
    websiteSSL.value = res[0];
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate(async (valid) => {
        if (!valid) {
            return;
        }
        if (website.value.type == 'deployment' && website.value.appType === 'new') {
            const isValid = await installFormRef.value?.validate();
            if (!isValid) return;
        }
        if (website.value.type === 'runtime' && website.value.runtimeType !== 'php' && website.value.port == 0) {
            MsgError(i18n.global.t('website.runtimePortWarn'));
            return;
        }
        loading.value = true;
        try {
            const res = await preCheck({});
            if (res.data) {
                preCheckRef.value.acceptParams({ items: res.data });
            } else {
                if (website.value.type === 'proxy') {
                    website.value.proxy = website.value.proxyProtocol + website.value.proxyAddress;
                }
                if (!website.value.enableFtp) {
                    website.value.ftpUser = '';
                    website.value.ftpPassword = '';
                }
                const taskID = uuidv4();
                website.value.taskID = taskID;
                await createWebsite(website.value);
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                handleClose();
                openTaskLog(taskID);
            }
        } catch (error) {
        } finally {
            loading.value = false;
        }
    });
};

watch(
    () => website.value.domains,
    (value) => {
        if (value.length > 0) {
            const firstDomain = value[0].domain;
            changeAlias(firstDomain);
        }
    },
    { deep: true },
);

const changeAlias = (value: string) => {
    const domain = value.split(':')[0];
    website.value.alias = domain;
};

const listWebsites = async () => {
    try {
        const res = await getWebsiteOptions({ types: ['static', 'runtime'] });
        parentWebsites.value = res.data;
        if (res.data.length > 0) {
            website.value.parentWebsiteID = res.data[0].id;
            getDir(res.data[0].id);
        }
    } catch (error) {}
};

const getDir = async (websiteID: number) => {
    try {
        const res = await getDirConfig({ id: websiteID });
        dirs.value = res.data.dirs;
        if (res.data.dirs.length > 0) {
            website.value.siteDir = res.data.dirs[0];
        }
    } catch (error) {}
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.runtimeName {
    width: 250px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: inline-block;
}
</style>
