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
        accepts: number;
        handled: number;
        active: number;
        requests: number;
        reading: number;
        writing: number;
        waiting: number;
    }

    export interface NginxFileUpdate {
        content: string;
        backup: boolean;
    }

    export interface NginxBuildReq {
        taskID: string;
        mirror: string;
    }

    export interface NginxModule {
        name: string;
        script?: string;
        packages?: string;
        enable: boolean;
        params: string;
    }

    export interface NginxBuildConfig {
        mirror: string;
        modules: NginxModule[];
    }

    export interface NginxModuleUpdate extends NginxModule {
        operate: string;
    }

    export interface NginxHttpsStatus {
        https: boolean;
        sslRejectHandshake: boolean;
    }

    export interface NginxOperateReq {
        operate: string;
    }

    export interface NginxHttpsOperateReq extends NginxOperateReq {
        sslRejectHandshake: boolean;
    }
}
