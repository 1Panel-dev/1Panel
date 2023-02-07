import http from '@/api';
import { ReqPage, ResPage } from '../interface';
import { Setting } from '../interface/setting';

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

export const updatePort = (param: Setting.PortUpdate) => {
    return http.post(`/settings/port/update`, param);
};

export const handleExpired = (param: Setting.PasswordUpdate) => {
    return http.post(`/settings/expired/handle`, param);
};

export const syncTime = () => {
    return http.post<string>(`/settings/time/sync`, {});
};

export const cleanMonitors = () => {
    return http.post(`/settings/monitor/clean`, {});
};

export const getMFA = () => {
    return http.get<Setting.MFAInfo>(`/settings/mfa`, {});
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

// snapshot
export const snapshotCreate = (param: Setting.SnapshotCreate) => {
    return http.post(`/settings/snapshot`, param);
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
export const searchSnapshotPage = (param: ReqPage) => {
    return http.post<ResPage<Setting.SnapshotInfo>>(`/settings/snapshot/search`, param);
};

// upgrade
export const loadUpgradeInfo = () => {
    return http.get<Setting.UpgradeInfo>(`/settings/upgrade`);
};
export const upgrade = (version: string) => {
    return http.post(`/settings/upgrade`, { version: version });
};
