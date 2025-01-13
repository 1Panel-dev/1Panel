import http from '@/api';
import { ReqPage, ResPage } from '../interface';
import { Website } from '../interface/website';
import { File } from '../interface/file';
import { TimeoutEnum } from '@/enums/http-enum';
import { deepCopy } from '@/utils/util';
import { Base64 } from 'js-base64';

export const SearchWebsites = (req: Website.WebSiteSearch) => {
    return http.post<ResPage<Website.WebsiteRes>>(`/websites/search`, req);
};

export const ListWebsites = () => {
    return http.get<Website.WebsiteDTO>(`/websites/list`);
};

export const CreateWebsite = (req: Website.WebSiteCreateReq) => {
    let request = deepCopy(req) as Website.WebSiteCreateReq;
    if (request.ftpPassword) {
        request.ftpPassword = Base64.encode(request.ftpPassword);
    }
    return http.post<any>(`/websites`, request);
};

export const OpWebsite = (req: Website.WebSiteOp) => {
    return http.post<any>(`/websites/operate`, req);
};

export const OpWebsiteLog = (req: Website.WebSiteOpLog) => {
    return http.post<Website.WebSiteLog>(`/websites/log`, req);
};

export const UpdateWebsite = (req: Website.WebSiteUpdateReq) => {
    return http.post<any>(`/websites/update`, req);
};

export const GetWebsite = (id: number) => {
    return http.get<Website.WebsiteDTO>(`/websites/${id}`);
};

export const GetWebsiteOptions = () => {
    return http.get<Array<string>>(`/websites/options`);
};

export const GetWebsiteConfig = (id: number, type: string) => {
    return http.get<File.File>(`/websites/${id}/config/${type}`);
};

export const DeleteWebsite = (req: Website.WebSiteDel) => {
    return http.post<any>(`/websites/del`, req);
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

export const DeleteDnsAccount = (req: Website.DelReq) => {
    return http.post<any>(`/websites/dns/del`, req);
};

export const SearchAcmeAccount = (req: ReqPage) => {
    return http.post<ResPage<Website.AcmeAccount>>(`/websites/acme/search`, req);
};

export const CreateAcmeAccount = (req: Website.AcmeAccountCreate) => {
    return http.post<Website.AcmeAccount>(`/websites/acme`, req, TimeoutEnum.T_10M);
};

export const DeleteAcmeAccount = (req: Website.DelReq) => {
    return http.post<any>(`/websites/acme/del`, req);
};

export const SearchSSL = (req: ReqPage) => {
    return http.post<ResPage<Website.SSLDTO>>(`/websites/ssl/search`, req);
};

export const ListSSL = (req: Website.SSLReq) => {
    return http.post<Website.SSLDTO[]>(`/websites/ssl/search`, req);
};

export const CreateSSL = (req: Website.SSLCreate) => {
    return http.post<Website.SSLCreate>(`/websites/ssl`, req, TimeoutEnum.T_10M);
};

export const DeleteSSL = (req: Website.DelReq) => {
    return http.post<any>(`/websites/ssl/del`, req);
};

export const GetWebsiteSSL = (websiteId: number) => {
    return http.get<Website.SSL>(`/websites/ssl/website/${websiteId}`);
};

export const GetSSL = (id: number) => {
    return http.get<Website.SSL>(`/websites/ssl/${id}`);
};

export const ApplySSL = (req: Website.SSLApply) => {
    return http.post<Website.SSLApply>(`/websites/ssl/apply`, req);
};

export const ObtainSSL = (req: Website.SSLObtain) => {
    return http.post<any>(`/websites/ssl/obtain`, req);
};

export const UpdateSSL = (req: Website.SSLUpdate) => {
    return http.post<any>(`/websites/ssl/update`, req);
};

export const GetDnsResolve = (req: Website.DNSResolveReq) => {
    return http.post<Website.DNSResolve[]>(`/websites/ssl/resolve`, req, TimeoutEnum.T_5M);
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

export const UpdateNginxFile = (req: Website.NginxUpdate) => {
    return http.post<any>(`/websites/nginx/update`, req);
};

export const ChangeDefaultServer = (req: Website.DefaultServerUpdate) => {
    return http.post<any>(`/websites/default/server`, req);
};

export const GetPHPConfig = (id: number) => {
    return http.get<Website.PHPConfig>(`/websites/php/config/${id}`);
};

export const UpdatePHPConfig = (req: Website.PHPConfigUpdate) => {
    return http.post<any>(`/websites/php/config/`, req);
};

export const UpdatePHPFile = (req: Website.PHPUpdate) => {
    return http.post<any>(`/websites/php/update`, req);
};

export const GetRewriteConfig = (req: Website.RewriteReq) => {
    return http.post<Website.RewriteRes>(`/websites/rewrite`, req);
};

export const UpdateRewriteConfig = (req: Website.RewriteUpdate) => {
    return http.post<any>(`/websites/rewrite/update`, req);
};

export const UpdateWebsiteDir = (req: Website.DirUpdate) => {
    return http.post<any>(`/websites/dir/update`, req);
};

export const UpdateWebsiteDirPermission = (req: Website.DirPermissionUpdate) => {
    return http.post<any>(`/websites/dir/permission`, req);
};

export const GetProxyConfig = (req: Website.ProxyReq) => {
    return http.post<Website.ProxyConfig[]>(`/websites/proxies`, req);
};

export const OperateProxyConfig = (req: Website.ProxyReq) => {
    return http.post<any>(`/websites/proxies/update`, req);
};

export const UpdateProxyConfigFile = (req: Website.ProxyFileUpdate) => {
    return http.post<any>(`/websites/proxies/file`, req);
};

export const DelProxy = (req: Website.ProxyDel) => {
    return http.post<any>(`/websites/proxies/del`, req);
};

export const GetAuthConfig = (req: Website.AuthReq) => {
    return http.post<Website.AuthConfig>(`/websites/auths`, req);
};

export const OperateAuthConfig = (req: Website.NginxAuthConfig) => {
    return http.post<any>(`/websites/auths/update`, req);
};

export const GetAntiLeech = (req: Website.LeechReq) => {
    return http.post<Website.LeechConfig>(`/websites/leech`, req);
};

export const UpdateAntiLeech = (req: Website.LeechConfig) => {
    return http.post<any>(`/websites/leech/update`, req);
};

export const GetRedirectConfig = (req: Website.WebsiteReq) => {
    return http.post<Website.RedirectConfig[]>(`/websites/redirect`, req);
};

export const OperateRedirectConfig = (req: Website.WebsiteReq) => {
    return http.post<any>(`/websites/redirect/update`, req);
};

export const UpdateRedirectConfigFile = (req: Website.RedirectFileUpdate) => {
    return http.post<any>(`/websites/redirect/file`, req);
};

export const ChangePHPVersion = (req: Website.PHPVersionChange) => {
    return http.post<any>(`/websites/php/version`, req);
};

export const GetDirConfig = (req: Website.ProxyReq) => {
    return http.post<Website.DirConfig>(`/websites/dir`, req);
};

export const UploadSSL = (req: Website.SSLUpload) => {
    return http.post<any>(`/websites/ssl/upload`, req);
};

export const SearchCAs = (req: ReqPage) => {
    return http.post<ResPage<Website.CA>>(`/websites/ca/search`, req);
};

export const CreateCA = (req: Website.CACreate) => {
    return http.post<Website.CA>(`/websites/ca`, req);
};

export const ObtainSSLByCA = (req: Website.SSLObtainByCA) => {
    return http.post<any>(`/websites/ca/obtain`, req);
};

export const DeleteCA = (req: Website.DelReq) => {
    return http.post<any>(`/websites/ca/del`, req);
};

export const RenewSSLByCA = (req: Website.RenewSSLByCA) => {
    return http.post<any>(`/websites/ca/renew`, req);
};

export const DownloadFile = (params: Website.SSLDownload) => {
    return http.download<BlobPart>(`/websites/ssl/download`, params, {
        responseType: 'blob',
        timeout: TimeoutEnum.T_40S,
    });
};

export const GetCA = (id: number) => {
    return http.get<Website.CADTO>(`/websites/ca/${id}`);
};

export const GetDefaultHtml = (type: string) => {
    return http.get<Website.WebsiteHtml>(`/websites/default/html/${type}`);
};

export const UpdateDefaultHtml = (req: Website.WebsiteHtmlUpdate) => {
    return http.post(`/websites/default/html/update`, req);
};

export const DownloadCAFile = (params: Website.SSLDownload) => {
    return http.download<BlobPart>(`/websites/ca/download`, params, {
        responseType: 'blob',
        timeout: TimeoutEnum.T_40S,
    });
};
