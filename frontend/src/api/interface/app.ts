import { ReqPage, CommonModel } from '.';

export namespace App {
    export interface App extends CommonModel {
        name: string;
        icon: string;
        key: string;
        tags: Tag[];
        shortDescZh: string;
        shortDescEn: string;
        author: string;
        source: string;
        type: string;
    }

    export interface AppDTO extends App {
        versions: string[];
    }

    export interface Tag {
        key: string;
        name: string;
    }

    export interface AppResPage {
        total: number;
        items: App.App[];
    }

    export interface AppUpdateRes {
        version: string;
        canUpdate: boolean;
    }

    export interface AppDetail extends CommonModel {
        appId: string;
        icon: string;
        version: string;
        readme: string;
        params: AppParams;
        dockerCompose: string;
        enbale: boolean;
    }

    export interface AppReq extends ReqPage {
        name?: string;
        tags?: string[];
        type?: string;
        recommend?: boolean;
    }

    export interface AppParams {
        formFields: FromField[];
    }

    export interface FromField {
        type: string;
        labelZh: string;
        labelEn: string;
        required: boolean;
        default: any;
        envKey: string;
        key?: string;
        values?: ServiceParam[];
        child?: FromFieldChild;
        params?: FromParam[];
    }

    export interface FromFieldChild extends FromField {
        services: App.AppService[];
    }

    export interface FromParam {
        type: string;
        key: string;
        value: string;
        envKey: string;
    }

    export interface ServiceParam {
        label: '';
        value: '';
    }

    export interface AppInstall {
        appDetailId: number;
        params: any;
    }

    export interface AppInstallSearch extends ReqPage {
        name?: string;
        tags?: string[];
        update?: boolean;
        unused?: boolean;
    }
    export interface ChangePort {
        key: string;
        name: string;
        port: number;
    }

    export interface AppInstalled extends CommonModel {
        name: string;
        appId: string;
        appDetailId: string;
        env: string;
        status: string;
        description: string;
        message: string;
        icon: string;
        canUpdate: boolean;
        app: App;
    }

    export interface CheckInstalled {
        name: string;
        version: string;
        isExist: boolean;
        app: string;
        status: string;
        createdAt: string;
        lastBackupAt: string;
        appInstallId: number;
        containerName: string;
        installPath: string;
    }

    export interface AppInstallResource {
        type: string;
        name: string;
    }

    export interface AppInstalledOp {
        installId: number;
        operate: string;
        backupId?: number;
        detailId?: number;
        forceDelete?: boolean;
        deleteBackup?: boolean;
    }

    export interface AppInstalledSearch {
        type: string;
        unused?: boolean;
    }

    export interface AppService {
        label: string;
        value: string;
        config?: Object;
    }

    export interface VersionDetail {
        version: string;
        detailId: number;
    }

    export interface InstallParams {
        labelZh: string;
        labelEn: string;
        value: any;
        edit: boolean;
        key: string;
        rule: string;
        type: string;
    }
}
