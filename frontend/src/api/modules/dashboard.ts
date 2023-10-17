import http from '@/api';
import { Dashboard } from '../interface/dashboard';

export const loadBaseInfo = (ioOption: string, netOption: string) => {
    return http.get<Dashboard.BaseInfo>(`/dashboard/base/${ioOption}/${netOption}`);
};

export const loadCurrentInfo = (ioOption: string, netOption: string) => {
    return http.get<Dashboard.CurrentInfo>(`/dashboard/current/${ioOption}/${netOption}`);
};

export const systemRestart = (operation: string) => {
    return http.post(`/dashboard/system/restart/${operation}`);
};
