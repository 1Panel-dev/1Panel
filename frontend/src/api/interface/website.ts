import { CommonModel, ReqPage } from '.';

export namespace WebSite {
    export interface WebSite extends CommonModel {
        primaryDomain: string;
        type: string;
        alias: string;
        remark: string;
        domains: string[];
        appType: string;
        appInstallID?: number;
        webSiteGroupID: number;
        otherDomains: string;
        appinstall?: NewAppInstall;
    }

    export interface NewAppInstall {
        name: string;
        appDetailID: number;
        params: any;
    }

    export interface WebSiteSearch extends ReqPage {
        name: string;
    }

    export interface WebSiteDel {
        id: number;
        deleteApp: boolean;
        deleteBackup: boolean;
    }

    export interface WebSiteCreateReq {
        primaryDomain: string;
        type: string;
        alias: string;
        remark: string;
        appType: string;
        appInstallID: number;
        webSiteGroupID: number;
        otherDomains: string;
    }

    export interface Group extends CommonModel {
        name: string;
        default: boolean;
    }

    export interface GroupOp {
        name: string;
        id?: number;
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

    export interface NginxConfigReq {
        operate: string;
        websiteId: number;
        scope: string;
        params?: any;
    }

    export interface NginxParam {
        name: string;
        secondKey: string;
        isRepeatKey: string;
        params: string[];
    }
}
