<template>
    <div>
        <el-alert
            :title="$t('app.hostModeHelper')"
            class="common-prompt"
            :closable="false"
            type="warning"
            v-if="isHostMode"
        />
        <el-alert
            :title="$t('app.memoryRequiredHelper', [computeSizeFromMB(memoryRequired)])"
            class="common-prompt"
            :closable="false"
            type="warning"
            v-if="memoryRequired > 0"
        />

        <el-form
            @validate="(prop, valid) => expandInvalidFormSection(formRef, prop, valid)"
            scroll-to-error
            v-loading="loading"
            @submit.prevent
            ref="formRef"
            label-position="top"
            :model="formData"
            label-width="150px"
            :rules="formRules"
            :validate-on-rule-change="false"
        >
            <component :is="sectioned ? 'details' : 'section'" :open="sectioned" :class="{ 'form-section': sectioned }">
                <summary v-if="sectioned" class="form-section-header">{{ $t('home.baseInfo') }}</summary>
                <div :class="{ 'form-section-body': sectioned }">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model.trim="formData.name"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('app.version')" prop="version">
                        <el-input
                            v-if="clusterVersionLocked"
                            :model-value="formData.version?.replace(/-slave$/, '') || ''"
                            readonly
                        />
                        <el-select
                            v-else
                            v-model="formData.version"
                            :disabled="!!clusterVersionNotice"
                            @change="handleVersionChange"
                        >
                            <el-option
                                v-for="(version, index) in appVersions"
                                :key="index"
                                :label="version"
                                :value="version"
                            ></el-option>
                        </el-select>
                        <el-text v-if="clusterVersionNotice" class="mt-1 w-full" type="warning" role="status">
                            {{ $t(clusterVersionNotice) }}
                        </el-text>
                        <el-text v-else-if="clusterVersionLocked && formData.version" class="mt-1 w-full" type="info">
                            {{ $t('xpack.cluster.versionLocked') }}
                        </el-text>
                    </el-form-item>

                    <Params
                        :key="paramKey"
                        v-if="showParams"
                        v-model:form="formData.params"
                        v-model:params="installParams"
                        v-model:rules="formRules.params"
                        :fallback-descriptions="fallbackDescriptions"
                        :propStart="'params.'"
                    />
                </div>
            </component>

            <slot name="after-params" />

            <component :is="sectioned ? 'details' : 'section'" :open="sectioned" :class="{ 'form-section': sectioned }">
                <summary v-if="sectioned" class="form-section-header">{{ $t('app.advanced') }}</summary>
                <el-form-item v-if="!sectioned" prop="advanced">
                    <el-checkbox v-model="formData.advanced" :label="$t('app.advanced')" size="large" />
                </el-form-item>

                <div v-if="formData.advanced" :class="{ 'form-section-body': sectioned }">
                    <el-form-item :label="$t('app.containerName')" prop="containerName">
                        <el-input v-model.trim="formData.containerName" :placeholder="$t('app.containerNameHelper')" />
                    </el-form-item>

                    <el-form-item prop="allowPort" v-if="!isHostMode">
                        <el-checkbox v-model="formData.allowPort" :label="$t('app.allowPort')" size="large" />
                        <span class="input-help">{{ $t('app.allowPortHelper') }}</span>
                    </el-form-item>

                    <el-form-item :label="$t('app.specifyIP')" v-if="formData.allowPort" prop="specifyIP">
                        <el-input v-model="formData.specifyIP" />
                        <span class="input-help">{{ $t('app.specifyIPHelper') }}</span>
                    </el-form-item>

                    <el-form-item :label="$t('container.restartPolicy')" prop="restartPolicy">
                        <el-select v-model="formData.restartPolicy" class="p-w-300">
                            <el-option :label="$t('container.no')" value="no"></el-option>
                            <el-option :label="$t('container.always')" value="always"></el-option>
                            <el-option :label="$t('container.onFailure')" value="on-failure"></el-option>
                            <el-option :label="$t('container.unlessStopped')" value="unless-stopped"></el-option>
                        </el-select>
                    </el-form-item>

                    <el-form-item
                        :label="$t('container.cpuQuota')"
                        prop="cpuQuota"
                        :rules="checkNumberRange(0, limits.cpu)"
                    >
                        <el-input type="number" class="!w-2/5" v-model.number="formData.cpuQuota" maxlength="5">
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
                        <el-input class="!w-2/5" v-model.number="formData.memoryLimit" maxlength="10">
                            <template #append>
                                <el-select
                                    v-model="formData.memoryUnit"
                                    placeholder="Select"
                                    class="p-w-100"
                                    @change="changeUnit"
                                >
                                    <el-option label="MB" value="M" />
                                    <el-option label="GB" value="G" />
                                </el-select>
                            </template>
                        </el-input>
                        <span class="input-help">
                            {{ $t('container.limitHelper', [limits.memory]) }}{{ formData.memoryUnit }}B
                        </span>
                    </el-form-item>

                    <el-form-item pro="gpuConfig" v-if="gpuSupport && !hideGpuConfig">
                        <el-checkbox v-model="formData.gpuConfig" :label="$t('app.gpuConfig')" size="large" />
                        <span class="input-help">{{ $t('app.gpuConfigHelper') }}</span>
                    </el-form-item>

                    <el-form-item pro="pullImage">
                        <el-checkbox v-model="formData.pullImage" :label="$t('app.pullImage')" size="large" />
                        <span class="input-help">{{ $t('app.pullImageHelper') }}</span>
                    </el-form-item>

                    <PushToNode
                        v-if="isMaster && isXpackOrEE && batchInstallSupport"
                        :push-node="formData.pushNode"
                        :nodes="formData.nodes"
                        type="app"
                        @update:push-node="formData.pushNode = $event"
                        @update:nodes="formData.nodes = $event"
                    />

                    <el-form-item prop="editCompose">
                        <el-checkbox v-model="formData.editCompose" :label="$t('app.editCompose')" size="large" />
                        <span class="input-help">{{ $t('app.editComposeHelper') }}</span>
                    </el-form-item>

                    <div v-if="formData.editCompose">
                        <CodemirrorPro v-model="formData.dockerCompose" mode="yaml" />
                    </div>
                </div>
            </component>
        </el-form>
    </div>
</template>

<script lang="ts" setup>
import { expandInvalidFormSection } from '@/components/form-section';
import { App } from '@/api/interface/app';
import { getAppByKey, getAppDetail, getAppInstalledByID } from '@/api/modules/app';
import { getAppStoreConfig } from '@/api/modules/setting';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { FormInstance, FormRules } from 'element-plus';
import { onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import Params from '../params/index.vue';
import { Container } from '@/api/interface/container';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import { computeSizeFromMB } from '@/utils/size';
import { loadResourceLimit } from '@/api/modules/container';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { isMaster, isOffline, isXpackOrEE } = useGlobalStore();
const { locale, t } = useI18n();

const PushToNode = defineAsyncComponent(async () => {
    const modules = import.meta.glob('@/xpack/views/ssl/index.vue');
    const loader = modules['/src/xpack/views/ssl/index.vue'];
    if (loader) {
        return ((await loader()) as any).default;
    }
    return { template: '<div></div>' };
});

interface ClusterProps {
    key: string;
    node: string;
    masterNode: string;
    appInstallID: number;
    masterVersion: string;
    primaryLabel: string;
    loadReplicaSource?: () => Promise<App.ClusterReplicaSource>;
    role: 'master' | 'slave';
    excludedMasterVersions: readonly string[];
}
interface Props {
    loading?: boolean;
    modelValue?: any;
    batchInstallSupport?: boolean;
    sectioned?: boolean;
    hideGpuConfig?: boolean;
    fallbackDescriptions?: Record<string, string>;
}

const limits = ref<Container.ResourceLimit>({
    cpu: null as number,
    memory: null as number,
});

const props = withDefaults(defineProps<Props>(), {
    loading: false,
    batchInstallSupport: false,
    sectioned: false,
});

interface Emits {
    (e: 'update:modelValue', value: any): void;
}

const emit = defineEmits<Emits>();

const formRef = ref<FormInstance>();
const paramKey = ref(1);
const isHostMode = ref(false);
const memoryRequired = ref(0);
const gpuSupport = ref(false);
const installParams = ref<App.AppParams>();
const oldMemory = ref<number>(0);
const showParams = ref(false);
const currentApp = ref<any>({});
const appVersions = ref<string[]>([]);
const clusterVersionNotice = ref('');
const clusterVersionLocked = ref(false);
let clusterFormRequest = 0;
const operateNode = ref();
const env = ref();
const primaryLabel = ref('');
const replicaSource = ref<App.ClusterReplicaSource>();

const formRules = ref<FormRules>({
    name: [Rules.appName],
    params: [],
    version: [Rules.requiredSelect],
    containerName: [Rules.containerName],
    cpuQuota: [Rules.requiredInput, checkNumberRange(0, 99999)],
    memoryLimit: [Rules.requiredInput, checkNumberRange(0, 9999999999)],
    specifyIP: [Rules.ipv4orV6],
    restartPolicy: [Rules.requiredSelect],
    format: [Rules.requiredInput],
    nodes: [Rules.requiredSelect],
});

const initFormData = () => ({
    appDetailId: 0,
    params: {},
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
    taskID: '',
    gpuConfig: false,
    specifyIP: '',
    restartPolicy: 'always',
    pushNode: false,
    nodes: [],
});

const formData = ref(props.modelValue || initFormData());

watch(
    formData.value,
    (newVal) => {
        emit('update:modelValue', newVal);
    },
    { deep: true },
);

watch(
    () => props.modelValue,
    (newVal) => {
        if (newVal) {
            Object.assign(formData.value, newVal);
        }
    },
    { immediate: true, deep: true },
);

const changeUnit = () => {
    if (formData.value.memoryUnit == 'M') {
        limits.value.memory = oldMemory.value;
    } else {
        limits.value.memory = Number((oldMemory.value / 1024).toFixed(2));
    }
};

const handleVersionChange = async (version: string) => {
    await getVersionDetail(version);
};

const getVersionDetail = async (version: string) => {
    const request = clusterFormRequest;
    formData.value.appDetailId = 0;
    formData.value.dockerCompose = '';
    formData.value.params = {};
    formRules.value.params = [];
    showParams.value = false;
    isHostMode.value = false;
    memoryRequired.value = 0;
    gpuSupport.value = false;

    try {
        const res = await getAppDetail(currentApp.value.id, version, 'app', operateNode.value);
        if (request !== clusterFormRequest || formData.value.version !== version) {
            return;
        }
        formData.value.appDetailId = res.data.id;
        formData.value.dockerCompose = res.data.dockerCompose;
        isHostMode.value = res.data.hostMode;
        if (env.value) {
            installParams.value = addMasterParams(res.data.params);
        } else {
            installParams.value = res.data.params;
        }
        paramKey.value++;
        memoryRequired.value = res.data.memoryRequired;
        gpuSupport.value = res.data.gpuSupport;
        showParams.value = true;
    } catch (error) {}
};

const getInstallAllowPort = async (node?: string) => {
    try {
        const res = await getAppStoreConfig(node);
        return res.data.installAllowPort === 'Enable';
    } catch (error) {
        return false;
    }
};

const initForm = async (appKey: string, allowPort?: boolean) => {
    operateNode.value = undefined;
    env.value = undefined;
    replicaSource.value = undefined;
    formData.value.name = appKey.replace(/^local/, '');
    formData.value.allowPort = allowPort === undefined ? await getInstallAllowPort() : allowPort;
    const res = await getAppByKey(appKey);
    currentApp.value = res.data;
    appVersions.value = currentApp.value.versions;
    if (appVersions.value.length > 0) {
        const defaultVersion = appVersions.value[0];
        formData.value.version = defaultVersion;
        await getVersionDetail(defaultVersion);
    }
    if (isOffline.value) {
        formData.value.pullImage = false;
    }
};

const getMasterAppInstall = async (appInstallID: number, masterNode: string) => {
    try {
        const res = await getAppInstalledByID(appInstallID, masterNode);
        return res.data;
    } catch (error) {}
};

const addMasterParams = (appParams: App.AppParams) => {
    const source = replicaSource.value;
    const language = locale.value.toLowerCase();
    for (const key in appParams.formFields) {
        const field = appParams.formFields[key];
        if (field?.envKey === 'MASTER_ROOT_PASSWORD') {
            field.default = env.value['PANEL_DB_ROOT_PASSWORD']
                ? env.value['PANEL_DB_ROOT_PASSWORD']
                : env.value['PANEL_REDIS_ROOT_PASSWORD'];
        }
        if (field?.envKey === 'PANEL_DB_ROOT_PASSWORD') {
            field.default = env.value['PANEL_DB_ROOT_PASSWORD'];
        }
        if (field?.envKey === 'REPLICATION_USER') {
            field.default = env.value['REPLICATION_USER'];
        }
        if (field?.envKey === 'REPLICATION_PASSWORD') {
            field.default = env.value['REPLICATION_PASSWORD'];
        }
        if (field?.envKey === 'MASTER_HOST') {
            field.default = source?.host ?? '';
            field.disabled = false;
            field.label = {
                ...field.label,
                [language]: t('xpack.cluster.replicaSourceAddress', [primaryLabel.value]),
            };
        }
        if (field?.envKey === 'MASTER_PORT') {
            field.default = source ? String(source.port) : '';
            field.disabled = false;
            field.rule = 'paramPort';
            field.label = {
                ...field.label,
                [language]: t('xpack.cluster.replicaSourcePort', [primaryLabel.value]),
            };
        }
    }
    return appParams;
};

const initClusterForm = async (config: ClusterProps) => {
    const request = ++clusterFormRequest;
    resetForm();
    clusterVersionLocked.value = config.role === 'slave';
    appVersions.value = [];
    env.value = undefined;
    replicaSource.value = undefined;
    primaryLabel.value = config.primaryLabel;
    operateNode.value = config.node;
    formData.value.name = `${config.key}-${config.role}`;
    if (!config.node) return;

    const [appResponse, allowPort, masterInstall, source] = await Promise.all([
        getAppByKey(config.key, config.node).catch(() => undefined),
        getInstallAllowPort(config.node),
        config.appInstallID && config.masterNode
            ? getMasterAppInstall(config.appInstallID, config.masterNode)
            : undefined,
        config.loadReplicaSource?.().catch(() => undefined),
    ]);
    if (request !== clusterFormRequest || !appResponse) return;
    if (config.role === 'slave' && !masterInstall?.env) {
        clusterVersionNotice.value = 'xpack.cluster.masterUnavailable';
        return;
    }

    env.value = masterInstall?.env;
    replicaSource.value = source;
    formData.value.allowPort = allowPort;
    currentApp.value = appResponse.data;
    const versions = appResponse.data.versions ?? [];
    if (config.role === 'slave') {
        const version = (masterInstall?.version || config.masterVersion).replace(/-master$/, '');
        appVersions.value = version ? versions.filter((candidate) => candidate === `${version}-slave`) : [];
    } else {
        const suffix = '-master';
        appVersions.value = versions.filter(
            (version) =>
                version.endsWith(suffix) && !config.excludedMasterVersions.includes(version.slice(0, -suffix.length)),
        );
    }

    const [defaultVersion] = appVersions.value;
    if (!defaultVersion) {
        clusterVersionNotice.value =
            config.role === 'slave'
                ? 'xpack.cluster.noMatchingReplicaVersion'
                : 'xpack.cluster.noAvailableMasterVersion';
        return;
    }

    formData.value.version = defaultVersion;
    await getVersionDetail(defaultVersion);
};

const resetForm = () => {
    if (formRef.value) {
        formRef.value.clearValidate();
        formRef.value.resetFields();
    }
    Object.assign(formData.value, initFormData());
    clusterVersionNotice.value = '';
    clusterVersionLocked.value = false;
    isHostMode.value = false;
    memoryRequired.value = 0;
    gpuSupport.value = false;
    showParams.value = false;
};

const validate = async (): Promise<boolean> => {
    if (!formRef.value) return false;
    try {
        const isValid = await formRef.value.validate();
        return isValid;
    } catch (error) {
        return false;
    }
};

const clearValidate = () => {
    if (formRef.value) {
        formRef.value.clearValidate();
    }
};

const getFormData = () => {
    return { ...formData.value };
};

const setFormData = (data: any) => {
    Object.assign(formData.value, data);
};

const loadLimit = async () => {
    const res = await loadResourceLimit();
    limits.value = res.data;
    limits.value.memory = Number((limits.value.memory / 1024 / 1024).toFixed(2));
    oldMemory.value = limits.value.memory;
};

onMounted(() => {
    loadLimit();
});

onBeforeUnmount(() => {
    clusterFormRequest++;
});

defineExpose({
    formRef,
    formData,
    initForm,
    resetForm,
    validate,
    clearValidate,
    getFormData,
    setFormData,
    isHostMode: () => isHostMode.value,
    initClusterForm,
});
</script>

<style lang="scss" scoped>
@use '@/components/form-section/index.scss';
</style>
