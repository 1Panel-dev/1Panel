export namespace Nginx {
    export interface NginxScopeReq {
        scope: string;
    }
    export interface NginxParam {
        name: string;
        secondKey: string;
        params: string[];
    }

    export interface NginxConfigReq {
        operate: string;
        websiteId?: number;
        scope: string;
        params?: any;
    }
}
