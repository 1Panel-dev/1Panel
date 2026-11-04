import { ReqPage, ResPage } from '@/api/interface';

export namespace ModelDownloader {
    export interface FeaturedModelSearch extends ReqPage {
        family: string;
        weightProvider: string;
        info: string;
    }

    export interface FeaturedModelPage extends ResPage<FeaturedModel> {
        families: string[];
        weightProviders: string[];
    }

    export interface FeaturedModel {
        description: Record<string, string> | null;
        family: string;
        repoID: string;
        weightProvider: string;
        sources: Source[];
        sourceModelIds?: Partial<Record<Source, string>> | null;
        architecture: 'MoE' | 'Dense';
        backboneParams: string;
        activeParams: string;
        engram: string;
        weightGB: number;
    }

    export type Source = 'huggingface' | 'modelscope';
    export type DownloadType = 'all' | 'tokenizer';

    export interface ModelSearch extends Partial<ReqPage> {
        query: string;
        sort: 'trending' | 'downloads' | 'updated' | 'created' | 'likes' | 'most_params' | 'least_params';
        limit?: number;
    }

    export interface ModelInfoReq {
        repoID: string;
        revision?: string;
    }

    export interface ModelDownload {
        downloadType?: DownloadType;
        repoID: string;
        revision?: string;
    }

    export interface Setting {
        featuredModelsSync?: 'Enable' | 'Disable';
        modelDir: string;
        hfEndpoint: string;
        hfToken: string;
        modelScopeEndpoint: string;
        modelScopeToken: string;
    }

    export interface ModelItem {
        repoID: string;
        name: string;
        downloads: number;
        likes: number;
        size: number;
        sizeFormatted: string;
    }

    export interface ModelFile {
        name: string;
        size: number;
        sizeFormatted: string;
    }

    export interface ModelInfo extends ModelItem {
        modelCard: string;
        files: ModelFile[];
    }

    export interface TaskSearch extends ReqPage {
        ids?: number[];
        includeLocal?: boolean;
        downloadType?: DownloadType | '';
        info: string;
        status: string;
    }

    export interface TaskOperate {
        withFile?: boolean;
        id: number;
    }

    export interface TaskRetry {
        id: number;
    }

    export interface TaskItem {
        localModel?: LocalModelItem;
        downloadType: DownloadType;
        id: number;
        source: string;
        repoID: string;
        revision: string;
        modelName: string;
        targetDir: string;
        status: string;
        progress: number;
        totalSize: number;
        downloadedSize: number;
        errorMessage: string;
        createdAt: string;
        startedAt: string;
        completedAt: string;
        canCancel: boolean;
        canRetry: boolean;
        canRemoveRecord: boolean;
    }

    export interface LocalModelOption {
        name: string;
        organization: string;
        path: string;
    }

    export interface LocalModelSearch extends ReqPage {
        downloadType?: DownloadType | '';
        info: string;
    }

    export interface LocalModelItem {
        downloadType: DownloadType;
        repoID: string;
        source: Source;
        revision: string;
        name: string;
        path: string;
        size: number;
        sizeFormatted: string;
        createdAt: string;
    }

    export interface LocalModelDelete {
        withFile?: boolean;
        downloadType?: DownloadType;
        name: string;
    }
}
