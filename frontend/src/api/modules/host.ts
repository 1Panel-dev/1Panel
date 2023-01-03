import http from '@/api';
import { Host } from '../interface/host';

export const getHostTree = (params: Host.ReqSearch) => {
    return http.post<Array<Host.HostTree>>(`/hosts/search`, params);
};

export const getHostInfo = (id: number) => {
    return http.get<Host.Host>(`/hosts/` + id);
};

export const addHost = (params: Host.HostOperate) => {
    return http.post<Host.HostOperate>(`/hosts`, params);
};

export const testByInfo = (params: Host.HostConnTest) => {
    return http.post<boolean>(`/hosts/test/byinfo`, params);
};

export const testByID = (id: number) => {
    return http.post<boolean>(`/hosts/test/byid/${id}`);
};

export const editHost = (params: Host.HostOperate) => {
    return http.post(`/hosts/update`, params);
};

export const deleteHost = (id: number) => {
    return http.post(`/hosts/del`, { id: id });
};
