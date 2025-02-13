<template>
    <el-drawer
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        v-model="open"
        :title="$t('app.install')"
        size="50%"
        :before-close="handleClose"
    >
        <template #header>
            <Header :header="$t('app.install')" :back="handleClose"></Header>
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-alert
                    :title="$t('app.appInstallWarn')"
                    class="common-prompt"
                    :closable="false"
                    type="warning"
                    v-if="!isHostMode"
                />
                <el-alert
                    :title="$t('app.hostModeHelper')"
                    class="common-prompt"
                    :closable="false"
                    type="warning"
                    v-else
                />
                <el-form
                    @submit.prevent
                    ref="paramForm"
                    label-position="top"
                    :model="req"
                    label-width="150px"
                    :rules="rules"
                    :validate-on-rule-change="false"
                >
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model.trim="req.name"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('app.version')" prop="version">
                        <el-select v-model="req.version" @change="getAppDetail(req.version)">
                            <el-option
                                v-for="(version, index) in appVersions"
                                :key="index"
                                :label="version"
                                :value="version"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <Params
                        :key="paramKey"
                        v-if="open"
                        v-model:form="req.params"
                        v-model:params="installData.params"
                        v-model:rules="rules.params"
                        :propStart="'params.'"
                    ></Params>
                    <el-form-item prop="advanced">
                        <el-checkbox v-model="req.advanced" :label="$t('app.advanced')" size="large" />
                    </el-form-item>
                    <div v-if="req.advanced">
                        <el-form-item :label="$t('app.containerName')" prop="containerName">
                            <el-input
                                v-model.trim="req.containerName"
                                :placeholder="$t('app.containerNameHelper')"
                            ></el-input>
                        </el-form-item>
                        <el-form-item prop="allowPort" v-if="!isHostMode">
                            <el-checkbox v-model="req.allowPort" :label="$t('app.allowPort')" size="large" />
                            <span class="input-help">{{ $t('app.allowPortHelper') }}</span>
                        </el-form-item>
                        <el-form-item
                            :label="$t('container.cpuQuota')"
                            prop="cpuQuota"
                            :rules="checkNumberRange(0, limits.cpu)"
                        >
                            <el-input type="number" style="width: 40%" v-model.number="req.cpuQuota" maxlength="5">
                                <template #append>{{ $t('app.cpuCore') }}</template>
                            </el-input>
                            <span class="input-help">
                                {{ $t('container.limitHelper', [limits.cpu]) }}{{ $t('commons.units.core') }}
                            </span>
                        </el-form-item>
                        <el-form-item
                            :label="$t('container.memoryLimit')"
                            prop="memoryLimit"
                            :rules="checkNumberRange(0, limits.memory)"
                        >
                            <el-input style="width: 40%" v-model.number="req.memoryLimit" maxlength="10">
                                <template #append>
                                    <el-select
                                        v-model="req.memoryUnit"
                                        placeholder="Select"
                                        style="width: 85px"
                                        @change="changeUnit"
                                    >
                                        <el-option label="MB" value="M" />
                                        <el-option label="GB" value="G" />
                                    </el-select>
                                </template>
                            </el-input>
                            <span class="input-help">
                                {{ $t('container.limitHelper', [limits.memory]) }}{{ req.memoryUnit }}B
                            </span>
                        </el-form-item>
                        <el-form-item prop="editCompose">
                            <el-checkbox v-model="req.editCompose" :label="$t('app.editCompose')" size="large" />
                            <span class="input-help">{{ $t('app.editComposeHelper') }}</span>
                        </el-form-item>
                        <el-form-item pro="gpuConfig" v-if="gpuSupport">
                            <el-checkbox v-model="req.gpuConfig" :label="$t('app.gpuConfig')" size="large" />
                            <span class="input-help">{{ $t('app.gpuConfigHelper') }}</span>
                        </el-form-item>
                        <el-form-item pro="pullImage">
                            <el-checkbox v-model="req.pullImage" :label="$t('app.pullImage')" size="large" />
                            <span class="input-help">{{ $t('app.pullImageHelper') }}</span>
                        </el-form-item>
                        <div v-if="req.editCompose">
                            <codemirror
                                :autofocus="true"
                                placeholder=""
                                :indent-with-tab="true"
                                :tabSize="4"
                                style="height: 400px"
                                :lineWrapping="true"
                                :matchBrackets="true"
                                theme="cobalt"
                                :styleActiveLine="true"
                                :extensions="extensions"
                                v-model="req.dockerCompose"
                            />
                        </div>
                    </div>
                </el-form>
            </el-col>
        </el-row>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(paramForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup name="appInstall">
import { App } from '@/api/interface/app';
import { GetApp, GetAppDetail, InstallApp } from '@/api/modules/app';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { FormInstance, FormRules } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import Params from '../params/index.vue';
import Header from '@/components/drawer-header/index.vue';
import { Codemirror } from 'vue-codemirror';
import { yaml } from '@codemirror/lang-yaml';
import { oneDark } from '@codemirror/theme-one-dark';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
import { Container } from '@/api/interface/container';
import { loadResourceLimit } from '@/api/modules/container';

const extensions = [yaml(), oneDark];
const router = useRouter();

interface InstallProps {
    params?: App.AppParams;
    app: any;
}

const installData = ref<InstallProps>({
    app: {},
});
const open = ref(false);
const rules = ref<FormRules>({
    name: [Rules.appName],
    params: [],
    version: [Rules.requiredSelect],
    containerName: [Rules.containerName],
    cpuQuota: [Rules.requiredInput, checkNumberRange(0, 99999)],
    memoryLimit: [Rules.requiredInput, checkNumberRange(0, 9999999999)],
});
const loading = ref(false);
const paramForm = ref<FormInstance>();
const form = ref<{ [key: string]: any }>({});
const initData = () => ({
    appDetailId: 0,
    params: form.value,
    name: '',
    advanced: true,
    cpuQuota: 0,
    memoryLimit: 0,
    memoryUnit: 'M',
    containerName: '',
    allowPort: false,
    editCompose: false,
    dockerCompose: '',
    version: '',
    appID: '',
    pullImage: true,
    gpuConfig: false,
});
const req = reactive(initData());
const limits = ref<Container.ResourceLimit>({
    cpu: null as number,
    memory: null as number,
});
const oldMemory = ref<number>(0);
const appVersions = ref<string[]>([]);
const handleClose = () => {
    open.value = false;
    resetForm();
    if (router.currentRoute.value.query.install) {
        router.push({ name: 'AppAll' });
    }
};
const paramKey = ref(1);
const isHostMode = ref(false);
const gpuSupport = ref(false);

const changeUnit = () => {
    if (req.memoryUnit == 'M') {
        limits.value.memory = oldMemory.value;
    } else {
        limits.value.memory = Number((oldMemory.value / 1024).toFixed(2));
    }
};

const resetForm = () => {
    if (paramForm.value) {
        paramForm.value.clearValidate();
        paramForm.value.resetFields();
    }
    isHostMode.value = false;
    Object.assign(req, initData());
};

const acceptParams = async (props: InstallProps) => {
    resetForm();
    if (props.app.versions != undefined) {
        installData.value = props;
    } else {
        const res = await GetApp(props.app.key);
        installData.value.app = res.data;
    }

    const app = installData.value.app;
    appVersions.value = app.versions;
    if (appVersions.value.length > 0) {
        req.version = appVersions.value[0];
        getAppDetail(appVersions.value[0]);
    }

    req.name = props.app.key;
    open.value = true;
};

const getAppDetail = async (version: string) => {
    loading.value = true;
    try {
        const res = await GetAppDetail(installData.value.app.id, version, 'app');
        req.appDetailId = res.data.id;
        req.dockerCompose = res.data.dockerCompose;
        isHostMode.value = res.data.hostMode;
        installData.value.params = res.data.params;
        paramKey.value++;
        gpuSupport.value = res.data.gpuSupport;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        if (req.editCompose && req.dockerCompose == '') {
            MsgError(i18n.global.t('app.composeNullErr'));
            return;
        }
        if (req.cpuQuota < 0) {
            req.cpuQuota = 0;
        }
        if (req.memoryLimit < 0) {
            req.memoryLimit = 0;
        }
        if (!isHostMode.value && !req.allowPort) {
            ElMessageBox.confirm(i18n.global.t('app.installWarn'), i18n.global.t('app.checkTitle'), {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            }).then(async () => {
                install();
            });
        } else {
            install();
        }
    });
};

const install = () => {
    loading.value = true;
    InstallApp(req)
        .then(() => {
            handleClose();
            router.push({ path: '/apps/installed' });
        })
        .finally(() => {
            loading.value = false;
        });
};

const loadLimit = async () => {
    const res = await loadResourceLimit();
    limits.value = res.data;
    limits.value.memory = Number((limits.value.memory / 1024 / 1024).toFixed(2));
    oldMemory.value = limits.value.memory;
};

defineExpose({
    acceptParams,
});

onMounted(() => {
    loadLimit();
});
</script>
