<template>
    <el-dialog
        v-model="open"
        :before-close="handleClose"
        :title="$t('commons.button.create')"
        width="30%"
        @open="onOpen"
        v-loading="loading"
    >
        <el-form ref="fileForm" label-position="left" :model="form" label-width="100px">
            <el-form-item :label="$t('file.path')"> <el-input v-model="form.path" /></el-form-item>
            <el-checkbox v-model="isLink" :label="$t('file.link')"></el-checkbox>
        </el-form>
        <el-checkbox v-model="setRole" :label="$t('file.setRole')"></el-checkbox>
        <FileRole v-if="setRole" :mode="'0775'" @get-mode="getMode"></FileRole>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { toRefs, ref } from 'vue';
import { File } from '@/api/interface/file';
import { ElMessage, FormInstance } from 'element-plus';
import { CreateFile } from '@/api/modules/files';
import i18n from '@/lang';
import FileRole from '@/components/file-role/index.vue';

const fileForm = ref<FormInstance>();
let loading = ref<Boolean>(false);
let setRole = ref<Boolean>(false);
let isLink = ref<Boolean>(false);

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

const getMode = (val: number) => {
    form.value.mode = val;
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
</script>
