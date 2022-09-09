import { Layout } from '@/routers/constant';

const operationRouter = {
    sort: 3,
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
                requiresAuth: true,
                key: 'OperationLog',
            },
        },
    ],
};

export default operationRouter;
