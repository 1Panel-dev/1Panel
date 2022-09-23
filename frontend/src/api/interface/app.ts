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
        formFields: string;
        dockerCompose: string;
    }

    export interface AppReq extends ReqPage {
        name: string;
        tags: string[];
    }
}
