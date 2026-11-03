import http from '@/api';
import { ResPage } from '@/api/interface';
import { ModelDownloader } from '@/api/interface/model-downloader';
import { TimeoutEnum } from '@/enums/http-enum';

export const searchHFModels = (req: ModelDownloader.ModelSearch) => {
    return http.post<ResPage<ModelDownloader.ModelItem>>('/ai/model/downloader/hf/search', req, TimeoutEnum.T_40S);
};

export const getHFModelInfo = (req: ModelDownloader.ModelInfoReq) => {
    return http.post<ModelDownloader.ModelInfo>('/ai/model/downloader/hf/info', req, TimeoutEnum.T_40S);
};

export const startHFModelDownload = (req: ModelDownloader.ModelDownload) => {
    return http.post<ModelDownloader.TaskItem>('/ai/model/downloader/hf/download', req, TimeoutEnum.T_40S);
};

export const searchModelScopeModels = (req: ModelDownloader.ModelSearch) => {
    return http.post<ResPage<ModelDownloader.ModelItem>>(
        '/ai/model/downloader/modelscope/search',
        req,
        TimeoutEnum.T_40S,
    );
};

export const getModelScopeModelInfo = (req: ModelDownloader.ModelInfoReq) => {
    return http.post<ModelDownloader.ModelInfo>('/ai/model/downloader/modelscope/info', req, TimeoutEnum.T_40S);
};

export const startModelScopeModelDownload = (req: ModelDownloader.ModelDownload) => {
    return http.post<ModelDownloader.TaskItem>('/ai/model/downloader/modelscope/download', req, TimeoutEnum.T_40S);
};

export const getModelDownloaderSettings = () => {
    return http.get<ModelDownloader.Setting>('/ai/model/downloader/settings');
};

export const updateModelDownloaderSettings = (req: ModelDownloader.Setting) => {
    return http.post('/ai/model/downloader/settings', req);
};

export const pageModelDownloadTasks = (req: ModelDownloader.TaskSearch) => {
    return http.post<ResPage<ModelDownloader.TaskItem>>('/ai/model/downloader/tasks/search', req);
};

export const cancelModelDownloadTask = (req: ModelDownloader.TaskOperate) => {
    return http.post('/ai/model/downloader/tasks/cancel', req);
};

export const retryModelDownloadTask = (req: ModelDownloader.TaskRetry) => {
    return http.post<ModelDownloader.TaskItem>('/ai/model/downloader/tasks/retry', req);
};

export const removeModelDownloadTask = (req: ModelDownloader.TaskOperate) => {
    return http.post('/ai/model/downloader/tasks/remove', req);
};

export const pageLocalDownloadedModels = (req: ModelDownloader.LocalModelSearch) => {
    return http.post<ResPage<ModelDownloader.LocalModelItem>>('/ai/model/downloader/local/search', req);
};

export const listLocalModelOptions = () => {
    return http.get<ModelDownloader.LocalModelOption[]>('/ai/model/downloader/local/options');
};

export const deleteLocalDownloadedModel = (req: ModelDownloader.LocalModelDelete) => {
    return http.post('/ai/model/downloader/local/delete', req);
};

export const syncFeaturedModels = () => {
    return http.post('/ai/model/downloader/featured/sync', {}, TimeoutEnum.T_60S);
};

export const pageFeaturedModels = (req: ModelDownloader.FeaturedModelSearch) => {
    return http.post<ModelDownloader.FeaturedModelPage>('/ai/model/downloader/featured/search', req);
};

export const updateFeaturedModelsSync = (value: string) => {
    return http.post('/ai/model/downloader/featured/sync/status', { value });
};
