import http from '@/api';
import { ResPage, ReqPage } from '../interface';
import { Cronjob } from '../interface/cronjob';

export const getCronjobPage = (params: ReqPage) => {
    return http.post<ResPage<Cronjob.CronjobInfo>>(`/cronjobs/search`, params);
};

export const addCronjob = (params: Cronjob.CronjobCreate) => {
    return http.post<Cronjob.CronjobCreate>(`/cronjobs`, params);
};

export const editCronjob = (params: Cronjob.CronjobUpdate) => {
    return http.put(`/cronjobs/${params.id}`, params);
};

export const deleteCronjob = (params: { ids: number[] }) => {
    return http.post(`/cronjobs/del`, params);
};
