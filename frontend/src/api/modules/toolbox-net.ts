import http from '@/api';
import { TimeoutEnum } from '@/enums/http-enum';

export interface NetToolReq {
    tool: 'ping' | 'traceroute' | 'dns' | 'http';
    host: string;
    count?: number;
}

export interface NetToolRes {
    output: string;
}

export const runNetTool = (req: NetToolReq) => {
    return http.post<NetToolRes>('/toolbox/net', req, TimeoutEnum.T_60S);
};
