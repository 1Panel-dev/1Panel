import { Layout } from '@/routers/constant';

const hostRouter = {
    sort: 7,
    path: '/hosts',
    component: Layout,
    redirect: '/hosts/security',
    meta: {
        icon: 'p-host',
        title: 'menu.host',
    },
    children: [
        {
            path: '/hosts/files',
            name: 'File',
            component: () => import('@/views/host/file-management/index.vue'),
            meta: {
                title: 'menu.files',
            },
        },
        {
            path: '/hosts/monitor',
            name: 'Monitor',
            component: () => import('@/views/host/monitor/index.vue'),
            meta: {
                title: 'menu.monitor',
            },
        },
        {
            path: '/hosts/terminal',
            name: 'Terminal',
            component: () => import('@/views/host/terminal/index.vue'),
            meta: {
                title: 'menu.terminal',
                keepAlive: true,
            },
        },
    ],
};

export default hostRouter;
