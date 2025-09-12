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
                        parent: 'menu.app',
                        title: 'app.all',
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
                        parent: 'menu.app',
                        title: 'app.installed',
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
                        parent: 'menu.app',
                        title: 'app.canUpgrade',
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
                        requiresAuth: false,
                        parent: 'menu.app',
                        title: 'commons.button.set',
                    },
                },
            ],
        },
    ],
};

export default appStoreRouter;
