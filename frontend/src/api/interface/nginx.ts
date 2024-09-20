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

    export interface NginxBuildReq {
        taskID: string;
    }

    export interface NginxModule {
        name: string;
        script?: string;
        packages?: string[];
        enable: boolean;
        params: string;
    }

    export interface NginxModuleUpdate {
        operate: string;
        name: string;
        script?: string;
        packages?: string;
        enable: boolean;
        params: string;
    }
}
