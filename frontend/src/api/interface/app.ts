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
        name: string;
        tags: string[];
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
    }

    export interface AppInstall {
        appDetailId: number;
        params: any;
    }

    export interface AppInstalled extends CommonModel {
        name: string;
        containerName: string;
        version: string;
        appId: string;
        appDetailId: string;
        params: string;
        status: string;
        description: string;
        message: string;
        appName: string;
        total: number;
        ready: number;
        icon: string;
    }

    export interface AppInstalledOp {
        installId: number;
        operate: string;
    }
}
