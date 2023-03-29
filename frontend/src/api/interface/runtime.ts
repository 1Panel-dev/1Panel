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
}
