import http from '@/api';
import { Dashboard } from '../interface/dashboard';

export const loadOsInfo = () => {
    return http.get<Dashboard.OsInfo>(`/dashboard/base/os`);
};

export const loadAppLauncher = () => {
    return http.get<Array<Dashboard.AppLauncher>>(`/dashboard/app/launcher`);
};
export const loadAppLauncherOption = (filter: string) => {
    return http.post<Array<Dashboard.AppLauncherOption>>(`/dashboard/app/launcher/option`, { filter: filter });
};
export const changeLauncherStatus = (key: string, val: string) => {
    return http.post(`/core/launcher/change/show`, { key: key, value: val });
};
export const resetLauncherStatus = () => {
    return http.post(`/core/launcher/reset`);
};

export const loadBaseInfo = (ioOption: string, netOption: string) => {
    return http.get<Dashboard.BaseInfo>(`/dashboard/base/${ioOption}/${netOption}`);
};

export const loadCurrentInfo = (ioOption: string, netOption: string) => {
    return http.get<Dashboard.CurrentInfo>(`/dashboard/current/${ioOption}/${netOption}`);
};

export const systemRestart = (operation: string) => {
    return http.post(`/dashboard/system/restart/${operation}`);
};
