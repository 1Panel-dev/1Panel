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
    }

    export interface RuntimeReq extends ReqPage {
        name?: string;
        status?: string;
    }

    export interface RuntimeDTO extends Runtime {
        appParams: App.InstallParams[];
        appId: number;
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
    }

    export interface RuntimeDelete {
        id: number;
    }
}
