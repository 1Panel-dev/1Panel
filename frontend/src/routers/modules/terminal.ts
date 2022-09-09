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
            path: '/terminals/terminal',
            name: 'Terminal',
            component: () => import('@/views/terminal/index.vue'),
            meta: {
                requiresAuth: true,
                key: 'Terminal',
                title: 'terminal.conn',
                icon: 'Connection',
                activeMenu: '/terminals',
            },
        },
    ],
};

export default terminalRouter;
