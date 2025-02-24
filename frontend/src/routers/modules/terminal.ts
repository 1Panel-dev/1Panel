import { Layout } from '@/routers/constant';

const terminalRouter = {
    sort: 8,
    path: '/terminal',
    component: Layout,
    redirect: '/terminal',
    meta: {
        title: 'menu.terminal',
        icon: 'p-terminal2',
    },
    children: [
        {
            path: '/terminal',
            name: 'Terminal',
            props: true,
            component: () => import('@/views/terminal/index.vue'),
            meta: {
                keepAlive: true,
                requiresAuth: false,
            },
        },
    ],
};

export default terminalRouter;
