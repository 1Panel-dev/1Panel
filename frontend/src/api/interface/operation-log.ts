import { DateTimeFormats } from '@intlify/core-base';

export interface ResOperationLog {
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
