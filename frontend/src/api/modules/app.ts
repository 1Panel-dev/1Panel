import http from '@/api';
import { ResPage } from '../interface';
import { App } from '../interface/app';
import { TimeoutEnum } from '@/enums/http-enum';

export const syncApp = (req: App.AppStoreSync) => {
    return http.post('apps/sync/remote', req);
};

export const syncLocalApp = (req: App.AppStoreSync) => {
    return http.post('apps/sync/local', req);
};

export const searchApp = (req: App.AppReq) => {
    return http.post<App.AppResPage>('apps/search', req);
};

export const getAppByKey = (key: string) => {
    return http.get<App.AppDTO>('apps/' + key);
};

export const getAppByKeyWithNode = (key: string, node: string) => {
    return http.get<App.AppDTO>('apps/' + key + `?operateNode=${node}`);
};

export const getAppTags = () => {
    return http.get<App.Tag[]>('apps/tags');
};

export const getAppDetail = (appID: number, version: string, type: string) => {
    return http.get<App.AppDetail>(`apps/detail/${appID}/${version}/${type}`);
};

export const getAppDetailByID = (id: number) => {
    return http.get<App.AppDetail>(`apps/details/${id}`);
};

export const installApp = (install: App.AppInstall) => {
    return http.post<any>('apps/install', install);
};

export const changePort = (params: App.ChangePort) => {
    return http.post<any>('apps/installed/port/change', params);
};

export const searchAppInstalled = (search: App.AppInstallSearch) => {
    return http.post<ResPage<App.AppInstallDto>>('apps/installed/search', search);
};

export const listAppInstalled = () => {
    return http.get<Array<App.AppInstalledInfo>>('apps/installed/list');
};

export const getAppPort = (type: string, name: string) => {
    return http.post<number>(`apps/installed/loadport`, { type: type, name: name });
};

export const getAppConnInfo = (type: string, name: string) => {
    return http.post<App.DatabaseConnInfo>(`apps/installed/conninfo`, { type: type, name: name });
};

export const checkAppInstalled = (key: string, name: string) => {
    return http.post<App.CheckInstalled>(`apps/installed/check`, { key: key, name: name });
};

export const appInstalledDeleteCheck = (appInstallId: number, node?: string) => {
    const params = node ? `?operateNode=${node}` : '';
    return http.get<App.AppInstallResource[]>(`apps/installed/delete/check/${appInstallId}${params}`);
};

export const getAppInstalled = (search: App.AppInstalledSearch) => {
    return http.post<ResPage<App.AppInstalled>>('apps/installed/search', search);
};

export const getAppInstalledByID = (installID: number, node: string) => {
    return http.get<App.AppInstalledInfo>(`apps/installed/info/${installID}?operateNode=${node}`);
};

export const installedOp = (op: App.AppInstalledOp) => {
    return http.post<any>('apps/installed/op', op, TimeoutEnum.T_40S);
};

export const syncInstalledApp = () => {
    return http.post<any>('apps/installed/sync', {});
};

export const getAppService = (key: string | undefined) => {
    return http.get<App.AppService[]>(`apps/services/${key}`);
};

export const getAppUpdateVersions = (req: App.AppUpdateVersionReq) => {
    return http.post<any>(`apps/installed/update/versions`, req);
};

export const getAppDefaultConfig = (key: string, name: string) => {
    return http.post<string>(`apps/installed/conf`, { type: key, name: name });
};

export const getAppInstallParams = (id: number) => {
    return http.get<App.AppConfig>(`apps/installed/params/${id}`);
};

export const updateAppInstallParams = (req: any) => {
    return http.post<any>(`apps/installed/params/update`, req);
};

export const ignoreUpgrade = (req: App.AppIgnoreReq) => {
    return http.post<any>(`apps/installed/ignore`, req);
};

export const getIgnoredApp = () => {
    return http.get<App.IgnoredApp>(`apps/ignored/detail`);
};

export const cancelAppIgnore = (req: App.CancelAppIgnore) => {
    return http.post(`apps/ignored/cancel`, req);
};

export const updateInstallConfig = (req: App.AppConfigUpdate) => {
    return http.post(`apps/installed/config/update`, req);
};

export const syncCutomAppStore = (req: App.AppStoreSync) => {
    return http.post(`/custom/app/sync`, req);
};

export const getCurrentNodeCustomAppConfig = () => {
    return http.get<App.CustomAppStoreConfig>(`/custom/app/config`);
};
