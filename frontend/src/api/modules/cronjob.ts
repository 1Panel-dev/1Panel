import http from '@/api';
import { ResPage, SearchWithPage } from '../interface';
import { Cronjob } from '../interface/cronjob';
import { TimeoutEnum } from '@/enums/http-enum';

export const getCronjobPage = (params: SearchWithPage) => {
    return http.post<ResPage<Cronjob.CronjobInfo>>(`/cronjobs/search`, params);
};

export const getRecordLog = (id: number) => {
    return http.post<string>(`/cronjobs/records/log`, { id: id });
};

export const addCronjob = (params: Cronjob.CronjobCreate) => {
    return http.post<Cronjob.CronjobCreate>(`/cronjobs`, params);
};

export const editCronjob = (params: Cronjob.CronjobUpdate) => {
    return http.post(`/cronjobs/update`, params);
};

export const deleteCronjob = (params: Cronjob.CronjobDelete) => {
    return http.post(`/cronjobs/del`, params);
};

export const searchRecords = (params: Cronjob.SearchRecord) => {
    return http.post<ResPage<Cronjob.Record>>(`cronjobs/search/records`, params);
};

export const cleanRecords = (id: number, cleanData: boolean) => {
    return http.post(`cronjobs/records/clean`, { cronjobID: id, cleanData: cleanData });
};

export const getRecordDetail = (params: string) => {
    return http.post<string>(`cronjobs/search/detail`, { path: params });
};

export const updateStatus = (params: Cronjob.UpdateStatus) => {
    return http.post(`cronjobs/status`, params);
};

export const downloadRecordCheck = (params: Cronjob.Download) => {
    return http.post<string>(`cronjobs/download`, params, TimeoutEnum.T_40S);
};
export const downloadRecord = (params: Cronjob.Download) => {
    return http.download<BlobPart>(`cronjobs/download`, params, {
        responseType: 'blob',
        timeout: TimeoutEnum.T_40S,
    });
};

export const handleOnce = (id: number) => {
    return http.post(`cronjobs/handle`, { id: id });
};
