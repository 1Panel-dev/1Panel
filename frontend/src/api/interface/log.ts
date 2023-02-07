import { DateTimeFormats } from '@intlify/core-base';
import { ReqPage } from '.';

export namespace Log {
    export interface OperationLog {
        id: number;
        source: string;
        action: string;
        ip: string;
        path: string;
        method: string;
        userAgent: string;
        body: string;
        resp: string;

        status: number;
        latency: number;
        errorMessage: string;

        detail: string;
        createdAt: DateTimeFormats;
    }
    export interface SearchOpLog extends ReqPage {
        source: string;
        status: string;
        operation: string;
    }
    export interface SearchLgLog extends ReqPage {
        ip: string;
        status: string;
    }
    export interface LoginLogs {
        ip: string;
        address: string;
        agent: string;
        status: string;
        message: string;
        createdAt: DateTimeFormats;
    }
    export interface CleanLog {
        logType: string;
    }
}
