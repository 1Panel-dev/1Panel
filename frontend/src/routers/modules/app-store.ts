import { Layout } from '@/routers/constant';

const appStoreRouter = {
    sort: 2,
    path: '/apps',
    name: 'App-Menu',
    component: Layout,
    redirect: '/apps/all',
    meta: {
        icon: 'p-appstore',
        title: 'menu.apps',
        permission: 'app_view',
    },
    children: [
        {
            path: '/apps',
            name: 'App',
            redirect: '/apps/all',
            component: () => import('@/views/app-store/index.vue'),
            meta: {
                permission: 'app_view',
            },
            children: [
                {
                    path: 'all',
                    name: 'AppAll',
                    component: () => import('@/views/app-store/apps/index.vue'),
                    props: true,
                    hidden: true,
                    meta: {
                        activeMenu: '/apps',
                        parent: 'menu.app',
                        title: 'app.all',
                        permission: 'app_view',
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
                        parent: 'menu.app',
                        title: 'app.installed',
                        permission: 'app_view',
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
                        parent: 'menu.app',
                        title: 'app.canUpgrade',
                        permission: 'app_view',
                    },
                },
                {
                    path: 'setting',
                    name: 'AppStoreSetting',
                    component: () => import('@/views/app-store/setting/index.vue'),
                    props: true,
                    hidden: true,
                    meta: {
                        activeMenu: '/apps',
                        parent: 'menu.app',
                        title: 'commons.button.set',
                        permission: 'app_view',
                    },
                },
            ],
        },
    ],
};

export default appStoreRouter;
