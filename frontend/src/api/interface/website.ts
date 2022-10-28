import { CommonModel } from '.';

export namespace WebSite {
    export interface WebSiteCreateReq {
        primaryDomain: string;
        type: string;
        alias: string;
        remark: string;
        domains: string[];
        appType: string;
        appInstallID: number;
        webSiteGroupID: number;
        otherDomains: string;
    }

    export interface Group extends CommonModel {
        name: string;
    }
}
