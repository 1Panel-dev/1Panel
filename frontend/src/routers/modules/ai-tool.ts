import { Layout } from '@/routers/constant';

const databaseRouter = {
    sort: 4,
    path: '/ai-tools',
    component: Layout,
    redirect: '/ai-tools/model',
    meta: {
        icon: 'p-jiqiren2',
        title: 'menu.ai_tools',
    },
    children: [
        {
            path: '/ai-tools/model',
            name: 'OllamaModel',
            component: () => import('@/views/ai-tools/model/index.vue'),
            meta: {
                title: 'ai_tools.model.model',
                requiresAuth: true,
            },
        },
        {
            path: '/ai-tools/gpu',
            name: 'GPU',
            component: () => import('@/views/ai-tools/gpu/index.vue'),
            meta: {
                title: 'ai_tools.gpu.gpu',
                requiresAuth: true,
            },
        },
    ],
};

export default databaseRouter;
