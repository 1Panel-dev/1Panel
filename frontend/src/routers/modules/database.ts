import { Layout } from '@/routers/constant';

const databaseRouter = {
    sort: 5,
    path: '/databases',
    name: 'Database-Menu',
    component: Layout,
    redirect: '/databases/mysql',
    meta: {
        icon: 'p-database',
        title: 'menu.database',
        permission: 'database_view',
    },
    children: [
        {
            path: '/databases',
            name: 'Database',
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
                        parent: 'menu.database',
                        title: 'MySQL',
                        permission: 'database_view',
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
                        ignoreTab: true,
                        permission: 'database_view',
                    },
                },
                {
                    path: 'mysql/remote',
                    name: 'MySQL-Remote',
                    component: () => import('@/views/database/mysql/remote/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        parent: 'menu.database',
                        title: 'MySQL',
                        detail: 'database.remote',
                        permission: 'database_view',
                    },
                },
                {
                    path: 'postgresql',
                    name: 'PostgreSQL',
                    component: () => import('@/views/database/postgresql/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        parent: 'menu.database',
                        title: 'PostgreSQL',
                        permission: 'database_view',
                    },
                },
                {
                    path: 'postgresql/remote',
                    name: 'PostgreSQL-Remote',
                    component: () => import('@/views/database/postgresql/remote/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        parent: 'menu.database',
                        title: 'PostgreSQL',
                        detail: 'database.remote',
                        permission: 'database_view',
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
                        ignoreTab: true,
                        permission: 'database_view',
                    },
                },
                {
                    path: 'redis',
                    name: 'Redis',
                    component: () => import('@/views/database/redis/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        parent: 'menu.database',
                        title: 'Redis',
                        permission: 'database_view',
                    },
                },
                {
                    path: 'redis/remote',
                    name: 'Redis-Remote',
                    component: () => import('@/views/database/redis/remote/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        parent: 'menu.database',
                        title: 'Redis',
                        detail: 'database.remote',
                        permission: 'database_view',
                    },
                },
                {
                    path: 'mongodb',
                    name: 'MongoDB',
                    component: () => import('@/views/database/mongodb/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        requiresAuth: false,
                        parent: 'menu.database',
                        title: 'MongoDB',
                        permission: 'database_view',
                    },
                },
                {
                    path: 'mongodb/remote',
                    name: 'MongoDB-Remote',
                    component: () => import('@/views/database/mongodb/remote/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/databases',
                        requiresAuth: false,
                        parent: 'menu.database',
                        title: 'MongoDB',
                        detail: 'database.remote',
                        permission: 'database_view',
                    },
                },
            ],
        },
    ],
};

export default databaseRouter;
