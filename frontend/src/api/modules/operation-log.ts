import http from '@/api';
import { ResPage } from '../interface';
import { ResOperationLog } from '../interface/operation-log';

export const getOperationList = (currentPage: number, pageSize: number) => {
    return http.get<ResPage<ResOperationLog>>(`/operations?page=${currentPage}&pageSize=${pageSize}`);
};
