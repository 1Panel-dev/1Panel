import { ReqPage, CommonModel } from '.';

export namespace App {
    export interface App extends CommonModel {
        name: string;
        icon: string;
        key: string;
        tags: Tag[];
        shortDesc: string;
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
        canUpdate: boolean;
        version: string;
        items: App.App[];
        tags: App.Tag[];
    }

    export interface AppDetail extends CommonModel {
        appId: string;
        icon: string;
        version: string;
        readme: string;
        params: AppParams;
        dockerCompose: string;
    }

    export interface AppReq extends ReqPage {
        name?: string;
        tags?: string[];
        type?: string;
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
    }

    export interface AppInstall {
        appDetailId: number;
        params: any;
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
    }

    export interface AppInstalledOp {
        installId: number;
        operate: string;
        backupId?: number;
        detailId?: number;
    }

    export interface AppInstalledSearch {
        type: string;
    }

    export interface AppService {
        label: string;
        value: string;
    }

    export interface AppBackupReq extends ReqPage {
        appInstallId: number;
    }

    export interface AppBackupDelReq {
        ids: number[];
    }

    export interface AppBackup extends CommonModel {
        name: string;
        path: string;
        appInstallId: string;
        appDetail: AppDetail;
    }

    export interface VersionDetail {
        version: string;
        detailId: number;
    }
}
