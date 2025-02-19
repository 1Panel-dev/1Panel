import { AI } from '@/api/interface/ai';
import http from '@/api';
import { ResPage } from '../interface';

export const createOllamaModel = (name: string) => {
    return http.post(`/ai/ollama/model`, { name: name });
};
export const recreateOllamaModel = (name: string) => {
    return http.post(`/ai/ollama/model/recreate`, { name: name });
};
export const deleteOllamaModel = (ids: Array<number>, force: boolean) => {
    return http.post(`/ai/ollama/model/del`, { ids: ids, forceDelete: force });
};
export const searchOllamaModel = (params: AI.OllamaModelSearch) => {
    return http.post<ResPage<AI.OllamaModelInfo>>(`/ai/ollama/model/search`, params);
};
export const loadOllamaModel = (name: string) => {
    return http.post<string>(`/ai/ollama/model/load`, { name: name });
};
export const syncOllamaModel = () => {
    return http.post<Array<AI.OllamaModelDropInfo>>(`/ai/ollama/model/sync`);
};
export const closeOllamaModel = (name: string) => {
    return http.post(`/ai/ollama/close`, { name: name });
};

export const loadGPUInfo = () => {
    return http.get<any>(`/ai/gpu/load`);
};

export const bindDomain = (req: AI.BindDomain) => {
    return http.post(`/ai/domain/bind`, req);
};

export const getBindDomain = (req: AI.BindDomainReq) => {
    return http.post<AI.BindDomainRes>(`/ai/domain/get`, req);
};

export const updateBindDomain = (req: AI.BindDomain) => {
    return http.post(`/ai/domain/update`, req);
};
