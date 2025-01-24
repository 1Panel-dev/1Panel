import http from '@/api';
import { File } from '../interface/file';
import { Nginx } from '../interface/nginx';

export const getNginx = () => {
    return http.get<File.File>(`/openresty`);
};

export const getNginxConfigByScope = (req: Nginx.NginxScopeReq) => {
    return http.post<Nginx.NginxParam[]>(`/openresty/scope`, req);
};

export const updateNginxConfigByScope = (req: Nginx.NginxConfigReq) => {
    return http.post(`/openresty/update`, req);
};

export const getNginxStatus = () => {
    return http.get<Nginx.NginxStatus>(`/openresty/status`);
};

export const updateNginxConfigFile = (req: Nginx.NginxFileUpdate) => {
    return http.post(`/openresty/file`, req);
};

export const buildNginx = (req: Nginx.NginxBuildReq) => {
    return http.post(`/openresty/build`, req);
};

export const getNginxModules = () => {
    return http.get<Nginx.NginxBuildConfig>(`/openresty/modules`);
};

export const updateNginxModule = (req: Nginx.NginxModuleUpdate) => {
    return http.post(`/openresty/modules/update`, req);
};
