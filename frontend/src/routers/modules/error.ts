import { Layout } from '@/routers/constant';
// 错误页面模块

const errorRouter = {
    path: '/error',
    component: Layout,
    children: [
        {
            path: '403',
            name: '403',
            hidden: true,
            component: () => import('@/components/error-message/403.vue'),
            meta: {
                requiresAuth: true,
                title: '403页面',
                key: '403',
            },
        },
        {
            path: '404',
            name: '404',
            hidden: true,
            component: () => import('@/components/error-message/404.vue'),
            meta: {
                requiresAuth: false,
                title: '404页面',
                key: '404',
            },
        },
        {
            path: '500',
            name: '500',
            hidden: true,
            component: () => import('@/components/error-message/500.vue'),
            meta: {
                requiresAuth: false,
                title: '500页面',
                key: '500',
            },
        },
    ],
};
export default errorRouter;
