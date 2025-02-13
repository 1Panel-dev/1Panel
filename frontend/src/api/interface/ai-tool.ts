import { ReqPage } from '.';

export namespace AITool {
    export interface OllamaModelInfo {
        name: string;
        size: string;
        modified: string;
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
}
