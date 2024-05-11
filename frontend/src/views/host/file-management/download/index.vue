<template>
    <el-drawer
        v-model="open"
        :title="$t('file.download')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        :before-close="handleClose"
        size="40%"
    >
        <template #header>
            <DrawerHeader :header="$t('file.download')" :back="handleClose" />
        </template>
        <el-form ref="fileForm" label-position="top" :model="addForm" :rules="rules" v-loading="loading">
            <el-row type="flex" justify="center">
                <el-col :span="22">
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
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script setup lang="ts">
import { FormInstance, FormRules } from 'element-plus';
import { CompressExtension, CompressType } from '@/enums/files';
import { computed, reactive, ref } from 'vue';
import { DownloadFile } from '@/api/modules/files';
import { File } from '@/api/interface/file';
import { Rules } from '@/global/form-rules';
import DrawerHeader from '@/components/drawer-header/index.vue';

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
        DownloadFile(addItem as File.FileDownload)
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
