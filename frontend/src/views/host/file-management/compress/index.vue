<template>
    <el-dialog v-model="open" :title="title" :before-close="handleClose" width="30%" @open="onOpen">
        <el-form
            ref="fileForm"
            label-position="left"
            :model="form"
            label-width="100px"
            :rules="rules"
            v-loading="loading"
        >
            <el-form-item :label="$t('file.compressType')" prop="type">
                <el-select v-model="form.type">
                    <el-option v-for="item in options" :key="item" :label="item" :value="item" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('file.name')" prop="name">
                <el-input v-model="form.name">
                    <template #append>{{ extension }}</template>
                </el-input>
            </el-form-item>
            <el-form-item :label="$t('file.compressDst')" prop="dst">
                <el-input v-model="form.dst">
                    <template #append><FileList :path="props.dst" @choose="getLinkPath"></FileList></template>
                </el-input>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="form.replace" :label="$t('file.replace')"></el-checkbox>
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
import i18n from '@/lang';
import { computed, reactive, ref, toRefs } from 'vue';
import { File } from '@/api/interface/file';
import { ElMessage, FormInstance, FormRules } from 'element-plus';
import { Rules } from '@/global/form-rues';
import { CompressExtention, CompressType } from '@/enums/files';
import { CompressFile } from '@/api/modules/files';
import FileList from '@/components/file-list/index.vue';

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    files: {
        type: Array,
        default: function () {
            return [];
        },
    },
    type: {
        type: String,
        default: 'compress',
    },
    dst: {
        type: String,
        default: '',
    },
    name: {
        type: String,
        default: '',
    },
});

const rules = reactive<FormRules>({
    type: [Rules.requiredSelect],
    dst: [Rules.requiredInput],
    name: [Rules.requiredInput],
});

const { open, files, type, dst, name } = toRefs(props);
const fileForm = ref<FormInstance>();
let loading = ref(false);
let form = ref<File.FileCompress>({ files: [], type: 'zip', dst: '', name: '', replace: false });
let options = ref<string[]>([]);

const em = defineEmits(['close']);

const title = computed(() => {
    return i18n.global.t('file.' + type.value);
});

const extension = computed(() => {
    return CompressExtention[form.value.type];
});

const handleClose = () => {
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    em('close', open);
};

const getLinkPath = (path: string) => {
    form.value.dst = path;
};

const onOpen = () => {
    form.value = {
        dst: dst.value,
        type: 'zip',
        files: files.value as string[],
        name: name.value,
        replace: false,
    };
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
        Object.assign(addItem, form.value);
        addItem['name'] = form.value.name + extension.value;
        loading.value = true;
        CompressFile(addItem as File.FileCompress)
            .then(() => {
                ElMessage.success(i18n.global.t('file.compressSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};
</script>
