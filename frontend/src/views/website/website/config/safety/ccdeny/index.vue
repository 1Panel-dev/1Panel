<template>
    <el-row :gutter="20" v-loading="loading">
        <el-col :xs="24" :sm="18" :md="10" :lg="10" :xl="10">
            <el-form ref="wafForm" label-position="left" label-width="auto" :model="form" :rules="rules">
                <el-form-item prop="enable" :label="$t('website.enable')">
                    <el-switch v-model="form.enable" @change="updateEnable"></el-switch>
                </el-form-item>
                <el-form-item prop="cycle" :label="$t('website.cycle')">
                    <el-input v-model.number="form.cycle" maxlength="15">
                        <template #append>{{ $t('commons.units.second') }}</template>
                    </el-input>
                </el-form-item>
                <el-form-item prop="frequency" :label="$t('website.frequency')">
                    <el-input v-model.number="form.frequency" maxlength="15">
                        <template #append>{{ $t('commons.units.time') }}</template>
                    </el-input>
                </el-form-item>
                <el-alert
                    :title="$t('website.ccHelper', [form.cycle, form.frequency])"
                    type="info"
                    :closable="false"
                ></el-alert>
                <el-form-item></el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="submit(wafForm)">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </el-form-item>
            </el-form>
        </el-col>
    </el-row>
</template>

<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { GetWafConfig, UpdateWafEnable, UpdateWafFile } from '@/api/modules/website';
import { checkNumberRange, Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const id = computed(() => {
    return props.id;
});

const data = ref<Website.WafRes>();
const loading = ref(false);
const form = reactive({
    enable: false,
    cycle: 60,
    frequency: 120,
});
const req = ref<Website.WafReq>({
    websiteId: 0,
    key: '$CCDeny',
    rule: 'cc',
});
const enableUpdate = ref<Website.WafUpdate>({
    websiteId: 0,
    key: '$CCDeny',
    enable: false,
});
const fileUpdate = reactive({
    content: '',
    websiteId: 0,
    type: 'cc',
});
const rules = ref({
    cycle: [Rules.requiredInput, checkNumberRange(1, 9999999)],
    frequency: [Rules.requiredInput, checkNumberRange(1, 9999999)],
});
const wafForm = ref<FormInstance>();

const get = async () => {
    loading.value = true;
    const res = await GetWafConfig(req.value);
    loading.value = false;
    data.value = res.data;
    form.enable = data.value.enable;
    if (data.value.content != '') {
        const params = data.value.content.split('/');
        form.frequency = Number(params[0]);
        form.cycle = Number(params[1]);
    }
};

const updateEnable = async (enable: boolean) => {
    enableUpdate.value.enable = enable;
    loading.value = true;
    try {
        await UpdateWafEnable(enableUpdate.value);
    } catch (error) {
        form.enable = !enable;
    }
    loading.value = false;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        fileUpdate.content = String(form.frequency) + '/' + String(form.cycle);
        loading.value = true;
        UpdateWafFile(fileUpdate)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

onMounted(() => {
    req.value.websiteId = id.value;
    enableUpdate.value.websiteId = id.value;
    fileUpdate.websiteId = id.value;
    get();
});
</script>
