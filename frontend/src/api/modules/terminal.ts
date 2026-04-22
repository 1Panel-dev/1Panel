import http from '@/api';
import { ResPage } from '../interface';
import { Host } from '../interface/host';
import { encodeBase64Fields } from '@/utils/base64';
import { deepCopy } from '@/utils/misc';
export const searchHosts = (params: Host.SearchWithPage) => {
    return http.postLocalNode<ResPage<Host.Host>>(`/hosts/search`, params);
};
export const getHostByID = (id: number) => {
    return http.postLocalNode<Host.Host>(`/hosts/info`, { id: id });
};
export const getHostTree = (params: Host.ReqSearch) => {
    return http.postLocalNode<Array<Host.HostTree>>(`/hosts/tree`, params);
};
export const updateLocalConn = (param: { withReset: boolean; defaultConn: string }) => {
    return http.post(`/settings/ssh/default`, param);
};
export const addHost = (params: Host.HostOperate) => {
    let request = deepCopy(params) as Host.HostOperate;
    encodeBase64Fields(request, ['password', 'privateKey']);
    if (params.isLocal) {
        return http.post(`/settings/ssh`, request);
    }
    return http.postLocalNode<Host.HostOperate>(`/hosts`, request);
};
export const testByInfo = (params: Host.HostConnTest) => {
    let request = deepCopy(params) as Host.HostOperate;
    encodeBase64Fields(request, ['password', 'privateKey']);
    if (params.isLocal) {
        return http.post<boolean>(`/settings/ssh/check/info`, request);
    }
    return http.postLocalNode<boolean>(`/hosts/test/byinfo`, request);
};
export const testByID = (id: number) => {
    return http.postLocalNode<boolean>(`/hosts/test/byid`, { id: id });
};
export const editHost = (params: Host.HostOperate) => {
    let request = deepCopy(params) as Host.HostOperate;
    encodeBase64Fields(request, ['password', 'privateKey']);
    return http.postLocalNode(`/hosts/update`, request);
};
export const editHostGroup = (params: Host.GroupChange) => {
    return http.postLocalNode(`/hosts/update/group`, params);
};
export const deleteHost = (params: { ids: number[] }) => {
    return http.postLocalNode(`/hosts/del`, params);
};

// agent
export const loadLocalConn = () => {
    return http.get<Host.HostConnTest>(`/settings/ssh/conn`);
};
export const testLocalConn = () => {
    return http.post<boolean>(`/settings/ssh/check`);
};
