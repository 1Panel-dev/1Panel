import { Layout } from '@/routers/constant';

const terminalRouter = {
    sort: 8,
    path: '/terminal',
    component: Layout,
    redirect: '/terminal',
    meta: {
        icon: 'p-zhongduan',
        title: 'menu.terminal',
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
