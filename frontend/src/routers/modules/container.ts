import { Layout } from '@/routers/constant';

const containerRouter = {
    sort: 6,
    path: '/containers',
    name: 'Container-Menu',
    component: Layout,
    redirect: '/containers/container',
    meta: {
        icon: 'p-docker1',
        title: 'menu.container',
    },
    children: [
        {
            path: '/containers',
            name: 'Container',
            redirect: '/containers/dashboard',
            component: () => import('@/views/container/index.vue'),
            meta: {},
            children: [
                {
                    path: 'dashboard',
                    name: 'ContainerDashboard',
                    component: () => import('@/views/container/dashboard/index.vue'),
                    props: true,
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        requiresAuth: false,
                        parent: 'menu.container',
                        title: 'menu.home',
                    },
                },
                {
                    path: 'container',
                    name: 'ContainerItem',
                    component: () => import('@/views/container/container/index.vue'),
                    props: true,
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        requiresAuth: false,
                        parent: 'menu.container',
                        title: 'menu.container',
                    },
                },
                {
                    path: 'container/operate',
                    name: 'ContainerCreate',
                    component: () => import('@/views/container/container/operate/index.vue'),
                    props: true,
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        requiresAuth: false,
                        ignoreTab: true,
                    },
                },
                {
                    path: 'compose/detail',
                    name: 'ComposeDetail',
                    component: () => import('@/views/container/compose/detail/index.vue'),
                    props: true,
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        requiresAuth: false,
                        ignoreTab: true,
                    },
                },
                {
                    path: 'image',
                    name: 'Image',
                    component: () => import('@/views/container/image/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        requiresAuth: false,
                        parent: 'menu.container',
                        title: 'container.image',
                    },
                },
                {
                    path: 'network',
                    name: 'Network',
                    component: () => import('@/views/container/network/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        requiresAuth: false,
                        parent: 'menu.container',
                        title: 'container.network',
                    },
                },
                {
                    path: 'volume',
                    name: 'Volume',
                    component: () => import('@/views/container/volume/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        requiresAuth: false,
                        parent: 'menu.container',
                        title: 'container.volume',
                    },
                },
                {
                    path: 'repo',
                    name: 'Repo',
                    component: () => import('@/views/container/repo/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        requiresAuth: false,
                        parent: 'menu.container',
                        title: 'container.repo',
                    },
                },
                {
                    path: 'compose',
                    name: 'Compose',
                    component: () => import('@/views/container/compose/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        requiresAuth: false,
                        parent: 'menu.container',
                        title: 'container.compose',
                    },
                },
                {
                    path: 'template',
                    name: 'ComposeTemplate',
                    component: () => import('@/views/container/template/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        requiresAuth: false,
                        parent: 'menu.container',
                        title: 'container.composeTemplate',
                    },
                },
                {
                    path: 'setting',
                    name: 'ContainerSetting',
                    component: () => import('@/views/container/setting/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        requiresAuth: false,
                        parent: 'menu.container',
                        title: 'container.setting',
                    },
                },
            ],
        },
    ],
};

export default containerRouter;
