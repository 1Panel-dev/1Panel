<template>
    <DrawerPro v-model="open" :header="$t('commons.button.download')" :back="handleClose" size="normal">
        <el-form ref="fileForm" label-position="top" :model="addForm" :rules="rules" v-loading="loading">
            <el-form-item :label="$t('file.compressType')" prop="type">
                <el-select v-model="addForm.type">
                    <el-option v-for="item in options" :key="item" :label="item" :value="item" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model="addForm.name">
                    <template #append>{{ extension }}</template>
                </el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { FormInstance, FormRules } from 'element-plus';
import { CompressExtension, CompressType } from '@/enums/files';
import { computed, reactive, ref } from 'vue';
import { downloadFile } from '@/api/modules/files';
import { File } from '@/api/interface/file';
import { Rules } from '@/global/form-rules';

interface DownloadProps {
    paths: Array<string>;
    name: string;
}

const rules = reactive<FormRules>({
    name: [Rules.requiredInput],
    type: [Rules.requiredInput],
});

const fileForm = ref<FormInstance>();
const options = ref<string[]>([]);
let loading = ref(false);
let open = ref(false);

const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    em('close', open);
};

let addForm = ref({
    paths: [] as string[],
    type: '',
    name: '',
    compress: true,
});

const extension = computed(() => {
    return CompressExtension[addForm.value.type];
});

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        let addItem = {};
        Object.assign(addItem, addForm.value);
        addItem['name'] = addForm.value.name + extension.value;
        loading.value = true;
        downloadFile(addItem as File.FileDownload)
            .then((res) => {
                const downloadUrl = window.URL.createObjectURL(new Blob([res]));
                const a = document.createElement('a');
                a.style.display = 'none';
                a.href = downloadUrl;
                a.download = addItem['name'];
                const event = new MouseEvent('click');
                a.dispatchEvent(event);
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};
const acceptParams = (props: DownloadProps) => {
    addForm.value.paths = props.paths;
    addForm.value.name = props.name;
    addForm.value.type = 'zip';
    addForm.value.compress = true;
    options.value = [];
    for (const t in CompressType) {
        options.value.push(CompressType[t]);
    }
    open.value = true;
};

defineExpose({ acceptParams });
</script>
