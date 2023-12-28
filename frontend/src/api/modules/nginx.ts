import http from '@/api';
import { File } from '../interface/file';
import { Nginx } from '../interface/nginx';

export const GetNginx = () => {
    return http.get<File.File>(`/openresty`);
};

export const GetNginxConfigByScope = (req: Nginx.NginxScopeReq) => {
    return http.post<Nginx.NginxParam[]>(`/openresty/scope`, req);
};

export const UpdateNginxConfigByScope = (req: Nginx.NginxConfigReq) => {
    return http.post<any>(`/openresty/update`, req);
};

export const GetNginxStatus = () => {
    return http.get<Nginx.NginxStatus>(`/openresty/status`);
};

export const UpdateNginxConfigFile = (req: Nginx.NginxFileUpdate) => {
    return http.post<any>(`/openresty/file`, req);
};

export const ClearNginxCache = () => {
    return http.post<any>(`/openresty/clear`);
};
