<template>
    <DrawerPro
        v-model="open"
        :header="$t('runtime.' + mode)"
        size="large"
        :resource="mode === 'edit' ? runtime.name : ''"
        :back="handleClose"
    >
        <el-form
            ref="runtimeForm"
            label-position="top"
            :model="runtime"
            label-width="125px"
            :rules="rules"
            :validate-on-rule-change="false"
            v-loading="loading"
        >
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input :disabled="mode === 'edit'" v-model="runtime.name"></el-input>
            </el-form-item>
            <el-form-item :label="$t('app.source')" prop="resource">
                <el-radio-group
                    :disabled="mode === 'edit'"
                    v-model="runtime.resource"
                    @change="changeResource(runtime.resource)"
                >
                    <el-radio :value="'appstore'">
                        {{ $t('menu.apps') }}
                    </el-radio>
                    <el-radio :value="'local'">
                        {{ $t('commons.table.local') }}
                    </el-radio>
                </el-radio-group>
            </el-form-item>
            <div v-if="runtime.resource === 'appstore'">
                <el-form-item :label="$t('app.app')" prop="appID">
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
                    <el-form-item
                        :label="getLabel(formFields['PHP_VERSION'])"
                        :rules="rules.params.PHP_VERSION"
                        v-if="formFields['PHP_VERSION']"
                    >
                        <el-select
                            v-model="runtime.params['PHP_VERSION']"
                            filterable
                            default-first-option
                            @change="changePHPVersion(runtime.params['PHP_VERSION'])"
                        >
                            <el-option
                                v-for="service in formFields['PHP_VERSION'].values"
                                :key="service.label"
                                :value="service.value"
                                :label="service.label"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('container.image')" prop="image">
                        <el-input v-model="runtime.image"></el-input>
                    </el-form-item>
                    <el-form-item
                        :label="getLabel(formFields['CONTAINER_PACKAGE_URL'])"
                        :rules="rules.params.CONTAINER_PACKAGE_URL"
                        v-if="formFields['CONTAINER_PACKAGE_URL']"
                    >
                        <el-select v-model="runtime.source" filterable default-first-option>
                            <el-option
                                v-for="source in phpSources"
                                :key="source.label"
                                :value="source.value"
                                :label="source.label + ' [' + source.value + ']'"
                            ></el-option>
                        </el-select>
                    </el-form-item>

                    <el-form-item
                        :label="getLabel(formFields['PANEL_APP_PORT_HTTP'])"
                        prop="params.PANEL_APP_PORT_HTTP"
                        v-if="formFields['PANEL_APP_PORT_HTTP']"
                    >
                        <el-input
                            v-model.number="runtime.params['PANEL_APP_PORT_HTTP']"
                            maxlength="15"
                            :disabled="mode == 'edit'"
                        ></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('app.containerName')" prop="params.CONTAINER_NAME">
                        <el-input v-model.trim="runtime.params['CONTAINER_NAME']"></el-input>
                    </el-form-item>
                    <el-form-item>
                        <el-alert type="warning" :closable="false">
                            <template #default>
                                <div>{{ $t('runtime.buildHelper') }}</div>
                                <div>
                                    <span>{{ $t('runtime.extendHelper') }}</span>
                                    <el-link
                                        target="_blank"
                                        type="primary"
                                        :href="globalStore.docsUrl + '/user_manual/websites/php/#php_1'"
                                    >
                                        {{ $t('php.toExtensionsList') }}
                                    </el-link>
                                </div>
                            </template>
                        </el-alert>
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
                    <el-form-item :label="getLabel(formFields['PHP_EXTENSIONS'])" v-if="formFields['PHP_EXTENSIONS']">
                        <el-select v-model="runtime.params['PHP_EXTENSIONS']" multiple allowCreate filterable>
                            <el-option
                                v-for="service in formFields['PHP_EXTENSIONS'].values"
                                :key="service.label"
                                :value="service.value"
                                :label="service.label"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                </div>
            </div>
            <div v-else>
                <el-form-item>
                    <el-alert :title="$t('runtime.localHelper')" type="info" :closable="false" />
                </el-form-item>
                <el-form-item :label="$t('app.version')" prop="version">
                    <el-input v-model="runtime.version" :placeholder="$t('runtime.versionHelper')"></el-input>
                </el-form-item>
            </div>
        </el-form>
        <template #footer>
            <span>
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(runtimeForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { Runtime } from '@/api/interface/runtime';
import { getAppByKey, getAppDetail, searchApp } from '@/api/modules/app';
import { CreateRuntime, GetRuntime, ListPHPExtensions, UpdateRuntime } from '@/api/modules/runtime';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

interface OperateRrops {
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
              value: 'https://dl-cdn.alpinelinux.org',
          },
          {
              label: i18n.global.t('runtime.xtom'),
              value: 'https://mirrors.xtom.com',
          },
      ]
    : [
          {
              label: i18n.global.t('runtime.ustc'),
              value: 'https://mirrors.ustc.edu.cn',
          },
          {
              label: i18n.global.t('runtime.netease'),
              value: 'https://mirrors.163.com',
          },
          {
              label: i18n.global.t('runtime.aliyun'),
              value: 'https://mirrors.aliyun.com',
          },
          {
              label: i18n.global.t('runtime.tsinghua'),
              value: 'https://mirrors.tuna.tsinghua.edu.cn',
          },
          {
              label: i18n.global.t('runtime.xtomhk'),
              value: 'https://mirrors.xtom.com.hk',
          },
          {
              label: i18n.global.t('runtime.xtom'),
              value: 'https://mirrors.xtom.com',
          },
          {
              label: i18n.global.t('commons.table.default'),
              value: 'https://dl-cdn.alpinelinux.org',
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
const formFields = ref();

let runtime = reactive<Runtime.RuntimeCreate>(initData('php'));

const rules = ref<any>({
    name: [Rules.appName],
    resource: [Rules.requiredInput],
    appID: [Rules.requiredSelect],
    version: [Rules.requiredInput, Rules.paramCommon],
    image: [Rules.requiredInput, Rules.imageName],
    source: [Rules.requiredSelect],
    params: {
        PANEL_APP_PORT_HTTP: [Rules.requiredInput, Rules.port],
        PHP_VERSION: [Rules.requiredSelect],
        CONTAINER_PACKAGE_URL: [Rules.requiredSelect],
        CONTAINER_NAME: [Rules.containerName, Rules.requiredInput],
    },
});

const getLabel = (row: App.FromField): string => {
    const language = localStorage.getItem('lang') || 'zh';
    let lang = language == 'tw' ? 'zh-Hant' : language;
    if (row.label && row.label[lang] != '') {
        return row.label[lang];
    }
    if (language == 'zh' || language == 'tw') {
        return row.labelZh;
    } else {
        return row.labelEn;
    }
};

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
        searchAppList(null);
    }
};

const searchAppList = (appId: number) => {
    searchApp(appReq).then((res) => {
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

const changePHPVersion = (version: string) => {
    runtime.image = 'php:' + version;
};

const changeVersion = () => {
    loading.value = true;
    initParam.value = false;
    extensions.value = undefined;
    getAppDetail(runtime.appID, runtime.version, 'runtime')
        .then((res) => {
            runtime.appDetailID = res.data.id;
            runtime.image = res.data.image + ':' + runtime.version;
            appParams.value = res.data.params;
            const fileds = res.data.params.formFields;
            formFields.value = {};
            for (const index in fileds) {
                formFields.value[fileds[index]['envKey']] = fileds[index];
                runtime.params[fileds[index]['envKey']] = fileds[index]['default'];
                if (fileds[index]['envKey'] == 'PHP_VERSION') {
                    runtime.image = 'php:' + fileds[index]['default'];
                }
            }
            initParam.value = true;
        })
        .finally(() => {
            loading.value = false;
        });
};

const getApp = (appkey: string, mode: string) => {
    getAppByKey(appkey).then((res) => {
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
            params: data.params,
            type: data.type,
            resource: data.resource,
            appID: data.appID,
            version: data.version,
            rebuild: true,
            source: data.source,
        });

        const fileds = data.appParams;
        const forms = {};
        for (const index in fileds) {
            forms[fileds[index].key] = fileds[index];
        }
        formFields.value = forms;
        if (data.params['PHP_EXTENSIONS'] != '') {
            runtime.params['PHP_EXTENSIONS'] = runtime.params['PHP_EXTENSIONS'].split(',');
        }
        initParam.value = true;
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

const acceptParams = async (props: OperateRrops) => {
    mode.value = props.mode;
    initParam.value = false;
    if (props.mode === 'create') {
        Object.assign(runtime, initData(props.type));
        searchAppList(null);
    } else {
        searchAppList(props.appID);
        getRuntime(props.id);
    }
    extensions.value = '';
    open.value = true;
    listPHPExtensions();
};

watch(
    () => runtime.name,
    (newVal) => {
        if (newVal && mode.value == 'create') {
            runtime.params['CONTAINER_NAME'] = newVal;
        }
    },
    { deep: true },
);

defineExpose({
    acceptParams,
});
</script>
