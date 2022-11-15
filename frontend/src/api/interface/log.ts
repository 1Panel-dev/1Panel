import { DateTimeFormats } from '@intlify/core-base';

export namespace Log {
    export interface OperationLog {
        id: number;
        group: string;
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
    export interface LoginLogs {
        ip: string;
        address: string;
        agent: string;
        status: string;
        message: string;
        createdAt: DateTimeFormats;
    }
}
