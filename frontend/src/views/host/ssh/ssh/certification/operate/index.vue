<template>
    <DialogPro
        v-model="drawerVisible"
        :header="title"
        @close="handleClose"
        :resource="dialogData.title !== 'edit' ? '' : dialogData.rowData?.name"
        size="large"
        :autoClose="false"
        :fullScreen="true"
    >
        <el-form ref="formRef" label-position="top" :rules="rules" :model="dialogData.rowData" v-loading="loading">
            <el-row :gutter="20">
                <el-col :span="12">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model="dialogData.rowData.name" />
                    </el-form-item>
                </el-col>
            </el-row>
            <el-form-item :label="$t('ssh.createMode')" prop="mode" v-if="dialogData.title === 'create'">
                <el-radio-group v-model="dialogData.rowData.mode" @change="onModeChange">
                    <el-radio value="generate">{{ $t('ssh.generate') }}</el-radio>
                    <el-radio value="input">{{ $t('ssh.input') }}</el-radio>
                    <el-radio value="import">{{ $t('ssh.import') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-row :gutter="20">
                <el-col :span="12" v-if="isGenerate">
                    <el-form-item :label="$t('ssh.encryptionMode')" prop="encryptionMode">
                        <el-select v-model="dialogData.rowData.encryptionMode">
                            <el-option label="ED25519" value="ed25519" />
                            <el-option label="ECDSA" value="ecdsa" />
                            <el-option label="RSA" value="rsa" />
                            <el-option label="DSA" value="dsa" />
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col :span="12">
                    <el-form-item :label="$t(isGenerate ? 'ssh.password' : 'ssh.existingPassPhrase')" prop="passPhrase">
                        <el-input v-model="dialogData.rowData.passPhrase" type="password" show-password>
                            <template #append v-if="isGenerate">
                                <el-button @click="random">
                                    {{ $t('commons.button.random') }}
                                </el-button>
                            </template>
                        </el-input>
                        <span class="input-help" v-if="!isGenerate">{{ $t('ssh.existingPassPhraseHelper') }}</span>
                    </el-form-item>
                </el-col>
            </el-row>
            <div v-if="!isGenerate">
                <el-row :gutter="20">
                    <el-col :span="12">
                        <el-form-item :label="$t('ssh.privateKey')" prop="privateKey">
                            <el-input
                                v-if="dialogData.rowData.mode === 'input'"
                                type="textarea"
                                :rows="4"
                                v-model="dialogData.rowData.privateKey"
                            />
                            <el-upload
                                v-else
                                action="#"
                                :auto-upload="false"
                                ref="uploadPrivateRef"
                                class="upload mt-2 w-full"
                                :limit="1"
                                :on-change="(file) => onKeyChange('privateKey', file)"
                                :on-exceed="privateExceed"
                                :on-remove="() => onKeyRemove('privateKey')"
                            >
                                <el-button size="small" icon="Upload">
                                    {{ $t('commons.button.upload') }}
                                </el-button>
                            </el-upload>
                        </el-form-item>
                    </el-col>
                    <el-col :span="12">
                        <el-form-item :label="$t('ssh.publicKey')" prop="publicKey">
                            <el-input
                                v-if="dialogData.rowData.mode === 'input'"
                                type="textarea"
                                :rows="4"
                                v-model="dialogData.rowData.publicKey"
                            />
                            <el-upload
                                v-else
                                action="#"
                                :auto-upload="false"
                                ref="uploadPublicRef"
                                class="upload mt-2 w-full"
                                :limit="1"
                                :on-change="(file) => onKeyChange('publicKey', file)"
                                :on-exceed="publicExceed"
                                :on-remove="() => onKeyRemove('publicKey')"
                            >
                                <el-button size="small" icon="Upload">
                                    {{ $t('commons.button.upload') }}
                                </el-button>
                            </el-upload>
                        </el-form-item>
                    </el-col>
                </el-row>
            </div>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input v-model="dialogData.rowData.description" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onConfirm(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, reactive, ref } from 'vue';
import i18n from '@/lang';
import { ElForm, genFileId, UploadFile, UploadProps, UploadRawFile } from 'element-plus';
import { Host } from '@/api/interface/host';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import { getRandomStr } from '@/utils/id';
import { createCert, editCert } from '@/api/modules/host';
import { Base64 } from 'js-base64';

interface DialogProps {
    title: string;
    rowData?: Host.RootCertInfo;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const loading = ref();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref();
const uploadPrivateRef = ref();
const uploadPublicRef = ref();
const isGenerate = computed(() => dialogData.value.title === 'create' && dialogData.value.rowData?.mode === 'generate');

type KeyField = 'privateKey' | 'publicKey';
const keyReaders: Partial<Record<KeyField, FileReader>> = {};

const cancelKeyRead = (field?: KeyField) => {
    const fields: KeyField[] = field ? [field] : ['privateKey', 'publicKey'];
    for (const key of fields) {
        const reader = keyReaders[key];
        delete keyReaders[key];
        reader?.abort();
    }
};

onBeforeUnmount(() => cancelKeyRead());

const acceptParams = (params: DialogProps): void => {
    cancelKeyRead();
    dialogData.value = params;
    params.rowData.passPhrase ||= '';
    params.rowData.privateKey ||= '';
    params.rowData.publicKey ||= '';
    if (params.title === 'edit') {
        params.rowData.mode = 'input';
        dialogData.value.rowData.publicKey = Base64.decode(params.rowData.publicKey);
        dialogData.value.rowData.privateKey = Base64.decode(params.rowData.privateKey);
        if (params.rowData.passPhrase) {
            dialogData.value.rowData.passPhrase = Base64.decode(params.rowData.passPhrase);
            if (dialogData.value.rowData.passPhrase === '<UN-SET>') {
                dialogData.value.rowData.passPhrase = '';
            }
        }
    }
    if (!isGenerate.value) {
        dialogData.value.rowData.encryptionMode = '';
    }
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

function checkPassword(rule: any, value: any, callback: any) {
    if (isGenerate.value && value) {
        const reg = /^[A-Za-z0-9]{6,15}$/;
        if (!reg.test(value)) {
            return callback(new Error(i18n.global.t('ssh.passwordHelper')));
        }
    }
    callback();
}
const checkKey = (rule: { field: string }, value: string, callback: (error?: Error) => void) => {
    if (isGenerate.value) {
        return callback();
    }
    const content = value?.trim();
    if (!content) {
        return callback(new Error(i18n.global.t('commons.rule.requiredInput')));
    }
    const pattern =
        rule.field === 'publicKey'
            ? /^(?:ssh-(?:rsa|ed25519|dss)|ecdsa-sha2-nistp(?:256|384|521))[ \t]+[A-Za-z0-9+/]+={0,2}(?:[ \t]+[^\r\n]*)?$/
            : /^-----BEGIN ((?:OPENSSH |RSA |EC |DSA |ENCRYPTED )?PRIVATE KEY)-----\r?\n(?:Proc-Type: 4,ENCRYPTED\r?\nDEK-Info: [^\r\n]+\r?\n\r?\n)?[A-Za-z0-9+/][A-Za-z0-9+/=\r\n]*\r?\n-----END \1-----$/;
    if (!pattern.test(content)) {
        return callback(new Error(i18n.global.t('commons.rule.formatErr')));
    }
    callback();
};

const onModeChange = () => {
    cancelKeyRead();
    dialogData.value.rowData.encryptionMode = isGenerate.value ? 'ed25519' : '';
    dialogData.value.rowData.passPhrase = '';
    dialogData.value.rowData.privateKey = '';
    dialogData.value.rowData.publicKey = '';
    uploadPrivateRef.value?.clearFiles();
    uploadPublicRef.value?.clearFiles();
    formRef.value?.clearValidate();
};

const rules = reactive({
    name: Rules.simpleName,
    encryptionMode: Rules.requiredSelect,
    passPhrase: [{ validator: checkPassword, trigger: 'blur' }],
    privateKey: [{ required: true, validator: checkKey, trigger: ['blur', 'change'] }],
    publicKey: [{ required: true, validator: checkKey, trigger: ['blur', 'change'] }],
});

const onConfirm = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        const request = {
            ...dialogData.value.rowData,
            encryptionMode: isGenerate.value ? dialogData.value.rowData.encryptionMode : '',
        };
        if (dialogData.value.title === 'create') {
            await createCert(request)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    handleClose();
                    emit('search');
                })
                .catch(() => {
                    loading.value = false;
                });
        } else {
            await editCert(request)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    handleClose();
                    emit('search');
                })
                .catch(() => {
                    loading.value = false;
                });
        }
    });
};

const onKeyRemove = (field: KeyField) => {
    cancelKeyRead(field);
    dialogData.value.rowData[field] = '';
    formRef.value?.validateField(field).catch(() => {});
};

const onKeyChange = (field: KeyField, uploadFile: UploadFile) => {
    const rowData = dialogData.value.rowData;
    if (!drawerVisible.value || rowData.mode !== 'import') return;
    cancelKeyRead(field);
    rowData[field] = '';
    const reader = new FileReader();
    keyReaders[field] = reader;
    reader.onload = (e) => {
        if (keyReaders[field] !== reader || dialogData.value.rowData !== rowData) return;
        delete keyReaders[field];
        rowData[field] = e.target.result as string;
        formRef.value?.validateField(field).catch(() => {});
    };
    reader.onerror = () => {
        if (keyReaders[field] !== reader || dialogData.value.rowData !== rowData) return;
        delete keyReaders[field];
        MsgError(i18n.global.t('commons.msg.errImport'));
    };
    reader.readAsText(uploadFile.raw);
};
const privateExceed: UploadProps['onExceed'] = (files) => {
    onKeyRemove('privateKey');
    uploadPrivateRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadPrivateRef.value!.handleStart(file);
};
const publicExceed: UploadProps['onExceed'] = (files) => {
    onKeyRemove('publicKey');
    uploadPublicRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadPublicRef.value!.handleStart(file);
};

const random = async () => {
    dialogData.value.rowData.passPhrase = getRandomStr(10);
};

const handleClose = () => {
    cancelKeyRead();
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
