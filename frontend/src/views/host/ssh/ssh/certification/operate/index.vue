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
            <el-row :gutter="20">
                <el-col :span="12">
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
                    <el-form-item :label="$t('commons.login.password')" prop="passPhrase">
                        <el-input v-model="dialogData.rowData.passPhrase" type="password" show-password>
                            <template #append>
                                <el-button @click="random">
                                    {{ $t('commons.button.random') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-form-item :label="$t('ssh.createMode')" prop="mode" v-if="dialogData.title === 'create'">
                <el-radio-group v-model="dialogData.rowData.mode">
                    <el-radio value="generate">{{ $t('ssh.generate') }}</el-radio>
                    <el-radio value="input">{{ $t('ssh.input') }}</el-radio>
                    <el-radio value="import">{{ $t('ssh.import') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <div v-if="dialogData.rowData.mode === 'input'">
                <el-row :gutter="20">
                    <el-col :span="12">
                        <el-form-item :label="$t('ssh.privateKey')" prop="privateKey">
                            <el-input type="textarea" :rows="2" v-model="dialogData.rowData.privateKey" />
                        </el-form-item>
                    </el-col>
                    <el-col :span="12">
                        <el-form-item :label="$t('ssh.publicKey')" prop="publicKey">
                            <el-input type="textarea" :rows="2" v-model="dialogData.rowData.publicKey" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </div>
            <div v-if="dialogData.rowData.mode === 'import'">
                <el-row :gutter="20">
                    <el-col :span="12">
                        <el-form-item :label="$t('ssh.privateKey')" prop="privateKey">
                            <el-upload
                                action="#"
                                :auto-upload="false"
                                ref="uploadPrivateRef"
                                class="upload mt-2 w-full"
                                :limit="1"
                                :on-change="privateOnChange"
                                :on-exceed="privateExceed"
                            >
                                <el-button size="small" icon="Upload">
                                    {{ $t('commons.button.upload') }}
                                </el-button>
                            </el-upload>
                        </el-form-item>
                    </el-col>
                    <el-col :span="12">
                        <el-form-item :label="$t('ssh.publicKey')" prop="publicKey">
                            <el-upload
                                action="#"
                                :auto-upload="false"
                                ref="uploadPublicRef"
                                class="upload mt-2 w-full"
                                :limit="1"
                                :on-change="publicOnChange"
                                :on-exceed="publicExceed"
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
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onConfirm(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { ElForm, genFileId, UploadFile, UploadProps, UploadRawFile } from 'element-plus';
import { Host } from '@/api/interface/host';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import { getRandomStr } from '@/utils/util';
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

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    if (params.title === 'edit') {
        params.rowData.mode = 'input';
        dialogData.value.rowData.publicKey = Base64.decode(params.rowData.publicKey);
        dialogData.value.rowData.privateKey = Base64.decode(params.rowData.privateKey);
        if (params.rowData.passPhrase) {
            dialogData.value.rowData.passPhrase = Base64.decode(params.rowData.passPhrase);
        }
    }
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

function checkPassword(rule: any, value: any, callback: any) {
    if (dialogData.value.rowData.passPhrase !== '') {
        const reg = /^[A-Za-z0-9]{6,15}$/;
        if (!reg.test(dialogData.value.rowData.passPhrase)) {
            return callback(new Error(i18n.global.t('ssh.passwordHelper')));
        }
    }
    callback();
}
const rules = reactive({
    name: Rules.simpleName,
    encryptionMode: Rules.requiredSelect,
    passPhrase: [{ validator: checkPassword, trigger: 'blur' }],
    privateKey: [Rules.requiredInput],
    publicKey: [Rules.requiredInput],
});

const onConfirm = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        if (dialogData.value.title === 'create') {
            await createCert(dialogData.value.rowData)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    drawerVisible.value = false;
                    emit('search');
                })
                .catch(() => {
                    loading.value = false;
                });
        } else {
            await editCert(dialogData.value.rowData)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    drawerVisible.value = false;
                    emit('search');
                })
                .catch(() => {
                    loading.value = false;
                });
        }
    });
};

const privateOnChange = (_uploadFile: UploadFile) => {
    const reader = new FileReader();
    reader.onload = (e) => {
        try {
            dialogData.value.rowData.privateKey = e.target.result as string;
        } catch (error) {
            MsgError(i18n.global.t('commons.msg.errImport') + error.message);
        }
    };
    reader.readAsText(_uploadFile.raw);
};
const privateExceed: UploadProps['onExceed'] = (files) => {
    uploadPrivateRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadPrivateRef.value!.handleStart(file);
};
const publicOnChange = (_uploadFile: UploadFile) => {
    const reader = new FileReader();
    reader.onload = (e) => {
        try {
            dialogData.value.rowData.publicKey = e.target.result as string;
        } catch (error) {
            MsgError(i18n.global.t('commons.msg.errImport') + error.message);
        }
    };
    reader.readAsText(_uploadFile.raw);
};
const publicExceed: UploadProps['onExceed'] = (files) => {
    uploadPublicRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadPublicRef.value!.handleStart(file);
};

const random = async () => {
    dialogData.value.rowData.passPhrase = getRandomStr(10);
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
