<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="40%">
        <template #header>
            <Header :header="$t('app.param')" :back="handleClose">
                <template #buttons v-if="canEdit">
                    <el-button type="primary" plain @click="editParam" :disabled="loading">
                        {{ edit ? $t('app.detail') : $t('commons.button.edit') }}
                    </el-button>
                </template>
            </Header>
        </template>
        <el-descriptions border :column="1" v-if="!edit">
            <el-descriptions-item v-for="(param, key) in params" :label="getLabel(param)" :key="key">
                <span>{{ param.showValue && param.showValue != '' ? param.showValue : param.value }}</span>
            </el-descriptions-item>
        </el-descriptions>
        <el-row v-else v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-alert :title="$t('app.updateHelper')" type="warning" :closable="false" />
                <el-form @submit.prevent ref="paramForm" :model="paramModel" label-position="top" :rules="rules">
                    <div v-for="(p, index) in params" :key="index">
                        <el-form-item :prop="p.key" :label="getLabel(p)">
                            <el-input
                                v-if="p.type == 'number'"
                                type="number"
                                v-model.number="paramModel[p.key]"
                                :disabled="!p.edit"
                            ></el-input>
                            <el-select v-model="paramModel[p.key]" v-else-if="p.type == 'select'">
                                <el-option
                                    v-for="value in p.values"
                                    :key="value.label"
                                    :value="value.value"
                                    :label="value.label"
                                    :disabled="!p.edit"
                                ></el-option>
                            </el-select>
                            <el-input v-else v-model.trim="paramModel[p.key]" :disabled="!p.edit"></el-input>
                        </el-form-item>
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
import { useI18n } from 'vue-i18n';
import { FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

interface ParamProps {
    id: Number;
}
const paramData = ref<ParamProps>({
    id: 0,
});

interface EditForm extends App.InstallParams {
    default: any;
}

let open = ref(false);
let loading = ref(false);
const params = ref<EditForm[]>();
let edit = ref(false);
const paramForm = ref<FormInstance>();
let paramModel = ref<any>({});
let rules = reactive({});
let submitModel = ref<any>({});
let canEdit = ref(false);

const acceptParams = async (props: ParamProps) => {
    canEdit.value = false;
    submitModel.value.installId = props.id;
    params.value = [];
    paramData.value.id = props.id;
    edit.value = false;
    await get();
    open.value = true;
};

const handleClose = () => {
    open.value = false;
};
const editParam = () => {
    params.value.forEach((param: EditForm) => {
        paramModel.value[param.key] = param.value;
    });
    edit.value = !edit.value;
};

const get = async () => {
    try {
        loading.value = true;
        const res = await GetAppInstallParams(Number(paramData.value.id));
        if (res.data && res.data.length > 0) {
            res.data.forEach((d) => {
                if (d.edit) {
                    canEdit.value = true;
                }
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
                });
                rules[d.key] = [Rules.requiredInput];
                if (d.rule) {
                    rules[d.key].push(Rules[d.rule]);
                }
            });
        }
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const getLabel = (row: EditForm): string => {
    const language = useI18n().locale.value;
    if (language == 'zh') {
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
            submitModel.value.params = paramModel.value;
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
