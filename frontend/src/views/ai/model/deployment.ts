import type { ReqPage, ResPage, ResultData } from '@/api/interface';
import type { AI } from '@/api/interface/ai';
import { searchOllamaModel } from '@/api/modules/ai';
import { checkAppInstalled, searchAppInstalled } from '@/api/modules/app';

export interface Deployment {
    taskID?: string;
    id: number;
    engine: 'vllm' | 'ollama';
    name: string;
    version: string;
    ports: number[];
    containerName: string;
    modelDir: string;
    models: string[];
    status: string;
    upgradable?: boolean;
}

interface VllmItem {
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

export async function loadModelOverview(pro: boolean) {
    const [vllm, ollama] = await Promise.all([
        pro && vllmAvailable
            ? loadVllmAPI().then((api) => readPages((page) => api.pageVllm({ ...page, info: '' })))
            : Promise.resolve([]),
        checkAppInstalled('ollama', ''),
    ]);
    const deployments: Deployment[] = vllm.map((item) => ({
        id: item.id,
        engine: 'vllm',
        upgradable: item.upgradable,
        name: item.name,
        version: item.appVersion,
        ports: item.port > 0 ? [item.port] : [],
        containerName: item.containerName,
        modelDir: item.modelDir,
        models: item.servedModelName ? [item.servedModelName] : [],
        status: item.status,
        taskID: item.taskID,
    }));
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
            ports: ollama.data.httpPort > 0 ? [ollama.data.httpPort] : [],
            upgradable: updates.some((app) => app.id === ollama.data.appInstallId && app.canUpdate),
            containerName: ollama.data.containerName,
            modelDir: '',
            models: models.filter((model) => model.status === 'Success').map((model) => model.name),
            status: ollama.data.status,
            taskID: ollama.data.taskID,
        });
    }
    return deployments;
}
