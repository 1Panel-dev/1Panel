import { Layout } from '@/routers/constant';

const appStoreRouter = {
    sort: 2,
    path: '/apps',
    component: Layout,
    redirect: '/apps',
    meta: {
        icon: 'p-appstore',
        title: 'menu.apps',
    },
    children: [
        {
            path: '/apps',
            name: 'App',
            component: () => import('@/views/app-store/index.vue'),
            meta: {},
        },
        // {
        //     path: '/apps/detail/:name',
        //     name: 'AppDetail',
        //     component: () => import('@/views/app-store/detail/index.vue'),
        //     meta: {
        //         hidden: true,
        //         title: 'menu.apps',
        //     },
        // },
        {
            path: '/apps/detail/:name',
            name: 'AppDetail',
            props: true,
            hidden: true,
            component: () => import('@/views/app-store/detail/index.vue'),
            meta: {
                activeMenu: '/apps',
            },
        },
    ],
};

export default appStoreRouter;
