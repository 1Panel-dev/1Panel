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
            <el-form-item :label="$t('file.path')" prop="path"> <el-input v-model="getPath" disabled /></el-form-item>
            <el-form-item :label="$t('file.name')" prop="name"> <el-input v-model="form.name" /></el-form-item>
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

const props = defineProps({
    open: Boolean,
    file: Object,
});
const { open, file } = toRefs(props);
let addItem = ref<File.FileCreate>({ path: '', isDir: false, mode: 0o755 });
let form = ref({ name: '', path: '' });
const em = defineEmits(['close']);
const handleClose = () => {
    em('close', open);
};

const rules = reactive<FormRules>({
    name: [Rules.required],
    path: [Rules.required],
});

const getMode = (val: number) => {
    addItem.value.mode = val;
};

let getPath = computed(() => {
    if (form.value.path === '/') {
        return form.value.path + form.value.name;
    } else {
        return form.value.path + '/' + form.value.name;
    }
});

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        addItem.value.path = getPath.value;
        CreateFile(addItem.value)
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
    addItem.value.isDir = f.isDir;
    addItem.value.path = f.path;
    form.value.name = '';
    form.value.path = f.path;
};
</script>
