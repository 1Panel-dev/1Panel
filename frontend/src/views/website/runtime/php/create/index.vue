<template>
    <el-drawer :close-on-click-modal="false" :close-on-press-escape="false" v-model="open" size="50%">
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
                            <el-radio :value="'appstore'">
                                {{ $t('runtime.appstore') }}
                            </el-radio>
                            <el-radio :value="'local'">
                                {{ $t('runtime.local') }}
                            </el-radio>
                        </el-radio-group>
                    </el-form-item>
                    <div v-if="runtime.resource === 'appstore'">
                        <el-form-item :label="$t('runtime.app')" prop="appID">
                            <el-row :gutter="20">
                                <el-col :span="12">
                                    <el-select
                                        v-model="runtime.appID"
                                        :disabled="mode === 'edit'"
                                        @change="changeApp(runtime.appID)"
                                        class="p-w-200"
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
                                        class="p-w-200"
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
                        <div v-if="initParam">
                            <div v-if="runtime.type === 'php'">
                                <el-form-item :label="$t('runtime.image')" prop="image">
                                    <el-input v-model="runtime.image"></el-input>
                                </el-form-item>
                                <el-form-item :label="$t('runtime.source')" prop="source">
                                    <el-select v-model="runtime.source" filterable allow-create default-first-option>
                                        <el-option
                                            v-for="(source, index) in phpSources"
                                            :key="index"
                                            :label="source.label + ' [' + source.value + ']'"
                                            :value="source.value"
                                        ></el-option>
                                    </el-select>
                                    <span class="input-help">
                                        {{ $t('runtime.phpsourceHelper') }}
                                    </span>
                                </el-form-item>
                                <el-form-item :label="$t('php.extensions')">
                                    <el-select v-model="extensions" @change="changePHPExtension()" clearable>
                                        <el-option
                                            v-for="(extension, index) in phpExtensions"
                                            :key="index"
                                            :label="extension.name"
                                            :value="extension.extensions"
                                        ></el-option>
                                    </el-select>
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
                                <el-form-item>
                                    <el-alert :title="$t('runtime.buildHelper')" type="warning" :closable="false" />
                                    <span class="input-help">
                                        <span>{{ $t('runtime.extendHelper') }}</span>
                                        <el-link target="_blank" type="primary" :href="globalStore.docsUrl + phpDocURL">
                                            {{ $t('php.toExtensionsList') }}
                                        </el-link>
                                        <br />
                                    </span>
                                </el-form-item>
                                <div v-if="mode == 'edit'">
                                    <el-form-item>
                                        <el-checkbox v-model="runtime.rebuild">
                                            {{ $t('runtime.rebuild') }}
                                        </el-checkbox>
                                        <span class="input-help">{{ $t('runtime.rebuildHelper') }}</span>
                                    </el-form-item>
                                </div>
                            </div>
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
import { CreateRuntime, GetRuntime, ListPHPExtensions, UpdateRuntime } from '@/api/modules/runtime';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import Params from '../param/index.vue';
import EditParams from '../edit/index.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const phpDocURL = globalStore.isIntl ? `/user_manual/websites/runtime_php/` : '/user_manual/websites/php/#php_1';

interface OperateProps {
    id?: number;
    mode: string;
    type: string;
    appID?: number;
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
const phpExtensions = ref([]);
const appReq = reactive({
    type: 'php',
    page: 1,
    pageSize: 20,
});
const phpSources = globalStore.isIntl
    ? [
          {
              label: i18n.global.t('runtime.default'),
              value: 'dl-cdn.alpinelinux.org',
          },
          {
              label: i18n.global.t('runtime.xtom'),
              value: 'mirrors.xtom.com',
          },
      ]
    : [
          {
              label: i18n.global.t('runtime.ustc'),
              value: 'mirrors.ustc.edu.cn',
          },
          {
              label: i18n.global.t('runtime.netease'),
              value: 'mirrors.163.com',
          },
          {
              label: i18n.global.t('runtime.aliyun'),
              value: 'mirrors.aliyun.com',
          },
          {
              label: i18n.global.t('runtime.tsinghua'),
              value: 'mirrors.tuna.tsinghua.edu.cn',
          },
          {
              label: i18n.global.t('runtime.xtomhk'),
              value: 'mirrors.xtom.com.hk',
          },
          {
              label: i18n.global.t('runtime.xtom'),
              value: 'mirrors.xtom.com',
          },
          {
              label: i18n.global.t('runtime.default'),
              value: 'dl-cdn.alpinelinux.org',
          },
      ];

const initData = (type: string) => ({
    name: '',
    appDetailID: undefined,
    image: '',
    params: {},
    type: type,
    resource: 'appstore',
    rebuild: false,
    source: phpSources[0].value,
});
const extensions = ref();

let runtime = reactive<Runtime.RuntimeCreate>(initData('php'));

const rules = ref<any>({
    name: [Rules.appName],
    resource: [Rules.requiredInput],
    appID: [Rules.requiredSelect],
    version: [Rules.requiredInput, Rules.paramCommon],
    image: [Rules.requiredInput, Rules.imageName],
    source: [Rules.requiredSelect],
});

const em = defineEmits(['close', 'submit']);

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const changeResource = (resource: string) => {
    if (resource === 'local') {
        runtime.appDetailID = undefined;
        runtime.version = '';
        runtime.params = {};
        runtime.image = '';
    } else {
        runtime.version = '';
        searchApp(null);
    }
};

const searchApp = (appId: number) => {
    SearchApp(appReq).then((res) => {
        apps.value = res.data.items || [];
        if (res.data && res.data.items && res.data.items.length > 0) {
            if (appId == null) {
                runtime.appID = res.data.items[0].id;
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
    extensions.value = undefined;
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
    extensions.value = undefined;
    GetAppDetail(runtime.appID, runtime.version, 'runtime')
        .then((res) => {
            runtime.appDetailID = res.data.id;
            runtime.image = res.data.image + ':' + runtime.version;
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
            runtime.version = res.data.versions[0];
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
        if (mode.value == 'create') {
            loading.value = true;
            CreateRuntime(runtime)
                .then((res) => {
                    MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                    handleClose();
                    em('submit', res.data.id);
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
                    em('submit', runtime.id);
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
            appDetailID: data.appDetailID,
            image: data.image,
            params: {},
            type: data.type,
            resource: data.resource,
            appID: data.appID,
            version: data.version,
            rebuild: true,
            source: data.source,
        });
        editParams.value = data.appParams;
        if (mode.value == 'create') {
            searchApp(data.appID);
        } else {
            initParam.value = true;
        }
    } catch (error) {}
};

const listPHPExtensions = async () => {
    try {
        const res = await ListPHPExtensions({
            all: true,
            page: 1,
            pageSize: 100,
        });
        phpExtensions.value = res.data;
    } catch (error) {}
};

const changePHPExtension = () => {
    if (extensions.value == '') {
        return;
    }
    runtime.params['PHP_EXTENSIONS'] = extensions.value.split(',');
};

const acceptParams = async (props: OperateProps) => {
    mode.value = props.mode;
    initParam.value = false;
    if (props.mode === 'create') {
        Object.assign(runtime, initData(props.type));
        searchApp(null);
    } else {
        searchApp(props.appID);
        getRuntime(props.id);
    }
    extensions.value = '';
    listPHPExtensions();
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>
