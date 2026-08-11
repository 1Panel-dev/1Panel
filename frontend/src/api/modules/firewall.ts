import http from '@/api';
import { Firewall } from '@/api/interface/firewall';
import { TimeoutEnum } from '@/enums/http-enum';

export const searchFirewallRules = (request: Firewall.InventoryRequest) => {
    return http.post<Firewall.Inventory>('/hosts/firewall/rules/search', request, TimeoutEnum.T_40S);
};

export const loadFirewallNativeDetail = (request: Firewall.NativeDetailRequest) => {
    return http.post<string>('/hosts/firewall/rules/native/detail', request, TimeoutEnum.T_40S);
};

export const checkFirewallRule = (request: Firewall.CheckRequest) => {
    return http.post<Firewall.RuleCheckResult>('/hosts/firewall/rules/check', request, TimeoutEnum.T_40S);
};

export const checkFirewallRulesBatch = (request: Firewall.BatchCheckRequest) => {
    return http.post<Firewall.BatchCheckResponse>('/hosts/firewall/rules/check/batch', request, TimeoutEnum.T_3M);
};

export const createFirewallRule = (request: Firewall.CreateRequest) => {
    return http.post('/hosts/firewall/rules', request, TimeoutEnum.T_60S);
};

export const createFirewallRulesBatch = (request: Firewall.BatchCreateRequest) => {
    return http.post<Firewall.BatchCreateResponse>('/hosts/firewall/rules/batch', request, TimeoutEnum.T_10M);
};

export const deleteFirewallRule = (uuid: string) => {
    return http.delete(`/hosts/firewall/rules/${encodeURIComponent(uuid)}`, undefined, {
        timeout: TimeoutEnum.T_60S,
    });
};

export const updateFirewallRule = (uuid: string, request: Firewall.UpdateRequest) => {
    return http.put(`/hosts/firewall/rules/${encodeURIComponent(uuid)}`, request, {
        timeout: TimeoutEnum.T_60S,
    });
};

export const reorderFirewallRule = (uuid: string, request: Firewall.ReorderRequest) => {
    return http.post(`/hosts/firewall/rules/${encodeURIComponent(uuid)}/reorder`, request, TimeoutEnum.T_60S);
};
