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
        {
            path: '/apps/detail/:id',
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
