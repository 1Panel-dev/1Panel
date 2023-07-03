import http from '@/api';
import { ResPage } from '../interface';
import { Command } from '../interface/command';
import { Host } from '../interface/host';
import { Base64 } from 'js-base64';
import { deepCopy } from '@/utils/util';

export const searchHosts = (params: Host.SearchWithPage) => {
    return http.post<ResPage<Host.Host>>(`/hosts/search`, params);
};
export const getHostTree = (params: Host.ReqSearch) => {
    return http.post<Array<Host.HostTree>>(`/hosts/tree`, params);
};
export const addHost = (params: Host.HostOperate) => {
    let reqest = deepCopy(params) as Host.HostOperate;
    if (reqest.password) {
        reqest.password = Base64.encode(reqest.password);
    }
    if (reqest.privateKey) {
        reqest.privateKey = Base64.encode(reqest.privateKey);
    }
    return http.post<Host.HostOperate>(`/hosts`, reqest);
};
export const testByInfo = (params: Host.HostConnTest) => {
    let reqest = deepCopy(params) as Host.HostOperate;
    if (reqest.password) {
        reqest.password = Base64.encode(reqest.password);
    }
    if (reqest.privateKey) {
        reqest.privateKey = Base64.encode(reqest.privateKey);
    }
    return http.post<boolean>(`/hosts/test/byinfo`, reqest);
};
export const testByID = (id: number) => {
    return http.post<boolean>(`/hosts/test/byid/${id}`);
};
export const editHost = (params: Host.HostOperate) => {
    let reqest = deepCopy(params) as Host.HostOperate;
    if (reqest.password) {
        reqest.password = Base64.encode(reqest.password);
    }
    if (reqest.privateKey) {
        reqest.privateKey = Base64.encode(reqest.privateKey);
    }
    return http.post(`/hosts/update`, reqest);
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
export const getCommandPage = (params: Command.CommandSearch) => {
    return http.post<ResPage<Command.CommandInfo>>(`/hosts/command/search`, params);
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

// firewall
export const loadFireBaseInfo = () => {
    return http.get<Host.FirewallBase>(`/hosts/firewall/base`);
};
export const searchFireRule = (params: Host.RuleSearch) => {
    return http.post<ResPage<Host.RuleInfo>>(`/hosts/firewall/search`, params);
};
export const operateFire = (operation: string) => {
    return http.post(`/hosts/firewall/operate`, { operation: operation });
};
export const operatePortRule = (params: Host.RulePort) => {
    return http.post<Host.RulePort>(`/hosts/firewall/port`, params);
};
export const operateIPRule = (params: Host.RuleIP) => {
    return http.post<Host.RuleIP>(`/hosts/firewall/ip`, params);
};
export const updatePortRule = (params: Host.UpdatePortRule) => {
    return http.post(`/hosts/firewall/update/port`, params);
};
export const updateAddrRule = (params: Host.UpdateAddrRule) => {
    return http.post(`/hosts/firewall/update/addr`, params);
};
export const batchOperateRule = (params: Host.BatchRule) => {
    return http.post(`/hosts/firewall/batch`, params);
};

// ssh
export const getSSHInfo = () => {
    return http.post<Host.SSHInfo>(`/hosts/ssh/search`);
};
export const operateSSH = (operation: string) => {
    return http.post(`/hosts/ssh/operate`, { operation: operation });
};
export const updateSSH = (key: string, value: string) => {
    return http.post(`/hosts/ssh/update`, { key: key, value: value });
};
export const updateSSHByfile = (file: string) => {
    return http.post(`/hosts/ssh/conffile/update`, { file: file });
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
