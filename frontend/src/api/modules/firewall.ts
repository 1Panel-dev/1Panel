import http from '@/api';
import { ResPage } from '@/api/interface';
import { Firewall } from '@/api/interface/firewall';
import { TimeoutEnum } from '@/enums/http-enum';

export const loadFireBaseInfo = (tab: string) =>
    http.post<Firewall.FirewallBase>('/hosts/firewall/base', { name: tab }, TimeoutEnum.T_40S);

export const loadForwardBaseInfo = () =>
    http.post<Firewall.FirewallBase>('/hosts/firewall/forward/base', {}, TimeoutEnum.T_40S);

export const searchForwardRule = (request: Firewall.ForwardRuleSearch) =>
    http.post<ResPage<Firewall.RuleInfo>>('/hosts/firewall/forward/search', request, TimeoutEnum.T_40S);

export const operateFire = (operation: string, withDockerRestart: boolean) =>
    http.post('/hosts/firewall/operate', { operation, withDockerRestart }, TimeoutEnum.T_60S);

export const operateForwardRule = (request: { rules: Firewall.RuleForward[]; forceDelete?: boolean }) =>
    http.post('/hosts/firewall/forward/operate', request, TimeoutEnum.T_40S);

export const enableForwarding = () => http.post('/hosts/firewall/forward/enable', {}, TimeoutEnum.T_60S);

export const operateFilterChain = (name: string, operate: string) =>
    http.post('/hosts/firewall/filter/operate', { name, operate }, TimeoutEnum.T_60S);

export const searchFirewallRules = (request: Firewall.InventoryRequest) => {
    return http.post<Firewall.Inventory>('/hosts/firewall/rules/search', request, TimeoutEnum.T_40S);
};

export const resetFirewallRules = (request: Firewall.ResetRequest) => {
    return http.post<Firewall.ResetResponse>('/hosts/firewall/rules/reset', request, TimeoutEnum.T_10M);
};

export const loadFirewallNativeDetail = (request: Firewall.NativeDetailRequest) => {
    return http.post<string>('/hosts/firewall/rules/native/detail', request, TimeoutEnum.T_40S);
};

export const checkFirewallRules = (request: Firewall.CheckRequest) => {
    return http.post<Firewall.CheckResponse>('/hosts/firewall/rules/check', request, TimeoutEnum.T_3M);
};

export const createFirewallRules = (request: Firewall.CreateRequest) => {
    return http.post<Firewall.CreateResponse>('/hosts/firewall/rules', request, TimeoutEnum.T_10M);
};

export const previewFirewallRuleSync = (request: Firewall.RuleSyncRequest) => {
    return http.post<Firewall.RuleSyncPreview>('/hosts/firewall/rules/sync/preview', request, TimeoutEnum.T_3M);
};

export const loadFirewallRuleSyncTask = () => {
    return http.get<Firewall.RuleSyncTask>('/hosts/firewall/rules/sync/task');
};

export const syncFirewallRules = (request: Firewall.RuleSyncRequest) => {
    return http.post<Firewall.RuleSyncResult>('/hosts/firewall/rules/sync', request, TimeoutEnum.T_10M);
};

export const deleteFirewallRules = (request: Firewall.DeleteRequest) => {
    return http.post<Firewall.DeleteResponse>('/hosts/firewall/rules/delete', request, TimeoutEnum.T_10M);
};

export const updateFirewallRule = (uuid: string, request: Firewall.UpdateRequest) => {
    return http.post('/hosts/firewall/rules/update', { ...request, uuid }, TimeoutEnum.T_60S);
};

export const reorderFirewallRule = (uuid: string, request: Firewall.ReorderRequest) => {
    return http.post('/hosts/firewall/rules/reorder', { ...request, uuid }, TimeoutEnum.T_60S);
};

export const loadDockerPortGuard = () =>
    http.get<Firewall.DockerGuardList>('/hosts/firewall/docker/ports', {}, { timeout: TimeoutEnum.T_40S });

export const loadDockerPublishedPorts = () =>
    http.get<Firewall.DockerGuardContainer[]>('/hosts/firewall/docker/endpoints', {}, { timeout: TimeoutEnum.T_40S });

export const syncDockerPortGuard = () => http.post('/hosts/firewall/docker/sync', {}, TimeoutEnum.T_60S);

export const operateDockerPortGuard = (operation: 'initialize' | 'bind' | 'unbind') =>
    http.post('/hosts/firewall/docker/operate', { operation }, TimeoutEnum.T_60S);

export const upsertDockerPortGuardPolicies = (request: Firewall.DockerGuardPolicyBatch) =>
    http.post('/hosts/firewall/docker/policies/batch', request, TimeoutEnum.T_60S);

export const deleteDockerPortGuardPolicies = (request: Firewall.DockerGuardPolicyBatchDelete) =>
    http.post('/hosts/firewall/docker/policies/delete/batch', request, TimeoutEnum.T_60S);

export const loadFirewallSettings = () => http.get<Firewall.Settings>('/hosts/firewall/settings');

export const operateFirewallBackend = (request: Firewall.BackendOperateRequest) =>
    http.post('/hosts/firewall/settings/operate', request, TimeoutEnum.T_10M);
