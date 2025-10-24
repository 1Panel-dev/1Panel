import { ReqPage } from '.';

export namespace AI {
    export interface OllamaModelInfo {
        id: number;
        name: string;
        size: string;
        from: string;
        logFileExist: boolean;
        status: string;
        message: string;
        createdAt: Date;
    }
    export interface OllamaModelDropInfo {
        id: number;
        name: string;
    }
    export interface OllamaModelSearch extends ReqPage {
        info: string;
    }

    export interface Info {
        cudaVersion: string;
        driverVersion: string;
        type: string;
        gpu: GPU[];
    }
    export interface GPU {
        index: number;
        productName: string;
        persistenceMode: string;
        busID: string;
        displayActive: string;
        ecc: string;
        fanSpeed: string;

        temperature: string;
        performanceState: string;
        powerDraw: string;
        maxPowerLimit: string;
        memUsed: string;
        memTotal: string;
        gpuUtil: string;
        computeMode: string;
        migMode: string;
        processes: Process[];
    }
    export interface Process {
        pid: string;
        type: string;
        processName: string;
        usedMemory: string;
    }

    export interface XpuInfo {
        type: string;
        driverVersion: string;
        xpu: Xpu[];
    }

    interface Xpu {
        basic: Basic;
        stats: Stats;
        processes: XpuProcess[];
    }

    interface Basic {
        deviceID: number;
        deviceName: string;
        vendorName: string;
        driverVersion: string;
        memory: string;
        freeMemory: string;
        pciBdfAddress: string;
    }

    interface Stats {
        power: string;
        frequency: string;
        temperature: string;
        memoryUsed: string;
        memoryUtil: string;
    }

    interface XpuProcess {
        pid: number;
        command: string;
        shr: string;
        memory: string;
    }

    export interface BindDomain {
        domain: string;
        sslID: number;
        ipList: string;
        appInstallID: number;
        websiteID?: number;
    }

    export interface BindDomainReq {
        appInstallID: number;
    }

    export interface BindDomainRes {
        domain: string;
        sslID: number;
        allowIPs: string[];
        websiteID?: number;
        connUrl: string;
        acmeAccountID: number;
    }

    export interface Environment {
        key: string;
        value: string;
    }

    export interface Volume {
        source: string;
        target: string;
    }

    export interface McpServer {
        id: number;
        name: string;
        status: string;
        baseUrl: string;
        ssePath: string;
        command: string;
        port: number;
        message: string;
        createdAt?: string;
        containerName: string;
        environments: Environment[];
        volumes: Volume[];
        dir?: string;
        hostIP: string;
        protocol: string;
        url: string;
        outputTransport: string;
        streamableHttpPath: string;
        type: string;
    }

    export interface McpServerSearch extends ReqPage {
        name: string;
    }

    export interface McpServerDelete {
        id: number;
    }

    export interface McpServerOperate {
        id: number;
        operate: string;
    }

    export interface McpBindDomain {
        domain: string;
        sslID: number;
        ipList: string;
    }

    export interface McpDomainRes {
        domain: string;
        sslID: number;
        acmeAccountID: number;
        allowIPs: string[];
        websiteID?: number;
        connUrl: string;
    }

    export interface McpBindDomainUpdate {
        websiteID: number;
        sslID: number;
        ipList: string;
    }

    export interface ImportMcpServer {
        name: string;
        command: string;
        ssePath: string;
        containerName: string;
        environments: Environment[];
    }

    export interface ExposedPort {
        hostPort: number;
        containerPort: number;
        hostIP: string;
    }

    export interface Environment {
        key: string;
        value: string;
    }
    export interface Volume {
        source: string;
        target: string;
    }

    export interface ExtraHosts {
        hostname: string;
        ip: string;
    }

    export interface TensorRTLLM {
        id?: number;
        name: string;
        containerName: string;
        version: string;
        modelDir: string;
        status?: string;
        message?: string;
        createdAt?: string;
        exposedPorts?: ExposedPort[];
        environments?: Environment[];
        volumes?: Volume[];
        extraHosts?: ExtraHosts[];
    }

    export interface TensorRTLLMDTO extends TensorRTLLM {
        dir?: string;
    }

    export interface TensorRTLLMSearch extends ReqPage {
        name: string;
    }

    export interface TensorRTLLMDelete {
        id: number;
    }

    export interface TensorRTLLMOperate {
        id: number;
        operate: string;
    }
}
