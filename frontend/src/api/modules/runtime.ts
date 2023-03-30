import http from '@/api';
import { ResPage } from '../interface';
import { Runtime } from '../interface/runtime';

export const SearchRuntimes = (req: Runtime.RuntimeReq) => {
    return http.post<ResPage<Runtime.RuntimeDTO>>(`/runtimes/search`, req);
};

export const CreateRuntime = (req: Runtime.RuntimeCreate) => {
    return http.post<any>(`/runtimes`, req);
};
