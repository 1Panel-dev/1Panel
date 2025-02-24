import { Layout } from '@/routers/constant';

const databaseRouter = {
    sort: 5,
    path: '/databases',
    component: Layout,
    redirect: '/databases/mysql',
    meta: {
        icon: 'p-database',
        title: 'menu.database',
    },
    children: [
        {
            path: '/databases',
            name: 'Databases',
            redirect: '/databases/mysql',
            component: () => import('@/views/database/index.vue'),
            meta: {},
            children: [
                {
                    path: 'mysql',
                    name: 'MySQL',
                    component: () => import('@/views/database/mysql/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'mysql/setting/:type/:database',
                    name: 'MySQL-Setting',
                    component: () => import('@/views/database/mysql/setting/index.vue'),
                    props: true,
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'mysql/remote',
                    name: 'MySQL-Remote',
                    component: () => import('@/views/database/mysql/remote/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'postgresql',
                    name: 'PostgreSQL',
                    component: () => import('@/views/database/postgresql/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'postgresql/remote',
                    name: 'PostgreSQL-Remote',
                    component: () => import('@/views/database/postgresql/remote/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'postgresql/setting/:type/:database',
                    name: 'PostgreSQL-Setting',
                    component: () => import('@/views/database/postgresql/setting/index.vue'),
                    props: true,
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'redis',
                    name: 'Redis',
                    component: () => import('@/views/database/redis/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'redis/remote',
                    name: 'Redis-Remote',
                    component: () => import('@/views/database/redis/remote/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        requiresAuth: false,
                    },
                },
            ],
        },
    ],
};

export default databaseRouter;
