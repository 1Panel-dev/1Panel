import http from '@/api';
import { Monitor } from '../interface/monitor';

export const loadMonitor = (param: Monitor.MonitorSearch) => {
    return http.post<Array<Monitor.MonitorData>>(`/monitors/search`, param);
};

export const getNetworkOptions = () => {
    return http.get<Array<string>>(`/monitors/netoptions`);
};

export const getIOOptions = () => {
    return http.get<Array<string>>(`/monitors/iooptions`);
};
