import http from '@/api';
import { deepCopy } from '@/utils/util';
import { Base64 } from 'js-base64';
import { ResPage, DescriptionUpdate } from '../interface';
import { Database } from '../interface/database';
import { TimeoutEnum } from '@/enums/http-enum';

export const searchMysqlDBs = (params: Database.SearchDBWithPage) => {
    return http.post<ResPage<Database.MysqlDBInfo>>(`/databases/search`, params);
};
export const loadDatabaseFile = (type: string, database: string) => {
    return http.post<string>(`/databases/load/file`, { type: type, name: database });
};

export const addMysqlDB = (params: Database.MysqlDBCreate) => {
    let request = deepCopy(params) as Database.MysqlDBCreate;
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    return http.post(`/databases`, request);
};
export const updateMysqlAccess = (params: Database.ChangeInfo) => {
    return http.post(`/databases/change/access`, params);
};
export const updateMysqlPassword = (params: Database.ChangeInfo) => {
    let request = deepCopy(params) as Database.ChangeInfo;
    if (request.value) {
        request.value = Base64.encode(request.value);
    }
    return http.post(`/databases/change/password`, request);
};
export const updateMysqlDescription = (params: DescriptionUpdate) => {
    return http.post(`/databases/description/update`, params);
};
export const updateMysqlVariables = (params: Database.VariablesUpdate) => {
    return http.post(`/databases/variables/update`, params);
};
export const updateMysqlConfByFile = (params: Database.MysqlConfUpdateByFile) => {
    return http.post(`/databases/conffile/update`, params);
};
export const deleteCheckMysqlDB = (params: Database.MysqlDBDeleteCheck) => {
    return http.post<Array<string>>(`/databases/del/check`, params);
};
export const deleteMysqlDB = (params: Database.MysqlDBDelete) => {
    return http.post(`/databases/del`, params);
};

export const loadDBFromRemote = (id: number) => {
    return http.post(`/databases/load`, { id: id });
};
export const loadMysqlBaseInfo = (id: number) => {
    return http.post<Database.BaseInfo>(`/databases/baseinfo`, { id: id });
};
export const loadMysqlVariables = (id: number) => {
    return http.post<Database.MysqlVariables>(`/databases/variables`, { id: id });
};
export const loadMysqlStatus = (id: number) => {
    return http.post<Database.MysqlStatus>(`/databases/status`, { id: id });
};
export const loadRemoteAccess = (id: number) => {
    return http.post<boolean>(`/databases/remote`, { id: id });
};
export const loadDBOptions = () => {
    return http.get<Array<Database.MysqlOption>>(`/databases/options`);
};

// redis
export const loadRedisStatus = () => {
    return http.get<Database.RedisStatus>(`/databases/redis/status`);
};
export const loadRedisConf = () => {
    return http.get<Database.RedisConf>(`/databases/redis/conf`);
};
export const redisPersistenceConf = () => {
    return http.get<Database.RedisPersistenceConf>(`/databases/redis/persistence/conf`);
};
export const changeRedisPassword = (value: string) => {
    if (value) {
        value = Base64.encode(value);
    }
    return http.post(`/databases/redis/password`, { value: value });
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

// database
export const getDatabase = (name: string) => {
    return http.get<Database.DatabaseInfo>(`/databases/db/${name}`);
};
export const searchDatabases = (params: Database.SearchDatabasePage) => {
    return http.post<ResPage<Database.DatabaseInfo>>(`/databases/db/search`, params);
};
export const listDatabases = (type: string) => {
    return http.get<Array<Database.DatabaseOption>>(`/databases/db/list/${type}`);
};
export const checkDatabase = (params: Database.DatabaseCreate) => {
    return http.post<boolean>(`/databases/db/check`, params, TimeoutEnum.T_40S);
};
export const addDatabase = (params: Database.DatabaseCreate) => {
    return http.post(`/databases/db`, params, TimeoutEnum.T_40S);
};
export const editDatabase = (params: Database.DatabaseUpdate) => {
    return http.post(`/databases/db/update`, params, TimeoutEnum.T_40S);
};
export const deleteDatabase = (id: number) => {
    return http.post(`/databases/db/del`, { id: id });
};
