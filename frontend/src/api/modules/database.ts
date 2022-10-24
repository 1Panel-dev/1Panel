import http from '@/api';
import { ResPage, ReqPage } from '../interface';
import { Database } from '../interface/database';

export const searchMysqlDBs = (params: ReqPage) => {
    return http.post<ResPage<Database.MysqlDBInfo>>(`databases/search`, params);
};

export const addMysqlDB = (params: Database.MysqlDBCreate) => {
    return http.post(`/databases`, params);
};
export const updateMysqlDBInfo = (params: Database.ChangeInfo) => {
    return http.put(`/databases/${params.id}`, params);
};
export const deleteMysqlDB = (params: { ids: number[] }) => {
    return http.post(`/databases/del`, params);
};

export const loadMysqlVariables = () => {
    return http.get<Database.MysqlVariables>(`/databases/conf`);
};
export const loadMysqlStatus = () => {
    return http.get<Database.MysqlStatus>(`/databases/status`);
};
