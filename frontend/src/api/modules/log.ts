import http from '@/api';
import { ResPage } from '../interface';
import { Log } from '../interface/log';

export const getOperationLogs = (info: Log.SearchOpLog) => {
    return http.post<ResPage<Log.OperationLog>>(`/logs/operation`, info);
};

export const getLoginLogs = (info: Log.SearchLgLog) => {
    return http.post<ResPage<Log.OperationLog>>(`/logs/login`, info);
};

export const getSystemLogs = () => {
    return http.get<string>(`/logs/system`);
};

export const cleanLogs = (param: Log.CleanLog) => {
    return http.post(`/logs/clean`, param);
};
