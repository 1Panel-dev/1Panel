import http from '@/api';
import { Backup } from '../interface/backup';
import { ResPage, ReqPage } from '../interface';

export const getBackupList = (params: ReqPage) => {
    return http.post<ResPage<Backup.BackupInfo>>(`/backups/search`, params);
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

export const listBucket = (params: Backup.ForBucket) => {
    return http.post(`/backups/buckets`, params);
};
