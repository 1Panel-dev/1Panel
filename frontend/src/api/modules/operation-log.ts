import http from '@/api';
import { ResPage, ReqPage } from '../interface';
import { ResOperationLog } from '../interface/operation-log';

export const getOperationList = (info: ReqPage) => {
    return http.post<ResPage<ResOperationLog>>(`/operations`, info);
};

export const deleteOperation = (params: { ids: number[] }) => {
    return http.post(`/operations/del`, params);
};
