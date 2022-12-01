import http from '@/api';
import { ReqPage, ResPage } from '../interface';
import { WebSite } from '../interface/website';
import { File } from '../interface/file';

export const SearchWebSites = (req: WebSite.WebSiteSearch) => {
    return http.post<ResPage<WebSite.WebSite>>(`/websites/search`, req);
};

export const CreateWebsite = (req: WebSite.WebSiteCreateReq) => {
    return http.post<any>(`/websites`, req);
};

export const BackupWebsite = (id: number) => {
    return http.post(`/websites/backup/${id}`);
};
export const RecoverWebsite = (req: WebSite.WebSiteRecover) => {
    return http.post(`/websites/recover`, req);
};
export const RecoverWebsiteByUpload = (req: WebSite.WebsiteRecoverByUpload) => {
    return http.post(`/websites/recover/byupload`, req);
};

export const UpdateWebsite = (req: WebSite.WebSiteUpdateReq) => {
    return http.post<any>(`/websites/update`, req);
};

export const GetWebsite = (id: number) => {
    return http.get<WebSite.WebSiteDTO>(`/websites/${id}`);
};

export const GetWebsiteOptions = () => {
    return http.get<Array<string>>(`/websites/options`);
};

export const GetWebsiteNginx = (id: number) => {
    return http.get<File.File>(`/websites/${id}/nginx`);
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

export const CreateDomain = (req: WebSite.DomainCreate) => {
    return http.post<any>(`/websites/domains`, req);
};

export const GetNginxConfig = (req: WebSite.NginxConfigReq) => {
    return http.post<WebSite.NginxParam[]>(`/websites/config`, req);
};

export const UpdateNginxConfig = (req: WebSite.NginxConfigReq) => {
    return http.post<any>(`/websites/config/update`, req);
};

export const SearchDnsAccount = (req: ReqPage) => {
    return http.post<ResPage<WebSite.DnsAccount>>(`/websites/dns/search`, req);
};

export const CreateDnsAccount = (req: WebSite.DnsAccountCreate) => {
    return http.post<any>(`/websites/dns`, req);
};

export const UpdateDnsAccount = (req: WebSite.DnsAccountUpdate) => {
    return http.post<any>(`/websites/dns/update`, req);
};

export const DeleteDnsAccount = (id: number) => {
    return http.delete<any>(`/websites/dns/${id}`);
};

export const SearchAcmeAccount = (req: ReqPage) => {
    return http.post<ResPage<WebSite.AcmeAccount>>(`/websites/acme/search`, req);
};

export const CreateAcmeAccount = (req: WebSite.AcmeAccountCreate) => {
    return http.post<WebSite.AcmeAccount>(`/websites/acme`, req);
};

export const DeleteAcmeAccount = (id: number) => {
    return http.delete<any>(`/websites/acme/${id}`);
};

export const SearchSSL = (req: ReqPage) => {
    return http.post<ResPage<WebSite.SSL>>(`/websites/ssl/search`, req);
};

export const ListSSL = (req: WebSite.SSLReq) => {
    return http.post<WebSite.SSL[]>(`/websites/ssl/search`, req);
};

export const CreateSSL = (req: WebSite.SSLCreate) => {
    return http.post<WebSite.SSLCreate>(`/websites/ssl`, req);
};

export const DeleteSSL = (id: number) => {
    return http.delete<any>(`/websites/ssl/${id}`);
};

export const GetWebsiteSSL = (websiteId: number) => {
    return http.get<WebSite.SSL>(`/websites/ssl/${websiteId}`);
};

export const ApplySSL = (req: WebSite.SSLApply) => {
    return http.post<WebSite.SSLApply>(`/websites/ssl/apply`, req);
};

export const RenewSSL = (req: WebSite.SSLRenew) => {
    return http.post<any>(`/websites/ssl/renew`, req);
};

export const GetDnsResolve = (req: WebSite.DNSResolveReq) => {
    return http.post<WebSite.DNSResolve[]>(`/websites/ssl/resolve`, req);
};

export const GetHTTPSConfig = (id: number) => {
    return http.get<WebSite.HTTPSConfig>(`/websites/${id}/https`);
};

export const UpdateHTTPSConfig = (req: WebSite.HTTPSReq) => {
    return http.post<WebSite.HTTPSConfig>(`/websites/${req.websiteId}/https`, req);
};
