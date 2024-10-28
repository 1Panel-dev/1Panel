import http from '@/api';
import { ResPage } from '../interface';
import { Log } from '../interface/log';

export const getOperationLogs = (info: Log.SearchOpLog) => {
    return http.post<ResPage<Log.OperationLog>>(`/core/logs/operation`, info);
};

export const getLoginLogs = (info: Log.SearchLgLog) => {
    return http.post<ResPage<Log.OperationLog>>(`/core/logs/login`, info);
};

export const getSystemFiles = () => {
    return http.get<Array<string>>(`/logs/system/files`);
};
export const getSystemLogs = (name: string) => {
    return http.post<string>(`/logs/system`, { name: name });
};

export const cleanLogs = (param: Log.CleanLog) => {
    return http.post(`/core/logs/clean`, param);
};

export const searchTasks = (req: Log.SearchTaskReq) => {
    return http.post<ResPage<Log.Task>>(`/logs/tasks/search`, req);
};
