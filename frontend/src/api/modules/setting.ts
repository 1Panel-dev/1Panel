import http from '@/api';
import { deepCopy } from '@/utils/util';
import { Base64 } from 'js-base64';
import { ResPage, SearchWithPage, DescriptionUpdate } from '../interface';
import { Backup } from '../interface/backup';
import { Setting } from '../interface/setting';
import { TimeoutEnum } from '@/enums/http-enum';

// license
export const UploadFileData = (params: FormData) => {
    return http.upload('/licenses/upload', params);
};
export const getLicense = () => {
    return http.get<Setting.License>(`/licenses/get`);
};
export const getLicenseStatus = () => {
    return http.get<Setting.LicenseStatus>(`/licenses/get/status`);
};
export const syncLicense = () => {
    return http.post(`/licenses/sync`);
};
export const unbindLicense = () => {
    return http.post(`/licenses/unbind`);
};

// agent
export const loadBaseDir = () => {
    return http.get<string>(`/settings/basedir`);
};
export const loadDaemonJsonPath = () => {
    return http.get<string>(`/settings/daemonjson`, {});
};

// core
export const getSettingInfo = () => {
    return http.post<Setting.SettingInfo>(`/core/settings/search`);
};
export const getSystemAvailable = () => {
    return http.get(`/settings/search/available`);
};
export const updateSetting = (param: Setting.SettingUpdate) => {
    return http.post(`/core/settings/update`, param);
};
export const updateMenu = (param: Setting.SettingUpdate) => {
    return http.post(`/core/settings/menu/update`, param);
};
export const updateProxy = (params: Setting.ProxyUpdate) => {
    let request = deepCopy(params) as Setting.ProxyUpdate;
    if (request.proxyPasswd) {
        request.proxyPasswd = Base64.encode(request.proxyPasswd);
    }
    request.proxyType = request.proxyType === 'close' ? '' : request.proxyType;
    return http.post(`/core/settings/proxy/update`, request);
};
export const updatePassword = (param: Setting.PasswordUpdate) => {
    return http.post(`/core/settings/password/update`, param);
};
export const loadInterfaceAddr = () => {
    return http.get(`/core/settings/interface`);
};
export const updateBindInfo = (ipv6: string, bindAddress: string) => {
    return http.post(`/core/settings/bind/update`, { ipv6: ipv6, bindAddress: bindAddress });
};
export const updatePort = (param: Setting.PortUpdate) => {
    return http.post(`/core/settings/port/update`, param);
};
export const updateSSL = (param: Setting.SSLUpdate) => {
    return http.post(`/core/settings/ssl/update`, param);
};
export const loadSSLInfo = () => {
    return http.get<Setting.SSLInfo>(`/core/settings/ssl/info`);
};
export const downloadSSL = () => {
    return http.download<any>(`/core/settings/ssl/download`);
};
export const handleExpired = (param: Setting.PasswordUpdate) => {
    return http.post(`/core/settings/expired/handle`, param);
};
export const loadMFA = (param: Setting.MFARequest) => {
    return http.post<Setting.MFAInfo>(`/core/settings/mfa`, param);
};
export const bindMFA = (param: Setting.MFABind) => {
    return http.post(`/core/settings/mfa/bind`, param);
};

// backup-agent
export const handleBackup = (params: Backup.Backup) => {
    return http.post(`/settings/backup/backup`, params, TimeoutEnum.T_1H);
};
export const handleRecover = (params: Backup.Recover) => {
    return http.post(`/settings/backup/recover`, params, TimeoutEnum.T_1D);
};
export const handleRecoverByUpload = (params: Backup.Recover) => {
    return http.post(`/settings/backup/recover/byupload`, params, TimeoutEnum.T_1D);
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
export const getFilesFromBackup = (type: string) => {
    return http.post<Array<any>>(`/settings/backup/search/files`, { type: type });
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
export const getOneDriveInfo = () => {
    return http.get<Backup.OneDriveInfo>(`/core/backup/onedrive`);
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
export const snapshotDelete = (param: { ids: number[]; deleteWithFile: boolean }) => {
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
