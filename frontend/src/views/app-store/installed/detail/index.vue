<template>
    <el-drawer :close-on-click-modal="false" :close-on-press-escape="false" v-model="open" size="40%">
        <template #header>
            <Header :header="$t('app.param')" :back="handleClose">
                <template #buttons>
                    <el-button type="primary" plain @click="editParam" :disabled="loading">
                        {{ edit ? $t('app.detail') : $t('commons.button.edit') }}
                    </el-button>
                </template>
            </Header>
        </template>
        <el-row v-if="!edit">
            <el-col :span="22" :offset="1">
                <el-descriptions border :column="1">
                    <el-descriptions-item v-for="(param, key) in params" :label="getLabel(param)" :key="key">
                        <span>{{ param.showValue && param.showValue != '' ? param.showValue : param.value }}</span>
                    </el-descriptions-item>
                </el-descriptions>
            </el-col>
        </el-row>
        <el-row v-else v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-alert :title="$t('app.updateHelper')" type="warning" :closable="false" class="common-prompt" />
                <el-form @submit.prevent ref="paramForm" :model="paramModel" label-position="top" :rules="rules">
                    <div v-for="(p, index) in params" :key="index">
                        <el-form-item
                            :prop="p.key"
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
                        <el-form-item :prop="p.key" :label="getLabel(p)" v-else>
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
                            <el-checkbox v-model="paramModel.allowPort" :label="$t('app.allowPort')" size="large" />
                            <span class="input-help">{{ $t('app.allowPortHelper') }}</span>
                        </el-form-item>
                        <el-form-item
                            :label="$t('container.cpuQuota')"
                            prop="cpuQuota"
                            :rules="checkNumberRange(0, limits.cpu)"
                        >
                            <el-input
                                type="number"
                                style="width: 40%"
                                v-model.number="paramModel.cpuQuota"
                                maxlength="5"
                            >
                                <template #append>{{ $t('app.cpuCore') }}</template>
                            </el-input>
                            <span class="input-help">
                                {{ $t('container.limitHelper', [limits.cpu]) }} {{ $t('commons.units.core') }}
                            </span>
                        </el-form-item>
                        <el-form-item
                            :label="$t('container.memoryLimit')"
                            prop="memoryLimit"
                            :rules="checkNumberRange(0, limits.memory)"
                        >
                            <el-input style="width: 40%" v-model.number="paramModel.memoryLimit" maxlength="10">
                                <template #append>
                                    <el-select
                                        v-model="paramModel.memoryUnit"
                                        placeholder="Select"
                                        style="width: 85px"
                                        @change="changeUnit"
                                    >
                                        <el-option label="KB" value="K" />
                                        <el-option label="MB" value="M" />
                                        <el-option label="GB" value="G" />
                                    </el-select>
                                </template>
                            </el-input>
                            <span class="input-help">
                                {{ $t('container.limitHelper', [limits.memory]) }} {{ paramModel.memoryUnit }}B
                            </span>
                        </el-form-item>

                        <el-form-item prop="editCompose">
                            <el-checkbox v-model="paramModel.editCompose" :label="$t('app.editCompose')" size="large" />
                            <span class="input-help">{{ $t('app.editComposeHelper') }}</span>
                        </el-form-item>
                        <div v-if="paramModel.editCompose">
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
                                v-model="paramModel.dockerCompose"
                            />
                        </div>
                    </div>
                </el-form>
            </el-col>
        </el-row>
        <template #footer v-if="edit">
            <span>
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" :disabled="loading" @click="submit(paramForm)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>
<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { GetAppInstallParams, UpdateAppInstallParams } from '@/api/modules/app';
import { reactive, ref } from 'vue';
import Header from '@/components/drawer-header/index.vue';
import { FormInstance } from 'element-plus';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { Codemirror } from 'vue-codemirror';
import { yaml } from '@codemirror/lang-yaml';
import { oneDark } from '@codemirror/theme-one-dark';
import { getLanguage } from '@/utils/util';
import { Container } from '@/api/interface/container';
import { loadResourceLimit } from '@/api/modules/container';

const extensions = [yaml(), oneDark];

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
    containerName: [Rules.containerName],
});
const submitModel = ref<any>({});
const oldMemory = ref<number>(0);
const limits = ref<Container.ResourceLimit>({
    cpu: null as number,
    memory: null as number,
});

const changeUnit = () => {
    if (paramModel.value.memoryUnit == 'M') {
        limits.value.memory = oldMemory.value;
    } else {
        limits.value.memory = Number((oldMemory.value / 1024).toFixed(2));
    }
};

const loadLimit = async () => {
    const res = await loadResourceLimit();
    limits.value = res.data;
    limits.value.memory = Number((limits.value.memory / 1024 / 1024).toFixed(2));
    oldMemory.value = limits.value.memory;
};

const acceptParams = async (props: ParamProps) => {
    loadLimit();
    submitModel.value.installId = props.id;
    params.value = [];
    paramData.value.id = props.id;
    paramModel.value.params = {};
    edit.value = false;
    await get();
    open.value = true;
};

const handleClose = () => {
    open.value = false;
};
const editParam = () => {
    params.value.forEach((param: EditForm) => {
        paramModel.value.params[param.key] = param.value;
    });
    edit.value = !edit.value;
};

const get = async () => {
    try {
        loading.value = true;
        const res = await GetAppInstallParams(Number(paramData.value.id));
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
                });
                rules.params[d.key] = [Rules.requiredInput];
                if (d.rule) {
                    rules.params[d.key].push(Rules[d.rule]);
                }
            });
        }
        paramModel.value.memoryLimit = res.data.memoryLimit;
        paramModel.value.cpuQuota = res.data.cpuQuota;
        paramModel.value.memoryUnit = res.data.memoryUnit !== '' ? res.data.memoryUnit : 'M';
        paramModel.value.allowPort = res.data.allowPort;
        paramModel.value.containerName = res.data.containerName;
        paramModel.value.advanced = false;
        paramModel.value.dockerCompose = res.data.dockerCompose;
        paramModel.value.isHostMode = res.data.hostMode;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const getLabel = (row: EditForm): string => {
    const language = getLanguage();
    if (language == 'zh' || language == 'tw') {
        return row.labelZh;
    } else {
        return row.labelEn;
    }
};

const submit = async (formEl: FormInstance) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        ElMessageBox.confirm(i18n.global.t('app.updateWarn'), i18n.global.t('app.update'), {
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
                await UpdateAppInstallParams(submitModel.value);
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                handleClose();
            } catch (error) {
                loading.value = false;
            }
        });
    });
};

defineExpose({ acceptParams });
</script>

<style lang="scss">
.change-button {
    margin-top: 5px;
}
</style>
