import { CommonModel, ReqPage } from '.';
export namespace Runtime {
    export interface Runtime extends CommonModel {
        name: string;
        appDetailId: string;
        image: string;
        workDir: string;
        dockerCompose: string;
        env: string;
        params: string;
        type: string;
    }

    export interface RuntimeReq extends ReqPage {
        name: string;
    }

    export interface RuntimeDTO extends Runtime {
        websites: string[];
    }

    export interface RuntimeCreate {
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
