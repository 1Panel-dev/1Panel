import http from '@/api';
import { Setting } from '../interface/setting';

export const getSettingInfo = () => {
    return http.post<Setting.SettingInfo>(`/settings/search`);
};

export const updateSetting = (param: Setting.SettingUpdate) => {
    return http.put(`/settings`, param);
};
