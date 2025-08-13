<template>
    <DrawerPro v-model="open" :header="$t('file.deCompress')" :resource="name" @close="handleClose" size="normal">
        <el-form
            ref="fileForm"
            label-position="top"
            :model="form"
            label-width="100px"
            :rules="rules"
            v-loading="loading"
        >
            <el-form-item :label="$t('commons.table.name')">
                <el-input v-model="name" disabled></el-input>
            </el-form-item>
            <el-form-item :label="$t('file.deCompressDst')" prop="dst">
                <el-input v-model="form.dst">
                    <template #prepend>
                        <el-button icon="Folder" @click="fileRef.acceptParams({ path: form.dst, dir: true })" />
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item :label="$t('setting.compressPassword')" prop="secret" v-if="name.includes('tar.gz')">
                <el-input v-model="form.secret"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </DrawerPro>
    <FileList ref="fileRef" @choose="getLinkPath" />
</template>

<script setup lang="ts">
import i18n from '@/lang';
import { reactive, ref } from 'vue';
import { File } from '@/api/interface/file';
import { FormInstance, FormRules } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { deCompressFile } from '@/api/modules/files';
import { Mimetypes } from '@/global/mimetype';
import FileList from '@/components/file-list/index.vue';
import { MsgSuccess } from '@/utils/message';

interface CompressProps {
    files: Array<any>;
    dst: string;
    name: string;
    path: string;
    mimeType: string;
}

const rules = reactive<FormRules>({
    dst: [Rules.requiredInput],
});

const fileForm = ref<FormInstance>();
let loading = ref(false);
let form = ref<File.FileDeCompress>({ type: 'zip', dst: '', path: '', secret: '' });
let open = ref(false);
let name = ref('');
const fileRef = ref();

const em = defineEmits(['close']);

const handleClose = () => {
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    open.value = false;
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

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        deCompressFile(form.value)
            .then(() => {
                MsgSuccess(i18n.global.t('file.deCompressSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const acceptParams = (props: CompressProps) => {
    form.value.type = getFileType(props.mimeType);
    form.value.dst = props.dst;
    form.value.path = props.path;
    name.value = props.name;
    open.value = true;
};

defineExpose({ acceptParams });
</script>
