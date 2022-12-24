<template>
    <div>
        <Submenu activeName="backupaccount" />
        <el-card style="margin-top: 20px">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('setting.backup') }}</span>
                </div>
            </template>
            <el-button type="primary" icon="Plus" @click="onOpenDialog('create')">
                {{ $t('commons.button.create') }}
            </el-button>
            <el-row :gutter="20" class="row-box">
                <el-col v-for="item in data" :key="item.id" :span="8" style="margin-top: 20px">
                    <el-card class="el-card">
                        <template #header>
                            <div class="card-header">
                                <svg-icon style="font-size: 7px" :iconName="loadIconName(item.type)"></svg-icon>
                                <span style="font-size: 16px; font-weight: 500">
                                    &nbsp;{{ loadBackupName(item.type) }}
                                </span>
                                <div style="float: right">
                                    <el-button @click="onOpenDialog('edit', item)">
                                        {{ $t('commons.button.edit') }}
                                    </el-button>
                                    <el-button v-if="item.type !== 'LOCAL'" @click="onBatchDelete(item)">
                                        {{ $t('commons.button.delete') }}
                                    </el-button>
                                </div>
                            </div>
                        </template>
                        <el-form label-position="left" label-width="130px">
                            <el-form-item v-if="item.type === 'LOCAL'" :label="$t('setting.currentPath')">
                                {{ item.varsJson['dir'] }}
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
                            <el-form-item v-if="item.type === 'SFTP'" :label="$t('setting.path')">
                                {{ item.bucket }}
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFromat(0, 0, item.createdAt) }}
                            </el-form-item>
                        </el-form>
                    </el-card>
                </el-col>
            </el-row>
        </el-card>
        <DialogOperate ref="dialogRef" @search="search" />
    </div>
</template>
<script setup lang="ts">
import { dateFromat } from '@/utils/util';
import { onMounted, ref } from 'vue';
import { loadBackupName } from '@/views/setting/helper';
import { getBackupList, deleteBackup } from '@/api/modules/backup';
import DialogOperate from '@/views/setting/backup-account/operate/index.vue';
import Submenu from '@/views/setting/index.vue';
import { Backup } from '@/api/interface/backup';
import { ElForm } from 'element-plus';
import { useDeleteData } from '@/hooks/use-delete-data';

const data = ref();

const search = async () => {
    const res = await getBackupList();
    data.value = res.data;
    for (const bac of data.value) {
        bac.varsJson = JSON.parse(bac.vars);
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
    rowData: Partial<Backup.BackupInfo> = {
        id: 0,
        varsJson: {},
    },
) => {
    let types = [] as Array<string>;
    for (const item of data.value) {
        types.push(item.type);
    }
    let params = {
        title,
        types,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

function hasBucket(val: string) {
    return val === 'OSS' || val === 'S3' || val === 'MINIO';
}

const loadIconName = (type: string) => {
    switch (type) {
        case 'OSS':
            return 'p-oss';
            break;
        case 'S3':
            return 'p-aws';
            break;
        case 'SFTP':
            return 'p-SFTP';
            break;
        case 'MINIO':
            return 'p-minio';
            break;
        case 'LOCAL':
            return 'p-file-folder';
            break;
    }
};
onMounted(() => {
    search();
});
</script>
