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
            path: '/ai/agents/agent',
            name: 'Agents',
            component: () => import('@/views/ai/agents/agent/index.vue'),
            meta: {
                icon: 'p-jiqiren2',
                title: 'aiTools.agents.agents',
                requiresAuth: true,
            },
        },
        {
            path: '/ai/agents/model',
            name: 'AgentsModel',
            component: () => import('@/views/ai/agents/model/index.vue'),
            meta: {
                title: 'aiTools.agents.account',
                activeMenu: '/ai/agents/agent',
                requiresAuth: true,
            },
        },
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
        {
            path: '/ai/gpu/current',
            name: 'GPU',
            component: () => import('@/views/ai/gpu/current/index.vue'),
            meta: {
                icon: 'p-gpu-menu',
                title: 'aiTools.gpu.gpu',
                activeMenu: '/ai/gpu',
                requiresAuth: true,
            },
        },
        {
            path: '/ai/gpu/history',
            name: 'GPUHistory',
            component: () => import('@/views/ai/gpu/history/index.vue'),
            meta: {
                title: 'aiTools.gpu.history',
                activeMenu: '/ai/gpu',
                requiresAuth: true,
            },
        },
    ],
};

export default databaseRouter;
