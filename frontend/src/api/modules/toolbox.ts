import http from '@/api';
import { UpdateByFile } from '../interface';
import { Toolbox } from '../interface/toolbox';
import { Base64 } from 'js-base64';
import { TimeoutEnum } from '@/enums/http-enum';

// device
export const getDeviceBase = () => {
    return http.post<Toolbox.DeviceBaseInfo>(`/toolbox/device/base`, {}, TimeoutEnum.T_60S);
};
export const loadTimeZoneOptions = () => {
    return http.get<Array<string>>(`/toolbox/device/zone/options`);
};
export const updateDevice = (key: string, value: string) => {
    return http.post(`/toolbox/device/update/conf`, { key: key, value: value }, TimeoutEnum.T_60S);
};
export const updateDeviceHost = (param: Array<Toolbox.TimeZoneOptions>) => {
    return http.post(`/toolbox/device/update/host`, param, TimeoutEnum.T_60S);
};
export const updateDevicePasswd = (user: string, passwd: string) => {
    return http.post(`/toolbox/device/update/passwd`, { user: user, passwd: Base64.encode(passwd) }, TimeoutEnum.T_60S);
};
export const updateDeviceSwap = (params: Toolbox.SwapHelper) => {
    return http.post(`/toolbox/device/update/swap`, params, TimeoutEnum.T_60S);
};
export const updateDeviceByConf = (name: string, file: string) => {
    return http.post(`/toolbox/device/update/byconf`, { name: name, file: file }, TimeoutEnum.T_5M);
};
export const checkDNS = (key: string, value: string) => {
    return http.post(`/toolbox/device/check/dns`, { key: key, value: value });
};
export const loadDeviceConf = (name: string) => {
    return http.post(`/toolbox/device/conf`, { name: name });
};

// clean
export const scan = () => {
    return http.post<Toolbox.CleanData>(`/toolbox/scan`, {});
};
export const clean = (param: any) => {
    return http.post(`/toolbox/clean`, param);
};

// fail2ban
export const getFail2banBase = () => {
    return http.get<Toolbox.Fail2banBaseInfo>(`/toolbox/fail2ban/base`);
};
export const getFail2banConf = () => {
    return http.get<string>(`/toolbox/fail2ban/load/conf`);
};

export const searchFail2ban = (param: Toolbox.Fail2banSearch) => {
    return http.post<Array<string>>(`/toolbox/fail2ban/search`, param);
};

export const operateFail2ban = (operate: string) => {
    return http.post(`/toolbox/fail2ban/operate`, { operation: operate }, TimeoutEnum.T_5M);
};

export const operatorFail2banSSHD = (param: Toolbox.Fail2banSet) => {
    return http.post(`/toolbox/fail2ban/operate/sshd`, param, TimeoutEnum.T_5M);
};

export const updateFail2ban = (param: Toolbox.Fail2banUpdate) => {
    return http.post(`/toolbox/fail2ban/update`, param, TimeoutEnum.T_5M);
};

export const updateFail2banByFile = (param: UpdateByFile) => {
    return http.post(`/toolbox/fail2ban/update/byconf`, param, TimeoutEnum.T_5M);
};
