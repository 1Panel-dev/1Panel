import http from '@/api';
import { ResPage } from '../interface';
import { Host } from '../interface/host';

export const getHostList = (params: Host.ReqSearchWithPage) => {
    return http.post<ResPage<Host.Host>>(`/hosts/search`, params);
};

export const addHost = (params: Host.HostOperate) => {
    return http.post<Host.HostOperate>(`/hosts`, params);
};

export const editHost = (params: Host.HostOperate) => {
    console.log(params.id);
    return http.put(`/hosts/` + params.id, params);
};

export const deleteHost = (params: { ids: number[] }) => {
    return http.post(`/hosts/del`, params);
};
