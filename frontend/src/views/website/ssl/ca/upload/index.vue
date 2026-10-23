<template>
    <DrawerPro v-model="open" size="large" :header="$t('commons.button.upload')" @close="handleClose">
        <el-form ref="formRef" label-position="top" :model="form" :rules="rules" :disabled="loading" @submit.prevent>
            <el-form-item :label="$t('website.importType')" prop="type">
                <el-select v-model="form.type" :disabled="loading" @change="changeType">
                    <el-option :label="$t('website.pasteSSL')" :value="'paste'"></el-option>
                    <el-option :label="$t('website.localSSL')" :value="'local'"></el-option>
                    <el-option :label="$t('commons.button.upload') + ' ZIP'" :value="'upload'"></el-option>
                </el-select>
            </el-form-item>
            <el-form-item v-if="form.type !== 'upload'" :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="form.name" :maxlength="128" />
            </el-form-item>
            <div v-if="form.type === 'paste'">
                <el-form-item :label="$t('website.certificate')" prop="certificate">
                    <el-input v-model="form.certificate" :rows="6" type="textarea" />
                </el-form-item>
                <el-form-item :label="$t('website.privateKey')" prop="privateKey">
                    <el-input v-model="form.privateKey" :rows="6" type="textarea" />
                </el-form-item>
            </div>
            <div v-if="form.type === 'local'">
                <el-form-item :label="$t('website.certificatePath')" prop="certificatePath">
                    <el-input v-model="form.certificatePath">
                        <template #prepend>
                            <el-button icon="Folder" @click="certFileRef.acceptParams({ dir: false })" />
                        </template>
                    </el-input>
                </el-form-item>
                <el-form-item :label="$t('website.privateKeyPath')" prop="privateKeyPath">
                    <el-input v-model="form.privateKeyPath">
                        <template #prepend>
                            <el-button icon="Folder" @click="keyFileRef.acceptParams({ dir: false })" />
                        </template>
                    </el-input>
                </el-form-item>
            </div>
            <el-form-item v-if="form.type === 'upload'" :label="$t('file.file')" prop="file">
                <el-upload
                    ref="uploadRef"
                    accept=".zip"
                    :auto-upload="false"
                    :limit="1"
                    :disabled="loading"
                    :on-change="onChange"
                    :on-remove="onRemove"
                >
                    <el-button type="primary" icon="Upload" :disabled="loading">
                        {{ $t('file.selectFile') }}
                    </el-button>
                    <template #tip>
                        <div class="el-upload__tip">{{ $t('ssl.caUploadHelp') }}</div>
                    </template>
                </el-upload>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="loading" @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
            <el-button v-permission type="primary" :loading="loading" @click="submit">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
    <FileList ref="certFileRef" @choose="form.certificatePath = $event" />
    <FileList ref="keyFileRef" @choose="form.privateKeyPath = $event" />
</template>

<script lang="ts" setup>
import { nextTick, reactive, ref } from 'vue';
import type { FormInstance, UploadFile, UploadInstance, UploadRawFile } from 'element-plus';
import { Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import { uploadCA, importCA } from '@/api/modules/website';
import { MsgError, MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const open = ref(false);
const loading = ref(false);
const formRef = ref<FormInstance>();
const uploadRef = ref<UploadInstance>();
const certFileRef = ref();
const keyFileRef = ref();
const initialForm = () => ({
    name: '',
    type: 'upload',
    certificate: '',
    privateKey: '',
    certificatePath: '',
    privateKeyPath: '',
    file: undefined as UploadRawFile | undefined,
});
const form = reactive(initialForm());
const rules = {
    name: [Rules.requiredInput, Rules.name],
    type: [Rules.requiredSelect],
    file: [Rules.requiredSelect],
    certificate: [Rules.requiredInput],
    privateKey: [Rules.requiredInput],
    certificatePath: [Rules.requiredInput],
    privateKeyPath: [Rules.requiredInput],
};
const changeType = async () => {
    form.file = undefined;
    uploadRef.value?.clearFiles();
    await nextTick();
    formRef.value?.clearValidate();
};
const emit = defineEmits(['uploaded']);

const onRemove = () => {
    form.file = undefined;
};
const onChange = (file: UploadFile) => {
    if (!file.raw || !file.name.toLowerCase().endsWith('.zip') || file.raw.size > 10 * 1024 * 1024) {
        uploadRef.value?.clearFiles();
        form.file = undefined;
        MsgError(i18n.global.t('ssl.caUploadHelp'));
        return;
    }
    form.file = file.raw;
    formRef.value?.clearValidate('file');
};
const handleClose = () => {
    open.value = false;
};
const acceptParams = async () => {
    Object.assign(form, initialForm());
    open.value = true;
    await nextTick();
    uploadRef.value?.clearFiles();
    formRef.value?.clearValidate();
};
const submit = async () => {
    if (loading.value) return;
    loading.value = true;
    try {
        if (!(await formRef.value?.validate().catch(() => false))) return;
        if (form.type === 'upload') {
            if (!form.file) return;
            const data = new FormData();
            data.append('file', form.file);
            await uploadCA(data);
        } else {
            await importCA(
                form.type === 'paste'
                    ? { type: 'paste', name: form.name, certificate: form.certificate, privateKey: form.privateKey }
                    : {
                          type: 'local',
                          name: form.name,
                          certificatePath: form.certificatePath,
                          privateKeyPath: form.privateKeyPath,
                      },
            );
        }
        MsgSuccess(i18n.global.t('commons.msg.uploadSuccess'));
        handleClose();
        emit('uploaded');
    } finally {
        loading.value = false;
    }
};
defineExpose({ acceptParams });
</script>
