import { CommonModel, ReqPage } from '.';
import { App } from './app';
export namespace Runtime {
    export interface Runtime extends CommonModel {
        name: string;
        appDetailID: number;
        image: string;
        workDir: string;
        dockerCompose: string;
        env: string;
        params: string;
        type: string;
        resource: string;
        version: string;
        status: string;
        codeDir: string;
        port: number;
        appID: number;
    }

    export interface RuntimeReq extends ReqPage {
        name?: string;
        status?: string;
        type?: string;
    }

    export interface NodeReq {
        codeDir: string;
    }

    export interface NodeScripts {
        name: string;
        script: string;
    }

    export interface RuntimeDTO extends Runtime {
        appParams: App.InstallParams[];
        appID: number;
        source?: string;
        path?: string;
        exposedPorts?: ExposedPort[];
    }

    export interface RuntimeCreate {
        id?: number;
        name: string;
        appDetailID: number;
        image: string;
        params: Object;
        type: string;
        resource: string;
        appID?: number;
        version?: string;
        rebuild?: boolean;
        source?: string;
        codeDir?: string;
        port?: number;
        exposedPorts?: ExposedPort[];
    }

    export interface ExposedPort {
        hostPort: number;
        containerPort: number;
    }

    export interface RuntimeUpdate {
        name: string;
        appDetailID: number;
        image: string;
        params: object;
        type: string;
        resource: string;
        appID?: number;
        version?: string;
        rebuild?: boolean;
    }

    export interface RuntimeDelete {
        id: number;
        forceDelete: boolean;
    }

    export interface RuntimeOperate {
        ID: number;
        operate: string;
    }

    export interface NodeModule {
        name: string;
        version: string;
        description: string;
    }

    export interface NodeModuleReq {
        ID: number;
        Operate?: string;
        Module?: string;
        PkgManager?: string;
    }

    export interface PHPExtensions extends CommonModel {
        id: number;
        name: string;
        extensions: string;
    }

    export interface PHPExtensionsList extends ReqPage {
        all: boolean;
    }

    export interface PHPExtensionsCreate {
        name: string;
        extensions: string;
    }

    export interface PHPExtensionsUpdate {
        id: number;
        name: string;
        extensions: string;
    }

    export interface PHPExtensionsDelete {
        id: number;
    }
}
