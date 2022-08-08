import { Layout } from '@/routers/constant';

// 常用组件模块
const assemblyRouter = {
    sort: 2,
    path: '/assembly',
    component: Layout,
    redirect: '/assembly/uploadImg',
    meta: {
        title: '常用组件',
    },
    children: [
        {
            path: '/assembly/uploadImg',
            name: 'uploadImg',
            component: () => import('@/views/assembly/uploadImg/index.vue'),
            meta: {
                keepAlive: true,
                requiresAuth: true,
                title: '图片上传',
                key: 'uploadImg',
            },
        },
        {
            path: '/assembly/batchImport',
            name: 'batchImport',
            component: () => import('@/views/assembly/batchImport/index.vue'),
            meta: {
                keepAlive: true,
                requiresAuth: true,
                title: '批量上传数据',
                key: 'batchImport',
            },
        },
    ],
};

export default assemblyRouter;
