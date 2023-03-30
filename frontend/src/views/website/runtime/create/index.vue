<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader :header="$t('runtime.create')" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form
                    ref="runtimeForm"
                    label-position="top"
                    :model="runtimeCreate"
                    label-width="125px"
                    :rules="rules"
                    :validate-on-rule-change="false"
                >
                    <el-form-item :label="$t('runtime.name')" prop="name">
                        <el-input v-model="runtimeCreate.name"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('runtime.resource')" prop="resource">
                        <el-radio-group
                            v-model="runtimeCreate.resource"
                            @change="changeResource(runtimeCreate.resource)"
                        >
                            <el-radio :label="'AppStore'" :value="'AppStore'">
                                {{ $t('runtime.appStore') }}
                            </el-radio>
                            <el-radio :label="'Local'" :value="'Local'">
                                {{ $t('runtime.local') }}
                            </el-radio>
                        </el-radio-group>
                    </el-form-item>
                    <div v-if="runtimeCreate.resource === 'AppStore'">
                        <el-form-item :label="$t('runtime.app')" prop="appId">
                            <el-row :gutter="20">
                                <el-col :span="12">
                                    <el-select v-model="runtimeCreate.appId">
                                        <el-option
                                            v-for="(app, index) in apps"
                                            :key="index"
                                            :label="app.name"
                                            :value="app.id"
                                        ></el-option>
                                    </el-select>
                                </el-col>
                                <el-col :span="12">
                                    <el-select v-model="runtimeCreate.version">
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
                        <Params
                            v-if="initParam"
                            v-model:form="runtimeCreate"
                            v-model:params="appParams"
                            v-model:rules="rules"
                        ></Params>
                    </div>
                    <div v-else>
                        <el-alert :title="$t('runtime.localHelper')" type="info" :closable="false" />
                        <el-form-item :label="$t('runtime.version')" prop="version">
                            <el-input v-model="runtimeCreate.version"></el-input>
                        </el-form-item>
                    </div>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span>
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(runtimeForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { Runtime } from '@/api/interface/runtime';
import { GetApp, GetAppDetail, SearchApp } from '@/api/modules/app';
import { CreateRuntime } from '@/api/modules/runtime';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import Params from '../param/index.vue';

const open = ref(false);
const apps = ref<App.App[]>([]);
const runtimeForm = ref<FormInstance>();
const loading = ref(false);
const initParam = ref(false);
let appParams = ref<App.AppParams>();
let appVersions = ref<string[]>([]);
let appReq = reactive({
    type: 'php',
    page: 1,
    pageSize: 20,
});
const runtimeCreate = ref<Runtime.RuntimeCreate>({
    name: '',
    appDetailId: undefined,
    image: '',
    params: {},
    type: '',
    resource: 'AppStore',
});
let rules = ref<any>({
    name: [Rules.appName],
    resource: [Rules.requiredInput],
    appId: [Rules.requiredSelect],
    version: [Rules.requiredInput],
});

const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const changeResource = (resource: string) => {
    if (resource === 'Local') {
        runtimeCreate.value.appDetailId = undefined;
        runtimeCreate.value.version = '';
        runtimeCreate.value.params = {};
        runtimeCreate.value.image = '';
    } else {
        runtimeCreate.value.version = '';
    }
};

const searchApp = () => {
    SearchApp(appReq).then((res) => {
        apps.value = res.data.items || [];
        if (res.data && res.data.items && res.data.items.length > 0) {
            runtimeCreate.value.appId = res.data.items[0].id;
            getApp(res.data.items[0].key);
        }
    });
};
const getApp = (appkey: string) => {
    GetApp(appkey).then((res) => {
        appVersions.value = res.data.versions || [];
        if (res.data.versions.length > 0) {
            runtimeCreate.value.version = res.data.versions[0];
            GetAppDetail(runtimeCreate.value.appId, runtimeCreate.value.version, 'runtime').then((res) => {
                runtimeCreate.value.appDetailId = res.data.id;
                appParams.value = res.data.params;
                initParam.value = true;
            });
        }
    });
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        CreateRuntime(runtimeCreate.value)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const acceptParams = async () => {
    searchApp();
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>
