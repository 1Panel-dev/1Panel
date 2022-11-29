import http from '@/api';
import { Backup } from '../interface/backup';
import { ResPage } from '../interface';

export const getBackupList = () => {
    return http.get<Array<Backup.BackupInfo>>(`/backups/search`);
};

export const addBackup = (params: Backup.BackupOperate) => {
    return http.post<Backup.BackupOperate>(`/backups`, params);
};

export const editBackup = (params: Backup.BackupOperate) => {
    return http.put(`/backups/` + params.id, params);
};

export const deleteBackup = (params: { ids: number[] }) => {
    return http.post(`/backups/del`, params);
};

export const downloadBackupRecord = (params: Backup.RecordDownload) => {
    return http.download<BlobPart>(`/backups/record/download`, params, { responseType: 'blob' });
};
export const deleteBackupRecord = (params: { ids: number[] }) => {
    return http.post(`/backups/record/del`, params);
};
export const searchBackupRecords = (params: Backup.SearchBackupRecord) => {
    return http.post<ResPage<Backup.RecordInfo>>(`/backups/record/search`, params);
};

export const listBucket = (params: Backup.ForBucket) => {
    return http.post(`/backups/buckets`, params);
};
