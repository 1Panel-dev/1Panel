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
        isShow: boolean;
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
        prettyDistro: string;
        kernelArch: string;
        kernelVersion: string;
        virtualizationSystem: string;
        ipV4Addr: string;
        httpProxy: string;

        cpuCores: number;
        cpuLogicalCores: number;
        cpuModelName: string;
        cpuMhz: number;

        currentInfo: CurrentInfo;
        quickJump: Array<QuickJump>;
    }
    export interface CurrentInfo {
        uptime: number;
        timeSinceUptime: string;
        runningTime: RunningTime;
        procs: number;

        load1: number;
        load5: number;
        load15: number;
        loadUsagePercent: number;

        cpuPercent: Array<number>;
        cpuUsedPercent: number;
        cpuUsed: number;
        cpuTotal: number;
        cpuDetailedPercent: Array<number>; // [user, system, nice, idle, iowait, irq, softirq, steal]

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
        npuData: Array<NPUInfo>;
        xpuData: Array<XPUInfo>;

        topCPUItems?: Array<Process>;
        topMemItems?: Array<Process>;

        netBytesSent: number;
        netBytesRecv: number;

        shotTime: Date;
    }
    export interface RunningTime {
        days: number;
        hours: number;
        minutes: number;
        seconds: number;
    }
    export interface Process {
        name: string;
        pid: number;
        percent: number;
        memory: number;
        cmd: string;
        user: string;
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
        type: string;
        index: number;
        npuIndex: number;
        chipIndex: number;
        productName: string;
        busID: string;
        gpuUtil: string;
        temperature: string;
        performanceState: string;
        powerUsage: string;
        powerDraw: string;
        maxPowerLimit: string;
        memoryUsage: string;
        memUsed: string;
        memTotal: string;
        fanSpeed: string;
    }

    export interface NPUInfo {
        type: 'ascend';
        index: number;
        npuIndex: number;
        chipIndex: number;
        productName: string;
        busID: string;
        health: string;
        temperature: string;
        powerDraw: string;
        aiCore: string;
        memUsed: string;
        memTotal: string;
        memoryUsed: string;
        memoryTotal: string;
        hbmUsed: string;
        hbmTotal: string;
        hugepagesUsed: string;
        hugepagesTotal: string;
    }

    export interface XPUInfo {
        deviceID: number;
        deviceName: string;
        pciBdfAddress: string;
        memory: string;
        temperature: string;
        gpuUtil: string;
        memoryUsed: string;
        power: string;
        memoryUtil: string;
    }
}
