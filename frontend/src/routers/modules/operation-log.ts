import { Layout } from '@/routers/constant';

const operationRouter = {
    sort: 2,
    path: '/operations',
    component: Layout,
    redirect: '/operation',
    meta: {
        title: 'menu.operations',
        icon: 'notebook',
    },
    children: [
        {
            path: '/operation',
            name: 'OperationLog',
            component: () => import('@/views/operation-log/index.vue'),
            meta: {
                keepAlive: true,
                requiresAuth: true,
                key: 'OperationLog',
            },
        },
    ],
};

export default operationRouter;
