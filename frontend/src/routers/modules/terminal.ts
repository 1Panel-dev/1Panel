import { Layout } from '@/routers/constant';

const terminalRouter = {
    sort: 7,
    path: '/terminal',
    component: Layout,
    redirect: '/terminal',
    meta: {
        title: 'menu.terminal',
        icon: 'p-terminal3',
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
