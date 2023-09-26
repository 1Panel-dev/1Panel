<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader
                :header="$t('runtime.' + mode)"
                :hideResource="mode == 'create'"
                :resource="runtime.name"
                :back="handleClose"
            />
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
                    <el-form-item :label="$t('runtime.app')" prop="appID">
                        <el-row :gutter="20">
                            <el-col :span="12">
                                <el-select
                                    v-model="runtime.appID"
                                    :disabled="mode === 'edit'"
                                    @change="changeApp(runtime.appID)"
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
                    <el-form-item :label="$t('runtime.codeDir')" prop="codeDir">
                        <el-input v-model.trim="runtime.codeDir" :disabled="mode === 'edit'">
                            <template #prepend>
                                <FileList
                                    :disabled="mode === 'edit'"
                                    :path="runtime.codeDir"
                                    @choose="getPath"
                                    :dir="true"
                                ></FileList>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('runtime.runScript')" prop="params.EXEC_SCRIPT">
                        <el-select v-model="runtime.params['EXEC_SCRIPT']">
                            <el-option
                                v-for="(script, index) in scripts"
                                :key="index"
                                :label="script.name + ' 【 ' + script.script + ' 】'"
                                :value="script.name"
                            >
                                <el-row :gutter="10">
                                    <el-col :span="4">{{ script.name }}</el-col>
                                    <el-col :span="10">{{ ' 【 ' + script.script + ' 】' }}</el-col>
                                </el-row>
                            </el-option>
                        </el-select>
                        <span class="input-help">{{ $t('runtime.runScriptHelper') }}</span>
                    </el-form-item>
                    <el-row :gutter="20">
                        <el-col :span="10">
                            <el-form-item :label="$t('runtime.appPort')" prop="params.NODE_APP_PORT">
                                <el-input v-model.number="runtime.params['NODE_APP_PORT']" />
                                <span class="input-help">{{ $t('runtime.appPortHelper') }}</span>
                            </el-form-item>
                        </el-col>
                        <el-col :span="10">
                            <el-form-item :label="$t('runtime.externalPort')" prop="params.PANEL_APP_PORT_HTTP">
                                <el-input v-model.number="runtime.params['PANEL_APP_PORT_HTTP']" />
                                <span class="input-help">{{ $t('runtime.externalPortHelper') }}</span>
                            </el-form-item>
                        </el-col>

                        <el-col :span="4">
                            <el-form-item :label="$t('app.allowPort')" prop="params.HOST_IP">
                                <el-select v-model="runtime.params['HOST_IP']">
                                    <el-option :label="$t('runtime.open')" value="0.0.0.0"></el-option>
                                    <el-option :label="$t('runtime.close')" value="127.0.0.1"></el-option>
                                </el-select>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-form-item :label="$t('runtime.packageManager')" prop="params.PACKAGE_MANAGER">
                        <el-select v-model="runtime.params['PACKAGE_MANAGER']">
                            <el-option label="npm" value="npm"></el-option>
                            <el-option label="yarn" value="yarn"></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('app.containerName')" prop="params.CONTAINER_NAME">
                        <el-input v-model.trim="runtime.params['CONTAINER_NAME']"></el-input>
                    </el-form-item>
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
import { CreateRuntime, GetNodeScripts, GetRuntime, UpdateRuntime } from '@/api/modules/runtime';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { reactive, ref, watch } from 'vue';
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
const mode = ref('create');
const editParams = ref<App.InstallParams[]>();
const appVersions = ref<string[]>([]);
const appReq = reactive({
    type: 'node',
    page: 1,
    pageSize: 20,
    resource: 'remote',
});
const initData = (type: string) => ({
    name: '',
    appDetailID: undefined,
    image: '',
    params: {
        PACKAGE_MANAGER: 'npm',
        HOST_IP: '0.0.0.0',
    },
    type: type,
    resource: 'appstore',
    rebuild: false,
    codeDir: '/',
});
let runtime = reactive<Runtime.RuntimeCreate>(initData('node'));
const rules = ref<any>({
    name: [Rules.appName],
    appID: [Rules.requiredSelect],
    codeDir: [Rules.requiredInput],
    params: {
        NODE_APP_PORT: [Rules.requiredInput, Rules.port],
        PANEL_APP_PORT_HTTP: [Rules.requiredInput, Rules.port],
        PACKAGE_MANAGER: [Rules.requiredSelect],
        HOST_IP: [Rules.requiredSelect],
        EXEC_SCRIPT: [Rules.requiredSelect],
        CONTAINER_NAME: [Rules.requiredInput],
    },
});
const scripts = ref<Runtime.NodeScripts[]>([]);
const em = defineEmits(['close']);

watch(
    () => runtime.params['NODE_APP_PORT'],
    (newVal) => {
        if (newVal && mode.value == 'create') {
            runtime.params['PANEL_APP_PORT_HTTP'] = newVal;
        }
    },
    { deep: true },
);

watch(
    () => runtime.name,
    (newVal) => {
        if (newVal) {
            runtime.params['CONTAINER_NAME'] = newVal;
        }
    },
    { deep: true },
);

const handleClose = () => {
    open.value = false;
    em('close', false);
    runtimeForm.value?.resetFields();
};

const getPath = (codeDir: string) => {
    runtime.codeDir = codeDir;
    getScripts();
};

const getScripts = () => {
    GetNodeScripts({ codeDir: runtime.codeDir }).then((res) => {
        scripts.value = res.data;
        if (mode.value == 'create' && scripts.value.length > 0) {
            runtime.params['EXEC_SCRIPT'] = scripts.value[0].name;
        }
    });
};

const searchApp = (appID: number) => {
    SearchApp(appReq).then((res) => {
        apps.value = res.data.items || [];
        if (res.data && res.data.items && res.data.items.length > 0) {
            if (appID == null) {
                runtime.appID = res.data.items[0].id;
                getApp(res.data.items[0].key, mode.value);
            } else {
                res.data.items.forEach((item) => {
                    if (item.id === appID) {
                        getApp(item.key, mode.value);
                    }
                });
            }
        }
    });
};

const changeApp = (appID: number) => {
    for (const app of apps.value) {
        if (app.id === appID) {
            getApp(app.key, mode.value);
            break;
        }
    }
};

const changeVersion = () => {
    loading.value = true;
    GetAppDetail(runtime.appID, runtime.version, 'runtime')
        .then((res) => {
            runtime.appDetailID = res.data.id;
        })
        .finally(() => {
            loading.value = false;
        });
};

const getApp = (appkey: string, mode: string) => {
    GetApp(appkey).then((res) => {
        appVersions.value = res.data.versions || [];
        if (res.data.versions.length > 0) {
            runtime.version = res.data.versions[0];
            if (mode === 'create') {
                changeVersion();
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
        if (mode.value == 'create') {
            loading.value = true;
            CreateRuntime(runtime)
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                    handleClose();
                })
                .finally(() => {
                    loading.value = false;
                });
        } else {
            loading.value = true;
            UpdateRuntime(runtime)
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
        Object.assign(runtime, {
            id: data.id,
            name: data.name,
            appDetailId: data.appDetailID,
            image: data.image,
            type: data.type,
            resource: data.resource,
            appID: data.appID,
            version: data.version,
            rebuild: true,
            source: data.source,
            params: data.params,
            codeDir: data.codeDir,
        });
        editParams.value = data.appParams;
        if (mode.value == 'edit') {
            searchApp(data.appID);
        }
        getScripts();
    } catch (error) {}
};

const acceptParams = async (props: OperateRrops) => {
    mode.value = props.mode;
    scripts.value = [];
    if (props.mode === 'create') {
        Object.assign(runtime, initData(props.type));
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
