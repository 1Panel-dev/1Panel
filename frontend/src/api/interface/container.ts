export namespace Container {
    export interface ContainerOperate {
        containerID: string;
        operation: string;
        newName: string;
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
    export interface ContainerInspect {
        id: string;
        type: string;
    }

    export interface ImageInfo {
        id: string;
        createdAt: Date;
        name: string;
        tags: Array<string>;
        size: string;
    }
    export interface ImageBuild {
        from: string;
        dockerfile: string;
    }
    export interface ImagePull {
        repoID: number;
        imageName: string;
    }
    export interface ImageTag {
        repoID: number;
        sourceID: string;
        targetName: string;
    }
    export interface ImagePush {
        repoID: number;
        tagName: string;
    }
    export interface ImageLoad {
        path: string;
    }
    export interface ImageSave {
        tagName: string;
        path: string;
        name: string;
    }

    export interface NetworkInfo {
        id: string;
        name: string;
        labels: Array<string>;
        driver: string;
        ipamDriver: string;
        subnet: string;
        gateway: string;
        createdAt: string;
        attachable: string;
    }
    export interface NetworkCreate {
        name: string;
        labels: Array<string>;
        options: Array<string>;
        driver: string;
        subnet: string;
        gateway: string;
        scope: string;
    }

    export interface VolumeInfo {
        name: string;
        labels: Array<string>;
        driver: string;
        mountpoint: string;
        createdAt: string;
    }
    export interface VolumeCreate {
        name: string;
        driver: string;
        options: Array<string>;
        labels: Array<string>;
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

    export interface BatchDelete {
        ids: Array<string>;
    }
}
