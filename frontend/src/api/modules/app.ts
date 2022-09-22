// export const GetAppList = ()
import http from '@/api';
import { App } from '../interface/app';

export const SyncApp = () => {
    return http.post<any>('apps/sync', {});
};

export const SearchApp = (req: App.AppReq) => {
    return http.post<App.AppResPage>('apps/search', req);
};
