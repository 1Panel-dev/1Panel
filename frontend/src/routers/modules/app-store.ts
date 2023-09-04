import { Layout } from '@/routers/constant';

const appStoreRouter = {
    sort: 2,
    path: '/apps',
    component: Layout,
    redirect: '/apps/all',
    meta: {
        icon: 'p-appstore',
        title: 'menu.apps',
    },
    children: [
        {
            path: '/apps',
            name: 'App',
            redirect: '/apps/all',
            component: () => import('@/views/app-store/index.vue'),
            meta: {},
            children: [
                {
                    path: 'all',
                    name: 'AppAll',
                    component: () => import('@/views/app-store/apps/index.vue'),
                    props: true,
                    hidden: true,
                    meta: {
                        activeMenu: '/apps',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'installed',
                    name: 'AppInstalled',
                    component: () => import('@/views/app-store/installed/index.vue'),
                    props: true,
                    hidden: true,
                    meta: {
                        activeMenu: '/apps',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'upgrade',
                    name: 'AppUpgrade',
                    component: () => import('@/views/app-store/installed/index.vue'),
                    props: true,
                    hidden: true,
                    meta: {
                        activeMenu: '/apps',
                        requiresAuth: false,
                    },
                },
            ],
        },
    ],
};

export default appStoreRouter;
