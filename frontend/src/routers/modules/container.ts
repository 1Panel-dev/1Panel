import { Layout } from '@/routers/constant';

const containerRouter = {
    sort: 5,
    path: '/containers',
    component: Layout,
    redirect: '/containers',
    meta: {
        icon: 'p-docker',
        title: 'menu.container',
    },
    children: [
        {
            path: ':filters?',
            name: 'Container',
            component: () => import('@/views/container/container/index.vue'),
            props: true,
            hidden: true,
            meta: {
                activeMenu: '/containers',
            },
        },
        {
            path: 'image',
            name: 'Image',
            component: () => import('@/views/container/image/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/containers',
            },
        },
        {
            path: 'network',
            name: 'Network',
            component: () => import('@/views/container/network/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/containers',
            },
        },
        {
            path: 'volume',
            name: 'Volume',
            component: () => import('@/views/container/volume/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/containers',
            },
        },
        {
            path: 'repo',
            name: 'Repo',
            component: () => import('@/views/container/repo/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/containers',
            },
        },
        {
            path: 'compose',
            name: 'Compose',
            component: () => import('@/views/container/compose/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/containers',
            },
        },
        {
            path: 'template',
            name: 'ComposeTemplate',
            component: () => import('@/views/container/template/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/containers',
            },
        },
        {
            path: 'setting',
            name: 'ContainerSetting',
            component: () => import('@/views/container/setting/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/containers',
            },
        },
    ],
};

export default containerRouter;
