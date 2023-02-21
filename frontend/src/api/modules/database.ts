import http from '@/api';
import { SearchWithPage, ReqPage, ResPage, DescriptionUpdate } from '../interface';
import { Database } from '../interface/database';

export const searchMysqlDBs = (params: SearchWithPage) => {
    return http.post<ResPage<Database.MysqlDBInfo>>(`/databases/search`, params);
};

export const addMysqlDB = (params: Database.MysqlDBCreate) => {
    return http.post(`/databases`, params);
};
export const updateMysqlAccess = (params: Database.ChangeInfo) => {
    return http.post(`/databases/change/access`, params);
};
export const updateMysqlPassword = (params: Database.ChangeInfo) => {
    return http.post(`/databases/change/password`, params);
};
export const updateMysqlDescription = (params: DescriptionUpdate) => {
    return http.post(`/databases/description/update`, params);
};
export const updateMysqlVariables = (params: Array<Database.VariablesUpdate>) => {
    return http.post(`/databases/variables/update`, params);
};
export const updateMysqlConfByFile = (params: Database.MysqlConfUpdateByFile) => {
    return http.post(`/databases/conffile/update`, params);
};
export const deleteCheckMysqlDB = (id: number) => {
    return http.post<Array<string>>(`/databases/del/check`, { id: id });
};
export const deleteMysqlDB = (params: Database.MysqlDBDelete) => {
    return http.post(`/databases/del`, params);
};

export const loadMysqlBaseInfo = () => {
    return http.get<Database.BaseInfo>(`/databases/baseinfo`);
};
export const loadMysqlVariables = () => {
    return http.get<Database.MysqlVariables>(`/databases/variables`);
};
export const loadMysqlStatus = () => {
    return http.get<Database.MysqlStatus>(`/databases/status`);
};
export const loadRemoteAccess = () => {
    return http.get<boolean>(`/databases/remote`);
};
export const loadDBNames = () => {
    return http.get<Array<string>>(`/databases/options`);
};

// redis
export const loadRedisStatus = () => {
    return http.get<Database.RedisStatus>(`/databases/redis/status`);
};
export const loadRedisConf = () => {
    return http.get<Database.RedisConf>(`/databases/redis/conf`);
};
export const RedisPersistenceConf = () => {
    return http.get<Database.RedisPersistenceConf>(`/databases/redis/persistence/conf`);
};
export const changeRedisPassword = (params: Database.ChangeInfo) => {
    return http.post(`/databases/redis/password`, params);
};
export const updateRedisPersistenceConf = (params: Database.RedisConfPersistenceUpdate) => {
    return http.post(`/databases/redis/persistence/update`, params);
};
export const updateRedisConf = (params: Database.RedisConfUpdate) => {
    return http.post(`/databases/redis/conf/update`, params);
};
export const updateRedisConfByFile = (params: Database.RedisConfUpdateByFile) => {
    return http.post(`/databases/redis/conffile/update`, params);
};
export const recoverRedis = (param: Database.RedisRecover) => {
    return http.post(`/databases/redis/recover`, param);
};
export const redisBackupRedisRecords = (param: ReqPage) => {
    return http.post<ResPage<Database.FileRecord>>(`/databases/redis/backup/search`, param);
};
