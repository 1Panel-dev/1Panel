import type { ReqPage, ResPage, ResultData } from '@/api/interface';
import type { AI } from '@/api/interface/ai';

export interface Deployment {
    taskID?: string;
    appInstallId: number;
    agentAccountId: number;
    path: string;
    upgradable: boolean;
    id: number;
    servedModelName: string;
    name: string;
    appVersion: string;
    port: number;
    modelAccountBaseURL: string;
    modelAccountBaseURLType: AI.ModelAccountSync['modelAccountBaseURLType'];
    command: string;
    imageType: string;
    containerName: string;
    modelDir: string;
    status: string;
}

interface VllmAPI {
    listVllmCommandTemplates: () => Promise<ResultData<Array<{ model: string }>>>;
    pageVllm: (req: ReqPage & { info: string }) => Promise<ResultData<ResPage<Deployment>>>;
    operateVllm: (
        req: { id: number; operate: string; taskID: string; forceDelete?: boolean } & Partial<AI.ModelAccountSync>,
    ) => Promise<unknown>;
}

const vllmModules = import.meta.glob<VllmAPI>('@/xpack/api/modules/vllm.ts');
export const vllmAvailable = Object.keys(vllmModules).length > 0;

export function formatVllmVersion(version: string) {
    return String(version || '')
        .trim()
        .replace(/^(nvidia|intel|ascend)-/i, '');
}

export async function loadVllmAPI() {
    return Object.values(vllmModules)[0]();
}

export async function readPages<T>(search: (page: ReqPage) => Promise<ResultData<ResPage<T>>>) {
    const items: T[] = [];
    for (let page = 1; ; page++) {
        const res = await search({ page, pageSize: 100 });
        const batch = res.data.items || [];
        items.push(...batch);
        if (items.length >= res.data.total || batch.length === 0) {
            return items;
        }
    }
}

export function matchesModelRecipe(recipeModel: string, modelPath: string, repoID = '') {
    const path = modelPath.replace(/\/+$/, '');
    const name = path.split('/').pop() || '';
    const model = recipeModel.trim().replace(/\/+$/, '');
    return Boolean(
        model && [repoID || path.split('/').slice(-2).join('/'), path, name, `/models/${name}`].includes(model),
    );
}
