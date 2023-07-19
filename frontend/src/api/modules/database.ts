import http from '@/api';
import { deepCopy } from '@/utils/util';
import { Base64 } from 'js-base64';
import { SearchWithPage, ResPage, DescriptionUpdate } from '../interface';
import { Database } from '../interface/database';

export const searchMysqlDBs = (params: SearchWithPage) => {
    return http.post<ResPage<Database.MysqlDBInfo>>(`/databases/search`, params);
};

export const addMysqlDB = (params: Database.MysqlDBCreate) => {
    let reqest = deepCopy(params) as Database.MysqlDBCreate;
    if (reqest.password) {
        reqest.password = Base64.encode(reqest.password);
    }
    return http.post(`/databases`, reqest);
};
export const updateMysqlAccess = (params: Database.ChangeInfo) => {
    return http.post(`/databases/change/access`, params);
};
export const updateMysqlPassword = (params: Database.ChangeInfo) => {
    let reqest = deepCopy(params) as Database.ChangeInfo;
    if (reqest.value) {
        reqest.value = Base64.encode(reqest.value);
    }
    return http.post(`/databases/change/password`, reqest);
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
export const redisPersistenceConf = () => {
    return http.get<Database.RedisPersistenceConf>(`/databases/redis/persistence/conf`);
};
export const changeRedisPassword = (params: Database.ChangeInfo) => {
    let reqest = deepCopy(params) as Database.ChangeInfo;
    if (reqest.value) {
        reqest.value = Base64.encode(reqest.value);
    }
    return http.post(`/databases/redis/password`, reqest);
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

// remote
export const searchRemoteDBs = (params: SearchWithPage) => {
    return http.post<ResPage<Database.RemoteDBInfo>>(`/databases/remote/search`, params);
};
export const addRemoteDB = (params: Database.RemoteDBInfo) => {
    return http.post(`/databases/remote`, params);
};
export const editRemoteDB = (params: Database.RemoteDBInfo) => {
    return http.post(`/databases/remote/update`, params);
};
export const deleteRemoteDB = (id: number) => {
    return http.post(`/databases/remote/del`, { id: id });
};
