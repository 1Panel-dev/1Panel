import http from '@/api';
import { ReqPage, ResPage, UpdateByFile } from '../interface';
import { Toolbox } from '../interface/toolbox';
import { Base64 } from 'js-base64';
import { TimeoutEnum } from '@/enums/http-enum';
import { deepCopy } from '@/utils/util';

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

// ftp
export const getFtpBase = () => {
    return http.get<Toolbox.FtpBaseInfo>(`/toolbox/ftp/base`);
};
export const searchFtpLog = (param: Toolbox.FtpSearchLog) => {
    return http.post<ResPage<Toolbox.FtpLog>>(`/toolbox/ftp/log/search`, param);
};
export const searchFtp = (param: ReqPage) => {
    return http.post<ResPage<Toolbox.FtpInfo>>(`/toolbox/ftp/search`, param);
};
export const operateFtp = (operate: string) => {
    return http.post(`/toolbox/ftp/operate`, { operation: operate }, TimeoutEnum.T_5M);
};
export const syncFtp = () => {
    return http.post(`/toolbox/ftp/sync`);
};

export const createFtp = (params: Toolbox.FtpCreate) => {
    let request = deepCopy(params) as Toolbox.FtpCreate;
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    return http.post(`/toolbox/ftp`, request);
};

export const updateFtp = (params: Toolbox.FtpUpdate) => {
    let request = deepCopy(params) as Toolbox.FtpUpdate;
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    return http.post(`/toolbox/ftp/update`, request);
};

export const deleteFtp = (params: { ids: number[] }) => {
    return http.post(`/toolbox/ftp/del`, params);
};

// clam
export const cleanClamRecord = (id: number) => {
    return http.post(`/toolbox/clam/record/clean`, { id: id });
};
export const searchClamRecord = (param: Toolbox.ClamSearchLog) => {
    return http.post<ResPage<Toolbox.ClamLog>>(`/toolbox/clam/record/search`, param);
};
export const getClamRecordLog = (param: Toolbox.ClamRecordReq) => {
    return http.post<string>(`/toolbox/clam/record/log`, param);
};
export const searchClamFile = (name: string, tail: string) => {
    return http.post<string>(`/toolbox/clam/file/search`, { name: name, tail: tail });
};
export const updateClamFile = (name: string, file: string) => {
    return http.post(`/toolbox/clam/file/update`, { name: name, file: file }, TimeoutEnum.T_60S);
};
export const searchClamBaseInfo = () => {
    return http.post<Toolbox.ClamBaseInfo>(`/toolbox/clam/base`);
};
export const updateClamBaseInfo = (operate: string) => {
    return http.post(`/toolbox/clam/operate`, { Operation: operate }, TimeoutEnum.T_60S);
};
export const searchClam = (param: ReqPage) => {
    return http.post<ResPage<Toolbox.ClamInfo>>(`/toolbox/clam/search`, param);
};
export const createClam = (params: Toolbox.ClamCreate) => {
    return http.post(`/toolbox/clam`, params);
};
export const updateClam = (params: Toolbox.ClamUpdate) => {
    return http.post(`/toolbox/clam/update`, params);
};
export const updateClamStatus = (id: number, status: string) => {
    return http.post(`/toolbox/clam/status/update`, { id: id, status: status });
};
export const deleteClam = (params: { ids: number[]; removeRecord: boolean; removeInfected: boolean }) => {
    return http.post(`/toolbox/clam/del`, params);
};
export const handleClamScan = (id: number) => {
    return http.post(`/toolbox/clam/handle`, { id: id });
};
