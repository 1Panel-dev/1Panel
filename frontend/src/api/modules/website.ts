import http from '@/api';
import { WebSite } from '../interface/website';

export const listGroups = () => {
    return http.get<WebSite.Group[]>(`/websites/groups`);
};

export const CreateWebsite = (req: WebSite.WebSiteCreateReq) => {
    return http.post<any>(`/websites`, req);
};
