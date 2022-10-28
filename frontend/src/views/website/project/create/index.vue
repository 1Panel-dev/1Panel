<template>
    <el-dialog v-model="open" :title="$t('website.create')" width="40%" :before-close="handleClose">
        <el-form ref="websiteForm" label-position="right" :model="website" label-width="100px" :rules="rules">
            <el-form-item :label="$t('website.type')" prop="type">
                <el-select v-model="website.type">
                    <el-option :label="$t('website.deployment')" value="deployment"></el-option>
                    <el-option :label="$t('website.static')" value="static"></el-option>
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('website.group')" prop="webSiteGroupID">
                <el-select v-model="website.webSiteGroupID">
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
                    <el-radio-group v-model="website.appType">
                        <el-radio :label="'installed'" :value="'installed'">
                            {{ $t('website.app_installed') }}
                        </el-radio>
                        <el-radio :label="'new'">
                            {{ $t('website.app_new') }}
                        </el-radio>
                    </el-radio-group>
                </el-form-item>
                <el-form-item
                    v-if="website.appType == 'installed'"
                    :label="$t('website.app_installed')"
                    prop="appInstallID"
                >
                    <el-select v-model="website.appInstallID">
                        <el-option
                            v-for="(appInstall, index) in appInstalles"
                            :key="index"
                            :label="appInstall.name"
                            :value="appInstall.id"
                        ></el-option>
                    </el-select>
                </el-form-item>
            </div>
            <el-form-item :label="$t('website.primaryDomain')" prop="primaryDomain">
                <el-input v-model="website.primaryDomain"></el-input>
            </el-form-item>
            <el-form-item :label="$t('website.otherDomains')" prop="otherDomains">
                <el-input v-model="website.otherDomains"></el-input>
            </el-form-item>
            <el-form-item :label="$t('website.alias')" prop="alias">
                <el-input v-model="website.alias"></el-input>
            </el-form-item>
            <el-form-item :label="$t('website.remark')" prop="remark">
                <el-input v-model="website.remark"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(websiteForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup name="CreateWebSite">
import { App } from '@/api/interface/app';
import { WebSite } from '@/api/interface/website';
import { SearchAppInstalled } from '@/api/modules/app';
import { CreateWebsite, listGroups } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import { FormRules, ElDialog, ElForm, FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';

const websiteForm = ref<FormInstance>();
const website = reactive({
    primaryDomain: '',
    type: 'deployment',
    alias: '',
    remark: '',
    domains: [],
    appType: 'installed',
    appInstallID: 0,
    webSiteGroupID: 1,
    otherDomains: '',
});
let rules = reactive<FormRules>({
    primaryDomain: [Rules.requiredInput],
    alias: [Rules.requiredInput],
    type: [Rules.requiredInput],
    webSiteGroupID: [Rules.requiredInput],
    appInstallID: [Rules.requiredInput],
    appType: [Rules.requiredInput],
});

let open = ref(false);
let loading = ref(false);
let groups = ref<WebSite.Group[]>([]);
let appInstalles = ref<App.AppInstalled[]>([]);

const handleClose = () => {
    open.value = false;
};

const acceptParams = async () => {
    await listGroups().then((res) => {
        groups.value = res.data;
        open.value = true;
    });
    await SearchAppInstalled({ type: 'website' }).then((res) => {
        appInstalles.value = res.data;
        if (res.data.length > 0) {
            website.appInstallID = res.data[0].id;
        }
    });
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        CreateWebsite(website).then(() => {});
    });
};

defineExpose({
    acceptParams,
});
</script>
