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
                icon: 'p-file-menu',
                title: 'menu.files',
                permission: 'host_file_view',
            },
        },
        {
            path: '/hosts/monitor/monitor',
            name: 'Monitorx',
            component: () => import('@/views/host/monitor/monitor/index.vue'),
            meta: {
                icon: 'p-system-monitor-menu',
                title: 'menu.monitor',
                permission: 'host_monitor_view',
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
                permission: 'host_monitor_view',
            },
        },
        {
            path: '/hosts/firewall/port',
            name: 'FirewallPort',
            component: () => import('@/views/host/firewall/port/index.vue'),
            meta: {
                icon: 'p-firewalld-menu',
                activeMenu: '/hosts/firewall/port',
                title: 'menu.firewall',
                detail: 'firewall.portRule',
                permission: 'host_firewall_view',
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
                permission: 'host_firewall_view',
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
                permission: 'host_firewall_view',
            },
        },
        {
            path: '/hosts/firewall/advance',
            name: 'FirewallAdvance',
            component: () => import('@/views/host/firewall/advance/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/hosts/firewall/port',
                parent: 'menu.firewall',
                title: 'firewall.advancedControl',
                permission: 'host_firewall_view',
            },
        },
        {
            path: '/hosts/disk',
            name: 'Disk',
            props: true,
            component: () => import('@/views/host/disk-management/disk/index.vue'),
            meta: {
                icon: 'p-disk-menu',
                title: 'menu.disk',
                permission: 'host_disk_view',
            },
        },
        {
            path: '/hosts/process/process',
            name: 'Process',
            component: () => import('@/views/host/process/process/index.vue'),
            meta: {
                icon: 'p-process-menu',
                title: 'menu.processManage',
                detail: 'menu.process',
                activeMenu: '/hosts/process/process',
                keepAlive: true,
                permission: 'host_process_view',
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
                permission: 'host_process_view',
            },
        },
        {
            path: '/hosts/ssh/ssh',
            name: 'SSH',
            component: () => import('@/views/host/ssh/ssh/index.vue'),
            meta: {
                icon: 'p-ssh-menu',
                title: 'menu.ssh',
                detail: 'menu.config',
                activeMenu: '/hosts/ssh/ssh',
                keepAlive: true,
                permission: 'host_ssh_view',
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
                permission: 'host_ssh_view',
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
                permission: 'host_ssh_view',
            },
        },
    ],
};

export default hostRouter;
