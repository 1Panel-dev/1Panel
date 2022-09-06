<template>
    <el-dialog
        v-model="open"
        :before-close="handleClose"
        :title="$t('file.setRole')"
        width="30%"
        @open="onOpen"
        v-loading="loading"
    >
        <el-form ref="fileForm" label-position="left" :model="addForm" label-width="100px" :rules="rules">
            <el-form-item :label="$t('file.path')" prop="path">
                <el-input v-model="props.path" disabled
            /></el-form-item>
            <el-form-item :label="$t('file.name')" prop="newName"> <el-input v-model="addForm.newName" /></el-form-item
        ></el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { RenameRile } from '@/api/modules/files';
import { Rules } from '@/global/form-rues';
import { ElMessage, FormInstance, FormRules } from 'element-plus';
import { reactive, ref, toRefs } from 'vue';
import { File } from '@/api/interface/file';
import i18n from '@/lang';

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    oldName: {
        type: String,
        default: '',
    },
    path: {
        type: String,
        default: '',
    },
});

const { open } = toRefs(props);
const fileForm = ref<FormInstance>();
const loading = ref(false);

const addForm = reactive({
    newName: '',
    path: '',
});

const rules = reactive<FormRules>({
    newName: [Rules.requiredInput],
});

const em = defineEmits(['close']);
const handleClose = () => {
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    em('close', false);
};

const getPath = (path: string, name: string) => {
    return path + '/' + name;
};
const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }

        let addItem = {};
        Object.assign(addItem, addForm);
        addItem['oldName'] = getPath(props.path, props.oldName);
        addItem['newName'] = getPath(props.path, addForm.newName);
        loading.value = true;
        RenameRile(addItem as File.FileRename)
            .then(() => {
                ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const onOpen = () => {
    addForm.newName = props.oldName;
};
</script>
