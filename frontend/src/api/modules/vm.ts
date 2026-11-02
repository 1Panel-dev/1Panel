import http from '@/api';
import { ResPage, SearchWithPage } from '@/api/interface';
import { VM } from '@/api/interface/vm';
import { TimeoutEnum } from '@/enums/http-enum';

export const checkVMEnvironment = () => {
    return http.get<VM.Environment>('/vms/environment');
};

export const enableVMEnvironment = () => {
    return http.post('/vms/environment/enable', {}, TimeoutEnum.T_5M);
};

export const searchVMs = (params: SearchWithPage) => {
    return http.post<ResPage<VM.Item> & { limit: number }>('/vms/search', params);
};

export const syncVMs = () => {
    return http.post('/vms/sync');
};

export const getVM = (name: string, checkDisks = false) => {
    return http.post<VM.Item>('/vms/info', { name, checkDisks }, checkDisks ? TimeoutEnum.T_3M : undefined);
};

export const healthCheckVM = (name: string, signal?: AbortSignal) => {
    return http.postWithConfig<VM.HealthCheckResult>(
        '/vms/health/check',
        { name },
        { timeout: TimeoutEnum.T_5M, signal },
    );
};

export const getVMStats = (name: string) => {
    return http.post<VM.Stats>('/vms/stats', { name });
};

export const createVM = (params: VM.Create) => {
    return http.post<VM.TaskResp>('/vms', params, TimeoutEnum.T_5M);
};

export const listVMTemplates = () => {
    return http.post<VM.TemplateItem[]>('/vms/templates/list');
};

export const createVMTemplate = (params: VM.TemplateCreate) => {
    return http.post<VM.TaskResp>('/vms/templates', params, TimeoutEnum.T_5M);
};

export const deleteVMTemplate = (params: VM.TemplateDelete) => {
    return http.post<VM.TaskResp>('/vms/templates/del', params, TimeoutEnum.T_5M);
};

export const updateVM = (params: VM.Update) => {
    return http.post<VM.TaskResp>('/vms/update', params, TimeoutEnum.T_5M);
};

export const resizeVMDisk = (params: VM.DiskResize) => {
    return http.post<VM.TaskResp>('/vms/disks/resize', params, TimeoutEnum.T_5M);
};

export const renameVM = (params: VM.Rename) => {
    return http.post<VM.TaskResp>('/vms/rename', params, TimeoutEnum.T_5M);
};

export const deleteVM = (params: VM.Delete) => {
    return http.post<VM.TaskResp>('/vms/del', params, TimeoutEnum.T_5M);
};

export const listVMSnapshots = (name: string) => {
    return http.post<VM.SnapshotItem[]>('/vms/snapshots/list', { name });
};

export const createVMSnapshot = (params: VM.SnapshotCreate) => {
    return http.post<VM.TaskResp>('/vms/snapshots', params, TimeoutEnum.T_5M);
};

export const recoverVMSnapshot = (params: VM.SnapshotOperate) => {
    return http.post<VM.TaskResp>('/vms/snapshots/recover', params, TimeoutEnum.T_5M);
};

export const deleteVMSnapshot = (params: VM.SnapshotOperate) => {
    return http.post<VM.TaskResp>('/vms/snapshots/del', params, TimeoutEnum.T_5M);
};

export const operateVM = (name: string, action: string) => {
    return http.post(`/vms/operate`, { name: name, operate: action });
};

export const getVMConsole = (name: string) => {
    return http.post<VM.Graphics>('/vms/console', { name });
};

export const listVMIso = () => {
    return http.post<VM.IsoItem[]>('/vms/iso/list');
};
export const searchVMIso = (params: SearchWithPage) => {
    return http.post<ResPage<VM.IsoItem>>('/vms/iso/search', params);
};
export const createVMIso = (params: VM.IsoCreate, currentNode?: string) => {
    return http.post('/vms/iso', params, TimeoutEnum.T_5M, currentNode ? { CurrentNode: currentNode } : undefined);
};
export const startVMIsoImport = (params: VM.IsoImportStart, currentNode?: string) => {
    return http.post<{ id: number }>(
        '/vms/iso/import/start',
        params,
        undefined,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};
export const updateVMIsoImport = (params: VM.IsoImportUpdate, currentNode?: string) => {
    return http.post(
        '/vms/iso/import/update',
        params,
        undefined,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};
export const updateVMIso = (params: VM.IsoUpdate, currentNode?: string) => {
    return http.post(
        '/vms/iso/update',
        params,
        TimeoutEnum.T_5M,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};
export const deleteVMIso = (params: VM.IsoDelete) => {
    return http.post('/vms/iso/del', params, TimeoutEnum.T_5M);
};

export const listVMNetworks = () => {
    return http.post<VM.NetworkItem[]>('/vms/networks/list');
};
export const listVMNetworkBridges = () => {
    return http.get<VM.NetworkBridgeItem[]>('/vms/networks/bridges');
};
export const listVMNetworkInterfaces = () => {
    return http.get<VM.NetworkInterfaceItem[]>('/vms/networks/interfaces');
};
export const createVMNetwork = (params: VM.NetworkCreate) => {
    return http.post('/vms/networks', params, TimeoutEnum.T_5M);
};
export const updateVMNetwork = (params: VM.NetworkUpdate) => {
    return http.post('/vms/networks/update', params);
};
export const deleteVMNetwork = (params: VM.ResourceDelete) => {
    return http.post('/vms/networks/del', params, TimeoutEnum.T_5M);
};

export const listVMStorages = () => {
    return http.post<VM.StorageItem[]>('/vms/storages/list');
};
export const createVMStorage = (params: VM.StorageCreate) => {
    return http.post('/vms/storages', params, TimeoutEnum.T_5M);
};
export const updateVMStorage = (params: VM.StorageUpdate) => {
    return http.post('/vms/storages/update', params);
};
export const deleteVMStorage = (params: VM.ResourceDelete) => {
    return http.post('/vms/storages/del', params, TimeoutEnum.T_5M);
};

export const scanVMOrphans = () => {
    return http.post<VM.OrphanItem[]>('/vms/orphans/scan');
};
export const cleanVMOrphans = (items: VM.OrphanItem[]) => {
    return http.post<VM.OrphanItem[]>('/vms/orphans/clean', { items }, TimeoutEnum.T_5M);
};
