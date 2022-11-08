import http from '@/api';
import { ReqPage, ResPage } from '../interface';
import { Backup } from '../interface/backup';
import { Database } from '../interface/database';
import { File } from '@/api/interface/file';

export const searchMysqlDBs = (params: Database.Search) => {
    return http.post<ResPage<Database.MysqlDBInfo>>(`/databases/search`, params);
};
export const listDBByVersion = (params: string) => {
    return http.get(`/databases/dbs/${params}`);
};

export const searchUpList = (params: Database.Search) => {
    return http.post<ResPage<Database.FileRecord>>(`/databases/uplist`, params);
};
export const uploadFile = (mysqlName: string, params: FormData) => {
    return http.upload<File.File>(`/databases/uplist/upload/${mysqlName}`, params);
};
export const backup = (params: Database.Backup) => {
    return http.post(`/databases/backup`, params);
};
export const recover = (params: Database.Recover) => {
    return http.post(`/databases/recover`, params);
};
export const searchBackupRecords = (params: Database.SearchBackupRecord) => {
    return http.post<ResPage<Backup.RecordInfo>>(`/databases/backups/search`, params);
};

export const addMysqlDB = (params: Database.MysqlDBCreate) => {
    return http.post(`/databases`, params);
};
export const updateMysqlDBInfo = (params: Database.ChangeInfo) => {
    return http.put(`/databases/${params.id}`, params);
};
export const updateMysqlVariables = (mysqlName: string, params: Array<Database.VariablesUpdate>) => {
    return http.post(`/databases/variables/update/${mysqlName}`, params);
};
export const updateMysqlConfByFile = (params: Database.MysqlConfUpdateByFile) => {
    return http.post(`/databases/conf/update/byfile`, params);
};
export const deleteMysqlDB = (params: { ids: number[] }) => {
    return http.post(`/databases/del`, params);
};

export const loadMysqlBaseInfo = (param: string) => {
    return http.get<Database.BaseInfo>(`/databases/baseinfo/${param}`);
};
export const loadMysqlVariables = (param: string) => {
    return http.get<Database.MysqlVariables>(`/databases/variables/${param}`);
};
export const loadMysqlStatus = (param: string) => {
    return http.get<Database.MysqlStatus>(`/databases/status/${param}`);
};
export const loadVersions = () => {
    return http.get(`/databases/versions`);
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
export const updateRedisPersistenceConf = (params: Database.RedisConfPersistenceUpdate) => {
    return http.post(`/databases/redis/conf/update/persistence`, params);
};
export const updateRedisConf = (params: Database.RedisConfUpdate) => {
    return http.post(`/databases/redis/conf/update`, params);
};
export const updateRedisConfByFile = (params: Database.RedisConfUpdateByFile) => {
    return http.post(`/databases/redis/conf/update/byfile`, params);
};
export const backupRedis = () => {
    return http.post(`/databases/redis/backup`);
};
export const recoverRedis = (param: Database.RedisRecover) => {
    return http.post(`/databases/redis/recover`, param);
};
export const redisBackupRedisRecords = (param: ReqPage) => {
    return http.post<ResPage<Database.FileRecord>>(`/databases/redis/backup/records`, param);
};
export const deleteBackupRedis = (param: Database.RedisBackupDelete) => {
    return http.post(`/databases/redis/backup/del`, param);
};
