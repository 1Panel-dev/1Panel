<template>
    <div>
        <Submenu activeName="backupaccount" />
        <el-form label-position="left" label-width="130px" :v-key="reflash">
            <el-row :gutter="20" style="margin-top: 20px">
                <el-col :span="24">
                    <el-card>
                        <template #header>
                            <svg-icon style="font-size: 7px" iconName="p-file-folder"></svg-icon>
                            <span style="font-size: 16px; font-weight: 500">&nbsp;{{ $t('setting.serverDisk') }}</span>
                            <div style="float: right">
                                <el-button round @click="onOpenDialog('edit', 'local', localData)">
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                            </div>
                        </template>
                        <el-form-item :label="$t('setting.currentPath')">
                            {{ localData.varsJson['dir'] }}
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.createdAt')">
                            {{ dateFromat(0, 0, localData.createdAt) }}
                        </el-form-item>
                    </el-card>
                </el-col>
            </el-row>
            <el-row :gutter="20" style="margin-top: 20px">
                <el-col :span="12">
                    <el-card>
                        <template #header>
                            <svg-icon style="font-size: 7px" iconName="p-MINIO"></svg-icon>
                            <span style="font-size: 16px; font-weight: 500">&nbsp;MIMIO</span>
                            <div style="float: right">
                                <el-button :disabled="minioData.id === 0" round @click="onBatchDelete(minioData)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                                <el-button
                                    round
                                    :disabled="minioData.id === 0"
                                    @click="onOpenDialog('edit', 'MINIO', minioData)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                            </div>
                        </template>
                        <div v-if="minioData.id !== 0">
                            <el-form-item label="Endpoint">
                                {{ minioData.varsJson['endpoint'] }}
                            </el-form-item>
                            <el-form-item label="Bucket">
                                {{ minioData.bucket }}
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFromat(0, 0, minioData.createdAt) }}
                            </el-form-item>
                        </div>
                        <el-alert v-else center style="height: 127px; background-color: #e2e4ec" :closable="false">
                            <el-button size="large" round plain type="primary" @click="onOpenDialog('create', 'MINIO')">
                                {{ $t('setting.createBackupAccount', ['MINIO']) }}
                            </el-button>
                        </el-alert>
                    </el-card>
                </el-col>
                <el-col :span="12">
                    <el-card>
                        <template #header>
                            <svg-icon style="font-size: 7px" iconName="p-OSS"></svg-icon>
                            <span style="font-size: 16px; font-weight: 500">&nbsp;OSS</span>
                            <div style="float: right">
                                <el-button round :disabled="ossData.id === 0" @click="onBatchDelete(ossData)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                                <el-button
                                    round
                                    :disabled="ossData.id === 0"
                                    @click="onOpenDialog('edit', 'OSS', ossData)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                            </div>
                        </template>
                        <div v-if="ossData.id !== 0">
                            <el-form-item label="Endpoint">
                                {{ ossData.varsJson['endpoint'] }}
                            </el-form-item>
                            <el-form-item label="Bucket">
                                {{ ossData.bucket }}
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFromat(0, 0, ossData.createdAt) }}
                            </el-form-item>
                        </div>
                        <el-alert v-else center style="height: 127px; background-color: #e2e4ec" :closable="false">
                            <el-button size="large" round plain type="primary" @click="onOpenDialog('create', 'OSS')">
                                {{ $t('setting.createBackupAccount', ['OSS']) }}
                            </el-button>
                        </el-alert>
                    </el-card>
                </el-col>
            </el-row>
            <el-row :gutter="20" style="margin-top: 20px">
                <el-col :span="12">
                    <el-card>
                        <template #header>
                            <svg-icon style="font-size: 7px" iconName="p-aws"></svg-icon>
                            <span style="font-size: 16px; font-weight: 500">&nbsp;{{ $t('setting.S3') }}</span>
                            <div style="float: right">
                                <el-button round :disabled="s3Data.id === 0" @click="onBatchDelete(s3Data)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                                <el-button
                                    round
                                    :disabled="s3Data.id === 0"
                                    @click="onOpenDialog('edit', 'S3', s3Data)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                            </div>
                        </template>
                        <div v-if="s3Data.id !== 0">
                            <el-form-item label="Region">
                                {{ s3Data.varsJson['region'] }}
                            </el-form-item>
                            <el-form-item label="Endpoint">
                                {{ s3Data.varsJson['endpoint'] }}
                            </el-form-item>
                            <el-form-item label="Bucket">
                                {{ s3Data.bucket }}
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFromat(0, 0, s3Data.createdAt) }}
                            </el-form-item>
                        </div>
                        <el-alert v-else center style="height: 167px; background-color: #e2e4ec" :closable="false">
                            <el-button size="large" round plain type="primary" @click="onOpenDialog('create', 'S3')">
                                {{ $t('setting.createBackupAccount', ['S3']) }}
                            </el-button>
                        </el-alert>
                    </el-card>
                </el-col>
                <el-col :span="12">
                    <el-card>
                        <template #header>
                            <svg-icon style="font-size: 7px" iconName="p-SFTP"></svg-icon>
                            <span style="font-size: 16px; font-weight: 500">&nbsp;SFTP</span>
                            <div style="float: right">
                                <el-button round :disabled="sftpData.id === 0" @click="onBatchDelete(sftpData)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                                <el-button
                                    round
                                    plain
                                    :disabled="sftpData.id === 0"
                                    @click="onOpenDialog('edit', 'SFTP', sftpData)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                            </div>
                        </template>
                        <div v-if="sftpData.id !== 0">
                            <el-form-item :label="$t('setting.address')">
                                {{ sftpData.varsJson['address'] }}
                            </el-form-item>
                            <el-form-item :label="$t('setting.port')">
                                {{ sftpData.varsJson['port'] }}
                            </el-form-item>
                            <el-form-item :label="$t('setting.path')">
                                {{ sftpData.bucket }}
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFromat(0, 0, sftpData.createdAt) }}
                            </el-form-item>
                        </div>
                        <el-alert v-else center style="height: 167px; background-color: #e2e4ec" :closable="false">
                            <el-button size="large" round plain type="primary" @click="onOpenDialog('create', 'SFTP')">
                                {{ $t('setting.createBackupAccount', ['SFTP']) }}
                            </el-button>
                        </el-alert>
                    </el-card>
                </el-col>
            </el-row>
        </el-form>
        <DialogOperate ref="dialogRef" @search="search" />
    </div>
</template>
<script setup lang="ts">
import { dateFromat } from '@/utils/util';
import { onMounted, ref } from 'vue';
import { getBackupList, deleteBackup } from '@/api/modules/backup';
import DialogOperate from '@/views/setting/backup-account/operate/index.vue';
import Submenu from '@/views/setting/index.vue';
import { Backup } from '@/api/interface/backup';
import { ElForm } from 'element-plus';
import { useDeleteData } from '@/hooks/use-delete-data';

const data = ref();
const reflash = ref(false);
const localData = ref<Backup.BackupInfo>({
    id: 0,
    type: 'LOCAL',
    accessKey: '',
    bucket: '',
    credential: '',
    vars: '',
    varsJson: {
        dir: '',
    },
    createdAt: new Date(),
});
const ossData = ref<Backup.BackupInfo>({
    id: 0,
    type: 'OSS',
    accessKey: '',
    bucket: '',
    credential: '',
    vars: '',
    varsJson: {
        region: '',
        endpoint: '',
    },
    createdAt: new Date(),
});
const minioData = ref<Backup.BackupInfo>({
    id: 0,
    type: 'MINIO',
    accessKey: '',
    bucket: '',
    credential: '',
    vars: '',
    varsJson: {
        region: '',
        endpoint: '',
    },
    createdAt: new Date(),
});
const sftpData = ref<Backup.BackupInfo>({
    id: 0,
    type: 'SFTP',
    accessKey: '',
    bucket: '',
    credential: '',
    vars: '',
    varsJson: {
        address: '',
        port: 0,
    },
    createdAt: new Date(),
});
const s3Data = ref<Backup.BackupInfo>({
    id: 0,
    type: 'S3',
    accessKey: '',
    bucket: '',
    credential: '',
    vars: '',
    varsJson: {
        region: '',
        endpoint: '',
    },
    createdAt: new Date(),
});

const search = async () => {
    const res = await getBackupList();
    data.value = res.data || [];
    for (const bac of data.value) {
        if (bac.id !== 0) {
            bac.varsJson = JSON.parse(bac.vars);
        }
        switch (bac.type) {
            case 'LOCAL':
                localData.value = bac;
                break;
            case 'OSS':
                ossData.value = bac;
                break;
            case 'S3':
                s3Data.value = bac;
                break;
            case 'MINIO':
                minioData.value = bac;
                break;
            case 'SFTP':
                sftpData.value = bac;
                break;
        }
    }
};

const onBatchDelete = async (row: Backup.BackupInfo | null) => {
    let ids: Array<number> = [];
    ids.push(row.id);
    await useDeleteData(deleteBackup, { ids: ids }, 'commons.msg.delete');
    search();
};

const dialogRef = ref();
const onOpenDialog = async (
    title: string,
    accountType: string,
    rowData: Partial<Backup.BackupInfo> = {
        id: 0,
        type: accountType,
        varsJson: {},
    },
) => {
    console.log(rowData);
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

onMounted(() => {
    search();
});
</script>
