import http from '@/api';
import { ResPage, SearchWithPage } from '../interface';
import { Command } from '../interface/command';
import { Host } from '../interface/host';
import { Base64 } from 'js-base64';
import { deepCopy } from '@/utils/util';
import { TimeoutEnum } from '@/enums/http-enum';

export const searchHosts = (params: Host.SearchWithPage) => {
    return http.post<ResPage<Host.Host>>(`/hosts/search`, params);
};
export const getHostTree = (params: Host.ReqSearch) => {
    return http.post<Array<Host.HostTree>>(`/hosts/tree`, params);
};
export const addHost = (params: Host.HostOperate) => {
    let request = deepCopy(params) as Host.HostOperate;
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    if (request.privateKey) {
        request.privateKey = Base64.encode(request.privateKey);
    }
    return http.post<Host.HostOperate>(`/hosts`, request);
};
export const testByInfo = (params: Host.HostConnTest) => {
    let request = deepCopy(params) as Host.HostOperate;
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    if (request.privateKey) {
        request.privateKey = Base64.encode(request.privateKey);
    }
    return http.post<boolean>(`/hosts/test/byinfo`, request);
};
export const testByID = (id: number) => {
    return http.post<boolean>(`/hosts/test/byid/${id}`);
};
export const editHost = (params: Host.HostOperate) => {
    let request = deepCopy(params) as Host.HostOperate;
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    if (request.privateKey) {
        request.privateKey = Base64.encode(request.privateKey);
    }
    return http.post(`/hosts/update`, request);
};
export const editHostGroup = (params: Host.GroupChange) => {
    return http.post(`/hosts/update/group`, params);
};
export const deleteHost = (params: { ids: number[] }) => {
    return http.post(`/hosts/del`, params);
};

// command
export const getCommandList = () => {
    return http.get<Array<Command.CommandInfo>>(`/hosts/command`, {});
};
export const getCommandPage = (params: SearchWithPage) => {
    return http.post<ResPage<Command.CommandInfo>>(`/hosts/command/search`, params);
};
export const getCommandTree = () => {
    return http.get<any>(`/hosts/command/tree`);
};
export const addCommand = (params: Command.CommandOperate) => {
    return http.post<Command.CommandOperate>(`/hosts/command`, params);
};
export const editCommand = (params: Command.CommandOperate) => {
    return http.post(`/hosts/command/update`, params);
};
export const deleteCommand = (params: { ids: number[] }) => {
    return http.post(`/hosts/command/del`, params);
};

export const getRedisCommandList = () => {
    return http.get<Array<Command.RedisCommand>>(`/hosts/command/redis`, {});
};
export const getRedisCommandPage = (params: SearchWithPage) => {
    return http.post<ResPage<Command.RedisCommand>>(`/hosts/command/redis/search`, params);
};
export const saveRedisCommand = (params: Command.RedisCommand) => {
    return http.post(`/hosts/command/redis`, params);
};
export const deleteRedisCommand = (params: { ids: number[] }) => {
    return http.post(`/hosts/command/redis/del`, params);
};

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
export const operateForwardRule = (params: { rules: Host.RuleForward[] }) => {
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
