import { DateTimeFormats } from '@intlify/core-base';
import { ReqPage } from '.';

export namespace Log {
    export interface OperationLog {
        id: number;
        source: string;
        user: string;
        node: string;
        ip: string;
        path: string;
        method: string;
        userAgent: string;
        status: string;
        latency: number;
        message: string;
        detailZH: string;
        detailEN: string;
        createdAt: DateTimeFormats;
    }
    export interface SearchOpLog extends ReqPage {
        source: string;
        status: string;
        operation: string;
        node?: string;
    }
    export interface SearchLgLog extends ReqPage {
        info: string;
        status: string;
        startTime?: string | Date;
        endTime?: string | Date;
    }
    export interface LoginLogs {
        ip: string;
        user: string;
        address: string;
        agent: string;
        status: string;
        message: string;
        createdAt: DateTimeFormats;
    }
    export interface CleanLog {
        logType: string;
    }

    export interface SearchTaskReq extends ReqPage {
        type: string;
        status: string;
        taskID?: string;
    }

    export interface TaskLogReadReq {
        page: number;
        pageSize: number;
        latest?: boolean;
        taskID?: string;
        taskType?: string;
        taskOperate?: string;
        resourceID?: number;
    }

    export interface Task {
        id: string;
        name: string;
        type: string;
        logFile: string;
        status: string;
        errorMsg: string;
        operationLogID: number;
        resourceID: number;
        currentStep: string;
        progressCurrent: number;
        progressTotal: number;
        progressPercent: number;
        progressMessage: string;
        endAt: Date;
        createdAt: Date;
    }
}
