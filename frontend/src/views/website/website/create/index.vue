<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader :header="$t('website.create')" :back="handleClose">
                <template #buttons>
                    <span class="drawer-header-button">
                        <template
                            v-for="item in [
                                {
                                    label: i18n.global.t('website.deployment'),
                                    value: 'deployment',
                                },
                                {
                                    label: i18n.global.t('runtime.runtime'),
                                    value: 'runtime',
                                },

                                {
                                    label: i18n.global.t('website.proxy'),
                                    value: 'proxy',
                                },
                                {
                                    label: i18n.global.t('website.static'),
                                    value: 'static',
                                },
                            ]"
                            :key="item.value"
                        >
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
            </DrawerHeader>
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
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
                <br />
                <el-form
                    ref="websiteForm"
                    label-position="top"
                    :model="website"
                    label-width="125px"
                    :rules="rules"
                    :validate-on-rule-change="false"
                >
                    <el-form-item :label="$t('website.group')" prop="webSiteGroupId">
                        <el-select v-model="website.webSiteGroupId">
                            <el-option
                                v-for="(group, index) in groups"
                                :key="index"
                                :label="group.name"
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
                                <el-radio :label="'new'">
                                    {{ $t('website.appNew') }}
                                </el-radio>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item
                            v-if="website.appType == 'installed'"
                            :label="$t('website.appInstalled')"
                            prop="appInstallId"
                        >
                            <el-select v-model="website.appInstallId">
                                <el-option
                                    v-for="(appInstall, index) in appInstalles"
                                    :key="index"
                                    :label="appInstall.name"
                                    :value="appInstall.id"
                                ></el-option>
                            </el-select>
                        </el-form-item>
                        <div v-if="website.appType == 'new'">
                            <el-form-item :label="$t('app.app')" prop="appinstall.appId">
                                <el-row :gutter="20">
                                    <el-col :span="12">
                                        <el-select v-model="website.appinstall.appId" @change="changeApp()">
                                            <el-option
                                                v-for="(app, index) in apps"
                                                :key="index"
                                                :label="app.name"
                                                :value="app.id"
                                            ></el-option>
                                        </el-select>
                                    </el-col>
                                    <el-col :span="12">
                                        <el-select
                                            v-model="website.appinstall.version"
                                            @change="getAppDetail(website.appinstall.version)"
                                        >
                                            <el-option
                                                v-for="(version, index) in appVersions"
                                                :key="index"
                                                :label="version"
                                                :value="version"
                                            ></el-option>
                                        </el-select>
                                    </el-col>
                                </el-row>
                            </el-form-item>
                            <el-form-item :label="$t('app.name')" prop="appinstall.name">
                                <el-input v-model.trim="website.appinstall.name"></el-input>
                            </el-form-item>
                            <Params
                                :key="paramKey"
                                v-model:form="website.appinstall.params"
                                v-model:rules="rules.appinstall.params"
                                :params="appParams"
                                :propStart="'appinstall.params.'"
                            ></Params>
                        </div>
                    </div>
                    <div v-if="website.type === 'runtime'">
                        <el-form-item :label="$t('runtime.runtime')" prop="runtimeID">
                            <el-select
                                v-model="website.runtimeID"
                                @change="changeRuntime(website.runtimeID)"
                                filterable
                            >
                                <el-option
                                    v-for="run in runtimes"
                                    :key="run.name"
                                    :label="run.name + '(' + $t('runtime.' + run.resource) + ')'"
                                    :value="run.id"
                                ></el-option>
                            </el-select>
                        </el-form-item>
                        <Params
                            v-if="runtimeResource === 'appstore'"
                            :key="paramKey"
                            v-model:form="website.appinstall.params"
                            v-model:rules="rules.appinstall.params"
                            :params="appParams"
                            :propStart="'appinstall.params.'"
                        ></Params>
                        <div v-else>
                            <el-form-item :label="$t('website.proxyType')" prop="proxyType">
                                <el-select v-model="website.proxyType">
                                    <el-option :label="$t('website.tcp')" :value="'tcp'"></el-option>
                                    <el-option :label="$t('website.unix')" :value="'unix'"></el-option>
                                </el-select>
                            </el-form-item>
                            <el-form-item v-if="website.proxyType === 'tcp'" :label="$t('website.port')" prop="port">
                                <el-input v-model.number="website.port"></el-input>
                            </el-form-item>
                        </div>
                    </div>
                    <el-form-item :label="$t('website.primaryDomain')" prop="primaryDomain">
                        <el-input
                            v-model.trim="website.primaryDomain"
                            @input="changeAlias(website.primaryDomain)"
                        ></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.otherDomains')" prop="otherDomains">
                        <el-input
                            type="textarea"
                            :autosize="{ minRows: 2, maxRows: 6 }"
                            v-model="website.otherDomains"
                            :placeholder="$t('website.domainHelper')"
                        ></el-input>
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
                    <el-form-item v-if="website.type === 'proxy'" :label="$t('website.proxyAddress')" prop="proxy">
                        <el-input v-model="website.proxy" :placeholder="$t('website.proxyHelper')"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.remark')" prop="remark">
                        <el-input v-model="website.remark"></el-input>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span>
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(websiteForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
        <Check ref="preCheckRef"></Check>
    </el-drawer>
</template>

<script lang="ts" setup name="CreateWebSite">
import DrawerHeader from '@/components/drawer-header/index.vue';
import { App } from '@/api/interface/app';
import { GetApp, GetAppDetail, SearchApp, GetAppInstalled, GetAppDetailByID } from '@/api/modules/app';
import { CreateWebsite, PreCheck } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import Params from '@/views/app-store/detail/params/index.vue';
import Check from '../check/index.vue';
import { MsgSuccess } from '@/utils/message';
import { GetGroupList } from '@/api/modules/group';
import { Group } from '@/api/interface/group';
import { SearchRuntimes } from '@/api/modules/runtime';
import { Runtime } from '@/api/interface/runtime';

const websiteForm = ref<FormInstance>();
const website = ref({
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
    },
    proxyType: 'tcp',
    port: 9000,
});
let rules = ref<any>({
    primaryDomain: [Rules.domain],
    alias: [Rules.linuxName],
    type: [Rules.requiredInput],
    webSiteGroupId: [Rules.requiredSelectBusiness],
    appInstallId: [Rules.requiredSelectBusiness],
    appType: [Rules.requiredInput],
    proxy: [Rules.requiredInput],
    runtimeID: [Rules.requiredSelectBusiness],
    appinstall: {
        name: [Rules.appName],
        appId: [Rules.requiredSelectBusiness],
        params: {},
    },
    proxyType: [Rules.requiredSelect],
    port: [Rules.port],
});

let open = ref(false);
let loading = ref(false);
let groups = ref<Group.GroupInfo[]>([]);

let appInstalles = ref<App.AppInstalled[]>([]);
let appReq = reactive({
    type: 'website',
    page: 1,
    pageSize: 20,
});
let apps = ref<App.App[]>([]);
let appVersions = ref<string[]>([]);
let appDetail = ref<App.AppDetail>();
let appParams = ref<App.AppParams>();
let paramKey = ref(1);
let preCheckRef = ref();
let staticPath = ref('');
let runtimeResource = ref('appstore');
const runtimeReq = ref<Runtime.RuntimeReq>({
    page: 1,
    pageSize: 100,
    status: 'normal',
});
const runtimes = ref<Runtime.RuntimeDTO[]>([]);

const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const changeType = (type: string) => {
    if (type == 'deployemnt') {
        if (appInstalles.value && appInstalles.value.length > 0) {
            website.value.appInstallId = appInstalles.value[0].id;
        }
    } else if (type == 'runtime') {
        getRuntimes();
    } else {
        website.value.appInstallId = undefined;
    }
    website.value.type = type;
};

const searchAppInstalled = () => {
    GetAppInstalled({ type: 'website', unused: true }).then((res) => {
        appInstalles.value = res.data;
        if (res.data && res.data.length > 0) {
            website.value.appInstallId = res.data[0].id;
        }
    });
};

const searchApp = () => {
    SearchApp(appReq).then((res) => {
        apps.value = res.data.items;
        if (res.data.items.length > 0) {
            website.value.appinstall.appId = res.data.items[0].id;
            website.value.appinstall.appkey = res.data.items[0].key;
            getApp();
        }
    });
};

const changeApp = () => {
    apps.value.forEach((app) => {
        if (app.id === website.value.appinstall.appId) {
            website.value.appinstall.appkey = app.key;
            getApp();
        }
    });
};

const getApp = () => {
    GetApp(website.value.appinstall.appkey).then((res) => {
        appVersions.value = res.data.versions;
        if (res.data.versions.length > 0) {
            website.value.appinstall.version = res.data.versions[0];
            getAppDetail(res.data.versions[0]);
        }
    });
};

const getAppDetail = (version: string) => {
    GetAppDetail(website.value.appinstall.appId, version, 'app').then((res) => {
        website.value.appinstall.appDetailId = res.data.id;
        appDetail.value = res.data;
        appParams.value = res.data.params;
        paramKey.value++;
    });
};

const getAppDetailByID = (id: number) => {
    GetAppDetailByID(id).then((res) => {
        website.value.appinstall.appDetailId = res.data.id;
        appDetail.value = res.data;
        appParams.value = res.data.params;
        paramKey.value++;
    });
};

const changeRuntime = (runID: number) => {
    runtimes.value.forEach((item) => {
        if (item.id === runID) {
            runtimeResource.value = item.resource;
            if (item.type === 'appstore') {
                getAppDetailByID(item.appDetailId);
            }
        }
    });
};

const getRuntimes = async () => {
    try {
        const res = await SearchRuntimes(runtimeReq.value);
        runtimes.value = res.data.items || [];
        if (runtimes.value.length > 0) {
            const first = runtimes.value[0];
            website.value.runtimeID = first.id;
            runtimeResource.value = first.resource;
            if (first.resource === 'appstore') {
                getAppDetailByID(first.appDetailId);
            }
        }
    } catch (error) {}
};

const acceptParams = async (installPath: string) => {
    if (websiteForm.value) {
        websiteForm.value.resetFields();
    }
    staticPath.value = installPath + '/www/sites/';

    const res = await GetGroupList({ type: 'website' });
    groups.value = res.data;
    website.value.webSiteGroupId = res.data[0].id;
    website.value.type = 'deployment';
    runtimeResource.value = 'appstore';

    searchAppInstalled();

    open.value = true;
};

const changeAppType = (type: string) => {
    if (type === 'installed') {
        searchAppInstalled();
    } else {
        searchApp();
    }
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        PreCheck({})
            .then((res) => {
                if (res.data) {
                    loading.value = false;
                    preCheckRef.value.acceptParams({ items: res.data });
                } else {
                    CreateWebsite(website.value)
                        .then(() => {
                            MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                            handleClose();
                        })
                        .finally(() => {
                            loading.value = false;
                        });
                }
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const changeAlias = (value: string) => {
    website.value.alias = value;
};

defineExpose({
    acceptParams,
});
</script>
