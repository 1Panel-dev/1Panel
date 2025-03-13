import http from '@/api';
import { ResPage } from '../interface';
import { Host } from '../interface/host';
import { TimeoutEnum } from '@/enums/http-enum';

// firewall
export const loadFireBaseInfo = () => {
    return http.get<Host.FirewallBase>(`/hosts/firewall/base`);
};
export const searchFireRule = (params: Host.RuleSearch) => {
    return http.post<ResPage<Host.RuleInfo>>(`/hosts/firewall/search`, params, TimeoutEnum.T_40S);
};
export const operateFire = (operation: string) => {
    return http.post(`/hosts/firewall/operate`, { operation: operation }, TimeoutEnum.T_40S);
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
export const generateSecret = (params: Host.SSHGenerate) => {
    return http.post(`/hosts/ssh/generate`, params);
};
export const loadSecret = (mode: string) => {
    return http.post<string>(`/hosts/ssh/secret`, { encryptionMode: mode });
};
export const loadSSHLogs = (params: Host.searchSSHLog) => {
    return http.post<Host.sshLog>(`/hosts/ssh/log`, params);
};
