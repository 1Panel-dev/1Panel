import http from '@/api';
import { ResPage, ReqPage } from '../interface';
import { Host } from '../interface/host';
import { TimeoutEnum } from '@/enums/http-enum';
import { deepCopy } from '@/utils/util';
import { Base64 } from 'js-base64';

// firewall
export const loadFireBaseInfo = () => {
    return http.get<Host.FirewallBase>(`/hosts/firewall/base`);
};
export const searchFireRule = (params: Host.RuleSearch) => {
    return http.post<ResPage<Host.RuleInfo>>(`/hosts/firewall/search`, params, TimeoutEnum.T_40S);
};
export const operateFire = (operation: string, withDockerRestart: boolean) => {
    return http.post(
        `/hosts/firewall/operate`,
        {
            operation: operation,
            withDockerRestart: withDockerRestart,
        },
        TimeoutEnum.T_60S,
    );
};
export const operatePortRule = (params: Host.RulePort) => {
    return http.post<Host.RulePort>(`/hosts/firewall/port`, params, TimeoutEnum.T_40S);
};
export const operateForwardRule = (params: { rules: Host.RuleForward[]; forceDelete: boolean }) => {
    return http.post<Host.RulePort>(`/hosts/firewall/forward`, params, TimeoutEnum.T_40S);
};
export const operateIPRule = (params: Host.RuleIP) => {
    return http.post<Host.RuleIP>(`/hosts/firewall/ip`, params, TimeoutEnum.T_40S);
};
export const updatePortRule = (params: Host.UpdatePortRule) => {
    return http.post(`/hosts/firewall/update/port`, params, TimeoutEnum.T_40S);
};
export const updateAddrRule = (params: Host.UpdateAddrRule) => {
    return http.post(`/hosts/firewall/update/addr`, params, TimeoutEnum.T_40S);
};
export const updateFirewallDescription = (params: Host.UpdateDescription) => {
    return http.post(`/hosts/firewall/update/description`, params);
};
export const batchOperateRule = (params: Host.BatchRule) => {
    return http.post(`/hosts/firewall/batch`, params, TimeoutEnum.T_60S);
};

// monitors
export const loadMonitor = (param: Host.MonitorSearch) => {
    return http.post<Array<Host.MonitorData>>(`/hosts/monitor/search`, param);
};
export const getNetworkOptions = () => {
    return http.get<Array<string>>(`/hosts/monitor/netoptions`);
};
export const getIOOptions = () => {
    return http.get<Array<string>>(`/hosts/monitor/iooptions`);
};
export const cleanMonitors = () => {
    return http.post(`/hosts/monitor/clean`, {});
};
export const loadMonitorSetting = () => {
    return http.get<Host.MonitorSetting>(`/hosts/monitor/setting`, {});
};
export const updateMonitorSetting = (key: string, value: string) => {
    return http.post(`/hosts/monitor/setting/update`, { key: key, value: value });
};

// ssh
export const getSSHInfo = () => {
    return http.post<Host.SSHInfo>(`/hosts/ssh/search`);
};
export const getSSHConf = () => {
    return http.get<string>(`/hosts/ssh/conf`);
};
export const operateSSH = (operation: string) => {
    return http.post(`/hosts/ssh/operate`, { operation: operation }, TimeoutEnum.T_40S);
};
export const updateSSH = (params: Host.SSHUpdate) => {
    return http.post(`/hosts/ssh/update`, params, TimeoutEnum.T_40S);
};
export const updateSSHByfile = (file: string) => {
    return http.post(`/hosts/ssh/conffile/update`, { file: file }, TimeoutEnum.T_40S);
};
export const createCert = (params: Host.RootCert) => {
    let request = deepCopy(params) as Host.RootCert;
    if (request.passPhrase) {
        request.passPhrase = Base64.encode(request.passPhrase);
    }
    if (request.privateKey) {
        request.privateKey = Base64.encode(request.privateKey);
    }
    if (request.publicKey) {
        request.publicKey = Base64.encode(request.publicKey);
    }
    return http.post(`/hosts/ssh/cert`, request);
};
export const searchCert = (params: ReqPage) => {
    return http.post<ResPage<Host.RootCertInfo>>(`/hosts/ssh/cert/search`, params);
};
export const deleteCert = (ids: Array<number>, forceDelete: boolean) => {
    return http.post(`/hosts/ssh/cert/delete`, { ids: ids, forceDelete: forceDelete });
};
export const syncCert = () => {
    return http.post(`/hosts/ssh/cert/sync`);
};
export const loadSSHLogs = (params: Host.searchSSHLog) => {
    return http.post<ResPage<Host.sshHistory>>(`/hosts/ssh/log`, params);
};
export const exportSSHLogs = (params: Host.searchSSHLog) => {
    return http.post<string>(`/hosts/ssh/log/export`, params, TimeoutEnum.T_40S);
};

export const listDisks = () => {
    return http.get<Host.CompleteDiskInfo>(`/hosts/disks`);
};

export const partitionDisk = (params: Host.DiskPartition) => {
    return http.post(`/hosts/disks/partition`, params, TimeoutEnum.T_60S);
};

export const mountDisk = (params: Host.DiskMount) => {
    return http.post(`/hosts/disks/mount`, params, TimeoutEnum.T_60S);
};

export const unmountDisk = (params: Host.DiskUmount) => {
    return http.post(`/hosts/disks/unmount`, params, TimeoutEnum.T_60S);
};
