<template>
    <div v-loading="loading">
        <el-row>
            <el-col :xs="20" :sm="12" :md="10" :lg="10" :xl="8" :offset="1">
                <el-form :model="form" :rules="rules" ref="phpFormRef">
                    <el-form-item prop="timeout">
                        <el-input clearable type="number" v-model.number="form.timeout" maxlength="15">
                            <template #append>{{ $t('commons.units.second') }}</template>
                        </el-input>
                    </el-form-item>
                </el-form>
                <el-button type="primary" @click="openCreate(phpFormRef)">
                    {{ $t('commons.button.save') }}
                </el-button>
            </el-col>
        </el-row>
    </div>
</template>
<script setup lang="ts">
import { GetPHPConfig, UpdatePHPConfig } from '@/api/modules/runtime';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { computed, onMounted, reactive } from 'vue';
import { ref } from 'vue';
import { FormInstance } from 'element-plus';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const websiteID = computed(() => {
    return props.id;
});
const rules = reactive({
    timeout: [Rules.requiredInput, checkNumberRange(0, 999999999)],
});
const phpFormRef = ref();
const loading = ref(false);
const form = ref({
    timeout: 0,
});

const search = () => {
    loading.value = true;
    GetPHPConfig(websiteID.value)
        .then((res) => {
            form.value.timeout = parseFloat(res.data.maxExecutionTime.replace(/[^\d.]/g, ''));
        })
        .finally(() => {
            loading.value = false;
        });
};

const openCreate = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        const action = await ElMessageBox.confirm(
            i18n.global.t('database.restartNowHelper'),
            i18n.global.t('database.confChange'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        );
        if (action === 'confirm') {
            loading.value = true;
            submit();
        }
    });
};

const submit = () => {
    loading.value = true;
    UpdatePHPConfig({ scope: 'max_execution_time', id: websiteID.value, maxExecutionTime: String(form.value.timeout) })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search();
});
</script>
