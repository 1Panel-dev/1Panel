<template>
    <div>
        <DrawerPro v-model="drawerVisible" :header="$t('ssh.pubkey')" @close="handleClose" size="large">
            <div class="mb-4">
                <el-alert :closable="false">{{ $t('ssh.pubKeyHelper', [currentUser]) }}</el-alert>
            </div>
            <el-button type="primary" plain @click="onCreate()">
                {{ $t('commons.button.create') }}
            </el-button>
            <el-button plain @click="onSync()">
                {{ $t('commons.button.sync') }}
            </el-button>
            <el-button plain :disabled="selects.length === 0" @click="onDelete(null)">
                {{ $t('commons.button.delete') }}
            </el-button>
            <ComplexTable
                :pagination-config="paginationConfig"
                v-model:selects="selects"
                :data="data"
                @search="search"
                :heightDiff="370"
            >
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip prop="name" />
                <el-table-column :label="$t('ssh.encryptionMode')" prop="encryptionMode" />
                <el-table-column :label="$t('commons.table.description')" prop="description" />
                <fu-table-operations width="200px" :buttons="buttons" :label="$t('commons.table.operate')" />
            </ComplexTable>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </DrawerPro>

        <DialogPro v-model="formOpen" :title="$t('commons.button.create')" size="w-60">
            <div>
                <el-form ref="formRef" label-position="top" :rules="rules" :model="form" v-loading="loading">
                    <el-row :gutter="20">
                        <el-col :span="12">
                            <el-form-item :label="$t('commons.table.name')" prop="name">
                                <el-input v-model="form.name" />
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-row :gutter="20">
                        <el-col :span="12">
                            <el-form-item :label="$t('ssh.encryptionMode')" prop="encryptionMode">
                                <el-select v-model="form.encryptionMode">
                                    <el-option label="ED25519" value="ed25519" />
                                    <el-option label="ECDSA" value="ecdsa" />
                                    <el-option label="RSA" value="rsa" />
                                    <el-option label="DSA" value="dsa" />
                                </el-select>
                            </el-form-item>
                        </el-col>
                        <el-col :span="12">
                            <el-form-item :label="$t('commons.login.password')" prop="passPhrase">
                                <el-input v-model="form.passPhrase" type="password" show-password>
                                    <template #append>
                                        <el-button @click="random">
                                            {{ $t('commons.button.random') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-form-item :label="$t('ssh.createMode')" prop="privateKey">
                        <el-radio-group v-model="form.mode">
                            <el-radio value="generate">{{ $t('ssh.generate') }}</el-radio>
                            <el-radio value="input">{{ $t('ssh.input') }}</el-radio>
                            <el-radio value="import">{{ $t('ssh.import') }}</el-radio>
                        </el-radio-group>
                    </el-form-item>
                    <div v-if="form.mode === 'input'">
                        <el-row :gutter="20">
                            <el-col :span="12">
                                <el-form-item :label="$t('ssh.privateKey')" prop="privateKey">
                                    <el-input type="textarea" :rows="2" v-model="form.privateKey" />
                                </el-form-item>
                            </el-col>
                            <el-col :span="12">
                                <el-form-item :label="$t('ssh.publicKey')" prop="publicKey">
                                    <el-input type="textarea" :rows="2" v-model="form.publicKey" />
                                </el-form-item>
                            </el-col>
                        </el-row>
                    </div>
                    <div v-if="form.mode === 'import'">
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
                        <el-input v-model="form.description" />
                    </el-form-item>
                </el-form>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="onCancel">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button type="primary" :disabled="loading" @click="onConfirm(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </DialogPro>

        <DialogPro v-model="connOpen" :title="$t('ssh.pubkey')" size="small" :showClose="false">
            <el-descriptions class="margin-top" :column="1" border>
                <el-descriptions-item align="center" :label="$t('ssh.password')">
                    <div>
                        <span>{{ loadPassPhrase() }}</span>
                    </div>
                    <el-button
                        v-if="currentRow.passPhrase && Base64.decode(currentRow.passPhrase) !== '<UN-SET>'"
                        size="small"
                        icon="CopyDocument"
                        @click="onCopy(currentRow.passPhrase)"
                    >
                        {{ $t('commons.button.copy') }}
                    </el-button>
                </el-descriptions-item>
                <el-descriptions-item align="center" :label="$t('ssh.publicKey')">
                    <el-button-group size="small">
                        <el-button icon="CopyDocument" @click="onCopy(currentRow.publicKey)">
                            {{ $t('commons.button.copy') }}
                        </el-button>
                        <el-button icon="Download" @click="onDownload(currentRow, 'publicKey')">
                            {{ $t('commons.button.download') }}
                        </el-button>
                    </el-button-group>
                </el-descriptions-item>
                <el-descriptions-item align="center" :label="$t('ssh.privateKey')">
                    <el-button-group size="small">
                        <el-button icon="CopyDocument" @click="onCopy(currentRow.privateKey)">
                            {{ $t('commons.button.copy') }}
                        </el-button>
                        <el-button icon="Download" @click="onDownload(currentRow, 'privateKey')">
                            {{ $t('commons.button.download') }}
                        </el-button>
                    </el-button-group>
                </el-descriptions-item>
            </el-descriptions>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="connOpen = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                </span>
            </template>
        </DialogPro>
        <OpDialog ref="opRef" @search="search" @submit="onSubmitDelete()">
            <template #content>
                <el-form ref="deleteForm" label-position="left">
                    <el-form-item>
                        <el-checkbox v-model="forceDelete" :label="$t('website.forceDelete')" />
                        <span class="input-help">
                            {{ $t('website.forceDeleteHelper') }}
                        </span>
                    </el-form-item>
                </el-form>
            </template>
        </OpDialog>
    </div>
</template>
<script lang="ts" setup>
import { Host } from '@/api/interface/host';
import { createCert, deleteCert, searchCert, syncCert } from '@/api/modules/host';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { copyText, getRandomStr } from '@/utils/util';
import { FormInstance, genFileId, UploadFile, UploadProps, UploadRawFile } from 'element-plus';
import { Base64 } from 'js-base64';
import { reactive, ref } from 'vue';

const loading = ref();
const drawerVisible = ref();
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'login-log-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('login-log-page-size')) || 10,
    total: 0,
    small: true,
});

const forceDelete = ref();
const operateIDs = ref();
const opRef = ref();

const currentRow = ref();
const connOpen = ref();

const formOpen = ref();
const formRef = ref();
const uploadPrivateRef = ref();
const uploadPublicRef = ref();

const currentUser = ref();
const form = reactive({
    name: '',
    mode: 'generate',
    passPhrase: '',
    encryptionMode: '',
    privateKey: '',
    publicKey: '',
    description: '',
});
const rules = reactive({
    name: Rules.simpleName,
    encryptionMode: Rules.requiredSelect,
    passPhrase: [{ validator: checkPassword, trigger: 'blur' }],
});

function checkPassword(rule: any, value: any, callback: any) {
    if (form.passPhrase !== '') {
        const reg = /^[A-Za-z0-9]{6,15}$/;
        if (!reg.test(form.passPhrase)) {
            return callback(new Error(i18n.global.t('ssh.passwordHelper')));
        }
    }
    callback();
}

const acceptParams = async (user: string): Promise<void> => {
    search();
    currentUser.value = user || 'root';
    drawerVisible.value = true;
};

const random = async () => {
    form.passPhrase = getRandomStr(10);
};

const loadPassPhrase = () => {
    if (currentRow.value.passPhrase === '') {
        return '-';
    }
    let pass = Base64.decode(currentRow.value.passPhrase);
    return pass === '<UN-SET>' ? i18n.global.t('ssh.unSyncPass') : '';
};

const onCopy = async (content: string) => {
    content = Base64.decode(content);
    copyText(content);
};

const onCreate = () => {
    form.name = '';
    form.mode = 'generate';
    form.encryptionMode = 'ed25519';
    form.passPhrase = '';
    form.privateKey = '';
    form.publicKey = '';
    form.description = '';
    formOpen.value = true;
};

const onConfirm = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await createCert(form)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                formOpen.value = false;
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const privateOnChange = (_uploadFile: UploadFile) => {
    const reader = new FileReader();
    reader.onload = (e) => {
        try {
            form.privateKey = e.target.result as string;
        } catch (error) {
            MsgError(i18n.global.t('cronjob.errImport') + error.message);
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
            form.publicKey = e.target.result as string;
        } catch (error) {
            MsgError(i18n.global.t('cronjob.errImport') + error.message);
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

const onCancel = () => {
    formOpen.value = false;
};

const onDownload = async (row: Host.RootCertInfo, type: string) => {
    let name = row.name;
    let content;
    if (type === 'publicKey') {
        name = row.name + '.pub';
        content = Base64.decode(row.publicKey);
    } else {
        content = Base64.decode(row.privateKey);
    }
    const downloadUrl = window.URL.createObjectURL(new Blob([content], { type: 'application/octet-stream' }));
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = downloadUrl;
    a.download = name;
    const event = new MouseEvent('click');
    a.dispatchEvent(event);

    setTimeout(() => {
        document.body.removeChild(a);
        window.URL.revokeObjectURL(downloadUrl);
    }, 100);
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchCert(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items;
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onSync = async () => {
    ElMessageBox.confirm(i18n.global.t('ssh.syncHelper'), i18n.global.t('commons.button.sync'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await syncCert()
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onDelete = async (row: Host.RootCertInfo | null) => {
    let names = [];
    let ids = [];
    forceDelete.value = false;
    if (row) {
        ids = [row.id];
        names = [row.name + ' - ' + row.encryptionMode];
    } else {
        for (const item of selects.value) {
            names.push(item.name + ' - ' + item.encryptionMode);
            ids.push(item.id);
        }
    }
    operateIDs.value = ids;
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('menu.cronjob'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: null,
        params: null,
    });
};

const onSubmitDelete = async () => {
    loading.value = true;
    await deleteCert(operateIDs.value, forceDelete.value)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    drawerVisible.value = false;
};

const buttons = [
    {
        label: i18n.global.t('commons.button.view'),
        click: (row: Host.RootCertInfo) => {
            currentRow.value = row;
            connOpen.value = true;
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Host.RootCertInfo) => {
            onDelete(row);
        },
    },
];

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.marginTop {
    margin-top: 10px;
}
</style>
