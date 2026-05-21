import { Layout } from '@/routers/constant';

const cronRouter = {
    sort: 9,
    path: '/cronjobs',
    name: 'Cronjob-Menu',
    component: Layout,
    redirect: '/cronjobs/cronjob',
    meta: {
        icon: 'p-plan',
        title: 'menu.cronjob',
        permission: 'cronjob_view',
    },
    children: [
        {
            path: '/cronjobs',
            name: 'Cronjob',
            redirect: '/cronjobs/cronjob',
            component: () => import('@/views/cronjob/index.vue'),
            meta: {},
            children: [
                {
                    path: 'cronjob',
                    name: 'CronjobItem',
                    component: () => import('@/views/cronjob/cronjob/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/cronjobs',
                        title: 'menu.cronjob',
                        permission: 'cronjob_view',
                    },
                },
                {
                    path: 'cronjob/operate',
                    name: 'CronjobOperate',
                    component: () => import('@/views/cronjob/cronjob/operate/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/cronjobs',
                        ignoreTab: true,
                        permission: 'cronjob_view',
                    },
                },
                {
                    path: 'library',
                    name: 'Library',
                    component: () => import('@/views/cronjob/library/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/cronjobs',
                        title: 'cronjob.library.library',
                        permission: 'cronjob_view',
                    },
                },
            ],
        },
    ],
};

export default cronRouter;
