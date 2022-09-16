<template>
    <el-card style="margin-top: 20px">
        <template #header>
            <div class="card-header">
                <span>{{ $t('setting.backup') }}</span>
            </div>
        </template>
        <el-button type="primary" @click="onCreate">
            {{ $t('commons.button.create') }}
        </el-button>
        <el-row :gutter="20" class="row-box" style="margin-top: 10px; margin-bottom: 20px">
            <el-col v-for="item in data" :key="item.id" :span="8">
                <el-card class="el-card">
                    <template #header>
                        <div class="card-header">
                            <span style="font-size: 16px; font-weight: 500">
                                [{{ item.type === 'LOCAL' ? $t('setting.serverDisk') : item.type }}] {{ item.name }}
                            </span>
                            <div style="float: right">
                                <el-button @click="onEdit(item)">{{ $t('commons.button.edit') }}</el-button>
                                <el-button @click="onBatchDelete(item)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>
                    </template>
                    <el-form label-position="left" label-width="130px">
                        <el-form-item v-if="item.type === 'LOCAL'" label="Dir">
                            {{ item.varsJson['dir'] }}
                        </el-form-item>
                        <el-form-item v-if="hasBucket(item.type)" label="Access Key ID">
                            {{ item.varsJson['accessKey'] }}
                        </el-form-item>
                        <el-form-item v-if="item.type === 'S3'" label="Region">
                            {{ item.varsJson['region'] }}
                        </el-form-item>
                        <el-form-item v-if="hasBucket(item.type)" label="Endpoint">
                            {{ item.varsJson['endpoint'] }}
                        </el-form-item>
                        <el-form-item v-if="hasBucket(item.type)" label="Bucket">
                            {{ item.bucket }}
                        </el-form-item>
                        <el-form-item v-if="item.type === 'SFTP'" :label="$t('setting.address')">
                            {{ item.varsJson['address'] }}
                        </el-form-item>
                        <el-form-item v-if="item.type === 'SFTP'" :label="$t('setting.port')">
                            {{ item.varsJson['port'] }}
                        </el-form-item>
                        <el-form-item v-if="item.type === 'SFTP'" :label="$t('setting.username')">
                            {{ item.varsJson['username'] }}
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.createdAt')">
                            {{ dateFromat(0, 0, item.createdAt) }}
                        </el-form-item>
                    </el-form>
                </el-card>
            </el-col>
        </el-row>

        <el-dialog @close="search" v-model="backupVisiable" :title="$t('setting.backupAccount')" width="30%">
            <el-form ref="formRef" label-position="left" :model="form" label-width="160px">
                <el-form-item :label="$t('commons.table.name')" prop="name" :rules="Rules.name">
                    <el-input v-model="form.name" :disabled="operation === 'edit'" />
                </el-form-item>
                <el-form-item :label="$t('commons.table.type')" prop="type" :rules="Rules.requiredSelect">
                    <el-select style="width: 100%" v-model="form.type" :disabled="operation === 'edit'">
                        <el-option
                            v-for="item in typeOptions"
                            :key="item.label"
                            :value="item.value"
                            :label="item.label"
                        />
                    </el-select>
                </el-form-item>
                <el-form-item
                    v-if="form.type === 'LOCAL'"
                    label="Dir"
                    prop="varsJson['dir']"
                    :rules="Rules.requiredInput"
                >
                    <el-input v-model="form.varsJson['dir']">
                        <template #append>
                            <FileList @choose="loadDir" :dir="true"></FileList>
                        </template>
                    </el-input>
                </el-form-item>
                <el-form-item
                    v-if="hasBucket(form.type)"
                    label="Access Key ID"
                    prop="varsJson.accessKey"
                    :rules="Rules.requiredInput"
                >
                    <el-input v-model="form.varsJson['accessKey']" />
                </el-form-item>
                <el-form-item
                    v-if="hasBucket(form.type)"
                    label="Secret Access Key"
                    prop="credential"
                    :rules="Rules.requiredInput"
                >
                    <el-input show-password v-model="form.credential" />
                </el-form-item>
                <el-form-item
                    v-if="form.type === 'S3'"
                    label="Region"
                    prop="varsJson.region"
                    :rules="Rules.requiredInput"
                >
                    <el-input v-model="form.varsJson['region']" />
                </el-form-item>
                <el-form-item
                    v-if="hasBucket(form.type)"
                    label="Endpoint"
                    prop="varsJson.endpoint"
                    :rules="Rules.requiredInput"
                >
                    <el-input v-model="form.varsJson['endpoint']" />
                </el-form-item>
                <el-form-item
                    v-if="form.type !== '' && hasBucket(form.type)"
                    label="Bucket"
                    prop="bucket"
                    :rules="Rules.requiredSelect"
                >
                    <el-select style="width: 80%" v-model="form.bucket">
                        <el-option v-for="item in buckets" :key="item" :value="item" />
                    </el-select>
                    <el-button style="width: 20%" plain @click="getBuckets">
                        {{ $t('setting.loadBucket') }}
                    </el-button>
                </el-form-item>
                <div v-if="form.type === 'SFTP'">
                    <el-form-item :label="$t('setting.address')" prop="varsJson.address" :rules="Rules.requiredInput">
                        <el-input v-model="form.varsJson['address']" />
                    </el-form-item>
                    <el-form-item
                        :label="$t('setting.port')"
                        prop="varsJson.port"
                        :rules="[Rules.number, { max: 65535 }]"
                    >
                        <el-input v-model.number="form.varsJson['port']" />
                    </el-form-item>
                    <el-form-item
                        :label="$t('setting.username')"
                        prop="varsJson.username"
                        :rules="[Rules.requiredInput]"
                    >
                        <el-input v-model="form.varsJson['username']" />
                    </el-form-item>
                    <el-form-item :label="$t('setting.password')" prop="credential" :rules="[Rules.requiredInput]">
                        <el-input type="password" show-password v-model="form.credential" />
                    </el-form-item>
                    <el-form-item :label="$t('setting.path')" prop="bucket">
                        <el-input v-model="form.bucket" />
                    </el-form-item>
                </div>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="backupVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="onSubmit(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </el-card>
</template>
<script setup lang="ts">
import { dateFromat } from '@/utils/util';
import { onMounted, reactive, ref } from 'vue';
import { getBackupList, addBackup, editBackup, listBucket, deleteBackup } from '@/api/modules/backup';
import { Backup } from '@/api/interface/backup';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { ElForm, ElMessage } from 'element-plus';
import { useDeleteData } from '@/hooks/use-delete-data';
import FileList from '@/components/file-list/index.vue';

const data = ref();
const selects = ref<any>([]);
const backupVisiable = ref<boolean>(false);
const operation = ref<string>('create');

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 5,
    total: 0,
});
const backSearch = reactive({
    page: 1,
    pageSize: 5,
});

const form = reactive({
    id: 0,
    name: '',
    type: 'LOCAL',
    bucket: '',
    credential: '',
    vars: '',
    varsJson: {},
});
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const typeOptions = ref([
    { label: i18n.global.t('setting.serverDisk'), value: 'LOCAL' },
    { label: 'OSS', value: 'OSS' },
    { label: 'S3', value: 'S3' },
    { label: 'SFTP', value: 'SFTP' },
    { label: 'MINIO', value: 'MINIO' },
]);
const buckets = ref();

const search = async () => {
    backSearch.page = paginationConfig.currentPage;
    backSearch.pageSize = paginationConfig.pageSize;
    const res = await getBackupList(backSearch);
    data.value = res.data.items;
    for (const bac of data.value) {
        bac.varsJson = JSON.parse(bac.vars);
    }
    paginationConfig.total = res.data.total;
};

const onCreate = () => {
    operation.value = 'create';
    form.id = 0;
    form.name = '';
    form.type = 'LOCAL';
    form.bucket = '';
    form.credential = '';
    form.vars = '';
    form.varsJson = {};
    backupVisiable.value = true;
};

const onBatchDelete = async (row: Backup.BackupInfo | null) => {
    let ids: Array<number> = [];
    if (row === null) {
        selects.value.forEach((item: Backup.BackupInfo) => {
            ids.push(item.id);
        });
    } else {
        ids.push(row.id);
    }
    await useDeleteData(deleteBackup, { ids: ids }, 'commons.msg.delete', true);
    search();
    restForm();
};

const onEdit = (row: Backup.BackupInfo) => {
    restForm();
    form.id = row.id;
    form.name = row.name;
    form.type = row.type;
    form.bucket = row.bucket;
    form.varsJson = JSON.parse(row.vars);
    operation.value = 'edit';
    backupVisiable.value = true;
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        form.vars = JSON.stringify(form.varsJson);
        if (form.id !== 0 && operation.value === 'edit') {
            await editBackup(form);
        } else if (form.id === 0 && operation.value === 'create') {
            await addBackup(form);
        } else {
            ElMessage.success(i18n.global.t('commons.msg.notSupportOperation'));
            return;
        }
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        restForm();
        search();
        backupVisiable.value = false;
    });
};
function hasBucket(val: string) {
    return val === 'OSS' || val === 'S3' || val === 'MINIO';
}
function restForm() {
    if (formRef.value) {
        formRef.value.resetFields();
    }
}

const getBuckets = async () => {
    const res = await listBucket({ type: form.type, vars: JSON.stringify(form.varsJson), credential: form.credential });
    buckets.value = res.data;
};
const loadDir = async (path: string) => {
    console.log(path);
    form.varsJson['dir'] = path;
};

onMounted(() => {
    search();
});
</script>
