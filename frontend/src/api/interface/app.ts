import { ReqPage } from '.';

export namespace App {
    export interface App {
        name: string;
        icon: string;
        key: string;
        tags: Tag[];
        shortDesc: string;
        author: string;
        source: string;
        type: string;
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

    export interface AppDetail {
        name: string;
        icon: string;
        description: string;
        sourceLink: string;
        versions: string[];
        readme: string;
        athor: string;
    }

    export interface AppReq extends ReqPage {
        name: string;
        types: string[];
    }
}
