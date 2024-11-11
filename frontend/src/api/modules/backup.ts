import http from '@/api';
import { deepCopy } from '@/utils/util';
import { Base64 } from 'js-base64';
import { ResPage } from '../interface';
import { Backup } from '../interface/backup';
import { TimeoutEnum } from '@/enums/http-enum';

// backup-agent
export const handleBackup = (params: Backup.Backup) => {
    return http.post(`/backups/backup`, params, TimeoutEnum.T_1H);
};
export const handleRecover = (params: Backup.Recover) => {
    return http.post(`/backups/recover`, params, TimeoutEnum.T_1D);
};
export const handleRecoverByUpload = (params: Backup.Recover) => {
    return http.post(`/backups/recover/byupload`, params, TimeoutEnum.T_1D);
};
export const downloadBackupRecord = (params: Backup.RecordDownload) => {
    return http.post<string>(`/backups/record/download`, params, TimeoutEnum.T_10M);
};
export const deleteBackupRecord = (params: { ids: number[] }) => {
    return http.post(`/backups/record/del`, params);
};
export const searchBackupRecords = (params: Backup.SearchBackupRecord) => {
    return http.post<ResPage<Backup.RecordInfo>>(`/backups/record/search`, params, TimeoutEnum.T_5M);
};
export const searchBackupRecordsByCronjob = (params: Backup.SearchBackupRecordByCronjob) => {
    return http.post<ResPage<Backup.RecordInfo>>(`/backups/record/search/bycronjob`, params, TimeoutEnum.T_5M);
};
export const getFilesFromBackup = (id: number) => {
    return http.post<Array<any>>(`/backups/search/files`, { id: id });
};

// backup-core
export const refreshOneDrive = () => {
    return http.post(`/core/backup/refresh/onedrive`, {});
};
export const getBackupList = () => {
    return http.get<Array<Backup.BackupOption>>(`/core/backup/options`);
};
export const getLocalBackupDir = () => {
    return http.get<string>(`/core/backup/local`);
};
export const searchBackup = (params: Backup.SearchWithType) => {
    return http.post<ResPage<Backup.BackupInfo>>(`/core/backup/search`, params);
};
export const getClientInfo = (clientType: string) => {
    return http.get<Backup.OneDriveInfo>(`/core/backup/client/${clientType}`);
};
export const addBackup = (params: Backup.BackupOperate) => {
    let request = deepCopy(params) as Backup.BackupOperate;
    if (request.accessKey) {
        request.accessKey = Base64.encode(request.accessKey);
    }
    if (request.credential) {
        request.credential = Base64.encode(request.credential);
    }
    return http.post<Backup.BackupOperate>(`/core/backup`, request, TimeoutEnum.T_60S);
};
export const editBackup = (params: Backup.BackupOperate) => {
    let request = deepCopy(params) as Backup.BackupOperate;
    if (request.accessKey) {
        request.accessKey = Base64.encode(request.accessKey);
    }
    if (request.credential) {
        request.credential = Base64.encode(request.credential);
    }
    return http.post(`/core/backup/update`, request);
};
export const deleteBackup = (params: { id: number }) => {
    return http.post(`/core/backup/del`, params);
};
export const listBucket = (params: Backup.ForBucket) => {
    let request = deepCopy(params) as Backup.BackupOperate;
    if (request.accessKey) {
        request.accessKey = Base64.encode(request.accessKey);
    }
    if (request.credential) {
        request.credential = Base64.encode(request.credential);
    }
    return http.post(`/core/backup/buckets`, request);
};
