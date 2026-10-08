import http from '@/api';
import { ResPage } from '@/api/interface';
import { Firewall } from '@/api/interface/firewall';
import { TimeoutEnum } from '@/enums/http-enum';

export const createFirewallPortWhitelist = (rule: Firewall.PortWhitelist) =>
    http.post('/hosts/firewall/settings/whitelist', { rule });

export const updateFirewallPortWhitelist = (request: Firewall.PortWhitelistUpdate) =>
    http.post('/hosts/firewall/settings/whitelist/update', request);

export const deleteFirewallPortWhitelist = (rule: Firewall.PortWhitelist) =>
    http.post('/hosts/firewall/settings/whitelist/delete', { rule });

export const loadFireBaseInfo = (tab: string) =>
    http.post<Firewall.FirewallBase>('/hosts/firewall/base', { name: tab }, TimeoutEnum.T_40S);

export const loadForwardBaseInfo = () =>
    http.post<Firewall.FirewallBase>('/hosts/firewall/forward/base', {}, TimeoutEnum.T_40S);

export const searchForwardRule = (request: Firewall.ForwardRuleSearch) =>
    http.post<ResPage<Firewall.RuleInfo>>('/hosts/firewall/forward/search', request, TimeoutEnum.T_40S);

export const operateFire = (operation: string, withDockerRestart: boolean) =>
    http.post<Firewall.FilterChainOperationResult>(
        '/hosts/firewall/operate',
        { operation, withDockerRestart },
        TimeoutEnum.T_10M,
    );

export const operateForwardRule = (request: { rules: Firewall.RuleForward[]; import?: boolean }) =>
    http.postWithConfig<Firewall.FilterChainOperationResult>('/hosts/firewall/forward/operate', request, {
        skipErrorMessage: true,
    });

export const enableForwarding = (taskID?: string, backupFile?: string) =>
    http.post<Firewall.FilterChainOperationResult>(
        '/hosts/firewall/forward/enable',
        { taskID, backupFile },
        TimeoutEnum.T_60S,
    );

export const operateFilterChain = (name: string, operate: string, taskID?: string) =>
    http.post<Firewall.FilterChainOperationResult>(
        '/hosts/firewall/filter/operate',
        { name, operate, ...(taskID ? { taskID } : {}) },
        TimeoutEnum.T_60S,
    );

export const searchFirewallRules = (request: Firewall.InventoryRequest) => {
    return http.post<Firewall.Inventory>('/hosts/firewall/rules/search', request, TimeoutEnum.T_40S);
};

export const listFirewallRuleBackups = (subsystem: Firewall.BackendSubsystem = 'system') =>
    http.get<Firewall.RuleBackups>('/hosts/firewall/rules/backups', { subsystem });

export const resetFirewallRules = (request: Firewall.ResetRequest) => {
    return http.post<Firewall.ResetResponse>('/hosts/firewall/rules/reset', request, TimeoutEnum.T_10M);
};

export const loadFirewallNativeDetail = (request: Firewall.NativeDetailRequest) => {
    return http.post<string>('/hosts/firewall/rules/native/detail', request, TimeoutEnum.T_40S);
};

export const createFirewallRules = (request: Firewall.CreateRequest) => {
    return http.post<Firewall.CreateResponse>('/hosts/firewall/rules', request, TimeoutEnum.T_10M);
};

export const deleteFirewallRules = (request: Firewall.DeleteRequest) => {
    return http.post<Firewall.DeleteResponse>('/hosts/firewall/rules/delete', request);
};

export const updateFirewallRule = (target: Firewall.RuleTarget, request: Firewall.UpdateRequest) => {
    return http.post('/hosts/firewall/rules/update', { ...request, ...target }, TimeoutEnum.T_60S);
};

export const loadDockerPortGuard = () =>
    http.get<Firewall.DockerGuardList>('/hosts/firewall/docker/ports', {}, { timeout: TimeoutEnum.T_40S });

export const loadDockerPublishedPorts = () =>
    http.get<Firewall.DockerGuardContainer[]>('/hosts/firewall/docker/endpoints', {}, { timeout: TimeoutEnum.T_40S });

export const operateDockerPortGuard = (
    operation: 'initialize' | 'bind' | 'unbind',
    taskID?: string,
    backupFile?: string,
) =>
    http.postWithConfig<Firewall.FilterChainOperationResult>(
        '/hosts/firewall/docker/operate',
        { operation, taskID, backupFile },
        {
            timeout: TimeoutEnum.T_60S,
            skipErrorMessage: true,
        },
    );

export const upsertDockerPortGuardPolicies = (request: Firewall.DockerGuardPolicyBatch) =>
    http.postWithConfig<Firewall.FilterChainOperationResult>('/hosts/firewall/docker/policies/batch', request, {
        skipErrorMessage: true,
    });

export const deleteDockerPortGuardPolicies = (request: Firewall.DockerGuardPolicyBatchDelete) =>
    http.postWithConfig<Firewall.FilterChainOperationResult>('/hosts/firewall/docker/policies/delete/batch', request, {
        timeout: TimeoutEnum.T_60S,
        skipErrorMessage: true,
    });

export const loadFirewallSettings = () => http.get<Firewall.Settings>('/hosts/firewall/settings');

export const operateFirewallBackend = (request: Firewall.BackendOperateRequest, skipErrorMessage = false) =>
    http.postWithConfig('/hosts/firewall/settings/operate', request, {
        timeout: TimeoutEnum.T_10M,
        skipErrorMessage,
    });

export const operateFirewallFamily = (request: {
    subsystem: Firewall.BackendSubsystem;
    backend: Firewall.Provider;
    family: 'ipv4' | 'ipv6';
    operation: 'initialize' | 'repair' | 'bind';
}) => http.post<Firewall.FilterChainOperationResult>('/hosts/firewall/family/operate', request, TimeoutEnum.T_60S);

export const operateFirewallIPv6 = (enabled: boolean) =>
    http.post<Firewall.FilterChainOperationResult>('/hosts/firewall/settings/ipv6', {
        status: enabled ? 'Enable' : 'Disable',
    });
