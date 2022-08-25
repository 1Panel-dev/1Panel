<template>
    <el-dialog
        v-model="open"
        :before-close="handleClose"
        :title="$t('commons.button.create')"
        width="30%"
        @open="onOpen"
        v-loading="loading"
    >
        <el-form ref="fileForm" label-position="left" :model="form">
            <el-form-item :label="$t('file.path')"> <el-input v-model="form.path" /></el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { defineProps, defineEmits, toRefs, ref } from 'vue';
import { File } from '@/api/interface/file';
import { ElMessage, FormInstance } from 'element-plus';
import { CreateFile } from '@/api/modules/files';
import i18n from '@/lang';

const fileForm = ref<FormInstance>();
let loading = ref<Boolean>(false);

const props = defineProps({
    open: Boolean,
    file: Object,
});
const { open, file } = toRefs(props);
let form = ref<File.FileCreate>({ path: '', isDir: false, mode: 0o755 });
const em = defineEmits(['close']);
const handleClose = () => {
    em('close', open);
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        CreateFile(form.value)
            .then(() => {
                ElMessage.success(i18n.global.t('commons.msg.createSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const onOpen = () => {
    const f = file?.value as File.FileCreate;
    form.value.isDir = f.isDir;
    form.value.path = f.path;
};

// function PrefixInteger(num: number, length: number) {
//     return (Array(length).join('0') + num).slice(-length);
// }
</script>
