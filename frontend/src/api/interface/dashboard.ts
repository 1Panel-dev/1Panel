export namespace Dashboard {
    export interface BaseInfo {
        haloID: number;
        dateeaseID: number;
        jumpserverID: number;
        metersphereID: number;
        kubeoperatorID: number;
        kubepiID: number;

        websiteNumber: number;
        databaseNumber: number;
        cronjobNumber: number;
        appInstalldNumber: number;

        hostname: string;
        os: string;
        platform: string;
        platformFamily: string;
        platformVersion: string;
        kernelArch: string;
        kernelVersion: string;
        virtualizationSystem: string;
        uptime: string;
        timeSinceUptime: string;

        cpuCores: number;
        cpuLogicalCores: number;
        cpuModelName: string;

        currentInfo: CurrentInfo;
    }
    export interface CurrentInfo {
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
        MemoryUsedPercent: number;

        ioReadBytes: number;
        ioWriteBytes: number;
        ioTime: number;
        ioCount: number;

        total: number;
        free: number;
        used: number;
        usedPercent: number;

        inodesTotal: number;
        inodesUsed: number;
        inodesFree: number;
        inodesUsedPercent: number;

        netBytesSent: number;
        netBytesRecv: number;

        shotTime: Date;
    }
}
