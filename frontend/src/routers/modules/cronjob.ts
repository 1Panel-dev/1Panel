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
                        requiresAuth: false,
                        title: 'menu.cronjob',
                    },
                },
                {
                    path: 'cronjob/operate',
                    name: 'CronjobOperate',
                    component: () => import('@/views/cronjob/cronjob/operate/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/cronjobs',
                        requiresAuth: false,
                        title: 'menu.cronjob',
                        detail: 'commons.table.operate',
                    },
                },
                {
                    path: 'library',
                    name: 'Library',
                    component: () => import('@/views/cronjob/library/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/cronjobs',
                        requiresAuth: false,
                        title: 'cronjob.library.library',
                    },
                },
            ],
        },
    ],
};

export default cronRouter;
