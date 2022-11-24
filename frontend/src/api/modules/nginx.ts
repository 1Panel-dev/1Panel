import http from '@/api';
import { File } from '../interface/file';
import { Nginx } from '../interface/nginx';

export const GetNginx = () => {
    return http.get<File.File>(`/nginx`);
};

export const GetNginxConfigByScope = (req: Nginx.NginxScopeReq) => {
    return http.post<Nginx.NginxParam[]>(`/nginx/scope`, req);
};

export const UpdateNginxConfigByScope = (req: Nginx.NginxConfigReq) => {
    return http.post<any>(`/nginx/update`, req);
};
