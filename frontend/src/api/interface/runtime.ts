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
        port: string;
        appID: number;
        remark: string;
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
        environments?: Environment[];
        volumes?: Volume[];
        extraHosts?: ExtraHost[];
        container: string;
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
        environments?: Environment[];
        volumes?: Volume[];
        extraHosts?: ExtraHost[];
        remark?: string;
    }

    export interface ExposedPort {
        hostPort: number;
        containerPort: number;
        hostIP: string;
    }
    export interface Environment {
        key: string;
        value: string;
    }
    export interface Volume {
        source: string;
        target: string;
    }

    export interface ExtraHost {
        hostname: string;
        ip: string;
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

    export interface PHPExtensionsRes {
        extensions: string[];
        supportExtensions: SupportExtension[];
    }

    export interface SupportExtension {
        name: string;
        description: string;
        installed: boolean;
        check: string;
        versions: string[];
    }

    export interface PHPExtensionInstall {
        name: string;
        id: number;
        taskID?: string;
    }

    export interface PHPConfig {
        params: any;
        disableFunctions: string[];
        uploadMaxSize: string;
        maxExecutionTime: string;
    }

    export interface PHPConfigUpdate {
        id: number;
        params?: any;
        disableFunctions?: string[];
        scope: string;
        uploadMaxSize?: string;
        maxExecutionTime?: string;
    }

    export interface PHPUpdate {
        id: number;
        content: string;
        type: string;
    }

    export interface PHPFileReq {
        id: number;
        type: string;
    }

    export interface FPMConfig {
        id: number;
        params: any;
    }

    export interface ProcessReq {
        operate: string;
        name: string;
        id: number;
    }

    export interface ProcessFileReq {
        operate: string;
        name: string;
        content?: string;
        file: string;
        id: number;
    }

    export interface SupersivorProcess {
        operate: string;
        name: string;
        command: string;
        user: string;
        dir: string;
        numprocs: string;
        id: number;
        environment: string;
    }

    export interface PHPContainerConfig {
        id: number;
        containerName: string;
        exposedPorts: ExposedPort[];
        environments: Environment[];
        volumes: Volume[];
        extraHosts: ExtraHost[];
    }

    export interface RemarkUpdate {
        id: number;
        remark: string;
    }

    export interface FpmStatus {
        key: string;
        value: any;
    }
}
