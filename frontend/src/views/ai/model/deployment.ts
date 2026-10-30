import type { ReqPage, ResPage, ResultData } from '@/api/interface';
import type { AI } from '@/api/interface/ai';
import { pageTensorRTLLM, searchOllamaModel } from '@/api/modules/ai';
import { checkAppInstalled, searchAppInstalled } from '@/api/modules/app';

export interface Deployment {
    id: number;
    engine: 'vllm' | 'ollama' | 'tensorrt';
    name: string;
    version: string;
    containerName: string;
    modelDir: string;
    models: string[];
    status: string;
    upgradable?: boolean;
}

interface VllmItem {
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
    pageVllm: (req: ReqPage & { info: string }) => Promise<ResultData<ResPage<VllmItem>>>;
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

export async function loadModelOverview(pro: boolean, fxplay: boolean) {
    const [vllm, tensors, ollama] = await Promise.all([
        pro && vllmAvailable
            ? loadVllmAPI().then((api) => readPages((page) => api.pageVllm({ ...page, info: '' })))
            : Promise.resolve([]),
        fxplay ? readPages<AI.TensorRTLLMDTO>((page) => pageTensorRTLLM({ ...page, name: '' })) : Promise.resolve([]),
        checkAppInstalled('ollama', ''),
    ]);
    const deployments: Deployment[] = vllm.map((item) => ({
        id: item.id,
        engine: 'vllm',
        upgradable: item.upgradable,
        name: item.name,
        version: item.appVersion,
        containerName: item.containerName,
        modelDir: item.modelDir,
        models: item.servedModelName ? [item.servedModelName] : [],
        status: item.status,
    }));
    deployments.push(
        ...tensors.map((item): Deployment => ({
            id: item.id!,
            engine: 'tensorrt',
            name: item.name,
            version: item.version,
            containerName: item.containerName,
            modelDir: item.modelDir,
            models: [],
            status: item.status || '',
        })),
    );
    if (ollama.data.isExist) {
        const [models, updates] = await Promise.all([
            readPages<AI.OllamaModelInfo>((page) => searchOllamaModel(page)),
            readPages((page) => searchAppInstalled({ ...page, name: ollama.data.name, update: true })),
        ]);
        deployments.push({
            id: ollama.data.appInstallId,
            engine: 'ollama',
            name: ollama.data.name || 'Ollama',
            version: ollama.data.version,
            upgradable: updates.some((app) => app.id === ollama.data.appInstallId && app.canUpdate),
            containerName: ollama.data.containerName,
            modelDir: '',
            models: models.filter((model) => model.status === 'Success').map((model) => model.name),
            status: ollama.data.status,
        });
    }
    return deployments;
}
