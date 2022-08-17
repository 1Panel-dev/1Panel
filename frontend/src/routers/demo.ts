import { Layout } from '@/routers/constant';

// demo
const demoRouter = {
    sort: 1,
    path: '/demos',
    component: Layout,
    redirect: '/demos/table',
    meta: {
        icon: 'apple',
        title: 'menu.demo',
    },
    children: [
        {
            path: '/demos/table',
            name: 'Table',
            component: () => import('@/views/demos/table/index.vue'),
        },
        {
            path: '/demos/table/:op/:id?',
            name: 'DemoOperate',
            props: true,
            hidden: true,
            component: () => import('@/views/demos/table/operate/index.vue'),
            meta: {
                activeMenu: '/demos/table',
            },
        },
    ],
};

export default demoRouter;
