import http from '@/api';
import { ReqPage, ResPage } from '../interface';
import { Website } from '../interface/website';
import { File } from '../interface/file';
import { TimeoutEnum } from '@/enums/http-enum';
import { deepCopy } from '@/utils/util';
import { Base64 } from 'js-base64';

export const searchWebsites = (req: Website.WebSiteSearch) => {
    return http.post<ResPage<Website.WebsiteRes>>(`/websites/search`, req);
};

export const listWebsites = () => {
    return http.get<Website.WebsiteDTO[]>(`/websites/list`);
};

export const createWebsite = (req: Website.WebSiteCreateReq) => {
    let request = deepCopy(req) as Website.WebSiteCreateReq;
    if (request.ftpPassword) {
        request.ftpPassword = Base64.encode(request.ftpPassword);
    }
    return http.post<any>(`/websites`, request, TimeoutEnum.T_10M);
};

export const opWebsite = (req: Website.WebSiteOp) => {
    return http.post<any>(`/websites/operate`, req);
};

export const opWebsiteLog = (req: Website.WebSiteOpLog) => {
    return http.post<Website.WebSiteLog>(`/websites/log`, req);
};

export const updateWebsite = (req: Website.WebSiteUpdateReq) => {
    return http.post<any>(`/websites/update`, req);
};

export const getWebsite = (id: number) => {
    return http.get<Website.WebsiteDTO>(`/websites/${id}`);
};

export const getWebsiteOptions = (req: Website.OptionReq) => {
    return http.post<any>(`/websites/options`, req);
};

export const getWebsiteConfig = (id: number, type: string) => {
    return http.get<File.File>(`/websites/${id}/config/${type}`);
};

export const deleteWebsite = (req: Website.WebSiteDel) => {
    return http.post<any>(`/websites/del`, req);
};

export const listDomains = (id: number) => {
    return http.get<Website.Domain[]>(`/websites/domains/${id}`);
};

export const deleteDomain = (req: Website.DomainDelete) => {
    return http.post<any>(`/websites/domains/del/`, req);
};

export const createDomain = (req: Website.DomainCreate) => {
    return http.post<any>(`/websites/domains`, req);
};

export const updateDomain = (req: Website.DomainUpdate) => {
    return http.post<any>(`/websites/domains/update`, req);
};

export const getNginxConfig = (req: Website.NginxScopeReq) => {
    return http.post<Website.NginxScopeConfig>(`/websites/config`, req);
};

export const updateNginxConfig = (req: Website.NginxConfigReq) => {
    return http.post<any>(`/websites/config/update`, req);
};

export const searchDnsAccount = (req: ReqPage) => {
    return http.post<ResPage<Website.DnsAccount>>(`/websites/dns/search`, req);
};

export const createDnsAccount = (req: Website.DnsAccountCreate) => {
    return http.post<any>(`/websites/dns`, req);
};

export const updateDnsAccount = (req: Website.DnsAccountUpdate) => {
    return http.post<any>(`/websites/dns/update`, req);
};

export const deleteDnsAccount = (req: Website.DelReq) => {
    return http.post<any>(`/websites/dns/del`, req);
};

export const searchAcmeAccount = (req: ReqPage) => {
    return http.post<ResPage<Website.AcmeAccount>>(`/websites/acme/search`, req);
};

export const createAcmeAccount = (req: Website.AcmeAccountCreate) => {
    return http.post<Website.AcmeAccount>(`/websites/acme`, req, TimeoutEnum.T_10M);
};

export const deleteAcmeAccount = (req: Website.DelReq) => {
    return http.post<any>(`/websites/acme/del`, req);
};

export const updateAcmeAccount = (req: Website.AcmeAccountUpdate) => {
    return http.post<Website.AcmeAccount>(`/websites/acme/update`, req, TimeoutEnum.T_10M);
};

export const searchSSL = (req: ReqPage) => {
    return http.post<ResPage<Website.SSLDTO>>(`/websites/ssl/search`, req);
};

export const listSSL = (req: Website.SSLReq) => {
    return http.post<Website.SSLDTO[]>(`/websites/ssl/search`, req);
};

export const listLocalNodeSSL = (req: Website.SSLReq) => {
    return http.postLocalNode<Website.SSLDTO[]>(`/websites/ssl/search`, req);
};

export const createSSL = (req: Website.SSLCreate) => {
    return http.post<Website.SSLCreate>(`/websites/ssl`, req, TimeoutEnum.T_10M);
};

export const deleteSSL = (req: Website.DelReq) => {
    return http.post<any>(`/websites/ssl/del`, req);
};

export const getSSL = (id: number) => {
    return http.get<Website.SSL>(`/websites/ssl/${id}`);
};

export const obtainSSL = (req: Website.SSLObtain) => {
    return http.post<any>(`/websites/ssl/obtain`, req);
};

export const updateSSL = (req: Website.SSLUpdate) => {
    return http.post<any>(`/websites/ssl/update`, req);
};

export const getDnsResolve = (req: Website.DNSResolveReq) => {
    return http.post<Website.DNSResolve[]>(`/websites/ssl/resolve`, req, TimeoutEnum.T_5M);
};

export const getHTTPSConfig = (id: number) => {
    return http.get<Website.HTTPSConfig>(`/websites/${id}/https`);
};

export const updateHTTPSConfig = (req: Website.HTTPSReq) => {
    return http.post<Website.HTTPSConfig>(`/websites/${req.websiteId}/https`, req);
};

export const preCheck = (req: Website.CheckReq) => {
    return http.post<Website.CheckRes[]>(`/websites/check`, req);
};

export const updateNginxFile = (req: Website.NginxUpdate) => {
    return http.post<any>(`/websites/nginx/update`, req);
};

export const changeDefaultServer = (req: Website.DefaultServerUpdate) => {
    return http.post<any>(`/websites/default/server`, req);
};

export const getRewriteConfig = (req: Website.RewriteReq) => {
    return http.post<Website.RewriteRes>(`/websites/rewrite`, req);
};

export const updateRewriteConfig = (req: Website.RewriteUpdate) => {
    return http.post<any>(`/websites/rewrite/update`, req);
};

export const updateWebsiteDir = (req: Website.DirUpdate) => {
    return http.post<any>(`/websites/dir/update`, req);
};

export const updateWebsiteDirPermission = (req: Website.DirPermissionUpdate) => {
    return http.post<any>(`/websites/dir/permission`, req);
};

export const getProxyConfig = (req: Website.ProxyReq) => {
    return http.post<Website.ProxyConfig[]>(`/websites/proxies`, req);
};

export const operateProxyConfig = (req: Website.ProxyReq) => {
    return http.post<any>(`/websites/proxies/update`, req);
};

export const updateProxyConfigFile = (req: Website.ProxyFileUpdate) => {
    return http.post<any>(`/websites/proxies/file`, req);
};

export const clearProxyCache = (req: Website.WebsiteReq) => {
    return http.post(`/websites/proxy/clear`, req);
};

export const getAuthConfig = (req: Website.AuthReq) => {
    return http.post<Website.AuthConfig>(`/websites/auths`, req);
};

export const operateAuthConfig = (req: Website.NginxAuthConfig) => {
    return http.post<any>(`/websites/auths/update`, req);
};

export const getPathAuthConfig = (req: Website.AuthReq) => {
    return http.post<Website.NginxPathAuthConfig[]>(`/websites/auths/path`, req);
};

export const operatePathAuthConfig = (req: Website.NginxPathAuthConfig) => {
    return http.post(`/websites/auths/path/update`, req);
};

export const getAntiLeech = (req: Website.LeechReq) => {
    return http.post<Website.LeechConfig>(`/websites/leech`, req);
};

export const updateAntiLeech = (req: Website.LeechConfig) => {
    return http.post<any>(`/websites/leech/update`, req);
};

export const getRedirectConfig = (req: Website.WebsiteReq) => {
    return http.post<Website.RedirectConfig[]>(`/websites/redirect`, req);
};

export const operateRedirectConfig = (req: Website.WebsiteReq) => {
    return http.post<any>(`/websites/redirect/update`, req);
};

export const updateRedirectConfigFile = (req: Website.RedirectFileUpdate) => {
    return http.post<any>(`/websites/redirect/file`, req);
};

export const changePHPVersion = (req: Website.PHPVersionChange) => {
    return http.post<any>(`/websites/php/version`, req);
};

export const getDirConfig = (req: Website.ProxyReq) => {
    return http.post<Website.DirConfig>(`/websites/dir`, req);
};

export const uploadSSL = (req: Website.SSLUpload) => {
    return http.post<any>(`/websites/ssl/upload`, req);
};

export const uploadSSLFile = (params: FormData) => {
    return http.upload<File.File>(`/websites/ssl/upload/file`, params, {});
};

export const searchCAs = (req: ReqPage) => {
    return http.post<ResPage<Website.CA>>(`/websites/ca/search`, req);
};

export const createCA = (req: Website.CACreate) => {
    return http.post<Website.CA>(`/websites/ca`, req);
};

export const obtainSSLByCA = (req: Website.SSLObtainByCA) => {
    return http.post<any>(`/websites/ca/obtain`, req);
};

export const deleteCA = (req: Website.DelReq) => {
    return http.post<any>(`/websites/ca/del`, req);
};

export const renewSSLByCA = (req: Website.RenewSSLByCA) => {
    return http.post<any>(`/websites/ca/renew`, req);
};

export const downloadFile = (params: Website.SSLDownload) => {
    return http.download<BlobPart>(`/websites/ssl/download`, params, {
        responseType: 'blob',
        timeout: TimeoutEnum.T_40S,
    });
};

export const getCA = (id: number) => {
    return http.get<Website.CADTO>(`/websites/ca/${id}`);
};

export const getDefaultHtml = (type: string) => {
    return http.get<Website.WebsiteHtml>(`/websites/default/html/${type}`);
};

export const updateDefaultHtml = (req: Website.WebsiteHtmlUpdate) => {
    return http.post(`/websites/default/html/update`, req);
};

export const downloadCAFile = (params: Website.SSLDownload) => {
    return http.download<BlobPart>(`/websites/ca/download`, params, {
        responseType: 'blob',
        timeout: TimeoutEnum.T_40S,
    });
};

export const getLoadBalances = (id: number) => {
    return http.get<Website.NginxUpstream[]>(`/websites/${id}/lbs`);
};

export const createLoadBalance = (req: Website.LoadBalanceReq) => {
    return http.post(`/websites/lbs/create`, req);
};

export const deleteLoadBalance = (req: Website.LoadBalanceDel) => {
    return http.post(`/websites/lbs/del`, req);
};

export const updateLoadBalance = (req: Website.LoadBalanceReq) => {
    return http.post(`/websites/lbs/update`, req);
};

export const updateLoadBalanceFile = (req: Website.WebsiteLBUpdateFile) => {
    return http.post(`/websites/lbs/file`, req);
};

export const updateCacheConfig = (req: Website.WebsiteCacheConfig) => {
    return http.post(`/websites/proxy/config`, req);
};

export const getCacheConfig = (id: number) => {
    return http.get<Website.WebsiteCacheConfig>(`/websites/proxy/config/${id}`);
};

export const updateRealIPConfig = (req: Website.WebsiteRealIPConfig) => {
    return http.post(`/websites/realip/config`, req);
};

export const getRealIPConfig = (id: number) => {
    return http.get<Website.WebsiteRealIPConfig>(`/websites/realip/config/${id}`);
};

export const getWebsiteResource = (id: number) => {
    return http.get<Website.WebsiteResource[]>(`/websites/resource/${id}`);
};

export const getWebsiteDatabase = () => {
    return http.get<Website.WebsiteDatabase[]>(`/websites/databases`);
};

export const changeDatabase = (req: Website.ChangeDatabase) => {
    return http.post(`/websites/databases`, req);
};

export const operateCustomRewrite = (req: Website.CustomRewirte) => {
    return http.post(`/websites/rewrite/custom`, req);
};

export const listCustomRewrite = () => {
    return http.get<string[]>(`/websites/rewrite/custom`);
};

export const operateCrossSiteAccess = (req: Website.CrossSiteAccessOp) => {
    return http.post(`/websites/crosssite`, req);
};

export const execComposer = (req: Website.ExecComposer) => {
    return http.post(`/websites/exec/composer`, req);
};

export const batchOpreate = (req: Website.BatchOperate) => {
    return http.post(`/websites/batch/operate`, req);
};

export const getCorsConfig = (id: number) => {
    return http.get<Website.CorsConfig>(`/websites/cors/${id}`);
};

export const updateCorsConfig = (req: Website.CorsConfigReq) => {
    return http.post(`/websites/cors/update`, req);
};

export const batchSetGroup = (req: Website.BatchSetGroup) => {
    return http.post(`/websites/batch/group`, req);
};
