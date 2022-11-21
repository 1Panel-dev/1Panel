<template>
    <el-dialog v-model="open" :title="$t('website.create')" width="40%" :before-close="handleClose">
        <el-form ref="websiteForm" label-position="right" :model="website" label-width="130px" :rules="rules">
            <el-form-item :label="$t('website.type')" prop="type">
                <el-select v-model="website.type">
                    <el-option :label="$t('website.deployment')" value="deployment"></el-option>
                    <el-option :label="$t('website.static')" value="static"></el-option>
                </el-select>
            </el-form-item>
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
                                <el-select v-model="website.appinstall.appId" @change="getApp()">
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
                        <el-input v-model="website.appinstall.name"></el-input>
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
            <el-form-item :label="$t('website.primaryDomain')" prop="primaryDomain">
                <el-input v-model="website.primaryDomain"></el-input>
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
                <el-input v-model="website.alias" :placeholder="$t('website.aliasHelper')"></el-input>
            </el-form-item>
            <el-form-item :label="$t('website.remark')" prop="remark">
                <el-input v-model="website.remark"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(websiteForm)" :loading="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup name="CreateWebSite">
import { App } from '@/api/interface/app';
import { WebSite } from '@/api/interface/website';
import { GetApp, GetAppDetail, SearchApp, SearchAppInstalled } from '@/api/modules/app';
import { CreateWebsite, ListGroups } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElDialog, ElForm, FormInstance, ElMessage } from 'element-plus';
import { reactive, ref } from 'vue';
import Params from '@/views/app-store/detail/params/index.vue';

const websiteForm = ref<FormInstance>();
const website = ref({
    primaryDomain: '',
    type: 'deployment',
    alias: '',
    remark: '',
    appType: 'installed',
    appInstallId: 0,
    webSiteGroupId: 1,
    otherDomains: '',
    appinstall: {
        appId: 0,
        name: '',
        appDetailId: 0,
        params: {},
        version: '',
    },
});
let rules = ref({
    primaryDomain: [Rules.requiredInput],
    alias: [Rules.requiredInput],
    type: [Rules.requiredInput],
    webSiteGroupId: [Rules.requiredSelectBusiness],
    appInstallId: [Rules.requiredSelectBusiness],
    appType: [Rules.requiredInput],
    appinstall: {
        name: [Rules.requiredInput],
        appId: [Rules.requiredSelectBusiness],
        params: {},
    },
});

let open = ref(false);
let loading = ref(false);
let groups = ref<WebSite.Group[]>([]);
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

const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const searchAppInstalled = () => {
    SearchAppInstalled({ type: 'website' }).then((res) => {
        appInstalles.value = res.data;
        if (res.data.length > 0) {
            website.value.appInstallId = res.data[0].id;
        }
    });
};

const searchApp = () => {
    SearchApp(appReq).then((res) => {
        apps.value = res.data.items;
        if (res.data.items.length > 0) {
            website.value.appinstall.appId = res.data.items[0].id;
            getApp();
        }
    });
};

const getApp = () => {
    GetApp(website.value.appinstall.appId).then((res) => {
        appVersions.value = res.data.versions;
        if (res.data.versions.length > 0) {
            website.value.appinstall.version = res.data.versions[0];
            getAppDetail(res.data.versions[0]);
        }
    });
};

const getAppDetail = (version: string) => {
    GetAppDetail(website.value.appinstall.appId, version).then((res) => {
        website.value.appinstall.appDetailId = res.data.id;
        appDetail.value = res.data;
        appParams.value = res.data.params;
        paramKey.value++;
    });
};

const acceptParams = async () => {
    if (websiteForm.value) {
        websiteForm.value.resetFields();
    }

    await ListGroups().then((res) => {
        groups.value = res.data;
        open.value = true;
    });
    searchAppInstalled();
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
        CreateWebsite(website.value)
            .then(() => {
                ElMessage.success(i18n.global.t('commons.msg.createSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
