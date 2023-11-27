export namespace Toolbox {
    export interface DeviceBaseInfo {
        dns: Array<string>;
        hosts: Array<HostHelper>;
        hostname: string;
        ntp: string;
        user: string;
        timeZone: string;
        localTime: string;

        swapMemoryTotal: number;
        swapMemoryAvailable: number;
        swapMemoryUsed: number;

        swapDetails: Array<SwapHelper>;
    }
    export interface SwapHelper {
        path: string;
        size: number;
        used: string;

        isNew: boolean;
    }
    export interface HostHelper {
        ip: string;
        host: string;
    }
    export interface TimeZoneOptions {
        from: string;
        zones: Array<string>;
    }

    export interface Fail2banBaseInfo {
        isEnable: boolean;
        isActive: boolean;
        isExist: boolean;
        version: string;

        port: number;
        maxRetry: number;
        banTime: string;
        findTime: string;
        banAction: string;
        logPath: string;
    }

    export interface Fail2banSearch {
        status: string;
    }

    export interface Fail2banUpdate {
        key: string;
        value: string;
    }

    export interface Fail2banSet {
        ips: Array<string>;
        operate: string;
    }
}
