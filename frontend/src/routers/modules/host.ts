import { Layout } from '@/routers/constant';

const hostRouter = {
    sort: 6,
    path: '/host',
    component: Layout,
    redirect: '/host/security',
    meta: {
        icon: 'p-host',
        title: 'menu.host',
    },
    children: [
        {
            path: '/host/security',
            name: 'Security',
            component: () => import('@/views/host/security/index.vue'),
            meta: {
                title: 'menu.security',
                keepAlive: true,
            },
        },
        {
            path: '/host/files',
            name: 'File',
            component: () => import('@/views/host/file-management/index.vue'),
            meta: {
                title: 'menu.files',
                keepAlive: true,
            },
        },
        // {
        //     path: '/host/terminal',
        //     name: 'Terminal',
        //     component: () => import('@/views/host/terminal/index.vue'),
        //     meta: {
        //         title: 'menu.terminal',
        //         keepAlive: true,
        //     },
        // },
    ],
};

export default hostRouter;
