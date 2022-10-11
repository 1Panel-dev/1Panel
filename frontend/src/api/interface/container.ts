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

    export interface ImageInfo {
        id: string;
        createdAt: Date;
        name: string;
        version: string;
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

    export interface NetworkInfo {
        id: string;
        name: string;
        labels: Array<string>;
        driver: string;
        ipamDriver: string;
        ipv4Subnet: string;
        ipv4Gateway: string;
        ipv6Subnet: string;
        ipv6Gateway: string;
        createdAt: string;
        attachable: string;
    }
    export interface NetworkCreate {
        name: string;
        labels: Array<string>;
        options: Array<string>;
        driver: string;
        ipv4Subnet: string;
        ipv4Gateway: string;
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
        label: Array<string>;
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
