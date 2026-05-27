import http from '@/api';
import { ResPage } from '../interface';
import { Log } from '../interface/log';
import { TimeoutEnum } from '@/enums/http-enum';

export const getOperationLogs = (info: Log.SearchOpLog) => {
    return http.post<ResPage<Log.OperationLog>>(`/core/logs/operation`, info);
};

export const getLoginLogs = (info: Log.SearchLgLog, currentNode?: string) => {
    return http.post<ResPage<Log.LoginLogs>>(
        `/core/logs/login`,
        info,
        undefined,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};

export const getSystemFiles = (node?: string) => {
    const params = node ? `?operateNode=${node}` : '';
    return http.get<Array<string>>(`/logs/system/files${params}`);
};

export const cleanLogs = (param: Log.CleanLog) => {
    return http.post(`/core/logs/clean`, param);
};

export const searchTasks = (req: Log.SearchTaskReq, node?: string) => {
    const params = node ? `?operateNode=${node}` : '';
    return http.post<ResPage<Log.Task>>(`/logs/tasks/search${params}`, req);
};

export const readTaskLogByLine = (req: Log.TaskLogReadReq, node?: string) => {
    const params = node ? `?operateNode=${node}` : '';
    return http.post<any>(`/logs/tasks/read${params}`, req, TimeoutEnum.T_40S);
};

export const countExecutingTask = () => {
    return http.get<number>(`/logs/tasks/executing/count`);
};
