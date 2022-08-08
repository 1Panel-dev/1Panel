import { Layout } from '@/routers/constant';

// 超级表格模块
const proTableRouter = {
    sort: 5,
    path: '/proTable',
    component: Layout,
    redirect: '/proTable/useHooks',
    meta: {
        title: '超级表格',
    },
    children: [
        {
            path: '/proTable/useHooks',
            name: 'useHooks',
            component: () => import('@/views/proTable/useHooks/index.vue'),
            meta: {
                keepAlive: true,
                requiresAuth: true,
                title: '使用 Hooks',
                key: 'useHooks',
            },
        },
        {
            path: '/proTable/useComponent',
            name: 'useComponent',
            component: () => import('@/views/proTable/useComponent/index.vue'),
            meta: {
                keepAlive: true,
                requiresAuth: true,
                title: '使用 Component',
                key: 'useComponent',
            },
        },
    ],
};

export default proTableRouter;
