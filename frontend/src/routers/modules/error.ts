import { Layout } from '@/routers/constant';

const errorRouter = {
    path: '/error',
    component: Layout,
    children: [
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
    ],
};
export default errorRouter;
