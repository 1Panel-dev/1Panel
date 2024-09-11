import http from '@/api';
import { Dashboard } from '../interface/dashboard';

export const loadOsInfo = () => {
    return http.get<Dashboard.OsInfo>(`/dashboard/base/os`);
};

export const loadBaseInfo = (ioOption: string, netOption: string) => {
    return http.get<Dashboard.BaseInfo>(`/dashboard/base/${ioOption}/${netOption}`);
};

export const loadCurrentInfo = (req: Dashboard.DashboardReq) => {
    return http.post<Dashboard.CurrentInfo>(`/dashboard/current`, req);
};

export const systemRestart = (operation: string) => {
    return http.post(`/dashboard/system/restart/${operation}`);
};
