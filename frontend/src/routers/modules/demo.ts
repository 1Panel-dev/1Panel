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
            meta: {
                keepAlive: true,
            },
        },
        {
            path: '/demos/table/:op',
            name: 'DemoCreate',
            props: true,
            hidden: true,
            component: () => import('@/views/demos/table/operate/index.vue'),
            meta: {
                activeMenu: '/demos/table',
                keepAlive: true,
            },
        },
    ],
};

export default demoRouter;
