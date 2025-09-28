import http from '@/api';
import { ResPage, SearchWithPage } from '../interface';
import { Cronjob } from '../interface/cronjob';
import { TimeoutEnum } from '@/enums/http-enum';

export const searchCronjobPage = (params: Cronjob.Search) => {
    return http.post<ResPage<Cronjob.CronjobInfo>>(`/cronjobs/search`, params);
};

export const loadNextHandle = (spec: string) => {
    return http.post<Array<String>>(`/cronjobs/next`, { spec: spec });
};

export const editCronjobGroup = (id: number, groupID: number) => {
    return http.post(`/cronjobs/group/update`, { id: id, groupID: groupID });
};

export const importCronjob = (trans: Array<Cronjob.CronjobTrans>) => {
    return http.post('cronjobs/import', { cronjobs: trans }, TimeoutEnum.T_60S);
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

export const addCronjob = (params: Cronjob.CronjobOperate) => {
    return http.post<Cronjob.CronjobOperate>(`/cronjobs`, params);
};

export const editCronjob = (params: Cronjob.CronjobOperate) => {
    return http.post(`/cronjobs/update`, params);
};

export const deleteCronjob = (params: Cronjob.CronjobDelete) => {
    return http.post(`/cronjobs/del`, params);
};

export const searchRecords = (params: Cronjob.SearchRecord) => {
    return http.post<ResPage<Cronjob.Record>>(`cronjobs/search/records`, params);
};

export const stopCronjob = (id: number) => {
    return http.post(`cronjobs/stop`, { id: id });
};

export const cleanRecords = (id: number, cleanData: boolean, cleanRemoteData: boolean) => {
    return http.post(`cronjobs/records/clean`, { cronjobID: id, cleanData: cleanData, cleanRemoteData });
};

export const updateStatus = (params: Cronjob.UpdateStatus) => {
    return http.post(`cronjobs/status`, params);
};

export const handleOnce = (id: number) => {
    return http.post(`cronjobs/handle`, { id: id });
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
