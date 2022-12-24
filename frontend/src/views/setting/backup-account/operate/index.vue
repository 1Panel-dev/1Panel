<template>
    <el-dialog v-model="dialogVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="50%">
        <template #header>
            <div class="card-header">
                <span>{{ title }}{{ $t('setting.backupAccount') }}</span>
            </div>
        </template>
        <el-form ref="formRef" v-loading="loading" :model="dialogData.rowData" label-width="120px">
            <el-form-item :label="$t('commons.table.type')" prop="type" :rules="Rules.requiredSelect">
                <el-select
                    style="width: 100%"
                    v-model="dialogData.rowData!.type"
                    @change="changeType"
                    :disabled="title === $t('commons.button.edit')"
                >
                    <el-option v-for="item in typeOptions" :key="item.label" :value="item.value" :label="item.label" />
                </el-select>
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'LOCAL'"
                :label="$t('setting.currentPath')"
                prop="varsJson['dir']"
                :rules="Rules.requiredInput"
            >
                <el-input v-model="dialogData.rowData!.varsJson['dir']">
                    <template #append>
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
                <el-input v-model="dialogData.rowData!.accessKey" />
            </el-form-item>
            <el-form-item
                v-if="hasBucket(dialogData.rowData!.type)"
                label="Secret Key"
                prop="credential"
                :rules="Rules.requiredInput"
            >
                <el-input show-password v-model="dialogData.rowData!.credential" />
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'S3'"
                label="Region"
                prop="varsJson.region"
                :rules="Rules.requiredInput"
            >
                <el-input v-model="dialogData.rowData!.varsJson['region']" />
            </el-form-item>
            <el-form-item
                v-if="hasBucket(dialogData.rowData!.type) && dialogData.rowData!.type !== 'MINIO'"
                label="Endpoint"
                prop="varsJson.endpoint"
                :rules="Rules.requiredInput"
            >
                <el-input v-model="dialogData.rowData!.varsJson['endpoint']" />
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'MINIO'"
                label="Endpoint"
                prop="varsJson.endpointItem"
                :rules="Rules.requiredInput"
            >
                <el-input v-model="dialogData.rowData!.varsJson['endpointItem']">
                    <template #prepend>
                        <el-select v-model="endpoints" style="width: 80px">
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
                :rules="Rules.requiredSelect"
            >
                <el-select style="width: 80%" v-model="dialogData.rowData!.bucket">
                    <el-option v-for="item in buckets" :key="item" :value="item" />
                </el-select>
                <el-button style="width: 20%" plain @click="getBuckets">
                    {{ $t('setting.loadBucket') }}
                </el-button>
            </el-form-item>
            <div v-if="dialogData.rowData!.type === 'SFTP'">
                <el-form-item :label="$t('setting.address')" prop="varsJson.address" :rules="Rules.requiredInput">
                    <el-input v-model="dialogData.rowData!.varsJson['address']" />
                </el-form-item>
                <el-form-item :label="$t('setting.port')" prop="varsJson.port" :rules="[Rules.number]">
                    <el-input-number :min="0" :max="65535" v-model.number="dialogData.rowData!.varsJson['port']" />
                </el-form-item>
                <el-form-item :label="$t('setting.username')" prop="accessKey" :rules="[Rules.requiredInput]">
                    <el-input v-model="dialogData.rowData!.accessKey" />
                </el-form-item>
                <el-form-item :label="$t('setting.password')" prop="credential" :rules="[Rules.requiredInput]">
                    <el-input type="password" show-password v-model="dialogData.rowData!.credential" />
                </el-form-item>
                <el-form-item :label="$t('setting.path')" prop="bucket">
                    <el-input v-model="dialogData.rowData!.bucket" />
                </el-form-item>
            </div>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { Backup } from '@/api/interface/backup';
import { addBackup, editBackup, listBucket } from '@/api/modules/backup';
import { deepCopy } from '@/utils/util';

const loading = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const typeOptions = ref();
const buckets = ref();

const endpoints = ref('http');

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    title: string;
    types: Array<string>;
    rowData?: Backup.BackupInfo;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const dialogVisiable = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
    types: [],
});
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    if (dialogData.value.title === 'edit' && dialogData.value.rowData!.type === 'MINIO') {
        if (dialogData.value.rowData!.varsJson['endpoint'].indexOf('://') !== 0) {
            endpoints.value = dialogData.value.rowData!.varsJson['endpoint'].split('://')[0];
            dialogData.value.rowData!.varsJson['endpointItem'] =
                dialogData.value.rowData!.varsJson['endpoint'].split('://')[1];
        }
    }
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    loadOption(params.types);
    dialogVisiable.value = true;
};

const loadOption = (existTypes: Array<string>) => {
    let options = [
        { label: i18n.global.t('setting.serverDisk'), value: 'LOCAL' },
        { label: i18n.global.t('setting.OSS'), value: 'OSS' },
        { label: i18n.global.t('setting.S3'), value: 'S3' },
        { label: 'SFTP', value: 'SFTP' },
        { label: 'MinIO', value: 'MINIO' },
    ];
    for (const item of existTypes) {
        for (let i = 0; i < options.length; i++) {
            if (item === options[i].value) {
                options.splice(i, 1);
            }
        }
    }
    typeOptions.value = options;
};

const changeType = async (val: string) => {
    let itemType = val;
    buckets.value = [];
    if (formRef.value) {
        formRef.value.resetFields();
    }
    dialogData.value.rowData!.type = itemType;
};
const loadDir = async (path: string) => {
    dialogData.value.rowData!.varsJson['dir'] = path;
};
function hasBucket(val: string) {
    return val === 'OSS' || val === 'S3' || val === 'MINIO';
}

const getBuckets = async () => {
    loading.value = true;
    let item = deepCopy(dialogData.value.rowData!.varsJson);
    if (dialogData.value.rowData!.type === 'MINIO') {
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
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!dialogData.value.rowData) return;
        if (dialogData.value.rowData!.type === 'MINIO') {
            dialogData.value.rowData!.varsJson['endpoint'] =
                endpoints.value + '://' + dialogData.value.rowData!.varsJson['endpointItem'];
            dialogData.value.rowData!.varsJson['endpointItem'] = undefined;
        }
        dialogData.value.rowData.vars = JSON.stringify(dialogData.value.rowData!.varsJson);
        if (dialogData.value.title === 'create') {
            await addBackup(dialogData.value.rowData);
        }
        if (dialogData.value.title === 'edit') {
            await editBackup(dialogData.value.rowData);
        }

        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        emit('search');
        dialogVisiable.value = false;
    });
};

defineExpose({
    acceptParams,
});
</script>
