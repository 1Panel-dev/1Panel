import http from '@/api';
import { ResPage, ReqPage } from '../interface';
import { Runtime } from '../interface/runtime';
import { TimeoutEnum } from '@/enums/http-enum';
import { App } from '@/api/interface/app';

export const SearchRuntimes = (req: Runtime.RuntimeReq) => {
    return http.post<ResPage<Runtime.RuntimeDTO>>(`/runtimes/search`, req);
};

export const CreateRuntime = (req: Runtime.RuntimeCreate) => {
    return http.post<Runtime.Runtime>(`/runtimes`, req);
};

export const DeleteRuntime = (req: Runtime.RuntimeDelete) => {
    return http.post<any>(`/runtimes/del`, req);
};

export const RuntimeDeleteCheck = (runTimeId: number) => {
    return http.get<App.AppInstallResource[]>(`runtimes/installed/delete/check/${runTimeId}`);
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

export const SearchPHPExtensions = (req: ReqPage) => {
    return http.post<ResPage<Runtime.PHPExtensions>>(`/runtimes/php/extensions/search`, req);
};

export const ListPHPExtensions = (req: Runtime.PHPExtensionsList) => {
    return http.post<ResPage<Runtime.PHPExtensions>>(`/runtimes/php/extensions/search`, req);
};

export const CreatePHPExtensions = (req: Runtime.PHPExtensionsCreate) => {
    return http.post<any>(`/runtimes/php/extensions`, req);
};

export const UpdatePHPExtensions = (req: Runtime.PHPExtensionsUpdate) => {
    return http.post<any>(`/runtimes/php/extensions/update`, req);
};

export const DeletePHPExtensions = (req: Runtime.PHPExtensionsDelete) => {
    return http.post<any>(`/runtimes/php/extensions/del`, req);
};

export const SyncRuntime = () => {
    return http.post(`/runtimes/sync`, {});
};
