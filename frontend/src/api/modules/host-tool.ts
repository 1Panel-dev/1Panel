import http from '@/api';
import { HostTool } from '../interface/host-tool';
import { TimeoutEnum } from '@/enums/http-enum';

export const getSupervisorStatus = () => {
    return http.post<HostTool.HostTool>(`/hosts/tool`, { type: 'supervisord', operate: 'status' });
};

export const operateSupervisor = (operate: string) => {
    return http.post<any>(`/hosts/tool/operate`, { type: 'supervisord', operate: operate });
};

export const operateSupervisorConfig = (req: HostTool.SupervisorConfig) => {
    return http.post<HostTool.SupervisorConfigRes>(`/hosts/tool/config`, req);
};

export const getSupervisorLog = () => {
    return http.post<any>(`/hosts/tool/log`, { type: 'supervisord' });
};

export const initSupervisor = (req: HostTool.SupervisorInit) => {
    return http.post<any>(`/hosts/tool/init`, req);
};

export const createSupervisorProcess = (req: HostTool.SupervisorProcess) => {
    return http.post<any>(`/hosts/tool/supervisor/process`, req);
};

export const operateSupervisorProcess = (req: HostTool.ProcessReq) => {
    return http.post<any>(`/hosts/tool/supervisor/process`, req, TimeoutEnum.T_60S);
};

export const getSupervisorProcess = () => {
    return http.get<HostTool.SupervisorProcess[]>(`/hosts/tool/supervisor/process`, {}, { timeout: TimeoutEnum.T_3M });
};

export const operateSupervisorProcessFile = (req: HostTool.ProcessFileReq) => {
    return http.post<any>(`/hosts/tool/supervisor/process/file`, req, TimeoutEnum.T_60S);
};
