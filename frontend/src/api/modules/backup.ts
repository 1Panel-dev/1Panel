import http from '@/api';
import { deepCopy } from '@/utils/util';
import { Base64 } from 'js-base64';
import { ResPage } from '../interface';
import { Backup } from '../interface/backup';
import { TimeoutEnum } from '@/enums/http-enum';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

// backup-agent
export const getLocalBackupDir = (node?: string) => {
    const params = node ? `?operateNode=${node}` : '';
    return http.get<string>(`/backups/local${params}`);
};
export const searchBackup = (params: Backup.SearchWithType) => {
    return http.post<ResPage<Backup.BackupInfo>>(`/backups/search`, params);
};
export const checkBackup = (params: Backup.BackupOperate) => {
    let request = deepCopy(params) as Backup.BackupOperate;
    if (request.accessKey) {
        request.accessKey = Base64.encode(request.accessKey);
    }
    if (request.credential) {
        request.credential = Base64.encode(request.credential);
    }
    if (!params.isPublic || !globalStore.isProductPro) {
        return http.postLocalNode<Backup.CheckResult>(`/backups/conn/check`, request);
    }
    return http.post<Backup.CheckResult>(`/backups/conn/check`, request);
};
export const listBucket = (params: Backup.ForBucket) => {
    let request = deepCopy(params) as Backup.BackupOperate;
    if (request.accessKey) {
        request.accessKey = Base64.encode(request.accessKey);
    }
    if (request.credential) {
        request.credential = Base64.encode(request.credential);
    }
    if (!params.isPublic || !globalStore.isProductPro) {
        return http.postLocalNode('/backups/buckets', request, TimeoutEnum.T_40S);
    }
    return http.post('/backups/buckets', request, TimeoutEnum.T_40S);
};
export const handleBackup = (params: Backup.Backup, node?: string) => {
    const query = node ? `?operateNode=${node}` : '';
    return http.post(`/backups/backup${query}`, params, TimeoutEnum.T_10M);
};
export const listBackupOptions = () => {
    return http.get<Array<Backup.BackupOption>>(`/backups/options`);
};
export const handleRecover = (params: Backup.Recover, node?: string) => {
    const query = node ? `?operateNode=${node}` : '';
    return http.post(`/backups/recover${query}`, params, TimeoutEnum.T_10M);
};
export const handleRecoverByUpload = (params: Backup.Recover) => {
    return http.post(`/backups/recover/byupload`, params, TimeoutEnum.T_10M);
};
export const downloadBackupRecord = (params: Backup.RecordDownload, node?: string) => {
    const query = node ? `?operateNode=${node}` : '';
    return http.post<string>(`/backups/record/download${query}`, params, TimeoutEnum.T_10M);
};
export const deleteBackupRecord = (params: { ids: number[] }, node?: string) => {
    const query = node ? `?operateNode=${node}` : '';
    return http.post(`/backups/record/del${query}`, params);
};
export const updateRecordDescription = (id: Number, description: String, node?: string) => {
    const query = node ? `?operateNode=${node}` : '';
    return http.post(`/backups/record/description/update${query}`, { id: id, description: description });
};
export const uploadByRecover = (filePath: string, targetDir: String) => {
    return http.post(`/backups/upload`, { filePath: filePath, targetDir: targetDir });
};
export const searchBackupRecords = (params: Backup.SearchBackupRecord, node?: string) => {
    const query = node ? `?operateNode=${node}` : '';
    return http.post<ResPage<Backup.RecordInfo>>(`/backups/record/search${query}`, params, TimeoutEnum.T_5M);
};
export const loadRecordSize = (param: Backup.SearchForSize, node?: string) => {
    const query = node ? `?operateNode=${node}` : '';
    return http.post<Array<Backup.RecordFileSize>>(`/backups/record/size${query}`, param);
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
    return http.post(urlItem, request, TimeoutEnum.T_60S);
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
    return http.post(urlItem, request, TimeoutEnum.T_60S);
};
export const deleteBackup = (params: { id: number; name: string; isPublic: boolean }) => {
    if (!params.isPublic) {
        return http.post('/backups/del', { id: params.id });
    }
    return http.post('/core/backups/del', { name: params.name });
};
