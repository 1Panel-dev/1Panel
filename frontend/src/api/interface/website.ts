import { CommonModel, ReqPage } from '.';

export namespace Website {
    export interface Website extends CommonModel {
        primaryDomain: string;
        type: string;
        alias: string;
        remark: string;
        domains: string[];
        appType: string;
        appInstallId?: number;
        webSiteGroupId: number;
        otherDomains: string;
        defaultServer: boolean;
        protocol: string;
        autoRenew: boolean;
        appinstall?: NewAppInstall;
        webSiteSSL: SSL;
        runtimeID: number;
        rewrite: string;
        user: string;
        group: string;
        IPV6: boolean;
    }

    export interface WebsiteDTO extends Website {
        errorLogPath: string;
        accessLogPath: string;
        sitePath: string;
        appName: string;
        runtimeName: string;
    }

    export interface NewAppInstall {
        name: string;
        appDetailId: number;
        params: any;
    }

    export interface WebSiteSearch extends ReqPage {
        name: string;
        orderBy: string;
        order: string;
        websiteGroupId: number;
    }

    export interface WebSiteDel {
        id: number;
        deleteApp: boolean;
        deleteBackup: boolean;
        forceDelete: boolean;
    }

    export interface WebSiteCreateReq {
        primaryDomain: string;
        type: string;
        alias: string;
        remark: string;
        appType: string;
        appInstallId: number;
        webSiteGroupId: number;
        otherDomains: string;
        proxy: string;
        proxyType: string;
    }

    export interface WebSiteUpdateReq {
        id: number;
        primaryDomain: string;
        remark: string;
        webSiteGroupId: number;
        expireDate?: string;
        IPV6: boolean;
    }

    export interface WebSiteOp {
        id: number;
        operate: string;
    }

    export interface WebSiteOpLog {
        id: number;
        operate: string;
        logType: string;
    }

    export interface WebSiteLog {
        enable: boolean;
        content: string;
    }

    export interface Domain {
        websiteId: number;
        port: number;
        id: number;
        domain: string;
    }

    export interface DomainCreate {
        websiteId: number;
        port: number;
        domain: string;
    }

    export interface DomainDelete {
        id: number;
    }

    export interface NginxConfigReq {
        operate: string;
        websiteId: number;
        scope: string;
        params?: any;
    }

    export interface NginxScopeReq {
        websiteId: number;
        scope: string;
    }

    export interface NginxParam {
        name: string;
        params: string[];
    }

    export interface NginxScopeConfig {
        enable: boolean;
        params: NginxParam[];
    }

    export interface DnsAccount extends CommonModel {
        name: string;
        type: string;
        authorization: Object;
    }

    export interface DnsAccountCreate {
        name: string;
        type: string;
        authorization: Object;
    }

    export interface DnsAccountUpdate {
        id: number;
        name: string;
        type: string;
        authorization: Object;
    }

    export interface SSL extends CommonModel {
        primaryDomain: string;
        privateKey: string;
        pem: string;
        otherDomains: string;
        certURL: string;
        type: string;
        issuerName: string;
        expireDate: string;
        startDate: string;
        provider: string;
        websites?: Website.Website[];
        autoRenew: boolean;
        acmeAccountId?: number;
    }

    export interface SSLCreate {
        primaryDomain: string;
        otherDomains: string;
        provider: string;
        acmeAccountId: number;
        dnsAccountId: number;
    }

    export interface SSLApply {
        websiteId: number;
        SSLId: number;
    }

    export interface SSLRenew {
        SSLId: number;
    }

    export interface SSLUpdate {
        id: number;
        autoRenew: boolean;
    }

    export interface AcmeAccount extends CommonModel {
        email: string;
        url: string;
    }

    export interface AcmeAccountCreate {
        email: string;
    }

    export interface DNSResolveReq {
        domains: string[];
        acmeAccountId: number;
    }

    export interface DNSResolve {
        resolve: string;
        value: string;
        domain: string;
        err: string;
    }

    export interface SSLReq {
        name?: string;
        acmeAccountID?: string;
    }

    export interface HTTPSReq {
        websiteId: number;
        enable: boolean;
        websiteSSLId?: number;
        type: string;
        certificate?: string;
        privateKey?: string;
        httpConfig: string;
        SSLProtocol: string[];
        algorithm: string;
    }

    export interface HTTPSConfig {
        enable: boolean;
        SSL: SSL;
        httpConfig: string;
        SSLProtocol: string[];
        algorithm: string;
    }

    export interface CheckReq {
        installIds?: number[];
    }

    export interface CheckRes {
        name: string;
        status: string;
        version: string;
        appName: string;
    }

    export interface WafReq {
        websiteId: number;
        key: string;
        rule: string;
    }

    export interface WafRes {
        enable: boolean;
        filePath: string;
        content: string;
    }

    export interface WafUpdate {
        enable: boolean;
        websiteId: number;
        key: string;
    }

    export interface DelReq {
        id: number;
    }

    export interface NginxUpdate {
        id: number;
        content: string;
    }

    export interface DefaultServerUpdate {
        id: number;
    }

    export interface PHPConfig {
        params: any;
        disableFunctions: string[];
        uploadMaxSize: string;
    }

    export interface PHPConfigUpdate {
        id: number;
        params?: any;
        disableFunctions?: string[];
        scope: string;
        uploadMaxSize?: string;
    }

    export interface PHPUpdate {
        id: number;
        content: string;
        type: string;
    }

    export interface RewriteReq {
        websiteID: number;
        name: string;
    }

    export interface RewriteRes {
        content: string;
    }

    export interface RewriteUpdate {
        websiteID: number;
        name: string;
        content: string;
    }

    export interface DirUpdate {
        id: number;
        siteDir: string;
    }

    export interface DirPermissionUpdate {
        id: number;
        user: string;
        group: string;
    }

    export interface ProxyReq {
        id: number;
    }

    export interface ProxyConfig {
        id: number;
        operate: string;
        enable: boolean;
        cache: boolean;
        cacheTime: number;
        cacheUnit: string;
        name: string;
        modifier: string;
        match: string;
        proxyPass: string;
        proxyHost: string;
        filePath?: string;
        replaces?: ProxReplace;
        content?: string;
    }

    export interface ProxReplace {
        [key: string]: string;
    }

    export interface ProxyFileUpdate {
        websiteID: number;
        name: string;
        content: string;
    }

    export interface AuthReq {
        websiteID: number;
    }

    export interface NginxAuth {
        username: string;
        remark: string;
    }

    export interface AuthConfig {
        enable: boolean;
        items: NginxAuth[];
    }

    export interface NginxAuthConfig {
        websiteID: number;
        operate: string;
        username: string;
        password: string;
        remark: string;
    }

    export interface LeechConfig {
        enable: boolean;
        cache: boolean;
        cacheTime: number;
        cacheUint: string;
        extends: string;
        return: string;
        serverNames: string[];
        noneRef: boolean;
        logEnable: boolean;
        blocked: boolean;
        websiteID?: number;
    }

    export interface LeechReq {
        websiteID: number;
    }
}
