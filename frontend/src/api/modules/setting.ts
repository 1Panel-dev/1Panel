import http from '@/api';
import { Setting } from '../interface/setting';

export const getSettingInfo = () => {
    return http.post<Setting.SettingInfo>(`/settings/search`);
};

export const updateSetting = (param: Setting.SettingUpdate) => {
    return http.post(`/settings/update`, param);
};

export const updatePassword = (param: Setting.PasswordUpdate) => {
    return http.post(`/settings/password/update`, param);
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
