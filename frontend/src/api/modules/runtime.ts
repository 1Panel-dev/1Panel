import http from '@/api';
import { ResPage } from '../interface';
import { Runtime } from '../interface/runtime';

export const SearchRuntimes = (req: Runtime.RuntimeReq) => {
    return http.post<ResPage<Runtime.RuntimeDTO>>(`/runtimes/search`, req);
};

export const CreateRuntime = (req: Runtime.RuntimeCreate) => {
    return http.post<any>(`/runtimes`, req);
};

export const DeleteRuntime = (req: Runtime.RuntimeDelete) => {
    return http.post<any>(`/runtimes/del`, req);
};

export const GetRuntime = (id: number) => {
    return http.get<Runtime.RuntimeDTO>(`/runtimes/${id}`);
};

export const UpdateRuntime = (req: Runtime.RuntimeUpdate) => {
    return http.post<any>(`/runtimes/update`, req);
};
