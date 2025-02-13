import { Layout } from '@/routers/constant';

const hostRouter = {
    sort: 7,
    path: '/hosts',
    component: Layout,
    redirect: '/hosts/security',
    meta: {
        icon: 'p-host',
        title: 'menu.system',
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
            path: '/hosts/monitor/monitor',
            name: 'Monitorx',
            component: () => import('@/views/host/monitor/monitor/index.vue'),
            meta: {
                title: 'menu.monitor',
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/monitor/setting',
            name: 'HostMonitorSetting',
            component: () => import('@/views/host/monitor/setting/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/hosts/monitor/monitor',
                title: 'menu.monitor',
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/terminal',
            name: 'Terminal',
            props: true,
            component: () => import('@/views/host/terminal/index.vue'),
            meta: {
                title: 'menu.terminal',
                keepAlive: true,
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
            path: '/hosts/firewall/forward',
            name: 'FirewallForward',
            component: () => import('@/views/host/firewall/forward/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/hosts/firewall/port',
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/firewall/ip',
            name: 'FirewallIP',
            component: () => import('@/views/host/firewall/ip/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/hosts/firewall/port',
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/process/process',
            name: 'Process',
            component: () => import('@/views/host/process/process/index.vue'),
            meta: {
                title: 'menu.processManage',
                activeMenu: '/hosts/process/process',
                keepAlive: true,
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/process/network',
            name: 'ProcessNetwork',
            hidden: true,
            component: () => import('@/views/host/process/network/index.vue'),
            meta: {
                activeMenu: '/hosts/process/process',
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/ssh/ssh',
            name: 'SSH',
            component: () => import('@/views/host/ssh/ssh/index.vue'),
            meta: {
                title: 'menu.ssh',
                activeMenu: '/hosts/ssh/ssh',
                keepAlive: true,
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/ssh/log',
            name: 'SSHLog',
            component: () => import('@/views/host/ssh/log/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/hosts/ssh/ssh',
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/ssh/session',
            name: 'SSHSession',
            component: () => import('@/views/host/ssh/session/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/hosts/ssh/ssh',
                requiresAuth: false,
            },
        },
    ],
};

export default hostRouter;
