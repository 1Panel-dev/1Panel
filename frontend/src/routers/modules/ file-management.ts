import { Layout } from '@/routers/constant';

// demo
const demoRouter = {
    sort: 3,
    path: '/files',
    component: Layout,
    redirect: '/files',
    meta: {
        icon: 'files',
        title: 'menu.files',
    },
    children: [
        {
            path: '/files',
            name: 'File',
            component: () => import('@/views/file-management/index.vue'),
            meta: {
                keepAlive: true,
            },
        },
    ],
};

export default demoRouter;
