import { ReqPage } from '.';

export namespace Container {
    export interface ContainerOperate {
        containerID: string;
        operation: string;
        newName: string;
    }
    export interface ContainerSearch extends ReqPage {
        status: string;
    }
    export interface ContainerInfo {
        containerID: string;
        name: string;
        imageName: string;
        createTime: string;
        state: string;
        runTime: string;
    }
    export interface ContainerLogSearch {
        containerID: string;
        mode: string;
    }
}
