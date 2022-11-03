import http from '@/api';
import { ResPage } from '../interface';
import { WebSite } from '../interface/website';

export const SearchWebSites = (req: WebSite.WebSiteSearch) => {
    return http.post<ResPage<WebSite.WebSite>>(`/websites/search`, req);
};

export const CreateWebsite = (req: WebSite.WebSiteCreateReq) => {
    return http.post<any>(`/websites`, req);
};

export const DeleteWebsite = (req: WebSite.WebSiteDel) => {
    return http.post<any>(`/websites/del`, req);
};

export const ListGroups = () => {
    return http.get<WebSite.Group[]>(`/websites/groups`);
};

export const CreateGroup = (req: WebSite.GroupOp) => {
    return http.post<any>(`/websites/groups`, req);
};

export const UpdateGroup = (req: WebSite.GroupOp) => {
    return http.put<any>(`/websites/groups`, req);
};

export const DeleteGroup = (id: number) => {
    return http.delete<any>(`/websites/groups/${id}`);
};

export const ListDomains = (id: number) => {
    return http.get<WebSite.Domain[]>(`/websites/domains/${id}`);
};

export const DeleteDomain = (id: number) => {
    return http.delete<any>(`/websites/domains/${id}`);
};
