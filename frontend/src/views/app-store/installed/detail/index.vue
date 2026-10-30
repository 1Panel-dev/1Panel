<template>
    <DrawerPro
        v-model="open"
        :header="$t('app.param')"
        @close="handleClose"
        :size="isMobile ? 'full' : appKey === 'ollama' ? 'large' : 'normal'"
    >
        <template #buttons>
            <el-button type="primary" plain @click="editParam" :disabled="loading" :icon="edit ? '' : 'Edit'">
                {{ edit ? $t('app.detail') : $t('commons.button.edit') }}
            </el-button>
        </template>
        <div v-if="!edit">
            <el-descriptions border :column="1">
                <el-descriptions-item :label="$t('app.webUI')">
                    <span v-if="!openConfig">
                        {{ appConfigUpdate.webUI }}
                        <el-button size="small" @click="openConfig = true">
                            {{ $t('container.custom') }}
                        </el-button>
                    </span>
                    <span class="flex" v-else>
                        <el-input v-model="webUI.domain" :placeholder="$t('app.webUIPlaceholder')">
                            <template #prepend>
                                <el-select v-model="webUI.protocol" class="pre-select">
                                    <el-option label="http" value="http://" />
                                    <el-option label="https" value="https://" />
                                </el-select>
                            </template>
                        </el-input>
                        <el-button type="primary" @click="updateAppConfig" class="ml-2">
                            {{ $t('commons.button.confirm') }}
                        </el-button>
                    </span>
                </el-descriptions-item>
                <el-descriptions-item
                    label-class-name="system-label"
                    v-for="(param, key) in params"
                    :label="getLabel(param)"
                    :key="key"
                >
                    <span v-if="param.type === 'password' && !param.showPassword">********</span>
                    <span v-else class="break-all whitespace-normal">
                        {{ param.showValue && param.showValue != '' ? param.showValue : param.value }}
                    </span>
                    <template v-if="param.type === 'password'">
                        <el-button
                            v-if="!param.showPassword"
                            link
                            @click="param.showPassword = true"
                            icon="View"
                            class="-mt-1 ml-3"
                        ></el-button>
                        <el-button
                            v-if="param.showPassword"
                            link
                            @click="param.showPassword = false"
                            icon="Hide"
                            class="-mt-1 ml-3"
                        ></el-button>
                    </template>
                    <CopyButton class="-mt-1" v-if="showCopyButton(param.key)" :content="param.value" />
                </el-descriptions-item>
            </el-descriptions>
        </div>
        <div v-else v-loading="loading">
            <el-alert :title="$t('app.updateHelper')" type="warning" :closable="false" class="common-prompt" />
            <el-form
                @validate="(prop, valid) => expandInvalidFormSection(paramForm, prop, valid)"
                scroll-to-error
                @submit.prevent
                ref="paramForm"
                :model="paramModel"
                label-position="top"
                :rules="rules"
            >
                <component
                    :is="appKey === 'ollama' ? 'details' : 'section'"
                    :open="appKey === 'ollama'"
                    :class="{ 'form-section': appKey === 'ollama' }"
                >
                    <summary v-if="appKey === 'ollama'" class="form-section-header">{{ $t('home.baseInfo') }}</summary>
                    <div :class="{ 'form-section-body': appKey === 'ollama' }">
                        <div v-for="(p, index) in params" :key="index">
                            <el-form-item :prop="'params.' + p.key" :label="getLabel(p)">
                                <el-input
                                    v-if="p.type == 'number'"
                                    type="number"
                                    v-model.number="paramModel.params[p.key]"
                                    :disabled="!p.edit"
                                ></el-input>
                                <el-input
                                    v-else-if="p.type == 'password'"
                                    :type="p.showPassword ? 'text' : 'password'"
                                    v-model="paramModel.params[p.key]"
                                    :disabled="!p.edit"
                                >
                                    <template #append>
                                        <el-button
                                            :icon="p.showPassword ? 'Hide' : 'View'"
                                            @click="p.showPassword = !p.showPassword"
                                        />
                                    </template>
                                </el-input>
                                <el-select
                                    v-model="paramModel.params[p.key]"
                                    v-else-if="p.type == 'select'"
                                    :multiple="p.multiple"
                                >
                                    <el-option
                                        v-for="value in p.values"
                                        :key="value.label"
                                        :value="value.value"
                                        :label="value.label"
                                        :disabled="!p.edit"
                                    ></el-option>
                                </el-select>
                                <el-input v-else v-model.trim="paramModel.params[p.key]" :disabled="!p.edit"></el-input>
                            </el-form-item>
                        </div>
                    </div>
                </component>
                <GPUConfig v-if="showModelGPU" ref="gpuRef" v-model="paramModel.gpu" />
                <details v-if="appKey === 'ollama'" ref="syncSectionRef" open class="form-section">
                    <summary class="form-section-header">
                        <span>{{ $t('aiTools.model.syncModelAccount') }}</span>
                        <el-switch
                            @click.stop
                            @keydown.stop
                            v-model="paramModel.syncModelAccount"
                            :aria-label="$t('aiTools.model.syncModelAccount')"
                            @change="(enabled) => enabled && syncSectionRef && (syncSectionRef.open = true)"
                        />
                    </summary>
                    <div v-show="paramModel.syncModelAccount" class="form-section-body">
                        <AccountAddress
                            ref="accountRef"
                            show-sync
                            hide-sync-toggle
                            :form="paramModel"
                            :host-port="Number(paramModel.params.PANEL_APP_PORT_HTTP) || paramModel.httpPort"
                            container-port="11434"
                            :host-mode="paramModel.isHostMode"
                        />
                    </div>
                </details>
                <component
                    :is="appKey === 'ollama' ? 'details' : 'section'"
                    :open="appKey === 'ollama'"
                    :class="{ 'form-section': appKey === 'ollama' }"
                >
                    <summary v-if="appKey === 'ollama'" class="form-section-header">{{ $t('app.advanced') }}</summary>
                    <el-form-item v-if="appKey !== 'ollama'" prop="advanced">
                        <el-checkbox v-model="paramModel.advanced" :label="$t('app.advanced')" size="large" />
                    </el-form-item>
                    <div
                        v-if="appKey === 'ollama' || paramModel.advanced"
                        :class="{ 'form-section-body': appKey === 'ollama' }"
                    >
                        <el-form-item :label="$t('app.containerName')" prop="containerName">
                            <el-input
                                v-model.trim="paramModel.containerName"
                                :placeholder="$t('app.containerNameHelper')"
                            ></el-input>
                        </el-form-item>
                        <el-form-item prop="allowPort" v-if="!paramModel.isHostMode">
                            <el-checkbox
                                v-model="paramModel.allowPort"
                                :label="$t('app.allowPort')"
                                size="large"
                                @change="changeAllowPort"
                            />
                            <span class="input-help">{{ $t('app.allowPortHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('app.specifyIP')" v-if="paramModel.allowPort" prop="specifyIP">
                            <el-input v-model="paramModel.specifyIP"></el-input>
                            <span class="input-help">{{ $t('app.specifyIPHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('container.restartPolicy')" prop="restartPolicy">
                            <el-select v-model="paramModel.restartPolicy" class="p-w-300">
                                <el-option :label="$t('container.no')" value="no"></el-option>
                                <el-option :label="$t('container.always')" value="always"></el-option>
                                <el-option :label="$t('container.onFailure')" value="on-failure"></el-option>
                                <el-option :label="$t('container.unlessStopped')" value="unless-stopped"></el-option>
                            </el-select>
                        </el-form-item>
                        <el-form-item :label="$t('container.cpuQuota')" prop="cpuQuota">
                            <el-input type="number" class="!w-2/5" v-model.number="paramModel.cpuQuota" maxlength="5">
                                <template #append>{{ $t('app.cpuCore') }}</template>
                            </el-input>
                            <span class="input-help">
                                {{ $t('container.limitHelper', [limits.cpu]) }}{{ $t('commons.units.core') }}
                            </span>
                        </el-form-item>
                        <el-form-item :label="$t('container.memoryLimit')" prop="memoryLimit">
                            <el-input class="!w-2/5" v-model.number="paramModel.memoryLimit" maxlength="10">
                                <template #append>
                                    <el-select
                                        v-model="paramModel.memoryUnit"
                                        placeholder="Select"
                                        class="!w-[85px]"
                                        @change="changeUnit"
                                    >
                                        <el-option label="MB" value="M" />
                                        <el-option label="GB" value="G" />
                                    </el-select>
                                </template>
                            </el-input>
                            <span class="input-help">
                                {{ $t('container.limitHelper', [limits.memory]) }}{{ paramModel.memoryUnit }}B
                            </span>
                        </el-form-item>

                        <el-form-item prop="editCompose">
                            <el-checkbox v-model="paramModel.editCompose" :label="$t('app.editCompose')" size="large" />
                            <span class="input-help">{{ $t('app.editComposeHelper') }}</span>
                        </el-form-item>
                        <div v-if="paramModel.editCompose">
                            <el-button @click="openDiff()" type="primary" class="!mb-2">
                                {{ $t('app.showDiff') }}
                            </el-button>
                            <CodemirrorPro v-model="paramModel.dockerCompose" mode="yaml"></CodemirrorPro>
                        </div>
                    </div>
                </component>
            </el-form>
        </div>
        <template #footer v-if="edit">
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" :disabled="loading || gpuRef?.loading" @click="submit(paramForm)">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
        <Diff ref="composeDiffRef" @confirm="getNewCompose" />
    </DrawerPro>
</template>
<script lang="ts" setup>
import { expandInvalidFormSection } from '@/views/ai/model/components/form-section';
import AccountAddress from '@/views/ai/model/components/account-address/index.vue';
import GPUConfig from '@/views/ai/model/components/gpu-config/index.vue';
import Diff from '@/views/app-store/installed/upgrade/diff/index.vue';

import { App } from '@/api/interface/app';
import { getAppInstallParams, updateAppInstallParams, updateInstallConfig } from '@/api/modules/app';
import { computed, nextTick, onBeforeUnmount, reactive, ref } from 'vue';
import { FormInstance } from 'element-plus';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { MsgError, MsgSuccess } from '@/utils/message';
import { getLabel } from '@/utils/app-store';
import { splitHttp, checkIpV4V6, checkDomain } from '@/utils/validate';
import i18n from '@/lang';
import { loadResourceLimit } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { useGlobalStore } from '@/composables/useGlobalStore';

interface ParamProps {
    id: number;
    app?: any;
    edit?: boolean;
}
const paramData = ref<ParamProps>({
    id: 0,
    app: {},
});

interface EditForm extends App.InstallParams {
    default: any;
}

const props = defineProps({ modelGPU: Boolean });
const emit = defineEmits(['close']);
const { currentNode, isMobile } = useGlobalStore();
const open = ref(false);
const loading = ref(false);
const params = ref<EditForm[]>();
const edit = ref(false);
const paramForm = ref<FormInstance>();
const paramModel = reactive<any>({
    params: {},
});
const advancedFields = [
    'memoryLimit',
    'cpuQuota',
    'memoryUnit',
    'allowPort',
    'specifyIP',
    'containerName',
    'restartPolicy',
];
let initialAdvancedConfig: Record<string, unknown> = {};
const rules = reactive({
    params: {},
    cpuQuota: [Rules.requiredInput, checkNumberRange(0, 999)],
    memoryLimit: [Rules.requiredInput, checkNumberRange(0, 9999999999)],
    containerName: [Rules.containerName],
    restartPolicy: [Rules.requiredSelect],
});
const appType = ref('');
const appKey = ref('');
const showModelGPU = computed(() => props.modelGPU && appKey.value === 'ollama');
const accountRef = ref<InstanceType<typeof AccountAddress>>();
const syncSectionRef = ref<HTMLDetailsElement>();
const gpuRef = ref<InstanceType<typeof GPUConfig>>();
const appConfigUpdate = ref<App.AppConfigUpdate>({
    installID: 0,
    webUI: '',
});
const openConfig = ref(false);
const webUI = reactive({
    protocol: 'http://',
    domain: '',
});
const limits = ref<Container.ResourceLimit>({
    cpu: null as number,
    memory: null as number,
});
const oldMemory = ref<number>(0);
const composeDiffRef = ref();
const rawCompose = ref('');

const getNewCompose = (compose: string) => {
    if (compose != '') {
        paramModel.dockerCompose = compose;
    }
};

const openDiff = () => {
    composeDiffRef.value.acceptParams(rawCompose.value, paramModel.dockerCompose);
};

function checkWebUI() {
    if (webUI.domain !== '') {
        let domain = webUI.domain;
        let port = null;

        if (domain.includes('/')) {
            domain = domain.split('/')[0];
        }

        if (domain.includes(':')) {
            const parts = domain.split(':');
            domain = parts[0];
            port = parts[1];

            if (!checkPort(port)) {
                return false;
            }
        }

        if (checkIpV4V6(domain) && checkDomain(domain)) {
            return false;
        }
    }
    return true;
}

function checkPort(port: string) {
    const portNum = parseInt(port, 10);
    return !isNaN(portNum) && portNum > 0 && portNum <= 65535;
}

const acceptParams = async (props: ParamProps) => {
    params.value = [];
    paramData.value.id = props.id;
    paramModel.params = {};
    edit.value = false;
    rules.params = {};
    if (!(await get())) return;
    open.value = true;
    openConfig.value = false;
    if (props.edit) await editParam();
};

const handleClose = () => {
    emit('close');
    open.value = false;
};
const editParam = async () => {
    params.value.forEach((param: EditForm) => {
        paramModel.params[param.key] = param.value;
    });

    edit.value = !edit.value;
    loadLimit();
    if (edit.value && showModelGPU.value) {
        await nextTick();
        await gpuRef.value?.init(false);
    }
};

const changeAllowPort = () => {
    if (paramModel.allowPort) {
        paramModel.specifyIP = '';
    }
};

const get = async () => {
    try {
        loading.value = true;
        const res = await getAppInstallParams(Number(paramData.value.id));
        rawCompose.value = res.data.rawCompose;
        const configParams = res.data.params || [];
        if (configParams && configParams.length > 0) {
            configParams.forEach((d) => {
                let value = d.value;
                if (d.type === 'number') {
                    value = Number(value);
                }
                params.value.push({
                    default: value,
                    labelEn: d.labelEn,
                    labelZh: d.labelZh,
                    rule: d.rule,
                    value: value,
                    edit: d.edit,
                    key: d.key,
                    type: d.type,
                    values: d.values,
                    showValue: d.showValue,
                    multiple: d.multiple,
                    label: d.label,
                    required: d.required,
                    showPassword: false,
                });
                if (d.required) {
                    rules.params[d.key] = [Rules.requiredInput];
                    if (d.rule) {
                        rules.params[d.key].push(Rules[d.rule]);
                    }
                } else {
                    rules.params[d.key] = [];
                }
            });
        }
        paramModel.memoryLimit = res.data.memoryLimit;
        paramModel.cpuQuota = res.data.cpuQuota;
        paramModel.memoryUnit = res.data.memoryUnit !== '' ? res.data.memoryUnit : 'MB';
        paramModel.allowPort = res.data.allowPort;
        paramModel.containerName = res.data.containerName;
        paramModel.advanced = false;
        paramModel.editCompose = false;
        paramModel.dockerCompose = res.data.dockerCompose;
        paramModel.isHostMode = res.data.hostMode;
        paramModel.specifyIP = res.data.specifyIP;
        paramModel.restartPolicy = res.data.restartPolicy || 'no';
        if (paramModel.restartPolicy === 'on-failure:5') {
            paramModel.restartPolicy = 'on-failure';
        }
        appConfigUpdate.value.webUI = res.data.webUI;
        if (res.data.webUI != '') {
            const httpConfig = splitHttp(res.data.webUI);
            webUI.domain = httpConfig.url;
            webUI.protocol = httpConfig.proto + '://';
        }
        appType.value = res.data.type;
        appKey.value = res.data.appKey;
        if (showModelGPU.value) {
            paramModel.gpu = res.data.gpu || {
                enabled: false,
                mode: 'devices',
                deviceIDs: [],
                capabilities: ['gpu', 'utility', 'compute'],
                error: i18n.global.t('aiTools.model.loadFailed'),
            };
        }
        paramModel.name = res.data.name;
        paramModel.httpPort = res.data.httpPort;
        paramModel.syncModelAccount = res.data.syncModelAccount;
        paramModel.modelAccountBaseURLType = res.data.modelAccountBaseURLType || 'container';
        paramModel.modelAccountBaseURL = res.data.modelAccountBaseURL;
        initialAdvancedConfig = Object.fromEntries(advancedFields.map((key) => [key, paramModel[key]]));
        return true;
    } catch (error) {
        return false;
    } finally {
        loading.value = false;
    }
};

const submit = async (formEl: FormInstance) => {
    if (!formEl || loading.value || gpuRef.value?.loading) return;
    const node = currentNode.value;
    const installId = paramData.value.id;
    await formEl.validate((valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('app.updateWarn'), i18n.global.t('commons.button.update'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        }).then(async () => {
            if (!open.value || node !== currentNode.value || installId !== paramData.value.id) return;
            const submitModel: Record<string, any> = { installId, params: paramModel.params };
            if (showModelGPU.value) submitModel.gpu = paramModel.gpu;
            if (appKey.value === 'ollama') {
                submitModel.syncModelAccount = paramModel.syncModelAccount;
                submitModel.modelAccountBaseURLType = paramModel.modelAccountBaseURLType;
                submitModel.modelAccountBaseURL = accountRef.value.baseURL;
            }
            const updateAdvanced =
                appKey.value === 'ollama'
                    ? paramModel.editCompose ||
                      advancedFields.some((key) => paramModel[key] !== initialAdvancedConfig[key])
                    : paramModel.advanced;
            if (updateAdvanced) {
                submitModel.advanced = true;
                for (const key of advancedFields) {
                    submitModel[key] = paramModel[key];
                }
                if (paramModel.editCompose) {
                    submitModel.editCompose = paramModel.editCompose;
                    submitModel.dockerCompose = paramModel.dockerCompose;
                }
            }
            try {
                loading.value = true;
                await updateAppInstallParams(submitModel);
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                handleClose();
            } catch (error) {
                loading.value = false;
            }
        });
    });
};

const updateAppConfig = async () => {
    try {
        let req = {
            installID: Number(paramData.value.id),
            webUI: webUI.protocol + webUI.domain,
        };
        if (!webUI.domain || webUI.domain === '') {
            req.webUI = '';
        }
        if (!checkWebUI()) {
            MsgError(i18n.global.t('commons.rule.host'));
            return;
        }
        await updateInstallConfig(req);
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        handleClose();
    } catch (error) {}
};

const showCopyButton = (key: string) => {
    const keys = [
        'PANEL_DB_ROOT_PASSWORD',
        'PANEL_DB_NAME',
        'PANEL_DB_USER',
        'PANEL_DB_USER_PASSWORD',
        'PANEL_REDIS_ROOT_PASSWORD',
        'PANEL_DB_ROOT_USER',
    ];
    for (let i = 0; i < keys.length; i++) {
        if (key === keys[i]) {
            return true;
        }
    }
    return false;
};

const loadLimit = async () => {
    const res = await loadResourceLimit();
    limits.value = res.data;
    limits.value.memory = Number((limits.value.memory / 1024 / 1024).toFixed(2));
    oldMemory.value = limits.value.memory;
};

const changeUnit = () => {
    if (paramModel.memoryUnit == 'M') {
        limits.value.memory = oldMemory.value;
    } else {
        limits.value.memory = Number((oldMemory.value / 1024).toFixed(2));
    }
};

defineExpose({ acceptParams });

onBeforeUnmount(() => {
    open.value = false;
});
</script>

<style lang="scss" scoped>
@use '@/views/ai/model/components/form-section.scss';
.change-button {
    margin-top: 5px;
}
:deep(.system-label) {
    width: 40% !important;
    white-space: nowrap !important;
}
</style>
