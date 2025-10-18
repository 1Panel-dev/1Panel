import { AI } from '@/api/interface/ai';
import http from '@/api';
import { ResPage } from '../interface';

export const createOllamaModel = (name: string, taskID: string) => {
    return http.post(`/ai/ollama/model`, { name: name, taskID: taskID });
};
export const recreateOllamaModel = (name: string, taskID: string) => {
    return http.post(`/ai/ollama/model/recreate`, { name: name, taskID: taskID });
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

export const pageMcpServer = (req: AI.McpServerSearch) => {
    return http.post<ResPage<AI.McpServer>>(`/ai/mcp/search`, req);
};

export const createMcpServer = (req: AI.McpServer) => {
    return http.post(`/ai/mcp/server`, req);
};

export const updateMcpServer = (req: AI.McpServer) => {
    return http.post(`/ai/mcp/server/update`, req);
};

export const deleteMcpServer = (req: AI.McpServerDelete) => {
    return http.post(`/ai/mcp/server/del`, req);
};

export const operateMcpServer = (req: AI.McpServerOperate) => {
    return http.post(`/ai/mcp/server/op`, req);
};

export const bindMcpDomain = (req: AI.McpBindDomain) => {
    return http.post(`/ai/mcp/domain/bind`, req);
};

export const getMcpDomain = () => {
    return http.get<AI.McpDomainRes>(`/ai/mcp/domain/get`);
};

export const updateMcpDomain = (req: AI.McpBindDomainUpdate) => {
    return http.post(`/ai/mcp/domain/update`, req);
};

export const pageTensorRTLLM = (req: AI.TensorRTLLMSearch) => {
    return http.post<ResPage<AI.TensorRTLLMDTO>>(`/ai/tensorrt/search`, req);
};

export const createTensorRTLLM = (req: AI.TensorRTLLM) => {
    return http.post(`/ai/tensorrt/create`, req);
};

export const updateTensorRTLLM = (req: AI.TensorRTLLM) => {
    return http.post(`/ai/tensorrt/update`, req);
};

export const deleteTensorRTLLM = (req: AI.TensorRTLLMDelete) => {
    return http.post(`/ai/tensorrt/delete`, req);
};

export const operateTensorRTLLM = (req: AI.TensorRTLLMOperate) => {
    return http.post(`/ai/tensorrt/operate`, req);
};
