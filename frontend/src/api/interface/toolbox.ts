import { ReqPage } from '.';
import { Cronjob } from './cronjob';

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
        maxSize: number;

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

    export interface CleanData {
        systemClean: Array<CleanTree>;
        uploadClean: Array<CleanTree>;
        downloadClean: Array<CleanTree>;
        systemLogClean: Array<CleanTree>;
        containerClean: Array<CleanTree>;
    }
    export interface CleanTree {
        id: string;
        label: string;
        children: Array<CleanTree>;
        type: string;
        name: string;
        size: number;
        isCheck: boolean;
        isRecommend: boolean;
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

    export interface FtpBaseInfo {
        isActive: boolean;
        isExist: boolean;
    }
    export interface FtpInfo {
        id: number;
        user: string;
        password: string;
        status: string;
        path: string;
        description: string;
    }
    export interface FtpCreate {
        user: string;
        password: string;
        path: string;
        description: string;
    }
    export interface FtpUpdate {
        id: number;
        password: string;
        status: string;
        path: string;
        description: string;
    }
    export interface FtpSearchLog extends ReqPage {
        user: string;
        operation: string;
    }
    export interface FtpLog {
        ip: string;
        user: string;
        time: string;
        operation: string;
        status: string;
        size: string;
    }

    export interface ClamBaseInfo {
        version: string;
        isActive: boolean;
        isExist: boolean;

        freshVersion: string;
        freshIsExist: boolean;
        freshIsActive: boolean;
    }
    export interface ClamInfo {
        id: number;
        name: string;
        status: string;
        path: string;
        infectedStrategy: string;
        infectedDir: string;
        lastHandleDate: string;
        hasSpec: boolean;
        spec: string;
        specObj: Cronjob.SpecObj;
        description: string;
        hasAlert: boolean;
        alertCount: number;
        alertTitle: string;
    }
    export interface ClamCreate {
        name: string;
        path: string;
        infectedStrategy: string;
        infectedDir: string;
        spec: string;
        specObj: Cronjob.SpecObj;
        description: string;
    }
    export interface ClamUpdate {
        id: number;
        name: string;
        path: string;
        infectedStrategy: string;
        infectedDir: string;
        spec: string;
        specObj: Cronjob.SpecObj;
        description: string;
    }
    export interface ClamSearchLog extends ReqPage {
        clamID: number;
        startTime: Date;
        endTime: Date;
    }
    export interface ClamRecordReq {
        tail: string;
        clamName: string;
        recordName: string;
    }
    export interface ClamLog {
        name: string;
        scanDate: string;
        scanTime: string;
        totalError: string;
        infectedFiles: string;
        status: string;
    }
}
