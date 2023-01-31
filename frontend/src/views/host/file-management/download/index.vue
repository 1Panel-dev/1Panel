<template>
    <el-drawer
        v-model="open"
        :title="$t('file.download')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :before-close="handleClose"
        size="30%"
    >
        <el-row>
            <el-col :span="11" :offset="1">
                <el-form
                    ref="fileForm"
                    label-position="top"
                    :model="addForm"
                    label-width="100px"
                    :rules="rules"
                    v-loading="loading"
                >
                    <el-form-item :label="$t('file.compressType')" prop="type">
                        <el-select v-model="addForm.type">
                            <el-option v-for="item in options" :key="item" :label="item" :value="item" />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('file.name')" prop="name">
                        <el-input v-model="addForm.name">
                            <template #append>{{ extension }}</template>
                        </el-input>
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

<script setup lang="ts">
import { FormInstance, FormRules } from 'element-plus';
import { CompressExtention, CompressType } from '@/enums/files';
import { computed, reactive, ref } from 'vue';
import { DownloadFile } from '@/api/modules/files';
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
});

const extension = computed(() => {
    return CompressExtention[addForm.value.type];
});

// const onOpen = () => {
//     addForm.value = {
//         type: 'zip',
//         paths: props.paths,
//         name: props.name,
//     };
//     console.log(addForm);
//     options.value = [];
//     for (const t in CompressType) {
//         options.value.push(CompressType[t]);
//     }
// };

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
    options.value = [];
    for (const t in CompressType) {
        options.value.push(CompressType[t]);
    }
    open.value = true;
};

defineExpose({ acceptParams });
</script>
