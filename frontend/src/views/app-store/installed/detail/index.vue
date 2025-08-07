<template>
    <DrawerPro v-model="open" :header="$t('app.param')" @close="handleClose" size="normal">
        <template #buttons>
            <el-button type="primary" plain @click="editParam" :disabled="loading">
                {{ edit ? $t('app.detail') : $t('commons.button.edit') }}
            </el-button>
        </template>
        <div v-if="!edit">
            <el-descriptions border :column="1">
                <el-descriptions-item :label="$t('app.webUI')">
                    <span v-if="!openConfig">
                        {{ appConfigUpdate.webUI }}
                        <el-button type="primary" @click="openConfig = true">
                            {{ $t('commons.button.edit') }}
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
                <el-descriptions-item v-for="(param, key) in params" :label="getLabel(param)" :key="key">
                    <span>{{ param.showValue && param.showValue != '' ? param.showValue : param.value }}</span>
                    <CopyButton v-if="showCopyButton(param.key)" :content="param.value" type="icon" />
                </el-descriptions-item>
            </el-descriptions>
        </div>
        <div v-else v-loading="loading">
            <el-alert :title="$t('app.updateHelper')" type="warning" :closable="false" class="common-prompt" />
            <el-form @submit.prevent ref="paramForm" :model="paramModel" label-position="top" :rules="rules">
                <div v-for="(p, index) in params" :key="index">
                    <el-form-item
                        :prop="'params.' + p.key"
                        :label="getLabel(p)"
                        v-if="p.showValue == undefined || p.showValue == ''"
                    >
                        <el-input
                            v-if="p.type == 'number'"
                            type="number"
                            v-model.number="paramModel.params[p.key]"
                            :disabled="!p.edit"
                        ></el-input>
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
                    <el-form-item :prop="'params.' + p.key" :label="getLabel(p)" v-else>
                        <el-input v-model.trim="p.showValue" :disabled="!p.edit"></el-input>
                    </el-form-item>
                </div>
                <el-form-item prop="advanced">
                    <el-checkbox v-model="paramModel.advanced" :label="$t('app.advanced')" size="large" />
                </el-form-item>
                <div v-if="paramModel.advanced">
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
                    <el-form-item :label="$t('container.cpuQuota')" prop="cpuQuota">
                        <el-input type="number" class="!w-2/5" v-model.number="paramModel.cpuQuota" maxlength="5">
                            <template #append>{{ $t('app.cpuCore') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('container.limitHelper') }}</span>
                    </el-form-item>
                    <el-form-item :label="$t('container.memoryLimit')" prop="memoryLimit">
                        <el-input class="!w-2/5" v-model.number="paramModel.memoryLimit" maxlength="10">
                            <template #append>
                                <el-select v-model="paramModel.memoryUnit" placeholder="Select" style="width: 85px">
                                    <el-option label="KB" value="K" />
                                    <el-option label="MB" value="M" />
                                    <el-option label="GB" value="G" />
                                </el-select>
                            </template>
                        </el-input>
                        <span class="input-help">{{ $t('container.limitHelper') }}</span>
                    </el-form-item>

                    <el-form-item prop="editCompose">
                        <el-checkbox v-model="paramModel.editCompose" :label="$t('app.editCompose')" size="large" />
                        <span class="input-help">{{ $t('app.editComposeHelper') }}</span>
                    </el-form-item>
                    <div v-if="paramModel.editCompose">
                        <CodemirrorPro v-model="paramModel.dockerCompose" mode="yaml"></CodemirrorPro>
                    </div>
                </div>
            </el-form>
        </div>
        <template #footer v-if="edit">
            <span>
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" :disabled="loading" @click="submit(paramForm)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { getAppInstallParams, updateAppInstallParams, updateInstallConfig } from '@/api/modules/app';
import { reactive, ref } from 'vue';
import { FormInstance } from 'element-plus';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { MsgError, MsgSuccess } from '@/utils/message';
import { getLabel, splitHttp, checkIpV4V6, checkDomain } from '@/utils/util';
import i18n from '@/lang';

interface ParamProps {
    id: Number;
    app: any;
}
const paramData = ref<ParamProps>({
    id: 0,
    app: {},
});

interface EditForm extends App.InstallParams {
    default: any;
}

const emit = defineEmits(['close']);
const open = ref(false);
const loading = ref(false);
const params = ref<EditForm[]>();
const edit = ref(false);
const paramForm = ref<FormInstance>();
const paramModel = ref<any>({
    params: {},
});
const rules = reactive({
    params: {},
    cpuQuota: [Rules.requiredInput, checkNumberRange(0, 999)],
    memoryLimit: [Rules.requiredInput, checkNumberRange(0, 9999999999)],
    containerName: [Rules.containerName],
});
const submitModel = ref<any>({
    webUI: '',
});
const appType = ref('');
const appConfigUpdate = ref<App.AppConfigUpdate>({
    installID: 0,
    webUI: '',
});
const openConfig = ref(false);
const webUI = reactive({
    protocol: 'http://',
    domain: '',
});

function checkWebUI() {
    if (webUI.domain !== '') {
        let domain = webUI.domain;
        let port = null;
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
    submitModel.value.installId = props.id;
    params.value = [];
    paramData.value.id = props.id;
    paramModel.value.params = {};
    edit.value = false;
    await get();
    open.value = true;
    openConfig.value = false;
};

const handleClose = () => {
    emit('close');
    open.value = false;
};
const editParam = () => {
    params.value.forEach((param: EditForm) => {
        paramModel.value.params[param.key] = param.value;
    });
    edit.value = !edit.value;
};

const changeAllowPort = () => {
    if (paramModel.value.allowPort) {
        paramModel.value.specifyIP = '';
    }
};

const get = async () => {
    try {
        loading.value = true;
        const res = await getAppInstallParams(Number(paramData.value.id));
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
                });
                if (d.required) {
                    rules.params[d.key] = [Rules.requiredInput];
                } else {
                    rules.params[d.key] = [];
                }
                if (d.rule) {
                    rules.params[d.key].push(Rules[d.rule]);
                }
            });
        }
        paramModel.value.memoryLimit = res.data.memoryLimit;
        paramModel.value.cpuQuota = res.data.cpuQuota;
        paramModel.value.memoryUnit = res.data.memoryUnit !== '' ? res.data.memoryUnit : 'MB';
        paramModel.value.allowPort = res.data.allowPort;
        paramModel.value.containerName = res.data.containerName;
        paramModel.value.advanced = false;
        paramModel.value.dockerCompose = res.data.dockerCompose;
        paramModel.value.isHostMode = res.data.hostMode;
        paramModel.value.specifyIP = res.data.specifyIP;
        appConfigUpdate.value.webUI = res.data.webUI;
        if (res.data.webUI != '') {
            const httpConfig = splitHttp(res.data.webUI);
            webUI.domain = httpConfig.url;
            webUI.protocol = httpConfig.proto + '://';
        }
        appType.value = res.data.type;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const submit = async (formEl: FormInstance) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        ElMessageBox.confirm(i18n.global.t('app.updateWarn'), i18n.global.t('commons.button.update'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        }).then(async () => {
            submitModel.value.params = paramModel.value.params;
            if (paramModel.value.advanced) {
                submitModel.value.advanced = paramModel.value.advanced;
                submitModel.value.memoryLimit = paramModel.value.memoryLimit;
                submitModel.value.cpuQuota = paramModel.value.cpuQuota;
                submitModel.value.memoryUnit = paramModel.value.memoryUnit;
                submitModel.value.allowPort = paramModel.value.allowPort;
                submitModel.value.containerName = paramModel.value.containerName;
                if (paramModel.value.editCompose) {
                    submitModel.value.editCompose = paramModel.value.editCompose;
                    submitModel.value.dockerCompose = paramModel.value.dockerCompose;
                }
            }
            try {
                loading.value = true;
                await updateAppInstallParams(submitModel.value);
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

defineExpose({ acceptParams });
</script>

<style lang="scss">
.change-button {
    margin-top: 5px;
}
</style>
