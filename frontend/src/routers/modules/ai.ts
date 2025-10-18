import { Layout } from '@/routers/constant';

const databaseRouter = {
    sort: 4,
    path: '/ai',
    name: 'AI-Menu',
    component: Layout,
    redirect: '/ai/model',
    meta: {
        icon: 'p-jiqiren2',
        title: 'menu.aiTools',
    },
    children: [
        {
            path: '/ai/model/ollama',
            name: 'OllamaModel',
            component: () => import('@/views/ai/model/ollama/index.vue'),
            meta: {
                icon: 'p-moxing-menu',
                title: 'aiTools.model.model',
                requiresAuth: true,
            },
        },
        {
            path: '/ai/mcp',
            name: 'MCPServer',
            component: () => import('@/views/ai/mcp/server/index.vue'),
            meta: {
                icon: 'p-mcp-menu',
                title: 'menu.mcp',
                requiresAuth: true,
            },
        },
        {
            path: '/ai/gpu',
            name: 'GPU',
            component: () => import('@/views/ai/gpu/index.vue'),
            meta: {
                icon: 'p-gpu-menu',
                title: 'aiTools.gpu.gpu',
                requiresAuth: true,
            },
        },
        {
            path: '/ai/model/tensorrt',
            hidden: true,
            name: 'TensorRTLLm',
            component: () => import('@/views/ai/model/tensorrt/index.vue'),
            meta: {
                title: 'aiTools.tensorRT.llm',
                activeMenu: '/ai/model/ollama',
                requiresAuth: true,
            },
        },
    ],
};

export default databaseRouter;
