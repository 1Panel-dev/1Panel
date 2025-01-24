import http from '@/api';
import { Process } from '../interface/process';

export const stopProcess = (req: Process.StopReq) => {
    return http.post<any>(`/process/stop`, req);
};
