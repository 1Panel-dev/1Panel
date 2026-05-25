import { Layout } from '@/routers/constant';
import { GlobalStore } from '@/store';

const settingPermissions = ['alert_view', 'backup_view'];

const redirectToAvailableSetting = () => {
    const globalStore = GlobalStore();
    if (globalStore.isAdmin) {
        return '/settings/panel';
    }
    if (globalStore.hasPermission('alert_view')) {
        return '/settings/alert';
    }
    if (globalStore.hasPermission('backup_view')) {
        return '/settings/backupaccount';
    }
    return '/settings/panel';
};

const settingRouter = {
    sort: 12,
    path: '/settings',
    name: 'Setting-Menu',
    component: Layout,
    redirect: redirectToAvailableSetting,
    meta: {
        title: 'menu.settings',
        icon: 'p-config',
        permission: settingPermissions,
    },
    children: [
        {
            path: '/settings',
            name: 'Setting',
            redirect: redirectToAvailableSetting,
            component: () => import('@/views/setting/index.vue'),
            meta: {
                permission: settingPermissions,
            },
            children: [
                {
                    path: 'panel',
                    name: 'Panel',
                    component: () => import('@/views/setting/panel/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.settings',
                        title: 'setting.panel',
                        activeMenu: '/settings',
                        adminOnly: true,
                    },
                },
                {
                    path: 'alert',
                    name: 'Alert',
                    component: () => import('@/views/setting/alert/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.settings',
                        title: 'xpack.alert.alertNotice',
                        activeMenu: '/settings',
                        permission: 'alert_view',
                    },
                },
                {
                    path: 'backupaccount',
                    name: 'BackupAccount',
                    component: () => import('@/views/setting/backup-account/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.settings',
                        title: 'setting.backupAccount',
                        activeMenu: '/settings',
                        permission: 'backup_view',
                    },
                },
                {
                    path: 'license',
                    name: 'License',
                    component: () => import('@/views/setting/license/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.settings',
                        title: 'setting.license',
                        activeMenu: '/settings',
                        adminOnly: true,
                    },
                },
                {
                    path: 'about',
                    name: 'About',
                    component: () => import('@/views/setting/about/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.settings',
                        title: 'setting.about',
                        activeMenu: '/settings',
                        adminOnly: true,
                    },
                },
                {
                    path: 'safe',
                    name: 'Safe',
                    component: () => import('@/views/setting/safe/index.vue'),
                    hidden: true,
                    meta: {
                        parent: 'menu.settings',
                        title: 'setting.safe',
                        activeMenu: '/settings',
                        adminOnly: true,
                    },
                },
                {
                    path: 'snapshot',
                    name: 'Snapshot',
                    hidden: true,
                    component: () => import('@/views/setting/snapshot/index.vue'),
                    meta: {
                        parent: 'menu.settings',
                        title: 'setting.snapshot',
                        activeMenu: '/settings',
                        adminOnly: true,
                    },
                },
                {
                    path: 'expired',
                    name: 'Expired',
                    hidden: true,
                    component: () => import('@/views/setting/expired.vue'),
                    meta: {
                        activeMenu: '/settings',
                        ignoreTab: true,
                    },
                },
            ],
        },
    ],
};

export default settingRouter;
