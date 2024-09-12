export namespace Dashboard {
    export interface OsInfo {
        os: string;
        platform: string;
        platformFamily: string;
        kernelArch: string;
        kernelVersion: string;

        diskSize: number;
    }
    export interface BaseInfo {
        websiteNumber: number;
        databaseNumber: number;
        cronjobNumber: number;
        appInstalledNumber: number;

        hostname: string;
        os: string;
        platform: string;
        platformFamily: string;
        platformVersion: string;
        kernelArch: string;
        kernelVersion: string;
        virtualizationSystem: string;
        ipv4Addr: string;
        systemProxy: string;

        cpuCores: number;
        cpuLogicalCores: number;
        cpuModelName: string;

        currentInfo: CurrentInfo;
    }
    export interface CurrentInfo {
        uptime: number;
        timeSinceUptime: string;
        procs: number;

        load1: number;
        load5: number;
        load15: number;
        loadUsagePercent: number;

        cpuPercent: Array<number>;
        cpuUsedPercent: number;
        cpuUsed: number;
        cpuTotal: number;

        memoryTotal: number;
        memoryAvailable: number;
        memoryUsed: number;
        memoryUsedPercent: number;
        swapMemoryTotal: number;
        swapMemoryAvailable: number;
        swapMemoryUsed: number;
        swapMemoryUsedPercent: number;

        ioReadBytes: number;
        ioWriteBytes: number;
        ioCount: number;
        ioReadTime: number;
        ioWriteTime: number;

        diskData: Array<DiskInfo>;

        gpuData: Array<GPUInfo>;
        xpuData: Array<XPUInfo>;

        netBytesSent: number;
        netBytesRecv: number;

        shotTime: Date;
    }
    export interface DiskInfo {
        path: string;
        type: string;
        device: string;
        total: number;
        free: number;
        used: number;
        usedPercent: number;

        inodesTotal: number;
        inodesUsed: number;
        inodesFree: number;
        inodesUsedPercent: number;
    }
    export interface GPUInfo {
        index: number;
        productName: string;
        gpuUtil: string;
        temperature: string;
        performanceState: string;
        powerUsage: string;
        memoryUsage: string;
        fanSpeed: string;
    }

    export interface XPUInfo {
        deviceID: number;
        deviceName: string;
        memory: string;
        temperature: string;
        memoryUsed: string;
        power: string;
        memoryUtil: string;
    }

    export interface DashboardReq {
        scope: string;
        ioOption: string;
        netOption: string;
    }
}
