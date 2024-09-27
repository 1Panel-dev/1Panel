import { ReqPage } from '.';

export namespace Container {
    export interface ContainerOperate {
        names: Array<string>;
        operation: string;
    }
    export interface ContainerRename {
        name: string;
        newName: string;
    }
    export interface ContainerCommit {
        containerID: string;
        containerName: string;
        newImageName: string;
        comment: string;
        author: string;
        pause: boolean;
    }
    export interface ContainerSearch extends ReqPage {
        name: string;
        state: string;
        filters: string;
        orderBy: string;
        order: string;
    }
    export interface ResourceLimit {
        cpu: number;
        memory: number;
    }
    export interface ContainerHelper {
        containerID: string;
        name: string;
        image: string;
        imageInput: boolean;
        forcePull: boolean;
        network: string;
        ipv4: string;
        ipv6: string;
        cmdStr: string;
        entrypointStr: string;
        memoryItem: number;
        cmd: Array<string>;
        openStdin: boolean;
        tty: boolean;
        entrypoint: Array<string>;
        publishAllPorts: boolean;
        exposedPorts: Array<Port>;
        nanoCPUs: number;
        cpuShares: number;
        memory: number;
        volumes: Array<Volume>;
        privileged: boolean;
        autoRemove: boolean;
        labels: Array<string>;
        labelsStr: string;
        env: Array<string>;
        envStr: string;
        restartPolicy: string;
    }
    export interface Port {
        host: string;
        hostIP: string;
        containerPort: string;
        hostPort: string;
        protocol: string;
    }
    export interface Volume {
        type: string;
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
        network: Array<string>;
        ports: Array<string>;
        isFromApp: boolean;
        isFromCompose: boolean;

        hasLoad: boolean;
        cpuPercent: number;
        memoryPercent: number;
    }
    export interface ContainerListStats {
        containerID: string;
        cpuTotalUsage: number;
        systemUsage: number;
        cpuPercent: number;
        percpuUsage: number;
        memoryCache: number;
        memoryUsage: number;
        memoryLimit: number;
        memoryPercent: number;
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
    export interface ContainerInspect {
        id: string;
        type: string;
    }
    export interface ContainerPrune {
        pruneType: string;
        withTagAll: boolean;
    }
    export interface ContainerPruneReport {
        deletedNumber: number;
        spaceReclaimed: number;
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
        isUsed: boolean;
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
    export interface RepoDelete {
        ids: Array<number>;
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
        envStr: string;
        env: Array<string>;
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
        env: Array<string>;
        envStr: string;
    }
    export interface ComposeOperation {
        name: string;
        operation: string;
        path: string;
        withFile: boolean;
    }
    export interface ComposeUpdate {
        name: string;
        path: string;
        content: string;
        env: Array<string>;
        createdBy: string;
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
        names: Array<string>;
    }

    export interface DaemonJsonUpdateByFile {
        file: string;
    }
    export interface DaemonJsonConf {
        isSwarm: boolean;
        status: string;
        version: string;
        registryMirrors: Array<string>;
        insecureRegistries: Array<string>;
        liveRestore: boolean;
        iptables: boolean;
        cgroupDriver: string;

        ipv6: boolean;
        fixedCidrV6: string;
        ip6Tables: boolean;
        experimental: boolean;

        logMaxSize: string;
        logMaxFile: string;
    }

    export interface ContainerLogInfo {
        container: string;
        since: string;
        tail: number;
        containerType: string;
    }
}
