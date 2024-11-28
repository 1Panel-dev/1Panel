<template>
    <div v-loading="loading">
        <el-row>
            <el-col :xs="20" :sm="12" :md="10" :lg="10" :xl="8">
                <el-form :model="form" :rules="rules" ref="phpFormRef">
                    <el-form-item prop="uploadSize">
                        <el-input clearable type="number" v-model.number="form.uploadSize" maxlength="15">
                            <template #append>M</template>
                        </el-input>
                    </el-form-item>
                </el-form>
                <el-button type="primary" @click="openCreate(phpFormRef)">
                    {{ $t('commons.button.save') }}
                </el-button>
            </el-col>
        </el-row>
        <ConfirmDialog ref="confirmDialogRef" @confirm="submit"></ConfirmDialog>
    </div>
</template>
<script setup lang="ts">
import { GetPHPConfig, UpdatePHPConfig } from '@/api/modules/website';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { computed, onMounted, reactive } from 'vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
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
    uploadSize: [Rules.requiredInput, checkNumberRange(0, 999999999)],
});
const phpFormRef = ref();
const confirmDialogRef = ref();
const loading = ref(false);
const form = ref({
    uploadSize: 0,
});

const search = () => {
    loading.value = true;
    GetPHPConfig(websiteID.value)
        .then((res) => {
            form.value.uploadSize = parseFloat(res.data.uploadMaxSize.replace(/[^\d.]/g, ''));
        })
        .finally(() => {
            loading.value = false;
        });
};

const openCreate = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            header: i18n.global.t('database.confChange'),
            operationInfo: i18n.global.t('database.restartNowHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialogRef.value!.acceptParams(params);
    });
};

const submit = () => {
    loading.value = true;
    const uploadMaxSize = form.value.uploadSize + 'M';
    UpdatePHPConfig({ scope: 'upload_max_filesize', id: websiteID.value, uploadMaxSize: uploadMaxSize })
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
