<template>
    <el-drawer v-model="open" :before-close="handleClose" :title="$t('file.rename')" size="30%">
        <el-row>
            <el-col :span="22" :offset="1">
                <el-form
                    ref="fileForm"
                    label-position="top"
                    :model="addForm"
                    label-width="100px"
                    :rules="rules"
                    v-loading="loading"
                >
                    <el-form-item :label="$t('file.path')" prop="path">
                        <el-input v-model="addForm.path" disabled />
                    </el-form-item>
                    <el-form-item :label="$t('file.name')" prop="newName">
                        <el-input v-model="addForm.newName" />
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { RenameRile } from '@/api/modules/files';
import { Rules } from '@/global/form-rules';
import { ElMessage, FormInstance, FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import { File } from '@/api/interface/file';
import i18n from '@/lang';

interface RenameProps {
    path: string;
    oldName: string;
}

const fileForm = ref<FormInstance>();
const loading = ref(false);
let open = ref(false);
const oldName = ref('');

const addForm = reactive({
    newName: '',
    path: '',
});

const rules = reactive<FormRules>({
    newName: [Rules.requiredInput],
});

const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
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
        addItem['oldName'] = getPath(addForm.path, oldName.value);
        addItem['newName'] = getPath(addForm.path, addForm.newName);
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

const acceptParams = (props: RenameProps) => {
    oldName.value = props.oldName;
    addForm.newName = props.oldName;
    addForm.path = props.path;
    open.value = true;
};

defineExpose({ acceptParams });
</script>
