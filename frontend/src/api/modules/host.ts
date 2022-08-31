import http from '@/api';
import { Host } from '../interface/host';

export const getHostList = (params: Host.ReqSearch) => {
    return http.post<Array<Host.HostTree>>(`/hosts/search`, params);
};

export const getHostInfo = (id: number) => {
    return http.get<Host.Host>(`/hosts/` + id);
};

export const addHost = (params: Host.HostOperate) => {
    return http.post<Host.HostOperate>(`/hosts`, params);
};

export const editHost = (params: Host.HostOperate) => {
    console.log(params.id);
    return http.put(`/hosts/` + params.id, params);
};

export const deleteHost = (id: number) => {
    return http.delete(`/hosts/` + id);
};
