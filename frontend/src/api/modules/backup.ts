import http from '@/api';
import { deepCopy } from '@/utils/util';
import { Base64 } from 'js-base64';
import { ResPage } from '../interface';
import { Backup } from '../interface/backup';
import { TimeoutEnum } from '@/enums/http-enum';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

// backup-agent
export const getLocalBackupDir = () => {
    return http.get<string>(`/backups/local`);
};
export const searchBackup = (params: Backup.SearchWithType) => {
    return http.post<ResPage<Backup.BackupInfo>>(`/backups/search`, params);
};
export const handleBackup = (params: Backup.Backup) => {
    return http.post(`/backups/backup`, params, TimeoutEnum.T_1H);
};
export const listBackupOptions = () => {
    return http.get<Array<Backup.BackupOption>>(`/backups/options`);
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
export const updateRecordDescription = (id: Number, description: String) => {
    return http.post(`/backups/record/description/update`, { id: id, description: description });
};
export const searchBackupRecords = (params: Backup.SearchBackupRecord) => {
    return http.post<ResPage<Backup.RecordInfo>>(`/backups/record/search`, params, TimeoutEnum.T_5M);
};
export const loadRecordSize = (param: Backup.SearchForSize) => {
    return http.post<Array<Backup.RecordFileSize>>(`/backups/record/size`, param);
};
export const searchBackupRecordsByCronjob = (params: Backup.SearchBackupRecordByCronjob) => {
    return http.post<ResPage<Backup.RecordInfo>>(`/backups/record/search/bycronjob`, params, TimeoutEnum.T_5M);
};
export const getFilesFromBackup = (id: number) => {
    return http.post<Array<any>>(`/backups/search/files`, { id: id });
};

// backup-core
export const refreshToken = (params: { id: number; name: string; isPublic: boolean }) => {
    if (!params.isPublic) {
        return http.post('/backups/refresh/token', { id: params.id });
    }
    return http.post('/core/backups/refresh/token', { name: params.name });
};
export const getClientInfo = (clientType: string) => {
    return http.get<Backup.ClientInfo>(`/core/backups/client/${clientType}`);
};
export const addBackup = (params: Backup.BackupOperate) => {
    let request = deepCopy(params) as Backup.BackupOperate;
    if (request.accessKey) {
        request.accessKey = Base64.encode(request.accessKey);
    }
    if (request.credential) {
        request.credential = Base64.encode(request.credential);
    }
    let urlItem = '/core/backups';
    if (!params.isPublic) {
        urlItem = '/backups';
    }
    return http.post<Backup.BackupOperate>(urlItem, request, TimeoutEnum.T_60S);
};
export const editBackup = (params: Backup.BackupOperate) => {
    let request = deepCopy(params) as Backup.BackupOperate;
    if (request.accessKey) {
        request.accessKey = Base64.encode(request.accessKey);
    }
    if (request.credential) {
        request.credential = Base64.encode(request.credential);
    }
    let urlItem = '/core/backups/update';
    if (!params.isPublic) {
        urlItem = '/backups/update';
    }
    return http.post(urlItem, request);
};
export const deleteBackup = (params: { id: number; name: string; isPublic: boolean }) => {
    if (!params.isPublic) {
        return http.post('/backups/del', { id: params.id });
    }
    return http.post('/core/backups/del', { name: params.name });
};
export const listBucket = (params: Backup.ForBucket) => {
    let request = deepCopy(params) as Backup.BackupOperate;
    if (request.accessKey) {
        request.accessKey = Base64.encode(request.accessKey);
    }
    if (request.credential) {
        request.credential = Base64.encode(request.credential);
    }
    let urlItem = '/core/backups/buckets';
    if (!params.isPublic || !globalStore.isProductPro) {
        urlItem = '/backups/buckets';
    }
    return http.post(urlItem, request);
};
