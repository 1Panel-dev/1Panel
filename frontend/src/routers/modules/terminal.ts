import { Layout } from '@/routers/constant';

const terminalRouter = {
    sort: 8,
    path: '/terminal',
    name: 'Terminal-Menu',
    component: Layout,
    redirect: '/terminal',
    meta: {
        title: 'menu.terminal',
        icon: 'p-terminal2',
        protectedRoleOnly: true,
    },
    children: [
        {
            path: '/terminal',
            name: 'Terminal',
            props: true,
            component: () => import('@/views/terminal/index.vue'),
            meta: {
                keepAlive: true,
                protectedRoleOnly: true,
            },
        },
    ],
};

export default terminalRouter;
