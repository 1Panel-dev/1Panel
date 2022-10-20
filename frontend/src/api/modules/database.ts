import http from '@/api';
import { ResPage, ReqPage } from '../interface';
import { Database } from '../interface/database';

export const searchMysqlDBs = (params: ReqPage) => {
    return http.post<ResPage<Database.MysqlDBInfo>>(`databases/search`, params);
};

export const addMysqlDB = (params: Database.MysqlDBCreate) => {
    return http.post(`/databases`, params);
};

export const deleteMysqlDB = (params: { ids: number[] }) => {
    return http.post(`/databases/del`, params);
};
