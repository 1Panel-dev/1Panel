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
            path: '/hosts/security',
            name: 'Security',
            component: () => import('@/views/host/security/index.vue'),
            meta: {
                title: 'menu.security',
            },
        },
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
            component: () => import('@/views/monitor/index.vue'),
            meta: {
                title: 'menu.monitor',
            },
        },
        {
            path: '/host/terminal',
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
