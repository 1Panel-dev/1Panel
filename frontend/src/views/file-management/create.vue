<template>
    <el-dialog
        v-model="open"
        :before-close="handleClose"
        :title="$t('commons.button.create')"
        width="30%"
        @open="onOpen"
        v-loading="loading"
    >
        <el-form ref="fileForm" label-position="left" :model="form" label-width="100px" :rules="rules">
            <el-form-item :label="$t('file.path')"> <el-input v-model="getPath" disabled /></el-form-item>
            <el-form-item :label="$t('file.name')"> <el-input v-model="name" /></el-form-item>
            <el-checkbox v-model="isLink" :label="$t('file.link')"></el-checkbox>
        </el-form>
        <el-checkbox v-model="setRole" :label="$t('file.setRole')"></el-checkbox>
        <FileRole v-if="setRole" :mode="'0755'" @get-mode="getMode"></FileRole>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { toRefs, ref, reactive, computed } from 'vue';
import { File } from '@/api/interface/file';
import { ElMessage, FormInstance, FormRules } from 'element-plus';
import { CreateFile } from '@/api/modules/files';
import i18n from '@/lang';
import FileRole from '@/components/file-role/index.vue';
import { Rules } from '@/global/form-rues';

const fileForm = ref<FormInstance>();
let loading = ref(false);
let setRole = ref(false);
let isLink = ref(false);
let name = ref('');
let path = ref('');

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

const rules = reactive<FormRules>({
    name: [Rules.required],
});

const getMode = (val: number) => {
    form.value.mode = val;
};

let getPath = computed(() => {
    return path.value + '/' + name.value;
});

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        form.value.path = getPath.value;
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
    path.value = f.path;
    name.value = '';
};
</script>
