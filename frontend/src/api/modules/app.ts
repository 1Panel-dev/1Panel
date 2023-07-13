import http from '@/api';
import { ResPage } from '../interface';
import { App } from '../interface/app';

export const SyncApp = () => {
    return http.post<any>('apps/sync', {});
};

export const GetAppListUpdate = () => {
    return http.get<App.AppUpdateRes>('apps/checkupdate');
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

export const GetAppDetail = (appID: number, version: string, type: string) => {
    return http.get<App.AppDetail>(`apps/detail/${appID}/${version}/${type}`);
};

export const GetAppDetailByID = (id: number) => {
    return http.get<App.AppDetail>(`apps/details/${id}`);
};

export const InstallApp = (install: App.AppInstall) => {
    return http.post<any>('apps/install', install);
};

export const ChangePort = (params: App.ChangePort) => {
    return http.post<any>('apps/installed/port/change', params);
};

export const SearchAppInstalled = (search: App.AppInstallSearch) => {
    return http.post<ResPage<App.AppInstalled>>('apps/installed/search', search);
};

export const GetAppPort = (key: string) => {
    return http.get<number>(`apps/installed/loadport/${key}`);
};

export const GetAppConnInfo = (key: string) => {
    return http.get<App.DatabaseConnInfo>(`apps/installed/conninfo/${key}`);
};

export const CheckAppInstalled = (key: string) => {
    return http.get<App.CheckInstalled>(`apps/installed/check/${key}`);
};

export const AppInstalledDeleteCheck = (appInstallId: number) => {
    return http.get<App.AppInstallResource[]>(`apps/installed/delete/check/${appInstallId}`);
};

export const GetAppInstalled = (search: App.AppInstalledSearch) => {
    return http.post<App.AppInstalled[]>('apps/installed/search', search);
};

export const InstalledOp = (op: App.AppInstalledOp) => {
    return http.post<any>('apps/installed/op', op, 40000);
};

export const SyncInstalledApp = () => {
    return http.post<any>('apps/installed/sync', {});
};

export const GetAppService = (key: string | undefined) => {
    return http.get<App.AppService[]>(`apps/services/${key}`);
};

export const GetAppUpdateVersions = (id: number) => {
    return http.get<any>(`apps/installed/${id}/versions`);
};

export const GetAppDefaultConfig = (key: string) => {
    return http.get<string>(`apps/installed/conf/${key}`);
};

export const GetAppInstallParams = (id: number) => {
    return http.get<App.AppConfig>(`apps/installed/params/${id}`);
};

export const UpdateAppInstallParams = (req: any) => {
    return http.post<any>(`apps/installed/params/update`, req);
};

export const IgnoreUpgrade = (req: any) => {
    return http.post<any>(`apps/installed/ignore`, req);
};

export const GetIgnoredApp = () => {
    return http.get<App.IgnoredApp>(`apps/ignored/detail`);
};
