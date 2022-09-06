<template>
    <el-dialog v-model="open" :before-close="handleClose" :title="$t('file.download')" width="30%" @open="onOpen">
        <el-form
            ref="fileForm"
            label-position="left"
            :model="addForm"
            label-width="100px"
            :rules="rules"
            v-loading="loading"
        >
            <el-form-item :label="$t('file.downloadUrl')" prop="url">
                <el-input v-model="addForm.url" />
            </el-form-item>
            <el-form-item :label="$t('file.path')" prop="path">
                <el-input v-model="addForm.path">
                    <template #append> <FileList :path="path" @choose="getPath"></FileList> </template
                ></el-input>
            </el-form-item>
            <el-form-item :label="$t('file.name')" prop="name">
                <el-input v-model="addForm.name"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)" :disabled="loading">{{
                    $t('commons.button.confirm')
                }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { DownloadFile } from '@/api/modules/files';
import { Rules } from '@/global/form-rues';
import i18n from '@/lang';
import { ElMessage, FormInstance, FormRules } from 'element-plus';
import { reactive, ref, toRefs } from 'vue';

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    path: {
        type: String,
        default: '',
    },
});
const { open } = toRefs(props);
const fileForm = ref<FormInstance>();
const loading = ref(false);

const rules = reactive<FormRules>({
    name: [Rules.requiredInput],
    path: [Rules.requiredInput],
    url: [Rules.requiredInput],
});

const addForm = reactive({
    url: '',
    path: '',
    name: '',
});

const em = defineEmits(['close']);

const handleClose = () => {
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    em('close', open);
};

const getPath = (path: string) => {
    addForm.path = path;
};

const onOpen = () => {
    addForm.path = props.path;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        DownloadFile(addForm)
            .then(() => {
                ElMessage.success(i18n.global.t('file.downloadStart'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};
</script>
