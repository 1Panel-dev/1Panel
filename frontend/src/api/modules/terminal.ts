import http from '@/api';
import { ResPage } from '../interface';
import { Host } from '../interface/host';
import { Base64 } from 'js-base64';
import { deepCopy } from '@/utils/util';

export const searchHosts = (params: Host.SearchWithPage) => {
    return http.post<ResPage<Host.Host>>(`/core/hosts/search`, params);
};
export const getHostByID = (id: number) => {
    return http.post<Host.Host>(`/core/hosts/info`, { id: id });
};
export const getHostTree = (params: Host.ReqSearch) => {
    return http.post<Array<Host.HostTree>>(`/core/hosts/tree`, params);
};
export const addHost = (params: Host.HostOperate) => {
    let request = deepCopy(params) as Host.HostOperate;
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    if (request.privateKey) {
        request.privateKey = Base64.encode(request.privateKey);
    }
    if (params.isLocal) {
        return http.post(`/settings/ssh`, request);
    }
    return http.post<Host.HostOperate>(`/core/hosts`, request);
};
export const testByInfo = (params: Host.HostConnTest) => {
    let request = deepCopy(params) as Host.HostOperate;
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    if (request.privateKey) {
        request.privateKey = Base64.encode(request.privateKey);
    }
    if (params.isLocal) {
        return http.post<boolean>(`/settings/ssh/check/info`, request);
    }
    return http.post<boolean>(`/core/hosts/test/byinfo`, request);
};
export const testByID = (id: number) => {
    return http.post<boolean>(`/core/hosts/test/byid/${id}`);
};
export const editHost = (params: Host.HostOperate) => {
    let request = deepCopy(params) as Host.HostOperate;
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    if (request.privateKey) {
        request.privateKey = Base64.encode(request.privateKey);
    }
    return http.post(`/core/hosts/update`, request);
};
export const editHostGroup = (params: Host.GroupChange) => {
    return http.post(`/core/hosts/update/group`, params);
};
export const deleteHost = (params: { ids: number[] }) => {
    return http.post(`/core/hosts/del`, params);
};

// agent
export const loadLocalConn = () => {
    return http.get<Host.HostConnTest>(`/settings/ssh/conn`);
};
export const testLocalConn = () => {
    return http.post<boolean>(`/settings/ssh/check`);
};
