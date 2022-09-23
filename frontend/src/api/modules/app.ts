import http from '@/api';
import { App } from '../interface/app';

export const SyncApp = () => {
    return http.post<any>('apps/sync', {});
};

export const SearchApp = (req: App.AppReq) => {
    return http.post<App.AppResPage>('apps/search', req);
};

export const GetApp = (id: number) => {
    return http.get<App.AppDTO>('apps/' + id);
};

export const GetAppDetail = (id: number, version: string) => {
    return http.get<App.AppDetail>('apps/detail/' + id + '/' + version);
};
