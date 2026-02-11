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

    export interface AgentCreateReq {
        name: string;
        appVersion: string;
        webUIPort: number;
        bridgePort: number;
        provider: string;
        model: string;
        accountId: number;
        apiKey: string;
        baseURL: string;
        token: string;
        taskID: string;
        advanced: boolean;
        containerName: string;
        allowPort: boolean;
        specifyIP: string;
        restartPolicy: string;
        cpuQuota: number;
        memoryLimit: number;
        memoryUnit: string;
        pullImage: boolean;
        editCompose: boolean;
        dockerCompose: string;
    }

    export interface AgentItem {
        id: number;
        name: string;
        provider: string;
        providerName: string;
        model: string;
        baseUrl: string;
        apiKey: string;
        token: string;
        status: string;
        message: string;
        appInstallId: number;
        accountId: number;
        appVersion: string;
        containerName: string;
        webUIPort: number;
        bridgePort: number;
        path: string;
        configPath: string;
        upgradable: boolean;
        createdAt: string;
    }

    export interface AgentDeleteReq {
        id: number;
        taskID: string;
        forceDelete: boolean;
    }

    export interface AgentTokenResetReq {
        id: number;
    }

    export interface AgentModelConfigUpdateReq {
        agentId: number;
        accountId: number;
        model: string;
    }

    export interface ProviderModelInfo {
        id: string;
        name: string;
    }

    export interface ProviderInfo {
        provider: string;
        displayName: string;
        baseUrl: string;
        models: ProviderModelInfo[];
    }

    export interface AgentAccountCreateReq {
        provider: string;
        name: string;
        apiKey: string;
        baseURL: string;
        remark: string;
    }

    export interface AgentAccountUpdateReq {
        id: number;
        name: string;
        apiKey: string;
        baseURL: string;
        remark: string;
        syncAgents: boolean;
    }

    export interface AgentAccountSearch {
        page: number;
        pageSize: number;
        provider: string;
        name: string;
    }

    export interface AgentAccountItem {
        id: number;
        provider: string;
        providerName: string;
        name: string;
        apiKey: string;
        baseUrl: string;
        verified: boolean;
        remark: string;
        createdAt: string;
    }

    export interface AgentAccountVerifyReq {
        provider: string;
        apiKey: string;
        baseURL: string;
    }

    export interface AgentAccountDeleteReq {
        id: number;
    }

    export interface AgentFeishuConfigReq {
        agentId: number;
    }

    export interface AgentFeishuConfig {
        enabled: boolean;
        dmPolicy: string;
        botName: string;
        appId: string;
        appSecret: string;
    }

    export interface AgentFeishuConfigUpdateReq {
        agentId: number;
        enabled: boolean;
        dmPolicy: string;
        botName: string;
        appId: string;
        appSecret: string;
    }

    export interface AgentFeishuPairingApproveReq {
        agentId: number;
        pairingCode: string;
    }
}
