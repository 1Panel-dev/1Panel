import http from '@/api';
import { deepCopy } from '@/utils/util';
import { Base64 } from 'js-base64';
import { ResPage, DescriptionUpdate } from '../interface';
import { Database } from '../interface/database';
import { TimeoutEnum } from '@/enums/http-enum';

// common
export const loadDBBaseInfo = (type: string, database: string) => {
    return http.post<Database.BaseInfo>(`/databases/common/info`, { type: type, name: database });
};
export const loadDBFile = (type: string, database: string) => {
    return http.post<string>(`/databases/common/load/file`, { type: type, name: database });
};
export const updateDBFile = (params: Database.DBConfUpdate) => {
    return http.post(`/databases/common/update/conf`, params);
};

// pg
export const addPostgresqlDB = (params: Database.PostgresqlDBCreate) => {
    let request = deepCopy(params) as Database.PostgresqlDBCreate;
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    return http.post(`/databases/pg`, request, TimeoutEnum.T_40S);
};
export const bindPostgresqlUser = (params: Database.PgBind) => {
    return http.post(`/databases/pg/bind`, params, TimeoutEnum.T_40S);
};
export const changePrivileges = (params: Database.PgChangePrivileges) => {
    return http.post(`/databases/pg/privileges`, params, TimeoutEnum.T_40S);
};
export const searchPostgresqlDBs = (params: Database.SearchDBWithPage) => {
    return http.post<ResPage<Database.PostgresqlDBInfo>>(`/databases/pg/search`, params);
};
export const updatePostgresqlDescription = (params: DescriptionUpdate) => {
    return http.post(`/databases/pg/description`, params);
};
export const loadPgFromRemote = (database: string) => {
    return http.post(`/databases/pg/${database}/load`);
};
export const deleteCheckPostgresqlDB = (params: Database.PostgresqlDBDeleteCheck) => {
    return http.post<Array<string>>(`/databases/pg/del/check`, params, TimeoutEnum.T_40S);
};
export const updatePostgresqlPassword = (params: Database.ChangeInfo) => {
    let request = deepCopy(params) as Database.ChangeInfo;
    if (request.value) {
        request.value = Base64.encode(request.value);
    }
    return http.post(`/databases/pg/password`, request, TimeoutEnum.T_40S);
};
export const deletePostgresqlDB = (params: Database.PostgresqlDBDelete) => {
    return http.post(`/databases/pg/del`, params, TimeoutEnum.T_40S);
};

// mysql
export const searchMysqlDBs = (params: Database.SearchDBWithPage) => {
    return http.post<ResPage<Database.MysqlDBInfo>>(`/databases/search`, params);
};
export const addMysqlDB = (params: Database.MysqlDBCreate) => {
    let request = deepCopy(params) as Database.MysqlDBCreate;
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    return http.post(`/databases`, request);
};
export const bindUser = (params: Database.BindUser) => {
    let request = deepCopy(params) as Database.BindUser;
    if (request.password) {
        request.password = Base64.encode(request.password);
    }
    return http.post(`/databases/bind`, request);
};
export const loadDBFromRemote = (params: Database.MysqlLoadDB) => {
    return http.post(`/databases/load`, params);
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
export const deleteCheckMysqlDB = (params: Database.MysqlDBDeleteCheck) => {
    return http.post<Array<string>>(`/databases/del/check`, params);
};
export const deleteMysqlDB = (params: Database.MysqlDBDelete) => {
    return http.post(`/databases/del`, params);
};

export const loadMysqlVariables = (type: string, database: string) => {
    return http.post<Database.MysqlVariables>(`/databases/variables`, { type: type, name: database });
};
export const loadMysqlStatus = (type: string, database: string) => {
    return http.post<Database.MysqlStatus>(`/databases/status`, { type: type, name: database });
};
export const loadRemoteAccess = (type: string, database: string) => {
    return http.post<boolean>(`/databases/remote`, { type: type, name: database });
};

// redis
export const loadRedisStatus = (database: string) => {
    return http.post<Database.RedisStatus>(`/databases/redis/status`, { name: database });
};
export const loadRedisConf = (database: string) => {
    return http.post<Database.RedisConf>(`/databases/redis/conf`, { name: database });
};
export const redisPersistenceConf = (database: string) => {
    return http.post<Database.RedisPersistenceConf>(`/databases/redis/persistence/conf`, { name: database });
};
export const checkRedisCli = () => {
    return http.get<boolean>(`/databases/redis/check`);
};
export const installRedisCli = () => {
    return http.post(`/databases/redis/install/cli`, {}, TimeoutEnum.T_5M);
};
export const changeRedisPassword = (database: string, password: string) => {
    if (password) {
        password = Base64.encode(password);
    }
    return http.post(`/databases/redis/password`, { database: database, value: password });
};
export const updateRedisPersistenceConf = (params: Database.RedisConfPersistenceUpdate) => {
    return http.post(`/databases/redis/persistence/update`, params);
};
export const updateRedisConf = (params: Database.RedisConfUpdate) => {
    return http.post(`/databases/redis/conf/update`, params);
};
export const updateRedisConfByFile = (params: Database.DBConfUpdate) => {
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
export const listDbItems = (type: string) => {
    return http.get<Array<Database.DbItem>>(`/databases/db/item/${type}`);
};
export const checkDatabase = (params: Database.DatabaseCreate) => {
    let request = deepCopy(params) as Database.DatabaseCreate;
    if (request.ssl) {
        request.clientKey = Base64.encode(request.clientKey);
        request.clientCert = Base64.encode(request.clientCert);
        request.rootCert = Base64.encode(request.rootCert);
    }

    return http.post<boolean>(`/databases/db/check`, request, TimeoutEnum.T_40S);
};
export const addDatabase = (params: Database.DatabaseCreate) => {
    let request = deepCopy(params) as Database.DatabaseCreate;
    if (request.ssl) {
        request.clientKey = Base64.encode(request.clientKey);
        request.clientCert = Base64.encode(request.clientCert);
        request.rootCert = Base64.encode(request.rootCert);
    }

    return http.post(`/databases/db`, request, TimeoutEnum.T_40S);
};
export const editDatabase = (params: Database.DatabaseUpdate) => {
    let request = deepCopy(params) as Database.DatabaseCreate;
    if (request.ssl) {
        request.clientKey = Base64.encode(request.clientKey);
        request.clientCert = Base64.encode(request.clientCert);
        request.rootCert = Base64.encode(request.rootCert);
    }

    return http.post(`/databases/db/update`, request, TimeoutEnum.T_40S);
};
export const deleteCheckDatabase = (id: number) => {
    return http.post<Array<string>>(`/databases/db/del/check`, { id: id });
};
export const deleteDatabase = (params: Database.DatabaseDelete) => {
    return http.post(`/databases/db/del`, params);
};
