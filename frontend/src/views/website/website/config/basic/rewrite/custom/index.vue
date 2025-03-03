<template>
    <DrawerPro v-model="open" :header="$t('website.saveCustom')" @close="handleClose">
        <el-form ref="rewriteForm" label-position="top" :model="req" :rules="rules">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model="req.name"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(rewriteForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { operateCustomRewrite } from '@/api/modules/website';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { Rules } from '@/global/form-rules';

const rewriteForm = ref<FormInstance>();
const open = ref(false);
const loading = ref(false);
const req = ref({
    name: '',
    operate: 'create',
    content: '',
});
const rules = ref({
    name: [Rules.requiredInput],
});

const em = defineEmits(['close']);
const handleClose = () => {
    rewriteForm.value?.resetFields();
    open.value = false;
    em('close', false);
};

const acceptParams = async (conetnt: string) => {
    req.value.content = conetnt;
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        operateCustomRewrite(req.value)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
