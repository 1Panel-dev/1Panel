import { Layout } from '@/routers/constant';

const hostRouter = {
    sort: 7,
    path: '/hosts',
    name: 'System-Menu',
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
                detail: 'commons.button.set',
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/firewall/port',
            name: 'FirewallPort',
            component: () => import('@/views/host/firewall/port/index.vue'),
            meta: {
                title: 'menu.firewall',
                detail: 'firewall.portRule',
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
                parent: 'menu.firewall',
                title: 'firewall.forwardRule',
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
                parent: 'menu.firewall',
                title: 'firewall.ipRule',
                requiresAuth: false,
            },
        },
        {
            path: '/hosts/process/process',
            name: 'Process',
            component: () => import('@/views/host/process/process/index.vue'),
            meta: {
                title: 'menu.processManage',
                detail: 'menu.process',
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
                parent: 'menu.processManage',
                title: 'menu.network',
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
                detail: 'menu.config',
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
                parent: 'menu.ssh',
                title: 'ssh.loginLogs',
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
                parent: 'menu.ssh',
                title: 'ssh.session',
                activeMenu: '/hosts/ssh/ssh',
                requiresAuth: false,
            },
        },
    ],
};

export default hostRouter;
