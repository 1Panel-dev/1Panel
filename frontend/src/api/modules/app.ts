import http from '@/api';
import { ResPage } from '../interface';
import { App } from '../interface/app';

export const SyncApp = () => {
    return http.post<any>('apps/sync', {});
};

export const SearchApp = (req: App.AppReq) => {
    return http.post<App.AppResPage>('apps/search', req);
};

export const GetApp = (key: string) => {
    return http.get<App.AppDTO>('apps/' + key);
};

export const GetAppTags = () => {
    return http.get<App.Tag[]>('apps/tags');
};

export const GetAppDetail = (id: number, version: string) => {
    return http.get<App.AppDetail>(`apps/detail/${id}/${version}`);
};

export const InstallApp = (install: App.AppInstall) => {
    return http.post<any>('apps/install', install);
};

export const ChangePort = (params: App.ChangePort) => {
    return http.post<any>('apps/installed/port/change', params);
};

export const SearchAppInstalled = (search: App.AppInstallSearch) => {
    return http.post<ResPage<App.AppInstalled>>('apps/installed', search);
};

export const GetAppPort = (key: string) => {
    return http.get<number>(`apps/installed/loadport/${key}`);
};

export const GetAppPassword = (key: string) => {
    return http.get<string>(`apps/installed/loadpassword/${key}`);
};

export const CheckAppInstalled = (key: string) => {
    return http.get<App.CheckInstalled>(`apps/installed/check/${key}`);
};

export const AppInstalledDeleteCheck = (appInstallId: number) => {
    return http.get<App.AppInstallResource[]>(`apps/installed/delete/check/${appInstallId}`);
};

export const GetAppInstalled = (search: App.AppInstalledSearch) => {
    return http.post<App.AppInstalled[]>('apps/installed', search);
};

export const InstalledOp = (op: App.AppInstalledOp) => {
    return http.post<any>('apps/installed/op', op);
};

export const SyncInstalledApp = () => {
    return http.post<any>('apps/installed/sync', {});
};

export const GetAppService = (key: string | undefined) => {
    return http.get<App.AppService[]>(`apps/services/${key}`);
};

export const GetAppBackups = (info: App.AppBackupReq) => {
    return http.post<ResPage<App.AppBackup>>('apps/installed/backups', info);
};

export const DelAppBackups = (req: App.AppBackupDelReq) => {
    return http.post<any>('apps/installed/backups/del', req);
};

export const GetAppUpdateVersions = (id: number) => {
    return http.get<any>(`apps/installed/${id}/versions`);
};

export const GetAppDefaultConfig = (key: string) => {
    return http.get<string>(`apps/installed/conf/${key}`);
};

export const GetAppInstallParams = (id: number) => {
    return http.get<App.InstallParams[]>(`apps/installed/params/${id}`);
};
