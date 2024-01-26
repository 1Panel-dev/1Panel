import http from '@/api';
import { deepCopy } from '@/utils/util';
import { Base64 } from 'js-base64';
import { ResPage, SearchWithPage, DescriptionUpdate } from '../interface';
import { Backup } from '../interface/backup';
import { Setting } from '../interface/setting';
import { TimeoutEnum } from '@/enums/http-enum';

export const getSettingInfo = () => {
    return http.post<Setting.SettingInfo>(`/settings/search`);
};
export const getSystemAvailable = () => {
    return http.get(`/settings/search/available`);
};

export const updateSetting = (param: Setting.SettingUpdate) => {
    return http.post(`/settings/update`, param);
};

export const updatePassword = (param: Setting.PasswordUpdate) => {
    return http.post(`/settings/password/update`, param);
};

export const loadInterfaceAddr = () => {
    return http.get(`/settings/interface`);
};

export const updateBindInfo = (ipv6: string, bindAddress: string) => {
    return http.post(`/settings/bind/update`, { ipv6: ipv6, bindAddress: bindAddress });
};

export const updatePort = (param: Setting.PortUpdate) => {
    return http.post(`/settings/port/update`, param);
};

export const updateSSL = (param: Setting.SSLUpdate) => {
    return http.post(`/settings/ssl/update`, param);
};
export const loadSSLInfo = () => {
    return http.get<Setting.SSLInfo>(`/settings/ssl/info`);
};
export const downloadSSL = () => {
    return http.download<any>(`settings/ssl/download`);
};

export const handleExpired = (param: Setting.PasswordUpdate) => {
    return http.post(`/settings/expired/handle`, param);
};

export const loadTimeZone = () => {
    return http.get<Array<string>>(`/settings/time/option`);
};
export const syncTime = (ntpSite: string) => {
    return http.post<string>(`/settings/time/sync`, { ntpSite: ntpSite });
};

export const cleanMonitors = () => {
    return http.post(`/settings/monitor/clean`, {});
};

export const loadMFA = (param: Setting.MFARequest) => {
    return http.post<Setting.MFAInfo>(`/settings/mfa`, param);
};

export const loadDaemonJsonPath = () => {
    return http.get<string>(`/settings/daemonjson`, {});
};

export const bindMFA = (param: Setting.MFABind) => {
    return http.post(`/settings/mfa/bind`, param);
};

export const loadBaseDir = () => {
    return http.get<string>(`/settings/basedir`);
};

// backup
export const handleBackup = (params: Backup.Backup) => {
    return http.post(`/settings/backup/backup`, params, TimeoutEnum.T_1H);
};
export const handleRecover = (params: Backup.Recover) => {
    return http.post(`/settings/backup/recover`, params, TimeoutEnum.T_1D);
};
export const handleRecoverByUpload = (params: Backup.Recover) => {
    return http.post(`/settings/backup/recover/byupload`, params, TimeoutEnum.T_1D);
};
export const refreshOneDrive = () => {
    return http.post(`/settings/backup/refresh/onedrive`, {});
};
export const downloadBackupRecord = (params: Backup.RecordDownload) => {
    return http.post<string>(`/settings/backup/record/download`, params, TimeoutEnum.T_10M);
};
export const deleteBackupRecord = (params: { ids: number[] }) => {
    return http.post(`/settings/backup/record/del`, params);
};
export const searchBackupRecords = (params: Backup.SearchBackupRecord) => {
    return http.post<ResPage<Backup.RecordInfo>>(`/settings/backup/record/search`, params, TimeoutEnum.T_5M);
};
export const searchBackupRecordsByCronjob = (params: Backup.SearchBackupRecordByCronjob) => {
    return http.post<ResPage<Backup.RecordInfo>>(`/settings/backup/record/search/bycronjob`, params, TimeoutEnum.T_5M);
};

export const getBackupList = () => {
    return http.get<Array<Backup.BackupInfo>>(`/settings/backup/search`);
};
export const getOneDriveInfo = () => {
    return http.get<Backup.OneDriveInfo>(`/settings/backup/onedrive`);
};
export const getFilesFromBackup = (type: string) => {
    return http.post<Array<any>>(`/settings/backup/search/files`, { type: type });
};
export const addBackup = (params: Backup.BackupOperate) => {
    let request = deepCopy(params) as Backup.BackupOperate;
    if (request.accessKey) {
        request.accessKey = Base64.encode(request.accessKey);
    }
    if (request.credential) {
        request.credential = Base64.encode(request.credential);
    }
    return http.post<Backup.BackupOperate>(`/settings/backup`, request);
};
export const editBackup = (params: Backup.BackupOperate) => {
    let request = deepCopy(params) as Backup.BackupOperate;
    if (request.accessKey) {
        request.accessKey = Base64.encode(request.accessKey);
    }
    if (request.credential) {
        request.credential = Base64.encode(request.credential);
    }
    return http.post(`/settings/backup/update`, request);
};
export const deleteBackup = (params: { id: number }) => {
    return http.post(`/settings/backup/del`, params);
};
export const listBucket = (params: Backup.ForBucket) => {
    let request = deepCopy(params) as Backup.BackupOperate;
    if (request.accessKey) {
        request.accessKey = Base64.encode(request.accessKey);
    }
    if (request.credential) {
        request.credential = Base64.encode(request.credential);
    }
    return http.post(`/settings/backup/buckets`, request);
};

// snapshot
export const snapshotCreate = (param: Setting.SnapshotCreate) => {
    return http.post(`/settings/snapshot`, param);
};
export const loadSnapStatus = (id: number) => {
    return http.post<Setting.SnapshotStatus>(`/settings/snapshot/status`, { id: id });
};
export const snapshotImport = (param: Setting.SnapshotImport) => {
    return http.post(`/settings/snapshot/import`, param);
};
export const updateSnapshotDescription = (param: DescriptionUpdate) => {
    return http.post(`/settings/snapshot/description/update`, param);
};
export const snapshotDelete = (param: { ids: number[] }) => {
    return http.post(`/settings/snapshot/del`, param);
};
export const snapshotRecover = (param: Setting.SnapshotRecover) => {
    return http.post(`/settings/snapshot/recover`, param);
};
export const snapshotRollback = (param: Setting.SnapshotRecover) => {
    return http.post(`/settings/snapshot/rollback`, param);
};
export const searchSnapshotPage = (param: SearchWithPage) => {
    return http.post<ResPage<Setting.SnapshotInfo>>(`/settings/snapshot/search`, param);
};

// upgrade
export const loadUpgradeInfo = () => {
    return http.get<Setting.UpgradeInfo>(`/settings/upgrade`);
};
export const loadReleaseNotes = (version: string) => {
    return http.post<string>(`/settings/upgrade/notes`, { version: version });
};
export const upgrade = (version: string) => {
    return http.post(`/settings/upgrade`, { version: version });
};
