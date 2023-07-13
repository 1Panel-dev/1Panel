<template>
    <div v-loading="loading">
        <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
            <template #header>
                <DrawerHeader :header="title + $t('setting.backupAccount')" :back="handleClose" />
            </template>
            <el-form
                @submit.prevent
                ref="formRef"
                v-loading="loading"
                label-position="top"
                :model="dialogData.rowData"
                label-width="120px"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.table.type')" prop="type" :rules="Rules.requiredSelect">
                            <el-tag>{{ $t('setting.' + dialogData.rowData!.type) }}</el-tag>
                        </el-form-item>
                        <el-form-item
                            v-if="dialogData.rowData!.type === 'LOCAL'"
                            :label="$t('setting.currentPath')"
                            prop="varsJson['dir']"
                            :rules="Rules.requiredInput"
                        >
                            <el-input v-model="dialogData.rowData!.varsJson['dir']">
                                <template #prepend>
                                    <FileList @choose="loadDir" :dir="true"></FileList>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item
                            v-if="hasBucket(dialogData.rowData!.type)"
                            label="Access Key ID"
                            prop="accessKey"
                            :rules="Rules.requiredInput"
                        >
                            <el-input v-model.trim="dialogData.rowData!.accessKey" />
                        </el-form-item>
                        <el-form-item
                            v-if="hasBucket(dialogData.rowData!.type)"
                            label="Secret Key"
                            prop="credential"
                            :rules="Rules.requiredInput"
                        >
                            <el-input show-password clearable v-model.trim="dialogData.rowData!.credential" />
                        </el-form-item>
                        <el-form-item v-if="dialogData.rowData!.type === 'OneDrive'">
                            <el-checkbox v-model="dialogData.rowData!.varsJson['isCN']" :label="$t('setting.isCN')" />
                        </el-form-item>
                        <el-form-item
                            v-if="dialogData.rowData!.type === 'OneDrive'"
                            :label="$t('setting.code')"
                            prop="varsJson.code"
                            :rules="Rules.requiredInput"
                        >
                            <el-input clearable v-model.trim="dialogData.rowData!.varsJson['code']">
                                <template #append>
                                    <el-button @click="jumpAzure">{{ $t('setting.loadCode') }}</el-button>
                                </template>
                            </el-input>
                            <span class="input-help">
                                {{ $t('setting.codeHelper') }}
                                <el-link
                                    style="font-size: 12px; margin-left: 5px"
                                    icon="Position"
                                    @click="toDoc()"
                                    type="primary"
                                >
                                    {{ $t('firewall.quickJump') }}
                                </el-link>
                            </span>
                        </el-form-item>
                        <el-form-item
                            v-if="dialogData.rowData!.type === 'S3' || dialogData.rowData!.type === 'COS'"
                            label="Region"
                            prop="varsJson.region"
                            :rules="Rules.requiredInput"
                        >
                            <el-input v-model.trim="dialogData.rowData!.varsJson['region']" />
                        </el-form-item>
                        <el-form-item
                            v-if="hasEndpoint(dialogData.rowData!.type)"
                            label="Endpoint"
                            prop="varsJson.endpoint"
                            :rules="Rules.requiredInput"
                        >
                            <el-input v-model.trim="dialogData.rowData!.varsJson['endpoint']" />
                        </el-form-item>
                        <el-form-item
                            v-if="dialogData.rowData!.type === 'KODO'"
                            :label="$t('setting.domain')"
                            prop="varsJson.domain"
                            :rules="Rules.requiredInput"
                        >
                            <el-input v-model.trim="dialogData.rowData!.varsJson['domain']" />
                            <span class="input-help">{{ $t('setting.domainHelper') }}</span>
                        </el-form-item>
                        <el-form-item
                            v-if="dialogData.rowData!.type === 'MINIO'"
                            label="Endpoint"
                            prop="varsJson.endpointItem"
                            :rules="Rules.requiredInput"
                        >
                            <el-input v-model="dialogData.rowData!.varsJson['endpointItem']">
                                <template #prepend>
                                    <el-select v-model.trim="endpoints" style="width: 80px">
                                        <el-option label="http" value="http" />
                                        <el-option label="https" value="https" />
                                    </el-select>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item
                            v-if="dialogData.rowData!.type !== '' && hasBucket(dialogData.rowData!.type)"
                            label="Bucket"
                            prop="bucket"
                        >
                            <el-select
                                style="width: 80%"
                                @change="errBuckets = false"
                                v-model="dialogData.rowData!.bucket"
                            >
                                <el-option v-for="item in buckets" :key="item" :value="item" />
                            </el-select>
                            <el-button style="width: 20%" plain @click="getBuckets(formRef)">
                                {{ $t('setting.loadBucket') }}
                            </el-button>
                            <span v-if="errBuckets" class="input-error">{{ $t('commons.rule.requiredSelect') }}</span>
                        </el-form-item>
                        <el-form-item
                            v-if="dialogData.rowData!.type === 'COS'"
                            :label="$t('setting.scType')"
                            prop="varsJson.scType"
                            :rules="[Rules.requiredSelect]"
                        >
                            <el-select v-model="dialogData.rowData!.varsJson['scType']">
                                <el-option value="Standard" :label="$t('setting.scStandard')" />
                                <el-option value="Standard_IA" :label="$t('setting.scStandard_IA')" />
                                <el-option value="Archive" :label="$t('setting.scArchive')" />
                                <el-option value="Deep_Archive" :label="$t('setting.scDeep_Archive')" />
                            </el-select>
                            <el-alert
                                v-if="dialogData.rowData!.varsJson['scType'] === 'Archive' || dialogData.rowData!.varsJson['scType'] === 'Deep_Archive'"
                                style="margin-top: 10px"
                                :closable="false"
                                type="warning"
                                :title="$t('setting.archiveHelper')"
                            />
                        </el-form-item>
                        <el-form-item
                            v-if="dialogData.rowData!.type === 'OSS'"
                            :label="$t('setting.scType')"
                            prop="varsJson.scType"
                            :rules="[Rules.requiredSelect]"
                        >
                            <el-select v-model="dialogData.rowData!.varsJson['scType']">
                                <el-option value="Standard" :label="$t('setting.scStandard')" />
                                <el-option value="IA" :label="$t('setting.scStandard_IA')" />
                                <el-option value="Archive" :label="$t('setting.scArchive')" />
                                <el-option value="ColdArchive" :label="$t('setting.scDeep_Archive')" />
                            </el-select>
                            <el-alert
                                v-if="dialogData.rowData!.varsJson['scType'] === 'Archive' || dialogData.rowData!.varsJson['scType'] === 'ColdArchive'"
                                style="margin-top: 10px"
                                :closable="false"
                                type="warning"
                                :title="$t('setting.archiveHelper')"
                            />
                        </el-form-item>
                        <el-form-item
                            v-if="dialogData.rowData!.type === 'S3'"
                            :label="$t('setting.scType')"
                            prop="varsJson.scType"
                            :rules="[Rules.requiredSelect]"
                        >
                            <el-select v-model="dialogData.rowData!.varsJson['scType']">
                                <el-option value="STANDARD" :label="$t('setting.scStandard')" />
                                <el-option value="STANDARD_IA" :label="$t('setting.scStandard_IA')" />
                                <el-option value="GLACIER" :label="$t('setting.scArchive')" />
                                <el-option value="DEEP_ARCHIVE" :label="$t('setting.scDeep_Archive')" />
                            </el-select>
                            <el-alert
                                v-if="dialogData.rowData!.varsJson['scType'] === 'Archive' || dialogData.rowData!.varsJson['scType'] === 'ColdArchive'"
                                style="margin-top: 10px"
                                :closable="false"
                                type="warning"
                                :title="$t('setting.archiveHelper')"
                            />
                        </el-form-item>
                        <div v-if="dialogData.rowData!.type === 'SFTP'">
                            <el-form-item :label="$t('setting.address')" prop="varsJson.address" :rules="Rules.host">
                                <el-input v-model.trim="dialogData.rowData!.varsJson['address']" />
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.port')" prop="varsJson.port" :rules="[Rules.port]">
                                <el-input-number
                                    :min="0"
                                    :max="65535"
                                    v-model.number="dialogData.rowData!.varsJson['port']"
                                />
                            </el-form-item>
                            <el-form-item
                                :label="$t('commons.login.username')"
                                prop="accessKey"
                                :rules="[Rules.requiredInput]"
                            >
                                <el-input v-model.trim="dialogData.rowData!.accessKey" />
                            </el-form-item>
                            <el-form-item
                                :label="$t('commons.login.password')"
                                prop="credential"
                                :rules="[Rules.requiredInput]"
                            >
                                <el-input
                                    type="password"
                                    clearable
                                    show-password
                                    v-model.trim="dialogData.rowData!.credential"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('setting.path')" prop="bucket" :rules="[Rules.requiredInput]">
                                <el-input v-model.trim="dialogData.rowData!.bucket" />
                            </el-form-item>
                        </div>
                        <el-form-item
                            v-if="dialogData.rowData!.type !== 'LOCAL' && dialogData.rowData!.type !== 'SFTP'"
                            :label="$t('setting.backupDir')"
                            prop="backupPath"
                        >
                            <el-input clearable v-model.trim="dialogData.rowData!.backupPath" placeholder="/1panel" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="handleClose">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Backup } from '@/api/interface/backup';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { addBackup, editBackup, getOneDriveInfo, listBucket } from '@/api/modules/setting';
import { deepCopy } from '@/utils/util';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const buckets = ref();
const errBuckets = ref();

const endpoints = ref('http');

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    title: string;
    rowData?: Backup.BackupInfo;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisiable = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    buckets.value = [];
    dialogData.value = params;
    if (dialogData.value.title === 'edit' && dialogData.value.rowData!.type === 'MINIO') {
        if (dialogData.value.rowData!.varsJson['endpoint'].indexOf('://') !== 0) {
            endpoints.value = dialogData.value.rowData!.varsJson['endpoint'].split('://')[0];
            dialogData.value.rowData!.varsJson['endpointItem'] =
                dialogData.value.rowData!.varsJson['endpoint'].split('://')[1];
        }
    }
    if (dialogData.value.title === 'create' && dialogData.value.rowData!.type === 'SFTP') {
        dialogData.value.rowData.varsJson['port'] = 22;
    }
    if (dialogData.value.rowData!.type === 'COS' || dialogData.value.rowData!.type === 'OSS') {
        if (params.title === 'create' || (params.title === 'edit' && !dialogData.value.rowData.varsJson['scType'])) {
            dialogData.value.rowData.varsJson['scType'] = 'Standard';
        }
    }
    if (dialogData.value.rowData!.type === 'S3') {
        if (params.title === 'create' || (params.title === 'edit' && !dialogData.value.rowData.varsJson['scType'])) {
            dialogData.value.rowData.varsJson['scType'] = 'STANDARD';
        }
    }
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    drawerVisiable.value = true;
};

const handleClose = () => {
    emit('search');
    drawerVisiable.value = false;
};
const jumpAzure = async () => {
    const res = await getOneDriveInfo();
    let commonUrl = `response_type=code&client_id=${res.data}&redirect_uri=http://localhost/login/authorized&scope=offline_access+Files.ReadWrite.All+User.Read`;
    if (!dialogData.value.rowData!.varsJson['isCN']) {
        window.open('https://login.microsoftonline.com/common/oauth2/v2.0/authorize?' + commonUrl, '_blank');
    } else {
        window.open('https://login.chinacloudapi.cn/common/oauth2/v2.0/authorize?' + commonUrl, '_blank');
    }
};

const loadDir = async (path: string) => {
    dialogData.value.rowData!.varsJson['dir'] = path;
};
function hasBucket(val: string) {
    return val === 'OSS' || val === 'S3' || val === 'MINIO' || val === 'COS' || val === 'KODO';
}

function hasEndpoint(val: string) {
    return val === 'OSS' || val === 'S3';
}

const toDoc = () => {
    window.open('https://1panel.cn/docs/user_manual/settings/', '_blank');
};

const getBuckets = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        let item = deepCopy(dialogData.value.rowData!.varsJson);
        if (dialogData.value.rowData!.type === 'MINIO') {
            dialogData.value.rowData!.varsJson['endpointItem'] = dialogData.value
                .rowData!.varsJson['endpointItem'].replace('https://', '')
                .replace('http://', '');
            item['endpoint'] = endpoints.value + '://' + dialogData.value.rowData!.varsJson['endpointItem'];
            item['endpointItem'] = undefined;
        }
        listBucket({
            type: dialogData.value.rowData!.type,
            vars: JSON.stringify(item),
            accessKey: dialogData.value.rowData!.accessKey,
            credential: dialogData.value.rowData!.credential,
        })
            .then((res) => {
                loading.value = false;
                buckets.value = res.data;
            })
            .catch(() => {
                buckets.value = [];
                loading.value = false;
            });
    });
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (hasBucket(dialogData.value.rowData.type) && !dialogData.value.rowData.bucket) {
        errBuckets.value = true;
        return;
    }
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!dialogData.value.rowData) return;
        if (dialogData.value.rowData!.type === 'MINIO') {
            dialogData.value.rowData!.varsJson['endpointItem'].replace('https://', '').replace('http://', '');
            dialogData.value.rowData!.varsJson['endpoint'] =
                endpoints.value + '://' + dialogData.value.rowData!.varsJson['endpointItem'];
            dialogData.value.rowData!.varsJson['endpointItem'] = undefined;
        }
        dialogData.value.rowData.vars = JSON.stringify(dialogData.value.rowData!.varsJson);
        loading.value = true;
        if (dialogData.value.title === 'create') {
            await addBackup(dialogData.value.rowData)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisiable.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
            return;
        }
        await editBackup(dialogData.value.rowData)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisiable.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
