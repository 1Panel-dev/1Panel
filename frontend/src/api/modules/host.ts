import http from '@/api';
import { ResPage, ReqPage } from '../interface';
import { Host } from '../interface/host';
import { TimeoutEnum } from '@/enums/http-enum';
import { deepCopy } from '@/utils/misc';
import { encodeBase64Fields } from '@/utils/base64';

// monitors
export const loadMonitor = (param: Host.MonitorSearch, currentNode?: string) => {
    return http.post<Array<Host.MonitorData>>(
        `/hosts/monitor/search`,
        param,
        TimeoutEnum.T_60S,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};
export const getNetworkOptions = (currentNode?: string) => {
    return http.get<Array<string>>(
        `/hosts/monitor/netoptions`,
        {},
        currentNode ? { headers: { CurrentNode: currentNode } } : {},
    );
};
export const getIOOptions = (currentNode?: string) => {
    return http.get<Array<string>>(
        `/hosts/monitor/iooptions`,
        {},
        currentNode ? { headers: { CurrentNode: currentNode } } : {},
    );
};
export const cleanMonitors = () => {
    return http.post(`/hosts/monitor/clean`, {});
};
export const loadMonitorSetting = (currentNode?: string) => {
    return http.get<Host.MonitorSetting>(
        `/hosts/monitor/setting`,
        {},
        currentNode ? { headers: { CurrentNode: currentNode } } : {},
    );
};
export const updateMonitorSetting = (key: string, value: string) => {
    return http.post(`/hosts/monitor/setting/update`, { key: key, value: value });
};
export const loadRuntimeDiagnosticsSummary = (currentNode?: string) => {
    return http.get<Host.RuntimeDiagnosticsSummary>(
        `/hosts/diagnostics/summary`,
        {},
        currentNode ? { headers: { CurrentNode: currentNode } } : {},
    );
};
export const loadRuntimeGoroutines = (currentNode?: string) => {
    return http.get<Host.RuntimeGoroutineSnapshot>(
        `/hosts/diagnostics/goroutines`,
        {},
        currentNode ? { headers: { CurrentNode: currentNode } } : {},
    );
};
export class RuntimeProfileDownloadError extends Error {
    constructor(message = '') {
        super(message);
        this.name = 'RuntimeProfileDownloadError';
    }
}
const parseRuntimeProfileError = async (data: unknown) => {
    if (!(data instanceof Blob) || !data.type.includes('application/json')) {
        return;
    }
    try {
        const response = JSON.parse(await data.text()) as { message?: string };
        return new RuntimeProfileDownloadError(response.message);
    } catch {
        return new RuntimeProfileDownloadError();
    }
};
export const createRuntimeProfile = async (params: Host.RuntimeProfileCreate, currentNode?: string) => {
    try {
        const data = await http.download<Blob>(`/hosts/diagnostics/profiles`, params, {
            responseType: 'blob',
            timeout: TimeoutEnum.T_60S,
            headers: currentNode ? { CurrentNode: currentNode } : undefined,
        });
        const profileError = await parseRuntimeProfileError(data);
        if (profileError) {
            throw profileError;
        }
        return data;
    } catch (error) {
        if (error instanceof RuntimeProfileDownloadError) {
            throw error;
        }
        const responseData = (error as { response?: { data?: unknown } })?.response?.data;
        const profileError = await parseRuntimeProfileError(responseData);
        throw profileError || error;
    }
};
// ssh
export const getSSHInfo = (currentNode?: string) => {
    return http.post<Host.SSHInfo>(
        `/hosts/ssh/search`,
        {},
        undefined,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};
export const operateSSH = (operation: string) => {
    return http.post(`/hosts/ssh/operate`, { operation: operation }, TimeoutEnum.T_40S);
};
export const updateSSH = (params: Host.SSHUpdate) => {
    return http.post(`/hosts/ssh/update`, params, TimeoutEnum.T_40S);
};
export const loadSSHFile = (name: string) => {
    return http.post<string>(`/hosts/ssh/file`, { name: name });
};
export const updateSSHByFile = (key: string, value: string, path = '') => {
    return http.post(`/hosts/ssh/file/update`, { key, path, value }, TimeoutEnum.T_60S);
};
export const createCert = (params: Host.RootCert) => {
    let request = deepCopy(params) as Host.RootCert;
    encodeBase64Fields(request, ['passPhrase', 'privateKey', 'publicKey']);
    return http.post(`/hosts/ssh/cert`, request);
};
export const editCert = (params: Host.RootCert) => {
    let request = deepCopy(params) as Host.RootCert;
    encodeBase64Fields(request, ['passPhrase', 'privateKey', 'publicKey']);
    return http.post(`/hosts/ssh/cert/update`, request);
};
export const searchCert = (params: ReqPage) => {
    return http.post<ResPage<Host.RootCertInfo>>(`/hosts/ssh/cert/search`, params);
};
export const deleteCert = (ids: Array<number>, forceDelete: boolean) => {
    return http.post(`/hosts/ssh/cert/delete`, { ids: ids, forceDelete: forceDelete });
};
export const syncCert = () => {
    return http.post(`/hosts/ssh/cert/sync`);
};
export const loadSSHLogs = (params: Host.searchSSHLog, currentNode?: string) => {
    return http.post<ResPage<Host.sshHistory>>(
        `/hosts/ssh/log`,
        params,
        undefined,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};
export const exportSSHLogs = (params: Host.searchSSHLog) => {
    return http.post<string>(`/hosts/ssh/log/export`, params, TimeoutEnum.T_40S);
};
export const cleanSSHLogs = () => {
    return http.post(`/hosts/ssh/log/clean`, {});
};

export const listDisks = () => {
    return http.get<Host.CompleteDiskInfo>(`/hosts/disks`);
};

export const partitionDisk = (params: Host.DiskPartition) => {
    return http.post(`/hosts/disks/partition`, params, TimeoutEnum.T_60S);
};

export const mountDisk = (params: Host.DiskMount) => {
    return http.post(`/hosts/disks/mount`, params, TimeoutEnum.T_60S);
};

export const unmountDisk = (params: Host.DiskUmount) => {
    return http.post(`/hosts/disks/unmount`, params, TimeoutEnum.T_60S);
};

export const getComponentInfo = (name: string, operateNode?: string) => {
    const params = operateNode ? `?operateNode=${operateNode}` : '';
    return http.get<Host.ComponentInfo>(`/hosts/components/${name}${params}`);
};
