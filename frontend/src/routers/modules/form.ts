import { Layout } from '@/routers/constant';

// 表单 Form 模块
const formRouter = {
    sort: 4,
    path: '/form',
    component: Layout,
    redirect: '/form/proForm',
    meta: {
        title: '表单 Form',
    },
    children: [
        {
            path: '/form/proForm',
            name: 'proForm',
            component: () => import('@/views/form/proForm/index.vue'),
            meta: {
                keepAlive: true,
                requiresAuth: true,
                title: '超级 Form',
                key: 'proForm',
            },
        },
        {
            path: '/form/basicForm',
            name: 'basicForm',
            component: () => import('@/views/form/basicForm/index.vue'),
            meta: {
                keepAlive: true,
                requiresAuth: true,
                title: '基础 Form',
                key: 'basicForm',
            },
        },
        {
            path: '/form/validateForm',
            name: 'validateForm',
            component: () => import('@/views/form/validateForm/index.vue'),
            meta: {
                keepAlive: true,
                requiresAuth: true,
                title: '校验 Form',
                key: 'validateForm',
            },
        },
    ],
};

export default formRouter;
