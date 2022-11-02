import http from '@/api';
import { ResPage } from '../interface';
import { WebSite } from '../interface/website';

export const SearchWebSites = (req: WebSite.WebSiteSearch) => {
    return http.post<ResPage<WebSite.WebSite>>(`/websites/search`, req);
};

export const ListGroups = () => {
    return http.get<WebSite.Group[]>(`/websites/groups`);
};

export const CreateWebsite = (req: WebSite.WebSiteCreateReq) => {
    return http.post<any>(`/websites`, req);
};

export const DeleteWebsite = (req: WebSite.WebSiteDel) => {
    return http.post<any>(`/websites/del`, req);
};
