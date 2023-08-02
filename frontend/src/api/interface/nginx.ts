export namespace Nginx {
    export interface NginxScopeReq {
        scope: string;
    }
    export interface NginxParam {
        name: string;
        params: string[];
    }

    export interface NginxConfigReq {
        operate: string;
        websiteId?: number;
        scope: string;
        params?: any;
    }

    export interface NginxStatus {
        accepts: string;
        handled: string;
        active: string;
        requests: string;
        reading: string;
        writing: string;
        waiting: string;
    }

    export interface NginxFileUpdate {
        content: string;
        backup: boolean;
    }
}
