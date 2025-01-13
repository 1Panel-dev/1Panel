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
        accessLog?: boolean;
        errorLog?: boolean;
    }

    export interface WebsiteDTO extends Website {
        errorLogPath: string;
        accessLogPath: string;
        sitePath: string;
        appName: string;
        runtimeName: string;
        runtimeType: string;
    }
    export interface WebsiteRes extends CommonModel {
        protocol: string;
        primaryDomain: string;
        type: string;
        alias: string;
        remark: string;
        status: string;
        expireDate: string;
        sitePath: string;
        appName: string;
        runtimeName: string;
        sslExpireDate: Date;
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
        ftpUser: string;
        ftpPassword: string;
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
        page?: number;
        pageSize?: number;
    }

    export interface WebSiteLog {
        enable: boolean;
        content: string;
        end: boolean;
        path: string;
    }

    export interface Domain {
        websiteId: number;
        port: number;
        id: number;
        domain: string;
    }

    export interface DomainCreate {
        websiteID: number;
        domains: string;
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
        acmeAccountId: number;
        status: string;
        domains: string;
        description: string;
        dnsAccountId?: number;
        pushDir: boolean;
        dir: string;
        keyType: string;
        nameserver1: string;
        nameserver2: string;
        disableCNAME: boolean;
        skipDNS: boolean;
        execShell: boolean;
        shell: string;
    }

    export interface SSLDTO extends SSL {
        logPath: string;
    }

    export interface SSLCreate {
        primaryDomain: string;
        otherDomains: string;
        provider: string;
        acmeAccountId: number;
        dnsAccountId: number;
        id?: number;
        description: string;
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
        description: string;
        primaryDomain: string;
        otherDomains: string;
        acmeAccountId: number;
        provider: string;
        dnsAccountId?: number;
        keyType: string;
        pushDir: boolean;
        dir: string;
    }

    export interface AcmeAccount extends CommonModel {
        email: string;
        url: string;
        type: string;
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
        hsts: boolean;
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

    export interface ProxyDel {
        id: number;
        name: string;
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
        proxyAddress?: string;
        proxyProtocol?: string;
        sni: boolean;
        proxySSLName: string;
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

    export interface WebsiteReq {
        websiteID: number;
    }

    export interface RedirectConfig {
        operate: string;
        websiteID: number;
        domains?: string[];
        enable: boolean;
        name: string;
        keepPath: boolean;
        type: string;
        redirect: string;
        path?: string;
        target: string;
        redirectRoot?: boolean;
        filePath?: string;
        content?: string;
    }

    export interface RedirectFileUpdate {
        websiteID: number;
        name: string;
        content: string;
    }

    export interface PHPVersionChange {
        websiteID: number;
        runtimeID: number;
        retainConfig: boolean;
    }

    export interface DirConfig {
        dirs: string[];
        user: string;
        userGroup: string;
        msg: string;
    }

    export interface SSLUpload {
        privateKey: string;
        certificate: string;
        privateKeyPath: string;
        certificatePath: string;
        type: string;
        sslID: number;
    }

    export interface SSLObtain {
        ID: number;
    }

    export interface CA extends CommonModel {
        name: string;
        csr: string;
        privateKey: string;
        keyType: string;
    }

    export interface CACreate {
        name: string;
        commonName: string;
        country: string;
        organization: string;
        organizationUint: string;
        keyType: string;
        province: string;
        city: string;
    }

    export interface CADTO extends CA {
        commonName: string;
        country: string;
        organization: string;
        organizationUint: string;
        province: string;
        city: string;
    }

    export interface SSLObtainByCA {
        id: number;
        domains: string;
        keyType: string;
        time: number;
        unit: string;
        pushDir: boolean;
        dir: string;
        description: string;
    }

    export interface RenewSSLByCA {
        SSLID: number;
    }

    export interface SSLDownload {
        id: number;
    }

    export interface WebsiteHtml {
        content: string;
    }
    export interface WebsiteHtmlUpdate {
        type: string;
        content: string;
    }
}
