<template>
    <el-dialog
        v-model="open"
        :title="$t('file.deCompress')"
        :before-close="handleClose"
        width="30%"
        @open="onOpen"
        v-loading="loading"
    >
        <el-form ref="fileForm" label-position="left" :model="form" label-width="100px" :rules="rules">
            <el-form-item :label="$t('file.name')">
                <el-input v-model="name" disabled></el-input>
            </el-form-item>
            <el-form-item :label="$t('file.deCompressDst')" prop="dst">
                <el-input v-model="form.dst">
                    <template #append> <FileList :path="props.dst" @choose="getLinkPath"></FileList> </template
                ></el-input>
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
import { reactive, ref, toRefs } from 'vue';
import { File } from '@/api/interface/file';
import { ElMessage, FormInstance, FormRules } from 'element-plus';
import { Rules } from '@/global/form-rues';
import { DeCompressFile } from '@/api/modules/files';
import { Mimetypes } from '@/global/mimetype';
import FileList from '@/components/file-list/index.vue';

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    dst: {
        type: String,
        default: '',
    },
    path: {
        type: String,
        default: '',
    },
    name: {
        type: String,
        default: '',
    },
    mimeType: {
        type: String,
        default: '',
    },
});

const rules = reactive<FormRules>({
    dst: [Rules.required],
});

const { open, dst, path, name, mimeType } = toRefs(props);
const fileForm = ref<FormInstance>();
let loading = ref(false);
let form = ref<File.FileDeCompress>({ type: 'zip', dst: '', path: '' });

const em = defineEmits(['close']);

const handleClose = () => {
    em('close', open);
};

const getFileType = (mime: string): string => {
    if (Mimetypes.get(mime) != undefined) {
        return String(Mimetypes.get(mime));
    } else {
        return '';
    }
};

const getLinkPath = (path: string) => {
    form.value.dst = path;
};

const onOpen = () => {
    form.value = {
        dst: dst.value,
        type: getFileType(mimeType.value),
        path: path.value,
    };
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        DeCompressFile(form.value)
            .then(() => {
                ElMessage.success(i18n.global.t('file.deCompressSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};
</script>
