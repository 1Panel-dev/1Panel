import http from '@/api';
import { HostTool } from '../interface/host-tool';

export const GetSupervisorStatus = () => {
    return http.post<HostTool.HostTool>(`/hosts/tool`, { type: 'supervisord', operate: 'status' });
};

export const OperateSupervisor = (operate: string) => {
    return http.post<any>(`/hosts/tool/operate`, { type: 'supervisord', operate: operate });
};

export const OperateSupervisorConfig = (req: HostTool.SupersivorConfig) => {
    return http.post<HostTool.SupersivorConfigRes>(`/hosts/tool/config`, req);
};

export const GetSupervisorLog = () => {
    return http.post<any>(`/hosts/tool/log`, { type: 'supervisord' });
};

export const InitSupervisor = (req: HostTool.SupersivorInit) => {
    return http.post<any>(`/hosts/tool/init`, req);
};

export const OperateSupervisorProcess = (req: HostTool.SupersivorProcess) => {
    return http.post<any>(`/hosts/tool/supervisor/process`, req);
};
