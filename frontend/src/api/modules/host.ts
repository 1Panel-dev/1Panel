import http from '@/api';
import { ResPage, ReqPage } from '../interface';
import { Host } from '../interface/host';
import { TimeoutEnum } from '@/enums/http-enum';
import { deepCopy } from '@/utils/misc';
import { encodeBase64Fields } from '@/utils/base64';

// firewall
export const loadFireBaseInfo = (tab: string) => {
    return http.post<Host.FirewallBase>(`/hosts/firewall/base`, { name: tab }, TimeoutEnum.T_40S);
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
export const operateForwardRule = (params: { rules: Host.RuleForward[]; forceDelete?: boolean }) => {
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

// Iptables Filter
export const searchFilterRules = (params: Host.IptablesFilterRuleSearch) => {
    return http.post<Host.IptablesData>(`/hosts/firewall/filter/rule/search`, params);
};
export const loadChainStatus = (name: string) => {
    return http.post<Host.ChainStatus>(`/hosts/firewall/filter/chain/status`, { name: name }, TimeoutEnum.T_60S);
};
export const operateFilterRule = (params: Host.IptablesFilterRuleOp) => {
    return http.post(`/hosts/firewall/filter/rule/operate`, params, TimeoutEnum.T_40S);
};
export const batchOperateFilterRule = (params: { rules: Host.IptablesFilterRuleOp[] }) => {
    return http.post(`/hosts/firewall/filter/rule/batch`, params, TimeoutEnum.T_40S);
};
export const operateFilterChain = (name: string, op: string) => {
    return http.post(`/hosts/firewall/filter/operate`, { name: name, operate: op }, TimeoutEnum.T_60S);
};

// monitors
export const loadMonitor = (param: Host.MonitorSearch, currentNode?: string) => {
    return http.post<Array<Host.MonitorData>>(
        `/hosts/monitor/search`,
        param,
        TimeoutEnum.T_60S,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};
export const getNetworkOptions = (currentNode?: string) => {
    return http.get<Array<string>>(
        `/hosts/monitor/netoptions`,
        {},
        currentNode ? { headers: { CurrentNode: currentNode } } : {},
    );
};
export const getIOOptions = (currentNode?: string) => {
    return http.get<Array<string>>(
        `/hosts/monitor/iooptions`,
        {},
        currentNode ? { headers: { CurrentNode: currentNode } } : {},
    );
};
export const cleanMonitors = () => {
    return http.post(`/hosts/monitor/clean`, {});
};
export const loadMonitorSetting = (currentNode?: string) => {
    return http.get<Host.MonitorSetting>(
        `/hosts/monitor/setting`,
        {},
        currentNode ? { headers: { CurrentNode: currentNode } } : {},
    );
};
export const updateMonitorSetting = (key: string, value: string) => {
    return http.post(`/hosts/monitor/setting/update`, { key: key, value: value });
};

// ssh
export const getSSHInfo = (currentNode?: string) => {
    return http.post<Host.SSHInfo>(
        `/hosts/ssh/search`,
        {},
        undefined,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};
export const operateSSH = (operation: string) => {
    return http.post(`/hosts/ssh/operate`, { operation: operation }, TimeoutEnum.T_40S);
};
export const updateSSH = (params: Host.SSHUpdate) => {
    return http.post(`/hosts/ssh/update`, params, TimeoutEnum.T_40S);
};
export const loadSSHFile = (name: string) => {
    return http.post<string>(`/hosts/ssh/file`, { name: name });
};
export const updateSSHByFile = (key: string, value: string, path = '') => {
    return http.post(`/hosts/ssh/file/update`, { key, path, value }, TimeoutEnum.T_60S);
};
export const createCert = (params: Host.RootCert) => {
    let request = deepCopy(params) as Host.RootCert;
    encodeBase64Fields(request, ['passPhrase', 'privateKey', 'publicKey']);
    return http.post(`/hosts/ssh/cert`, request);
};
export const editCert = (params: Host.RootCert) => {
    let request = deepCopy(params) as Host.RootCert;
    encodeBase64Fields(request, ['passPhrase', 'privateKey', 'publicKey']);
    return http.post(`/hosts/ssh/cert/update`, request);
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
export const loadSSHLogs = (params: Host.searchSSHLog, currentNode?: string) => {
    return http.post<ResPage<Host.sshHistory>>(
        `/hosts/ssh/log`,
        params,
        undefined,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
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

export const getComponentInfo = (name: string, operateNode?: string) => {
    const params = operateNode ? `?operateNode=${operateNode}` : '';
    return http.get<Host.ComponentInfo>(`/hosts/components/${name}${params}`);
};
