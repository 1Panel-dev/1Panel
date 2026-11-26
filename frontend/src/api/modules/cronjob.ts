import http from '@/api';
import { GlobalStore } from '@/store';
import { MsgError } from '@/utils/message';
import i18n from '@/lang';
import { ResPage, SearchWithPage } from '../interface';
import { Cronjob } from '../interface/cronjob';
import { TimeoutEnum } from '@/enums/http-enum';

const requireWebsiteSelectionCapability = async () => {
    const node = GlobalStore().currentNode;
    try {
        const res = await http.get<{ websiteSelectionVersion: number }>('/cronjobs/capabilities');
        if (res.data?.websiteSelectionVersion !== 2 || GlobalStore().currentNode !== node)
            throw new Error('Unsupported agent');
    } catch (error) {
        MsgError(i18n.global.t('commons.msg.notSupportOperation'));
        throw error;
    }
};

export const searchCronjobPage = (params: Cronjob.Search, node?: string) => {
    const query = node ? `?operateNode=${node}` : '';
    return http.post<ResPage<Cronjob.CronjobInfo>>(`/cronjobs/search${query}`, params);
};

export const loadNextHandle = (spec: string) => {
    return http.post<Array<string>>(`/cronjobs/next`, { spec: spec });
};

export const editCronjobGroup = (id: number, groupID: number) => {
    return http.post(`/cronjobs/group/update`, { id: id, groupID: groupID });
};

export const importCronjob = async (trans: Array<Cronjob.CronjobTrans>) => {
    if (trans.some((item) => item.websiteSelectionVersion || item.websiteGroups?.length || item.allWebsites)) {
        await requireWebsiteSelectionCapability();
    }
    return http.post<{ warnings: string[] }>('cronjobs/import', { cronjobs: trans }, TimeoutEnum.T_60S);
};
export const exportCronjob = (params: { ids: Array<number> }) => {
    return http.download<BlobPart>('cronjobs/export', params, { responseType: 'blob', timeout: TimeoutEnum.T_40S });
};

export const loadCronjobInfo = (id: number) => {
    return http.post<Cronjob.CronjobOperate>(`/cronjobs/load/info`, { id: id });
};

export const loadScriptOptions = () => {
    return http.get<Array<Cronjob.ScriptOptions>>(`/cronjobs/script/options`);
};

export const getRecordLog = (id: number) => {
    return http.post<string>(`/cronjobs/records/log`, { id: id });
};

export const addCronjob = async (params: Cronjob.CronjobOperate) => {
    if (params.website?.split(',').some((value) => value.startsWith('group:')))
        await requireWebsiteSelectionCapability();
    return http.post<Cronjob.CronjobOperate>(`/cronjobs`, params);
};

export const editCronjob = async (params: Cronjob.CronjobOperate) => {
    if (params.website?.split(',').some((value) => value.startsWith('group:')))
        await requireWebsiteSelectionCapability();
    return http.post(`/cronjobs/update`, params);
};

export const deleteCronjob = (params: Cronjob.CronjobDelete) => {
    return http.post(`/cronjobs/del`, params);
};

export const searchRecords = (params: Cronjob.SearchRecord, timeout?: TimeoutEnum, currentNode?: string) => {
    return http.post<ResPage<Cronjob.Record>>(
        `cronjobs/search/records`,
        params,
        timeout,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};

export const stopCronjob = (id: number) => {
    return http.post(`cronjobs/stop`, { id: id });
};

export const cleanRecords = (id: number, cleanData: boolean, cleanRemoteData: boolean) => {
    return http.post(`cronjobs/records/clean`, { cronjobID: id, cleanData: cleanData, cleanRemoteData });
};

export const updateStatus = (params: Cronjob.UpdateStatus, node?: string) => {
    const query = node ? `?operateNode=${node}` : '';
    return http.post(`cronjobs/status${query}`, params);
};

export const handleOnce = (id: number, node?: string) => {
    const query = node ? `?operateNode=${node}` : '';
    return http.post(`cronjobs/handle${query}`, { id: id });
};

export const searchScript = (params: SearchWithPage) => {
    return http.post<ResPage<Cronjob.ScriptInfo>>(`/core/script/search`, params);
};
export const addScript = (params: Cronjob.ScriptOperate) => {
    return http.post(`/core/script`, params);
};
export const syncScript = (taskID: string) => {
    return http.post(`/core/script/sync`, { taskID: taskID }, TimeoutEnum.T_60S);
};
export const editScript = (params: Cronjob.ScriptOperate) => {
    return http.post(`/core/script/update`, params);
};
export const deleteScript = (ids: Array<number>) => {
    return http.post(`/core/script/del`, { ids: ids });
};
