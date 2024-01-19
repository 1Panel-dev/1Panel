<template>
    <div>
        <LayoutContent :title="$t('commons.button.backup')">
            <template #main>
                <el-form label-width="130px" :v-key="refresh">
                    <el-row :gutter="20">
                        <el-col :span="24">
                            <div>
                                <svg-icon class="card-logo" iconName="p-file-folder"></svg-icon>
                                <span class="card-title">&nbsp;{{ $t('setting.LOCAL') }}</span>
                                <div style="float: right">
                                    <el-button round @click="onOpenDialog('edit', 'LOCAL', localData)">
                                        {{ $t('commons.button.edit') }}
                                    </el-button>
                                </div>
                            </div>
                            <el-divider class="divider" />
                            <div style="margin-left: 20px">
                                <el-form-item :label="$t('setting.currentPath')">
                                    {{ localData.varsJson['dir'] }}
                                </el-form-item>
                                <el-form-item :label="$t('commons.table.createdAt')">
                                    {{ dateFormat(0, 0, localData.createdAt) }}
                                </el-form-item>
                            </div>
                        </el-col>
                    </el-row>
                </el-form>

                <div class="common-div">
                    <span class="card-title">{{ $t('setting.thirdParty') }}</span>
                </div>

                <el-alert type="info" :closable="false" class="common-div">
                    <template #default>
                        <div style="margin-bottom: 3px"><span v-html="$t('setting.backupAlert')"></span></div>
                    </template>
                </el-alert>

                <el-row :gutter="20" class="common-div">
                    <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                        <div>
                            <svg-icon class="card-logo" iconName="p-aws"></svg-icon>
                            <span class="card-title">&nbsp;{{ $t('setting.S3') }}</span>
                            <div style="float: right">
                                <el-button
                                    round
                                    :disabled="s3Data.id === 0"
                                    @click="onOpenDialog('edit', 'S3', s3Data)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                                <el-button round :disabled="s3Data.id === 0" @click="onDelete(s3Data)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                            <el-divider class="divider" />
                        </div>
                        <div v-if="s3Data.id !== 0" style="margin-left: 20px">
                            <el-form-item label="Region">
                                {{ s3Data.varsJson['region'] }}
                            </el-form-item>
                            <el-form-item label="Endpoint">
                                {{ s3Data.varsJson['endpoint'] }}
                            </el-form-item>
                            <el-form-item label="Bucket">
                                {{ s3Data.bucket }}
                            </el-form-item>
                            <el-form-item :label="$t('setting.scType')">
                                <span v-if="!s3Data.varsJson['scType'] || s3Data.varsJson['scType'] === 'STANDARD'">
                                    {{ $t('setting.typeStandard') }}
                                </span>
                                <span v-if="s3Data.varsJson['scType'] === 'STANDARD_IA'">
                                    {{ $t('setting.typeStandard_IA') }}
                                </span>
                                <span v-if="s3Data.varsJson['scType'] === 'GLACIER'">
                                    {{ $t('setting.typeArchive') }}
                                </span>
                                <span v-if="s3Data.varsJson['scType'] === 'DEEP_ARCHIVE'">
                                    {{ $t('setting.typeDeep_Archive') }}
                                </span>
                            </el-form-item>
                            <el-form-item :label="$t('setting.backupDir')">
                                <span v-if="s3Data.backupPath">{{ s3Data.backupPath }}</span>
                                <span v-else>{{ $t('setting.unSetting') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFormat(0, 0, s3Data.createdAt) }}
                            </el-form-item>
                        </div>
                        <el-alert v-else center class="alert" style="height: 257px" :closable="false">
                            <el-button size="large" round plain type="primary" @click="onOpenDialog('create', 'S3')">
                                {{ $t('setting.createBackupAccount', [$t('setting.S3')]) }}
                            </el-button>
                        </el-alert>
                    </el-col>
                    <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                        <div>
                            <svg-icon class="card-logo" iconName="p-oss"></svg-icon>
                            <span class="card-title">&nbsp;{{ $t('setting.OSS') }}</span>
                            <div style="float: right">
                                <el-button
                                    round
                                    :disabled="ossData.id === 0"
                                    @click="onOpenDialog('edit', 'OSS', ossData)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                                <el-button round :disabled="ossData.id === 0" @click="onDelete(ossData)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>

                        <el-divider class="divider" />
                        <div v-if="ossData.id !== 0" style="margin-left: 20px">
                            <el-form-item label="Endpoint">
                                {{ ossData.varsJson['endpoint'] }}
                            </el-form-item>
                            <el-form-item label="Bucket">
                                {{ ossData.bucket }}
                            </el-form-item>
                            <el-form-item :label="$t('setting.scType')">
                                <span v-if="!ossData.varsJson['scType'] || ossData.varsJson['scType'] === 'Standard'">
                                    {{ $t('setting.typeStandard') }}
                                </span>
                                <span v-if="ossData.varsJson['scType'] === 'IA'">
                                    {{ $t('setting.typeStandard_IA') }}
                                </span>
                                <span v-if="ossData.varsJson['scType'] === 'Archive'">
                                    {{ $t('setting.typeArchive') }}
                                </span>
                                <span v-if="ossData.varsJson['scType'] === 'ColdArchive'">
                                    {{ $t('setting.typeDeep_Archive') }}
                                </span>
                            </el-form-item>
                            <el-form-item :label="$t('setting.backupDir')">
                                <span v-if="ossData.backupPath">{{ ossData.backupPath }}</span>
                                <span v-else>{{ $t('setting.unSetting') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFormat(0, 0, ossData.createdAt) }}
                            </el-form-item>
                        </div>
                        <el-alert v-else center class="alert" style="height: 257px" :closable="false">
                            <el-button size="large" round plain type="primary" @click="onOpenDialog('create', 'OSS')">
                                {{ $t('setting.createBackupAccount', [$t('setting.OSS')]) }}
                            </el-button>
                        </el-alert>
                    </el-col>
                </el-row>
                <el-row :gutter="20" class="common-div">
                    <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                        <div>
                            <svg-icon class="card-logo" iconName="p-tengxunyun1"></svg-icon>
                            <span class="card-title">&nbsp;{{ $t('setting.COS') }}</span>
                            <div style="float: right">
                                <el-button
                                    round
                                    :disabled="cosData.id === 0"
                                    @click="onOpenDialog('edit', 'COS', cosData)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                                <el-button round :disabled="cosData.id === 0" @click="onDelete(cosData)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                            <el-divider class="divider" />
                        </div>
                        <div v-if="cosData.id !== 0" style="margin-left: 20px">
                            <el-form-item label="Region">
                                {{ cosData.varsJson['region'] }}
                            </el-form-item>
                            <el-form-item label="Bucket">
                                {{ cosData.bucket }}
                            </el-form-item>
                            <el-form-item :label="$t('setting.scType')">
                                <span v-if="!cosData.varsJson['scType'] || cosData.varsJson['scType'] === 'Standard'">
                                    {{ $t('setting.typeStandard') }}
                                </span>
                                <span v-if="cosData.varsJson['scType'] === 'Standard_IA'">
                                    {{ $t('setting.typeStandard_IA') }}
                                </span>
                                <span v-if="cosData.varsJson['scType'] === 'Archive'">
                                    {{ $t('setting.typeArchive') }}
                                </span>
                                <span v-if="cosData.varsJson['scType'] === 'Deep_Archive'">
                                    {{ $t('setting.typeDeep_Archive') }}
                                </span>
                            </el-form-item>
                            <el-form-item :label="$t('setting.backupDir')">
                                <span v-if="cosData.backupPath">{{ cosData.backupPath }}</span>
                                <span v-else>{{ $t('setting.unSetting') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFormat(0, 0, cosData.createdAt) }}
                            </el-form-item>
                        </div>
                        <el-alert v-else center class="alert" style="height: 257px" :closable="false">
                            <el-button size="large" round plain type="primary" @click="onOpenDialog('create', 'COS')">
                                {{ $t('setting.createBackupAccount', [$t('setting.COS')]) }}
                            </el-button>
                        </el-alert>
                    </el-col>
                    <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                        <div>
                            <svg-icon class="card-logo" iconName="p-onedrive"></svg-icon>
                            <span class="card-title">&nbsp;{{ $t('setting.OneDrive') }}</span>
                            <div style="float: right">
                                <el-button
                                    round
                                    plain
                                    :disabled="oneDriveData.id === 0"
                                    @click="onOpenDialog('edit', 'OneDrive', oneDriveData)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                                <el-button round :disabled="oneDriveData.id === 0" @click="onDelete(oneDriveData)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>
                        <el-divider class="divider" />
                        <div v-if="oneDriveData.id !== 0" style="margin-left: 20px">
                            <el-form-item :label="$t('setting.backupDir')">
                                <span v-if="oneDriveData.backupPath">{{ oneDriveData.backupPath }}</span>
                                <span v-else>{{ $t('setting.unSetting') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('setting.refreshTime')">
                                <span>{{ oneDriveData.varsJson['refresh_time'] }}</span>
                                <el-button @click="refreshToken" link type="primary" class="ml-2">
                                    {{ $t('commons.button.refresh') }}
                                </el-button>
                            </el-form-item>
                            <el-form-item :label="$t('setting.refreshStatus')">
                                <el-tag v-if="oneDriveData.varsJson['refresh_status'] === 'Success'" type="success">
                                    {{ $t('commons.status.success') }}
                                </el-tag>
                                <el-tooltip
                                    v-if="oneDriveData.varsJson['refresh_status'] === 'Failed'"
                                    :content="oneDriveData.varsJson['refresh_msg']"
                                    placement="top"
                                >
                                    <el-tag type="danger">
                                        {{ $t('commons.status.failed') }}
                                    </el-tag>
                                </el-tooltip>
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFormat(0, 0, oneDriveData.createdAt) }}
                            </el-form-item>
                        </div>
                        <el-alert v-else center class="alert" style="height: 257px" :closable="false">
                            <el-button
                                size="large"
                                round
                                plain
                                type="primary"
                                @click="onOpenDialog('create', 'OneDrive')"
                            >
                                {{ $t('setting.createBackupAccount', [$t('setting.OneDrive')]) }}
                            </el-button>
                        </el-alert>
                    </el-col>
                </el-row>
                <el-row :gutter="20" style="margin-top: 20px">
                    <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                        <div>
                            <svg-icon class="card-logo" iconName="p-qiniuyun"></svg-icon>
                            <span class="card-title">&nbsp;{{ $t('setting.KODO') }}</span>
                            <div style="float: right">
                                <el-button
                                    round
                                    :disabled="kodoData.id === 0"
                                    @click="onOpenDialog('edit', 'KODO', kodoData)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                                <el-button round :disabled="kodoData.id === 0" @click="onDelete(kodoData)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>

                        <el-divider class="divider" />
                        <div v-if="kodoData.id !== 0" style="margin-left: 20px">
                            <el-form-item :label="$t('setting.domain')">
                                {{ kodoData.varsJson['domain'] }}
                            </el-form-item>
                            <el-form-item label="Bucket">
                                {{ kodoData.bucket }}
                            </el-form-item>
                            <el-form-item :label="$t('setting.backupDir')">
                                <span v-if="kodoData.backupPath">{{ kodoData.backupPath }}</span>
                                <span v-else>{{ $t('setting.unSetting') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFormat(0, 0, kodoData.createdAt) }}
                            </el-form-item>
                        </div>
                        <el-alert v-else center class="alert" style="height: 257px" :closable="false">
                            <el-button size="large" round plain type="primary" @click="onOpenDialog('create', 'KODO')">
                                {{ $t('setting.createBackupAccount', [$t('setting.KODO')]) }}
                            </el-button>
                        </el-alert>
                    </el-col>
                    <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                        <div>
                            <svg-icon class="card-logo" iconName="p-minio"></svg-icon>
                            <span class="card-title">&nbsp;MINIO</span>
                            <div style="float: right">
                                <el-button
                                    round
                                    :disabled="minioData.id === 0"
                                    @click="onOpenDialog('edit', 'MINIO', minioData)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                                <el-button :disabled="minioData.id === 0" round @click="onDelete(minioData)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>
                        <el-divider class="divider" />
                        <div v-if="minioData.id !== 0" style="margin-left: 20px">
                            <el-form-item label="Endpoint">
                                {{ minioData.varsJson['endpoint'] }}
                            </el-form-item>
                            <el-form-item label="Bucket">
                                {{ minioData.bucket }}
                            </el-form-item>
                            <el-form-item :label="$t('setting.backupDir')">
                                <span v-if="minioData.backupPath">{{ minioData.backupPath }}</span>
                                <span v-else>{{ $t('setting.unSetting') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFormat(0, 0, minioData.createdAt) }}
                            </el-form-item>
                        </div>
                        <el-alert v-else center class="alert" style="height: 257px" :closable="false">
                            <el-button size="large" round plain type="primary" @click="onOpenDialog('create', 'MINIO')">
                                {{ $t('setting.createBackupAccount', [$t('setting.MINIO')]) }}
                            </el-button>
                        </el-alert>
                    </el-col>
                </el-row>
                <el-row :gutter="20" style="margin-top: 20px">
                    <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                        <div>
                            <svg-icon class="card-logo" iconName="p-SFTP"></svg-icon>
                            <span class="card-title">&nbsp;SFTP</span>
                            <div style="float: right">
                                <el-button
                                    round
                                    plain
                                    :disabled="sftpData.id === 0"
                                    @click="onOpenDialog('edit', 'SFTP', sftpData)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                                <el-button round :disabled="sftpData.id === 0" @click="onDelete(sftpData)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>
                        <el-divider class="divider" />
                        <div v-if="sftpData.id !== 0" style="margin-left: 20px">
                            <el-form-item :label="$t('setting.address')">
                                {{ sftpData.varsJson['address'] }}
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.port')">
                                {{ sftpData.varsJson['port'] }}
                            </el-form-item>
                            <el-form-item :label="$t('setting.path')">
                                {{ sftpData.bucket }}
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFormat(0, 0, sftpData.createdAt) }}
                            </el-form-item>
                        </div>
                        <el-alert v-else center class="alert" style="height: 257px" :closable="false">
                            <el-button size="large" round plain type="primary" @click="onOpenDialog('create', 'SFTP')">
                                {{ $t('setting.createBackupAccount', [$t('setting.SFTP')]) }}
                            </el-button>
                        </el-alert>
                    </el-col>
                    <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                        <div>
                            <svg-icon class="card-logo" iconName="p-webdav"></svg-icon>
                            <span class="card-title">&nbsp;WebDAV</span>
                            <div style="float: right">
                                <el-button
                                    round
                                    plain
                                    :disabled="webDAVData.id === 0"
                                    @click="onOpenDialog('edit', 'WebDAV', webDAVData)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                                <el-button round :disabled="webDAVData.id === 0" @click="onDelete(webDAVData)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>
                        <el-divider class="divider" />
                        <div v-if="webDAVData.id !== 0" style="margin-left: 20px">
                            <el-form-item :label="$t('setting.address')">
                                {{ webDAVData.varsJson['address'] }}
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.port')">
                                {{ webDAVData.varsJson['port'] }}
                            </el-form-item>
                            <el-form-item :label="$t('setting.path')">
                                {{ webDAVData.bucket }}
                            </el-form-item>
                            <el-form-item :label="$t('commons.table.createdAt')">
                                {{ dateFormat(0, 0, webDAVData.createdAt) }}
                            </el-form-item>
                        </div>
                        <el-alert v-else center class="alert" style="height: 257px" :closable="false">
                            <el-button
                                size="large"
                                round
                                plain
                                type="primary"
                                @click="onOpenDialog('create', 'WebDAV')"
                            >
                                {{ $t('setting.createBackupAccount', ['WebDAV']) }}
                            </el-button>
                        </el-alert>
                    </el-col>
                </el-row>
            </template>
        </LayoutContent>

        <localDialog ref="localRef" @search="search" />
        <s3Dialog ref="s3Ref" @search="search" />
        <ossDialog ref="ossRef" @search="search" />
        <cosDialog ref="cosRef" @search="search" />
        <oneDriveDialog ref="oneDriveRef" @search="search" />
        <kodoDialog ref="kodoRef" @search="search" />
        <minioDialog ref="minioRef" @search="search" />
        <sftpDialog ref="sftpRef" @search="search" />
        <webDavDialog ref="webDavRef" @search="search" />
        <OpDialog ref="opRef" @search="search" />
    </div>
</template>
<script setup lang="ts">
import { dateFormat } from '@/utils/util';
import { onMounted, ref } from 'vue';
import OpDialog from '@/components/del-dialog/index.vue';
import { getBackupList, deleteBackup, refreshOneDrive } from '@/api/modules/setting';
import localDialog from '@/views/setting/backup-account/local/index.vue';
import s3Dialog from '@/views/setting/backup-account/s3/index.vue';
import ossDialog from '@/views/setting/backup-account/oss/index.vue';
import cosDialog from '@/views/setting/backup-account/cos/index.vue';
import oneDriveDialog from '@/views/setting/backup-account/onedrive/index.vue';
import kodoDialog from '@/views/setting/backup-account/kodo/index.vue';
import minioDialog from '@/views/setting/backup-account/minio/index.vue';
import sftpDialog from '@/views/setting/backup-account/sftp/index.vue';
import webDavDialog from '@/views/setting/backup-account/webdav/index.vue';
import { Backup } from '@/api/interface/backup';
import { ElForm } from 'element-plus';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const data = ref();
const opRef = ref();
const refresh = ref(false);

const localRef = ref();
const s3Ref = ref();
const ossRef = ref();
const cosRef = ref();
const oneDriveRef = ref();
const kodoRef = ref();
const minioRef = ref();
const sftpRef = ref();
const webDavRef = ref();

const localData = ref<Backup.BackupInfo>({
    id: 0,
    type: 'LOCAL',
    accessKey: '',
    bucket: '',
    credential: '',
    backupPath: '',
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
    backupPath: '',
    vars: '',
    varsJson: {
        endpoint: '',
        scType: 'Standard',
    },
    createdAt: new Date(),
});
const minioData = ref<Backup.BackupInfo>({
    id: 0,
    type: 'MINIO',
    accessKey: '',
    bucket: '',
    credential: '',
    backupPath: '',
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
    backupPath: '',
    vars: '',
    varsJson: {
        address: '',
        port: 22,
    },
    createdAt: new Date(),
});
const webDAVData = ref<Backup.BackupInfo>({
    id: 0,
    type: 'WebDAV',
    accessKey: '',
    bucket: '',
    credential: '',
    backupPath: '',
    vars: '',
    varsJson: {
        address: '',
        port: 10080,
    },
    createdAt: new Date(),
});
const oneDriveData = ref<Backup.BackupInfo>({
    id: 0,
    type: 'OneDrive',
    accessKey: '',
    bucket: '',
    credential: '',
    backupPath: '',
    vars: '',
    varsJson: {
        refresh_msg: '',
        refresh_time: '',
        refresh_status: '',
    },
    createdAt: new Date(),
});
const s3Data = ref<Backup.BackupInfo>({
    id: 0,
    type: 'S3',
    accessKey: '',
    bucket: '',
    credential: '',
    backupPath: '',
    vars: '',
    varsJson: {
        region: '',
        scType: 'Standard',
        endpoint: '',
    },
    createdAt: new Date(),
});
const cosData = ref<Backup.BackupInfo>({
    id: 0,
    type: 'COS',
    accessKey: '',
    bucket: '',
    credential: '',
    backupPath: '',
    vars: '',
    varsJson: {
        region: '',
        scType: 'Standard',
        endpoint: '',
    },
    createdAt: new Date(),
});
const kodoData = ref<Backup.BackupInfo>({
    id: 0,
    type: 'KODO',
    accessKey: '',
    bucket: '',
    credential: '',
    backupPath: '',
    vars: '',
    varsJson: {
        domain: '',
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
            case 'COS':
                cosData.value = bac;
                break;
            case 'KODO':
                kodoData.value = bac;
                break;
            case 'OneDrive':
                oneDriveData.value = bac;
                break;
            case 'WebDAV':
                webDAVData.value = bac;
                break;
        }
    }
};

const onDelete = async (row: Backup.BackupInfo) => {
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [row.type],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('setting.backupAccount'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteBackup,
        params: { id: row.id },
    });
};

const onOpenDialog = async (
    title: string,
    accountType: string,
    rowData: Partial<Backup.BackupInfo> = {
        id: 0,
        type: accountType,
        varsJson: {},
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    switch (accountType) {
        case 'LOCAL':
            localRef.value.acceptParams(params);
            return;
        case 'S3':
            s3Ref.value.acceptParams(params);
            return;
        case 'OSS':
            ossRef.value.acceptParams(params);
            return;
        case 'COS':
            cosRef.value.acceptParams(params);
            return;
        case 'OneDrive':
            oneDriveRef.value.acceptParams(params);
            return;
        case 'KODO':
            kodoRef.value.acceptParams(params);
            return;
        case 'MINIO':
            minioRef.value.acceptParams(params);
            return;
        case 'SFTP':
            sftpRef.value.acceptParams(params);
            return;
        case 'WebDAV':
            webDavRef.value.acceptParams(params);
            return;
    }
};

const refreshToken = async () => {
    await refreshOneDrive();
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    search();
};

onMounted(() => {
    search();
});
</script>

<style scoped lang="scss">
.divider {
    display: block;
    height: 1px;
    width: 100%;
    margin: 12px 0;
    border-top: 1px var(--el-border-color) var(--el-border-style);
}
.alert {
    background-color: rgba(0, 94, 235, 0.03);
}

.common-div {
    margin-top: 20px;
}

.card-title {
    font-size: 14px;
    font-weight: 500;
    line-height: 25px;
}
.card-logo {
    font-size: 7px;
}
</style>
