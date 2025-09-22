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
            v-loading="loading"
            @submit.prevent
            ref="formRef"
            label-position="top"
            :model="formData"
            label-width="150px"
            :rules="formRules"
            :validate-on-rule-change="false"
        >
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="formData.name"></el-input>
            </el-form-item>
            <el-form-item :label="$t('app.version')" prop="version">
                <el-select v-model="formData.version" @change="handleVersionChange">
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
                v-if="showParams"
                v-model:form="formData.params"
                v-model:params="installParams"
                v-model:rules="formRules.params"
                :propStart="'params.'"
            />

            <el-form-item prop="advanced">
                <el-checkbox v-model="formData.advanced" :label="$t('app.advanced')" size="large" />
            </el-form-item>

            <div v-if="formData.advanced">
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

                <el-form-item pro="gpuConfig" v-if="gpuSupport">
                    <el-checkbox v-model="formData.gpuConfig" :label="$t('app.gpuConfig')" size="large" />
                    <span class="input-help">{{ $t('app.gpuConfigHelper') }}</span>
                </el-form-item>

                <el-form-item pro="pullImage">
                    <el-checkbox v-model="formData.pullImage" :label="$t('app.pullImage')" size="large" />
                    <span class="input-help">{{ $t('app.pullImageHelper') }}</span>
                </el-form-item>

                <el-form-item prop="editCompose">
                    <el-checkbox v-model="formData.editCompose" :label="$t('app.editCompose')" size="large" />
                    <span class="input-help">{{ $t('app.editComposeHelper') }}</span>
                </el-form-item>

                <div v-if="formData.editCompose">
                    <CodemirrorPro v-model="formData.dockerCompose" mode="yaml" />
                </div>
            </div>
        </el-form>
    </div>
</template>

<script lang="ts" setup name="AppInstallForm">
import { App } from '@/api/interface/app';
import { getAppByKey, getAppDetail, getAppInstalledByID } from '@/api/modules/app';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { FormInstance, FormRules } from 'element-plus';
import { ref, watch } from 'vue';
import Params from '../params/index.vue';
import { Container } from '@/api/interface/container';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import { computeSizeFromMB } from '@/utils/util';
import { loadResourceLimit } from '@/api/modules/container';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { isOffLine } = useGlobalStore();

interface ClusterProps {
    key: string;
    node: string;
    masterNode: string;
    appInstallID: number;
    masterVersion: string;
    masterNodeAddr: string;
    role: string;
}
interface Props {
    loading?: boolean;
    modelValue?: any;
}

const limits = ref<Container.ResourceLimit>({
    cpu: null as number,
    memory: null as number,
});

const props = withDefaults(defineProps<Props>(), {
    loading: false,
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
const operateNode = ref();
const env = ref();
const masterNodeAddr = ref();

const formRules = ref<FormRules>({
    name: [Rules.appName],
    params: [],
    version: [Rules.requiredSelect],
    containerName: [Rules.containerName],
    cpuQuota: [Rules.requiredInput, checkNumberRange(0, 99999)],
    memoryLimit: [Rules.requiredInput, checkNumberRange(0, 9999999999)],
    specifyIP: [Rules.ipv4orV6],
    restartPolicy: [Rules.requiredSelect],
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
    try {
        const res = await getAppDetail(currentApp.value.id, version, 'app', operateNode.value);
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

const initForm = async (appKey: string) => {
    formData.value.name = appKey.replace(/^local/, '');
    const res = await getAppByKey(appKey);
    currentApp.value = res.data;
    appVersions.value = currentApp.value.versions;
    if (appVersions.value.length > 0) {
        const defaultVersion = appVersions.value[0];
        formData.value.version = defaultVersion;
        getVersionDetail(defaultVersion);
    }
    if (isOffLine.value) {
        formData.value.pullImage = false;
    }
};

const getMasterAppInstall = async (appInstallID: number, masterNode: string) => {
    try {
        const res = await getAppInstalledByID(appInstallID, masterNode);
        env.value = res.data.env;
    } catch (error) {}
};

const addMasterParams = (appParams: App.AppParams) => {
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
        if (field?.envKey === 'MASTER_PORT') {
            field.default = env.value['PANEL_APP_PORT_HTTP'];
        }
        if (field?.envKey === 'MASTER_HOST' && masterNodeAddr.value != '127.0.0.1') {
            field.default = masterNodeAddr.value;
        }
    }
    return appParams;
};

const initClusterForm = async (props: ClusterProps) => {
    if (props.appInstallID && props.masterNode) {
        getMasterAppInstall(props.appInstallID, props.masterNode);
    }
    masterNodeAddr.value = props.masterNodeAddr;
    operateNode.value = props.node;
    const res = await getAppByKey(props.key, props.node);
    currentApp.value = res.data;
    appVersions.value = currentApp.value.versions;
    if (appVersions.value.length > 0) {
        appVersions.value = appVersions.value.filter((v: string) => {
            return v.includes(props.role) && v.includes(props.masterVersion);
        });
        const defaultVersion = appVersions.value[0];
        formData.value.version = defaultVersion;
        getVersionDetail(defaultVersion);
    }
    formData.value.name = props.key + '-' + props.role;
};

const resetForm = () => {
    if (formRef.value) {
        formRef.value.clearValidate();
        formRef.value.resetFields();
    }
    Object.assign(formData.value, initFormData());
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
