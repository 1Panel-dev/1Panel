<template>
    <el-dialog
        v-model="open"
        :before-close="handleClose"
        :title="$t('commons.button.create')"
        width="30%"
        @open="onOpen"
    >
        <el-form
            ref="fileForm"
            label-position="left"
            :model="addForm"
            label-width="100px"
            :rules="rules"
            v-loading="loading"
        >
            <el-form-item :label="$t('file.path')" prop="path"><el-input v-model="getPath" disabled /></el-form-item>
            <el-form-item :label="$t('file.name')" prop="name"><el-input v-model="addForm.name" /></el-form-item>
            <el-form-item v-if="!addForm.isDir">
                <el-checkbox v-model="addForm.isLink" :label="$t('file.link')"></el-checkbox>
            </el-form-item>
            <el-form-item :label="$t('file.linkType')" v-if="addForm.isLink" prop="linkType">
                <el-radio-group v-model="addForm.isSymlink">
                    <el-radio :label="true">{{ $t('file.softLink') }}</el-radio>
                    <el-radio :label="false">{{ $t('file.hardLink') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="addForm.isLink" :label="$t('file.linkPath')" prop="linkPath">
                <el-input v-model="addForm.linkPath">
                    <template #append>
                        <FileList @choose="getLinkPath"></FileList>
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-if="addForm.isDir" v-model="setRole" :label="$t('file.setRole')"></el-checkbox>
            </el-form-item>
            <el-form-item>
                <FileRole v-if="setRole" :mode="'0755'" @get-mode="getMode"></FileRole>
            </el-form-item>
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
import { toRefs, ref, reactive, computed } from 'vue';
import { File } from '@/api/interface/file';
import { ElMessage, FormInstance, FormRules } from 'element-plus';
import { CreateFile } from '@/api/modules/files';
import i18n from '@/lang';
import FileRole from '@/components/file-role/index.vue';
import { Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';

const fileForm = ref<FormInstance>();
let loading = ref(false);
let setRole = ref(false);

const props = defineProps({
    open: Boolean,
    file: Object,
});
const { open, file } = toRefs(props);

let addForm = reactive({ path: '', name: '', isDir: false, mode: 0o755, isLink: false, isSymlink: true, linkPath: '' });

const em = defineEmits(['close']);
const handleClose = () => {
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    em('close', open);
};

const rules = reactive<FormRules>({
    name: [Rules.requiredInput],
    path: [Rules.requiredInput],
    isSymlink: [Rules.requiredInput],
    linkPath: [Rules.requiredInput],
});

const getMode = (val: number) => {
    addForm.mode = val;
};

let getPath = computed(() => {
    if (addForm.path.endsWith('/')) {
        return addForm.path + addForm.name;
    } else {
        return addForm.path + '/' + addForm.name;
    }
});

const getLinkPath = (path: string) => {
    addForm.linkPath = path;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }

        let addItem = {};
        Object.assign(addItem, addForm);
        addItem['path'] = getPath.value;
        loading.value = true;
        CreateFile(addItem as File.FileCreate)
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
    addForm.isDir = f.isDir;
    addForm.path = f.path;
    addForm.name = '';
    addForm.isLink = false;
    init();
};

const init = () => {
    setRole.value = false;
};
</script>
