import { AITool } from '@/api/interface/ai-tool';
import http from '@/api';
import { ResPage } from '../interface';

export const createOllamaModel = (name: string) => {
    return http.post(`/aitools/ollama/model`, { name: name });
};
export const deleteOllamaModel = (name: string) => {
    return http.post(`/aitools/ollama/model/del`, { name: name });
};
export const searchOllamaModel = (params: AITool.OllamaModelSearch) => {
    return http.post<ResPage<AITool.OllamaModelInfo>>(`/aitools/ollama/model/search`, params);
};

export const loadGPUInfo = () => {
    return http.get<any>(`/aitools/gpu/load`);
};
