export namespace Dashboard {
    export interface OsInfo {
        os: string;
        platform: string;
        platformFamily: string;
        kernelArch: string;
        kernelVersion: string;

        diskSize: number;
    }
    export interface QuickJump {
        id: number;
        name: string;
        alias: string;
        title: string;
        detail: string;
        recommend: number;
        isShow: boolean ;
        router: string;
    }
    export interface AppLauncher {
        key: string;
        icon: string;
        limit: number;
        shortDescEn: string;
        shortDescZh: string;
        currentRow: InstallDetail;

        isInstall: boolean;
        isRecommend: boolean;
        detail: Array<InstallDetail>;
    }
    export interface AppLauncherOption {
        key: string;
        isShow: boolean;
    }
    export interface InstallDetail {
        installID: number;
        detailID: string;
        name: string;
        version: string;
        path: string;
        status: string;
        appType: string;
        webUI: string;
        httpPort: string;
        httpsPort: string;
    }
    export interface BaseInfo {
        hostname: string;
        os: string;
        platform: string;
        platformFamily: string;
        platformVersion: string;
        kernelArch: string;
        kernelVersion: string;
        virtualizationSystem: string;
        ipV4Addr: string;
        httpProxy: string;

        cpuCores: number;
        cpuLogicalCores: number;
        cpuModelName: string;

        currentInfo: CurrentInfo;
        quickJump: Array<QuickJump>;
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
        memoryFree: number;
        memoryShard: number;
        memoryCache: number;
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
}
