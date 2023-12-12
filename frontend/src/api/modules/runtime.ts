import http from '@/api';
import { ResPage } from '../interface';
import { Runtime } from '../interface/runtime';
import { TimeoutEnum } from '@/enums/http-enum';

export const SearchRuntimes = (req: Runtime.RuntimeReq) => {
    return http.post<ResPage<Runtime.RuntimeDTO>>(`/runtimes/search`, req);
};

export const CreateRuntime = (req: Runtime.RuntimeCreate) => {
    return http.post<Runtime.Runtime>(`/runtimes`, req);
};

export const DeleteRuntime = (req: Runtime.RuntimeDelete) => {
    return http.post<any>(`/runtimes/del`, req);
};

export const GetRuntime = (id: number) => {
    return http.get<Runtime.RuntimeDTO>(`/runtimes/${id}`);
};

export const UpdateRuntime = (req: Runtime.RuntimeUpdate) => {
    return http.post<any>(`/runtimes/update`, req);
};

export const GetNodeScripts = (req: Runtime.NodeReq) => {
    return http.post<Runtime.NodeScripts[]>(`/runtimes/node/package`, req);
};

export const OperateRuntime = (req: Runtime.RuntimeOperate) => {
    return http.post<any>(`/runtimes/operate`, req);
};

export const GetNodeModules = (req: Runtime.NodeModuleReq) => {
    return http.post<Runtime.NodeModule[]>(`/runtimes/node/modules`, req);
};

export const OperateNodeModule = (req: Runtime.NodeModuleReq) => {
    return http.post<any>(`/runtimes/node/modules/operate`, req, TimeoutEnum.T_10M);
};
