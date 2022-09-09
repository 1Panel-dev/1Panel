<template>
    <div>
        <el-dialog v-model="open" :title="$t('file.download')" :before-close="handleClose" width="30%" @open="onOpen">
            <el-form
                ref="fileForm"
                label-position="left"
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
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="submit(fileForm)">{{ $t('commons.button.confirm') }}</el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { FormInstance, FormRules } from 'element-plus';
import { CompressExtention, CompressType } from '@/enums/files';
import { computed, PropType, reactive, ref, toRefs } from 'vue';
import { DownloadFile } from '@/api/modules/files';
import { File } from '@/api/interface/file';
import { Rules } from '@/global/form-rues';

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    paths: {
        type: Array as PropType<string[]>,
        default: function () {
            return [];
        },
    },
    name: {
        type: String,
        default: '',
    },
});

const rules = reactive<FormRules>({
    name: [Rules.requiredInput],
    type: [Rules.requiredInput],
});

const { open } = toRefs(props);
const fileForm = ref<FormInstance>();
const options = ref<string[]>([]);
let loading = ref(false);

const em = defineEmits(['close']);
const handleClose = () => {
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

const onOpen = () => {
    addForm.value = {
        type: 'zip',
        paths: props.paths,
        name: props.name,
    };
    console.log(addForm);
    options.value = [];
    for (const t in CompressType) {
        options.value.push(CompressType[t]);
    }
};

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
</script>
