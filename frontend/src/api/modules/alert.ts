import http from '@/api';
import { ResPage } from '@/api/interface';
import { Alert } from '../interface/alert';
import { deepCopy } from '@/utils/misc';
import { GlobalStore } from '@/store';

const alertConfigHiddenTypes = ['sms'];

const resolveAlertConfigExcludeTypes = (excludeTypes: string[] = []) => {
    const globalStore = GlobalStore();
    const types = new Set(excludeTypes);
    if (globalStore.isIntl || globalStore.isEE) {
        alertConfigHiddenTypes.forEach((type) => types.add(type));
    }
    return Array.from(types);
};
export const SearchAlerts = (req: Alert.AlertSearch, currentNode?: string) => {
    return http.post<ResPage<Alert.AlertInfo>>(
        `/alert/search`,
        req,
        undefined,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};

export const CreateAlert = (req: Alert.AlertCreateReq) => {
    let request = deepCopy(req) as Alert.AlertCreateReq;
    return http.post<any>(`/alert`, request);
};

export const UpdateAlert = (req: Alert.AlertUpdateReq) => {
    return http.post<any>(`/alert/update`, req);
};

export const DeleteAlert = (req: Alert.DelReq) => {
    return http.post<any>(`/alert/del`, req);
};

export const UpdateAlertStatus = (req: Alert.AlertUpdateStatusReq) => {
    return http.post<any>(`/alert/status`, req);
};

export const ListDisks = () => {
    return http.get<Alert.DisksDTO[]>(`/alert/disks/list`);
};

export const SearchAlertLogs = (req: Alert.AlertLogSearch, currentNode?: string) => {
    return http.post<ResPage<Alert.AlertLog>>(
        `/alert/logs/search`,
        req,
        undefined,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};

export const CleanAlertLogs = () => {
    return http.post<any>(`/alert/logs/clean`);
};

export const ListClams = () => {
    return http.get<Alert.ClamsDTO[]>(`/alert/clams/list`);
};

export const ListCronJob = (req: Alert.CronJobReq) => {
    return http.post<Alert.CronJobDTO[]>(`/alert/cronjob/list`, req);
};

export const ListAlertConfigs = (req: Alert.AlertConfigFilterReq = {}, currentNode?: string) => {
    const request = {
        ...req,
        excludeTypes: resolveAlertConfigExcludeTypes(req.excludeTypes),
    };
    return http.post<Alert.AlertConfigInfo[]>(
        `/alert/config/info`,
        request,
        undefined,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};

export const PageAlertConfigs = (req: Alert.AlertConfigPageReq, currentNode?: string) => {
    const request = {
        ...req,
        excludeTypes: resolveAlertConfigExcludeTypes(req.excludeTypes),
    };
    return http.post<ResPage<Alert.AlertConfigInfo>>(
        `/alert/config/search`,
        request,
        undefined,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};

export const DeleteAlertConfig = (req: Alert.DelReq) => {
    return http.post<any>(`/alert/config/del`, req);
};

export const UpdateAlertConfig = (req: Alert.AlertConfigUpdateReq) => {
    return http.post<any>(`/alert/config/update`, req);
};

export const TestAlertConfig = (req: Alert.AlertConfigTest) => {
    return http.post<any>(`/alert/config/test`, req);
};

export const SyncAlertInfo = (req: Alert.AlertLogId) => {
    return http.post<any>(`/xpack/alert/logs/sync`, req);
};

export const SyncAlertAll = () => {
    return http.post<any>(`/xpack/alert/logs/sync/all`);
};

export const SyncOfflineAlert = () => {
    return http.post<any>(`/core/xpack/alert/offline/sync`);
};
