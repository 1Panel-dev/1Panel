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
        permission: 'container_view',
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
                        parent: 'menu.container',
                        title: 'menu.home',
                        permission: 'container_view',
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
                        parent: 'menu.container',
                        title: 'menu.container',
                        permission: 'container_view',
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
                        ignoreTab: true,
                        permission: 'container_view',
                    },
                },
                {
                    path: 'image',
                    name: 'Image',
                    component: () => import('@/views/container/image/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        parent: 'menu.container',
                        title: 'container.image',
                        permission: 'container_view',
                    },
                },
                {
                    path: 'network',
                    name: 'Network',
                    component: () => import('@/views/container/network/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        parent: 'menu.container',
                        title: 'container.network',
                        permission: 'container_view',
                    },
                },
                {
                    path: 'volume',
                    name: 'Volume',
                    component: () => import('@/views/container/volume/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        parent: 'menu.container',
                        title: 'container.volume',
                        permission: 'container_view',
                    },
                },
                {
                    path: 'repo',
                    name: 'Repo',
                    component: () => import('@/views/container/repo/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        parent: 'menu.container',
                        title: 'container.repo',
                        permission: 'container_view',
                    },
                },
                {
                    path: 'compose',
                    name: 'Compose',
                    component: () => import('@/views/container/compose/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        parent: 'menu.container',
                        title: 'container.compose',
                        permission: 'container_view',
                    },
                },
                {
                    path: 'template',
                    name: 'ComposeTemplate',
                    component: () => import('@/views/container/template/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        parent: 'menu.container',
                        title: 'container.composeTemplate',
                        permission: 'container_view',
                    },
                },
                {
                    path: 'setting',
                    name: 'ContainerSetting',
                    component: () => import('@/views/container/setting/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/containers',
                        parent: 'menu.container',
                        title: 'container.setting',
                        permission: 'container_view',
                    },
                },
            ],
        },
    ],
};

export default containerRouter;
