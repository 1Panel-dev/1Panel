import { Layout } from '@/routers/constant';

const hostRouter = {
    sort: 7,
    path: '/hosts',
    component: Layout,
    redirect: '/hosts/security',
    meta: {
        icon: 'p-host',
        title: 'menu.host',
    },
    children: [
        {
            path: '/hosts/files',
            name: 'File',
            props: true,
            component: () => import('@/views/host/file-management/index.vue'),
            meta: {
                title: 'menu.files',
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/monitor',
            name: 'Monitorx',
            component: () => import('@/views/host/monitor/index.vue'),
            meta: {
                title: 'menu.monitor',
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/firewall/port',
            name: 'FirewallPort',
            component: () => import('@/views/host/firewall/port/index.vue'),
            meta: {
                title: 'menu.firewall',
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/firewall/ip',
            name: 'FirewallIP',
            component: () => import('@/views/host/firewall/ip/index.vue'),
            hidden: true,
            meta: {
                title: 'menu.toolbox',
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/terminal',
            name: 'Terminal',
            component: () => import('@/views/host/terminal/index.vue'),
            meta: {
                title: 'menu.terminal',
                keepAlive: true,
                requiresAuth: false,
            },
        },
    ],
};

export default hostRouter;
