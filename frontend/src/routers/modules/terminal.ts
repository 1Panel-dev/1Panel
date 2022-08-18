import { Layout } from '@/routers/constant';

const terminalRouter = {
    sort: 2,
    path: '/terminals',
    component: Layout,
    redirect: '/terminal',
    meta: {
        title: 'menu.terminal',
        icon: 'monitor',
    },
    children: [
        {
            path: '/terminal',
            name: 'Terminal',
            component: () => import('@/views/terminal/index.vue'),
            meta: {
                keepAlive: true,
                requiresAuth: true,
                key: 'Terminal',
            },
        },
    ],
};

export default terminalRouter;
