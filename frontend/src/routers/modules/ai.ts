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
            path: '/ai/model',
            name: 'OllamaModel',
            component: () => import('@/views/ai/model/index.vue'),
            meta: {
                title: 'aiTools.model.model',
                requiresAuth: true,
            },
        },
        {
            path: '/ai/mcp',
            name: 'MCPServer',
            component: () => import('@/views/ai/mcp/server/index.vue'),
            meta: {
                title: 'menu.mcp',
                requiresAuth: true,
            },
        },
        {
            path: '/ai/gpu',
            name: 'GPU',
            component: () => import('@/views/ai/gpu/index.vue'),
            meta: {
                title: 'aiTools.gpu.gpu',
                requiresAuth: true,
            },
        },
    ],
};

export default databaseRouter;
