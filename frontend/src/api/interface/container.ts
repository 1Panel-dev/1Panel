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

    export interface ImageInfo {
        id: string;
        createdAt: Date;
        name: string;
        version: string;
        size: string;
    }
    export interface ImagePull {
        repoID: number;
        imageName: string;
    }
    export interface ImagePush {
        repoID: number;
        imageName: string;
        tagName: string;
    }
    export interface ImageRemove {
        imageName: string;
    }
    export interface ImageLoad {
        path: string;
    }
    export interface ImageSave {
        imageName: string;
        path: string;
        name: string;
    }

    export interface RepoCreate {
        name: string;
        downloadUrl: string;
        protocol: string;
        username: string;
        password: string;
        auth: boolean;
    }
    export interface RepoUpdate {
        id: number;
        downloadUrl: string;
        protocol: string;
        username: string;
        password: string;
        auth: boolean;
    }
    export interface RepoInfo {
        id: number;
        createdAt: Date;
        name: string;
        downloadUrl: string;
        protocol: string;
        username: string;
        password: string;
        auth: boolean;
    }
    export interface RepoOptions {
        id: number;
        name: string;
        downloadUrl: string;
    }
}
