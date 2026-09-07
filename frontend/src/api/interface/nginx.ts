export namespace Nginx {
    export interface NginxScopeReq {
        scope: string;
    }
    export interface NginxParam {
        name: string;
        params: string[];
    }

    export interface NginxBrotliRes {
        params: NginxParam[];
        managedExternally: boolean;
        managedUnavailable: boolean;
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
        modules?: string[];
        force?: boolean;
    }

    export interface NginxModuleArtifact {
        name: string;
        path: string;
        checksum: string;
    }

    export interface NginxModule {
        name: string;
        custom: boolean;
        script?: string;
        packages?: string;
        enable: boolean;
        params: string;
        buildMode: 'dynamic' | 'static';
        provider: 'local' | 'prebuilt';
        loadOrder: number;
        buildStatus: 'pending' | 'ready' | 'failed';
        loadStatus: 'enabled' | 'disabled';
        artifacts?: NginxModuleArtifact[];
        lastError?: string;
    }

    export interface NginxBuildConfig {
        mirror: string;
        modules: NginxModule[];
        dynamicSupported: boolean;
    }

    export interface NginxModuleUpdate {
        operate: string;
        name: string;
        script?: string;
        packages?: string;
        enable?: boolean;
        params?: string;
        buildMode?: 'dynamic' | 'static';
        provider?: 'local' | 'prebuilt';
        loadOrder?: number;
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
