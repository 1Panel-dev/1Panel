import { ReqPage } from '.';

export namespace Container {
    export interface ContainerOperate {
        containerID: string;
        operation: string;
        newName: string;
    }
    export interface ContainerSearch extends ReqPage {
        filters: string;
    }
    export interface ContainerCreate {
        name: string;
        image: string;
        cmd: Array<string>;
        publishAllPorts: boolean;
        exposedPorts: Array<Port>;
        nanoCPUs: number;
        memory: number;
        volumes: Array<Volume>;
        autoRemove: boolean;
        labels: Array<string>;
        labelsStr: string;
        env: Array<string>;
        envStr: string;
        restartPolicy: string;
    }
    export interface Port {
        containerPort: number;
        hostPort: number;
    }
    export interface Volume {
        sourceDir: string;
        containerDir: string;
        mode: string;
    }
    export interface ContainerInfo {
        containerID: string;
        name: string;
        imageName: string;
        createTime: string;
        state: string;
        runTime: string;
    }
    export interface ContainerStats {
        cpuPercent: number;
        memory: number;
        cache: number;
        ioRead: number;
        ioWrite: number;
        networkRX: number;
        networkTX: number;
        shotTime: Date;
    }
    export interface ContainerLogSearch {
        containerID: string;
        mode: string;
    }
    export interface ContainerInspect {
        id: string;
        type: string;
    }
    export interface Options {
        option: string;
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
        name: string;
        dockerfile: string;
        tags: Array<string>;
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
        isSystem: boolean;
        labels: Array<string>;
        driver: string;
        ipamDriver: string;
        subnet: string;
        gateway: string;
        createdAt: string;
        attachable: string;
        expand: boolean;
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

    export interface ComposeInfo {
        name: string;
        createdAt: string;
        createdBy: string;
        containerNumber: number;
        configFile: string;
        workdir: string;
        path: string;
        containers: Array<ComposeContainer>;
        expand: boolean;
    }
    export interface ComposeContainer {
        name: string;
        createTime: string;
        containerID: string;
        state: string;
    }
    export interface ComposeCreate {
        name: string;
        from: string;
        file: string;
        path: string;
        template: number;
    }
    export interface ComposeOpration {
        operation: string;
        path: string;
    }
    export interface ComposeUpdate {
        path: string;
        content: string;
    }

    export interface TemplateCreate {
        name: string;
        from: string;
        description: string;
        path: string;
        content: string;
    }
    export interface TemplateUpdate {
        id: number;
        from: string;
        description: string;
        path: string;
        content: string;
    }
    export interface TemplateInfo {
        id: number;
        createdAt: Date;
        name: string;
        from: string;
        description: string;
        path: string;
        content: string;
    }

    export interface BatchDelete {
        ids: Array<string>;
    }

    export interface DaemonJsonUpdateByFile {
        path: string;
        file: string;
    }

    export interface DaemonJsonConf {
        status: string;
        bip: string;
        registryMirrors: Array<string>;
        insecureRegistries: Array<string>;
        liveRestore: boolean;
        cgroupDriver: string;
    }
}
