<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader :header="$t('runtime.' + mode)" :resource="runtime.name" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form
                    ref="runtimeForm"
                    label-position="top"
                    :model="runtime"
                    label-width="125px"
                    :rules="rules"
                    :validate-on-rule-change="false"
                >
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input :disabled="mode === 'edit'" v-model="runtime.name"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('runtime.resource')" prop="resource">
                        <el-radio-group
                            :disabled="mode === 'edit'"
                            v-model="runtime.resource"
                            @change="changeResource(runtime.resource)"
                        >
                            <el-radio :label="'appstore'">
                                {{ $t('runtime.appstore') }}
                            </el-radio>
                            <el-radio :label="'local'">
                                {{ $t('runtime.local') }}
                            </el-radio>
                        </el-radio-group>
                    </el-form-item>
                    <div v-if="runtime.resource === 'appstore'">
                        <el-form-item :label="$t('runtime.app')" prop="appId">
                            <el-row :gutter="20">
                                <el-col :span="12">
                                    <el-select
                                        v-model="runtime.appId"
                                        :disabled="mode === 'edit'"
                                        @change="changeApp(runtime.appId)"
                                    >
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
                                        v-model="runtime.version"
                                        :disabled="mode === 'edit'"
                                        @change="changeVersion()"
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
                        <el-form-item :label="$t('runtime.image')" prop="image">
                            <el-input v-model="runtime.image"></el-input>
                        </el-form-item>
                        <div v-if="initParam">
                            <el-form-item v-if="runtime.type === 'php'">
                                <el-alert :title="$t('runtime.buildHelper')" type="warning" :closable="false" />
                            </el-form-item>
                            <Params
                                v-if="mode === 'create'"
                                v-model:form="runtime.params"
                                v-model:params="appParams"
                                v-model:rules="rules"
                            ></Params>
                            <EditParams
                                v-if="mode === 'edit'"
                                v-model:form="runtime.params"
                                v-model:params="editParams"
                                v-model:rules="rules"
                            ></EditParams>
                            <el-form-item v-if="runtime.type === 'php'">
                                <el-alert type="info" :closable="false">
                                    <span>{{ $t('runtime.extendHelper') }}</span>
                                    <span v-html="$t('runtime.phpPluginHelper')"></span>
                                    <br />
                                    <span v-if="mode == 'edit'">{{ $t('runtime.rebuildHelper') }}</span>
                                </el-alert>
                            </el-form-item>
                        </div>
                    </div>
                    <div v-else>
                        <el-form-item>
                            <el-alert :title="$t('runtime.localHelper')" type="info" :closable="false" />
                        </el-form-item>
                        <el-form-item :label="$t('runtime.version')" prop="version">
                            <el-input v-model="runtime.version" :placeholder="$t('runtime.versionHelper')"></el-input>
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
import { CreateRuntime, GetRuntime, UpdateRuntime } from '@/api/modules/runtime';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import Params from '../param/index.vue';
import EditParams from '../edit/index.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';

interface OperateRrops {
    id?: number;
    mode: string;
    type: string;
}

const open = ref(false);
const apps = ref<App.App[]>([]);
const runtimeForm = ref<FormInstance>();
const loading = ref(false);
const initParam = ref(false);
const mode = ref('create');
const appParams = ref<App.AppParams>();
const editParams = ref<App.InstallParams[]>();
const appVersions = ref<string[]>([]);
const appReq = reactive({
    type: 'php',
    page: 1,
    pageSize: 20,
});
const runtime = ref<Runtime.RuntimeCreate>({
    name: '',
    appDetailId: undefined,
    image: '',
    params: {},
    type: 'php',
    resource: 'appstore',
});
const rules = ref<any>({
    name: [Rules.appName],
    resource: [Rules.requiredInput],
    appId: [Rules.requiredSelect],
    version: [Rules.requiredInput, Rules.paramCommon],
    image: [Rules.requiredInput, Rules.imageName],
});

const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const changeResource = (resource: string) => {
    if (resource === 'local') {
        runtime.value.appDetailId = undefined;
        runtime.value.version = '';
        runtime.value.params = {};
        runtime.value.image = '';
    } else {
        runtime.value.version = '';
        searchApp(null);
    }
};

const searchApp = (appId: number) => {
    SearchApp(appReq).then((res) => {
        apps.value = res.data.items || [];
        if (res.data && res.data.items && res.data.items.length > 0) {
            if (appId == null) {
                runtime.value.appId = res.data.items[0].id;
                getApp(res.data.items[0].key, mode.value);
            } else {
                res.data.items.forEach((item) => {
                    if (item.id === appId) {
                        getApp(item.key, mode.value);
                    }
                });
            }
        }
    });
};

const changeApp = (appId: number) => {
    for (const app of apps.value) {
        if (app.id === appId) {
            initParam.value = false;
            getApp(app.key, mode.value);
            break;
        }
    }
};

const changeVersion = () => {
    loading.value = true;
    initParam.value = false;
    GetAppDetail(runtime.value.appId, runtime.value.version, 'runtime')
        .then((res) => {
            runtime.value.appDetailId = res.data.id;
            runtime.value.image = res.data.image + ':' + runtime.value.version;
            appParams.value = res.data.params;
            initParam.value = true;
        })
        .finally(() => {
            loading.value = false;
        });
};

const getApp = (appkey: string, mode: string) => {
    GetApp(appkey).then((res) => {
        appVersions.value = res.data.versions || [];
        if (res.data.versions.length > 0) {
            runtime.value.version = res.data.versions[0];
            if (mode === 'create') {
                changeVersion();
            } else {
                initParam.value = true;
            }
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
        if (mode.value == 'create') {
            CreateRuntime(runtime.value)
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                    handleClose();
                })
                .finally(() => {
                    loading.value = false;
                });
        } else {
            UpdateRuntime(runtime.value)
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                    handleClose();
                })
                .finally(() => {
                    loading.value = false;
                });
        }
    });
};

const getRuntime = async (id: number) => {
    try {
        const res = await GetRuntime(id);
        const data = res.data;
        runtime.value = {
            id: data.id,
            name: data.name,
            appDetailId: data.appDetailId,
            image: data.image,
            params: {},
            type: data.type,
            resource: data.resource,
            appId: data.appId,
            version: data.version,
        };
        editParams.value = data.appParams;
        if (mode.value == 'create') {
            searchApp(data.appId);
        } else {
            initParam.value = true;
        }
    } catch (error) {}
};

const acceptParams = async (props: OperateRrops) => {
    mode.value = props.mode;
    initParam.value = false;
    if (props.mode === 'create') {
        runtime.value = {
            name: '',
            appDetailId: undefined,
            image: '',
            params: {},
            type: props.type,
            resource: 'appstore',
        };
        searchApp(null);
    } else {
        getRuntime(props.id);
    }
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>
