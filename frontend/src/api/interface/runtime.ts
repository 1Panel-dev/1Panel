import { CommonModel, ReqPage } from '.';
import { App } from './app';
export namespace Runtime {
    export interface Runtime extends CommonModel {
        name: string;
        appDetailId: number;
        image: string;
        workDir: string;
        dockerCompose: string;
        env: string;
        params: string;
        type: string;
        resource: string;
        version: string;
        status: string;
    }

    export interface RuntimeReq extends ReqPage {
        name?: string;
        status?: string;
    }

    export interface RuntimeDTO extends Runtime {
        appParams: App.InstallParams[];
        appId: number;
        source?: string;
    }

    export interface RuntimeCreate {
        id?: number;
        name: string;
        appDetailId: number;
        image: string;
        params: object;
        type: string;
        resource: string;
        appId?: number;
        version?: string;
        rebuild?: boolean;
        source?: string;
    }

    export interface RuntimeUpdate {
        name: string;
        appDetailId: number;
        image: string;
        params: object;
        type: string;
        resource: string;
        appId?: number;
        version?: string;
        rebuild?: boolean;
    }

    export interface RuntimeDelete {
        id: number;
    }
}
