import { Layout } from '@/routers/constant';

const operationRouter = {
    sort: 10,
    path: '/operations',
    component: Layout,
    redirect: '/operation',
    meta: {
        title: 'menu.operations',
        icon: 'p-log',
    },
    children: [
        {
            path: '/operation',
            name: 'OperationLog',
            component: () => import('@/views/operation-log/index.vue'),
            meta: {},
        },
    ],
};

export default operationRouter;
