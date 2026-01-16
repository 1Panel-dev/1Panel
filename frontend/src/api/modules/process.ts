import http from '@/api';
import { Process } from '../interface/process';

export const stopProcess = (req: Process.StopReq) => {
    return http.post<any>(`/process/stop`, req);
};

export const getProcessByID = (pid: number) => {
    return http.get<Process.PsProcessData>(`/process/${pid}`);
};

export const getListeningProcess = () => {
    return http.post<Process.ListeningProcess[]>(`/process/listening`);
};
