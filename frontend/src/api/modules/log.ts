import http from '@/api';
import { ResPage, ReqPage } from '../interface';
import { Log } from '../interface/log';

export const getOperationLogs = (info: ReqPage) => {
    return http.post<ResPage<Log.OperationLog>>(`/logs/operation`, info);
};

export const getLoginLogs = (info: ReqPage) => {
    return http.post<ResPage<Log.OperationLog>>(`/logs/login`, info);
};

export const cleanLogs = (logtype: string) => {
    return http.post(`/logs/clean/${logtype}`);
};
