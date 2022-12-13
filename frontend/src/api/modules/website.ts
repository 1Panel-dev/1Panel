import http from '@/api';
import { ReqPage, ResPage } from '../interface';
import { Website } from '../interface/Website';
import { File } from '../interface/file';

export const SearchWebsites = (req: Website.WebSiteSearch) => {
    return http.post<ResPage<Website.Website>>(`/websites/search`, req);
};

export const CreateWebsite = (req: Website.WebSiteCreateReq) => {
    return http.post<any>(`/websites`, req);
};

export const BackupWebsite = (id: number) => {
    return http.post(`/websites/backup/${id}`);
};

export const RecoverWebsite = (req: Website.WebSiteRecover) => {
    return http.post(`/websites/recover`, req);
};

export const RecoverWebsiteByUpload = (req: Website.WebsiteRecoverByUpload) => {
    return http.post(`/websites/recover/byupload`, req);
};

export const UpdateWebsite = (req: Website.WebSiteUpdateReq) => {
    return http.post<any>(`/websites/update`, req);
};

export const GetWebsite = (id: number) => {
    return http.get<Website.Website>(`/websites/${id}`);
};

export const GetWebsiteOptions = () => {
    return http.get<Array<string>>(`/websites/options`);
};

export const GetWebsiteNginx = (id: number) => {
    return http.get<File.File>(`/websites/${id}/nginx`);
};

export const DeleteWebsite = (req: Website.WebSiteDel) => {
    return http.post<any>(`/websites/del`, req);
};

export const ListGroups = () => {
    return http.get<Website.Group[]>(`/websites/groups`);
};

export const CreateGroup = (req: Website.GroupOp) => {
    return http.post<any>(`/websites/groups`, req);
};

export const UpdateGroup = (req: Website.GroupOp) => {
    return http.put<any>(`/websites/groups`, req);
};

export const DeleteGroup = (id: number) => {
    return http.delete<any>(`/websites/groups/${id}`);
};

export const ListDomains = (id: number) => {
    return http.get<Website.Domain[]>(`/websites/domains/${id}`);
};

export const DeleteDomain = (req: Website.DomainDelete) => {
    return http.post<any>(`/websites/domains/del/`, req);
};

export const CreateDomain = (req: Website.DomainCreate) => {
    return http.post<any>(`/websites/domains`, req);
};

export const GetNginxConfig = (req: Website.NginxScopeReq) => {
    return http.post<Website.NginxScopeConfig>(`/websites/config`, req);
};

export const UpdateNginxConfig = (req: Website.NginxConfigReq) => {
    return http.post<any>(`/websites/config/update`, req);
};

export const SearchDnsAccount = (req: ReqPage) => {
    return http.post<ResPage<Website.DnsAccount>>(`/websites/dns/search`, req);
};

export const CreateDnsAccount = (req: Website.DnsAccountCreate) => {
    return http.post<any>(`/websites/dns`, req);
};

export const UpdateDnsAccount = (req: Website.DnsAccountUpdate) => {
    return http.post<any>(`/websites/dns/update`, req);
};

export const DeleteDnsAccount = (id: number) => {
    return http.delete<any>(`/websites/dns/${id}`);
};

export const SearchAcmeAccount = (req: ReqPage) => {
    return http.post<ResPage<Website.AcmeAccount>>(`/websites/acme/search`, req);
};

export const CreateAcmeAccount = (req: Website.AcmeAccountCreate) => {
    return http.post<Website.AcmeAccount>(`/websites/acme`, req);
};

export const DeleteAcmeAccount = (id: number) => {
    return http.delete<any>(`/websites/acme/${id}`);
};

export const SearchSSL = (req: ReqPage) => {
    return http.post<ResPage<Website.SSL>>(`/websites/ssl/search`, req);
};

export const ListSSL = (req: Website.SSLReq) => {
    return http.post<Website.SSL[]>(`/websites/ssl/search`, req);
};

export const CreateSSL = (req: Website.SSLCreate) => {
    return http.post<Website.SSLCreate>(`/websites/ssl`, req);
};

export const DeleteSSL = (id: number) => {
    return http.delete<any>(`/websites/ssl/${id}`);
};

export const GetWebsiteSSL = (websiteId: number) => {
    return http.get<Website.SSL>(`/websites/ssl/${websiteId}`);
};

export const ApplySSL = (req: Website.SSLApply) => {
    return http.post<Website.SSLApply>(`/websites/ssl/apply`, req);
};

export const RenewSSL = (req: Website.SSLRenew) => {
    return http.post<any>(`/websites/ssl/renew`, req);
};

export const GetDnsResolve = (req: Website.DNSResolveReq) => {
    return http.post<Website.DNSResolve[]>(`/websites/ssl/resolve`, req);
};

export const GetHTTPSConfig = (id: number) => {
    return http.get<Website.HTTPSConfig>(`/websites/${id}/https`);
};

export const UpdateHTTPSConfig = (req: Website.HTTPSReq) => {
    return http.post<Website.HTTPSConfig>(`/websites/${req.websiteId}/https`, req);
};

export const PreCheck = (req: Website.CheckReq) => {
    return http.post<Website.CheckRes[]>(`/websites/check`, req);
};

export const GetWafConfig = (req: Website.WafReq) => {
    return http.post<Website.WafRes>(`/websites/waf/config`, req);
};

export const UpdateWafEnable = (req: Website.WafUpdate) => {
    return http.post<any>(`/websites/waf/update`, req);
};
