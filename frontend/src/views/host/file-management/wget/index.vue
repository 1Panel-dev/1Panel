<template>
    <DrawerPro v-model="open" :header="$t('commons.button.download')" @close="handleClose" size="large">
        <el-form
            ref="fileForm"
            label-position="top"
            :model="addForm"
            label-width="100px"
            :rules="rules"
            v-loading="loading"
        >
            <el-form-item :label="$t('file.downloadUrl')" prop="url">
                <el-input v-model="addForm.url" @input="getFileName" />
            </el-form-item>
            <el-form-item :label="$t('file.path')" prop="path">
                <el-input v-model="addForm.path">
                    <template #prepend>
                        <el-button icon="Folder" @click="fileRef.acceptParams({ path: addForm.path })" />
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model="addForm.name"></el-input>
            </el-form-item>
            <el-form-item>
                <el-checkbox
                    v-model="addForm.useServerFilename"
                    :disabled="!preferenceReady || preferenceLoading || preferenceSaving || loading"
                    class="!h-auto [&_.el-checkbox__label]:whitespace-normal"
                    @change="saveServerFilenamePreference"
                >
                    {{ $t('file.useServerFilename') }}
                </el-checkbox>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="addForm.useProxy">
                    {{ $t('file.useProxy') }}
                </el-checkbox>
                <span class="input-help">{{ $t('file.useProxyHelper') }}</span>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="addForm.ignoreCertificate">
                    {{ $t('file.ignoreCertificate') }}
                </el-checkbox>
                <span class="input-help">{{ $t('file.ignoreCertificateHelper') }}</span>
            </el-form-item>
        </el-form>
        <el-alert
            v-if="isAppendOnly"
            class="mt-4"
            type="warning"
            :title="$t('xpack.tamper.tamperCreateHint')"
            :closable="false"
        />
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose()" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button
                    type="primary"
                    @click="submit(fileForm)"
                    :disabled="loading || preferenceLoading || preferenceSaving"
                >
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <FileList ref="fileRef" @choose="getPath" />
</template>

<script lang="ts" setup>
import { getFileDownloadPreference, updateFileDownloadPreference, wgetFile } from '@/api/modules/files';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance, FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import FileList from '@/components/file-list/index.vue';
import { MsgSuccess } from '@/utils/message';
import { getFilenameFromUrl } from '@/utils/file';

interface WgetProps {
    path: string;
    isAppendOnly?: boolean;
}

const fileForm = ref<FormInstance>();
const loading = ref(false);
const preferenceLoading = ref(false);
const preferenceSaving = ref(false);
const preferenceReady = ref(false);
let preferenceToken = 0;
let savedServerFilename = false;
const isAppendOnly = ref(false);
let open = ref(false);
let submitData = ref(false);
const fileRef = ref();

const validateWgetUrl = (_rule: unknown, value: string, callback: (e?: Error) => void) => {
    const v = (value || '').trim();
    if (!v) {
        callback();
        return;
    }
    try {
        const u = new URL(v);
        if (u.protocol !== 'http:' && u.protocol !== 'https:') {
            callback(new Error(i18n.global.t('file.wgetUrlInvalid')));
            return;
        }
        callback();
    } catch {
        callback(new Error(i18n.global.t('file.wgetUrlInvalid')));
    }
};

const rules = reactive<FormRules>({
    name: [Rules.requiredInput],
    path: [Rules.requiredInput],
    url: [Rules.requiredInput, { validator: validateWgetUrl, trigger: 'blur' }],
});

const addForm = reactive({
    url: '',
    path: '',
    name: '',
    ignoreCertificate: false,
    useProxy: false,
    useServerFilename: false,
});

const em = defineEmits(['close']);

const handleClose = () => {
    preferenceToken++;
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    open.value = false;
    em('close', submitData.value);
};

const getPath = (path: string) => {
    addForm.path = path;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl || preferenceLoading.value || preferenceSaving.value || loading.value) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        wgetFile(addForm)
            .then(() => {
                MsgSuccess(i18n.global.t('file.downloadStart'));
                submitData.value = true;
                handleClose();
            })
            .catch(() => {
                submitData.value = false;
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const getFileName = (url: string) => {
    addForm.name = getFilenameFromUrl(url);
};

const saveServerFilenamePreference = async () => {
    if (!preferenceReady.value || preferenceSaving.value) return;
    const token = preferenceToken;
    const value = addForm.useServerFilename;
    preferenceSaving.value = true;
    try {
        await updateFileDownloadPreference(value);
        if (token === preferenceToken) savedServerFilename = value;
    } catch {
        if (token === preferenceToken) addForm.useServerFilename = savedServerFilename;
    } finally {
        if (token === preferenceToken) preferenceSaving.value = false;
    }
};

const acceptParams = async (props: WgetProps) => {
    const token = ++preferenceToken;
    addForm.path = props.path;
    isAppendOnly.value = Boolean(props.isAppendOnly);
    open.value = true;
    submitData.value = false;
    addForm.ignoreCertificate = false;
    addForm.useProxy = false;
    addForm.useServerFilename = false;
    savedServerFilename = false;
    preferenceReady.value = false;
    preferenceSaving.value = false;
    preferenceLoading.value = true;
    try {
        const result = await getFileDownloadPreference();
        if (token !== preferenceToken || !open.value) return;
        savedServerFilename = result.data?.useServerFilename === true;
        addForm.useServerFilename = savedServerFilename;
        preferenceReady.value = true;
    } catch {
        // Older Core versions can still use the existing manual-name download flow.
    } finally {
        if (token === preferenceToken) preferenceLoading.value = false;
    }
};

defineExpose({ acceptParams });
</script>
