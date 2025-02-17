import { Layout } from '@/routers/constant';

const databaseRouter = {
    sort: 4,
    path: '/ai',
    component: Layout,
    redirect: '/ai/model',
    meta: {
        icon: 'p-jiqiren2',
        title: 'menu.ai_tools',
    },
    children: [
        {
            path: '/ai/model',
            name: 'OllamaModel',
            component: () => import('@/views/ai/model/index.vue'),
            meta: {
                title: 'ai_tools.model.model',
                requiresAuth: true,
            },
        },
        {
            path: '/ai/gpu',
            name: 'GPU',
            component: () => import('@/views/ai/gpu/index.vue'),
            meta: {
                title: 'ai_tools.gpu.gpu',
                requiresAuth: true,
            },
        },
    ],
};

export default databaseRouter;
