import http from '@/api';
import { ResPage, ReqPage } from '../interface';
import { Runtime } from '../interface/runtime';
import { TimeoutEnum } from '@/enums/http-enum';
import { App } from '@/api/interface/app';
import { File } from '../interface/file';
import { HostTool } from '../interface/host-tool';

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

export const GetPHPExtensions = (id: number) => {
    return http.get<Runtime.PHPExtensionsRes>(`/runtimes/php/${id}/extensions`);
};

export const InstallPHPExtension = (req: Runtime.PHPExtensionInstall) => {
    return http.post(`/runtimes/php/extensions/install`, req);
};

export const UnInstallPHPExtension = (req: Runtime.PHPExtensionInstall) => {
    return http.post(`/runtimes/php/extensions/uninstall`, req);
};

export const GetPHPConfig = (id: number) => {
    return http.get<Runtime.PHPConfig>(`/runtimes/php/config/${id}`);
};

export const UpdatePHPConfig = (req: Runtime.PHPConfigUpdate) => {
    return http.post<any>(`/runtimes/php/config/`, req);
};

export const UpdatePHPFile = (req: Runtime.PHPUpdate) => {
    return http.post<any>(`/runtimes/php/update`, req);
};

export const GetPHPConfigFile = (req: Runtime.PHPFileReq) => {
    return http.post<File.File>(`/runtimes/php/file`, req);
};

export const UpdateFPMConfig = (req: Runtime.FPMConfig) => {
    return http.post(`/runtimes/php/fpm/config`, req);
};

export const GetFPMConfig = (id: number) => {
    return http.get<Runtime.FPMConfig>(`/runtimes/php/fpm/config/${id}`);
};

export const GetSupervisorProcess = (id: number) => {
    return http.get<HostTool.ProcessStatus[]>(`/runtimes/supervisor/process/${id}`);
};

export const operateSupervisorProcess = (req: Runtime.ProcessReq) => {
    return http.post(`/runtimes/supervisor/process`, req, TimeoutEnum.T_60S);
};

export const operateSupervisorProcessFile = (req: Runtime.ProcessFileReq) => {
    return http.post<string>(`/runtimes/supervisor/process/file`, req, TimeoutEnum.T_60S);
};

export const createSupervisorProcess = (req: Runtime.SupersivorProcess) => {
    return http.post(`/runtimes/supervisor/process`, req);
};

export const getPHPContainerConfig = (id: number) => {
    return http.get<Runtime.PHPContainerConfig>(`/runtimes/php/container/${id}`);
};

export const updatePHPContainerConfig = (req: Runtime.PHPContainerConfig) => {
    return http.post<any>(`/runtimes/php/container/update`, req);
};

export const updateRemark = (req: Runtime.RemarkUpdate) => {
    return http.post(`/runtimes/remark`, req);
};

export const getFPMStatus = (id: number) => {
    return http.get<Runtime.FpmStatus[]>(`/runtimes/php/fpm/status/${id}`);
};
