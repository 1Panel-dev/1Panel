import { Layout } from '@/routers/constant';
import i18n from '@/lang';

const terminalRouter = {
    sort: 2,
    path: '/terminals',
    component: Layout,
    redirect: '/terminal',
    meta: {
        title: i18n.global.t('menu.terminal'),
        icon: 'monitor',
    },
    children: [
        {
            path: '/terminals/terminal',
            name: 'Terminal',
            component: () => import('@/views/terminal/index.vue'),
            meta: {
                requiresAuth: true,
                key: 'Terminal',
                title: i18n.global.t('terminal.conn'),
                icon: 'connection',
                activeMenu: '/terminals',
            },
        },
        {
            path: '/terminals/host',
            name: 'Host',
            component: () => import('@/views/terminal/host/index.vue'),
            meta: {
                requiresAuth: true,
                key: 'Host',
                title: i18n.global.t('terminal.hostList'),
                icon: 'platform',
                activeMenu: '/terminals',
            },
        },
        {
            path: '/terminals/command',
            name: 'Command',
            component: () => import('@/views/terminal/command/index.vue'),
            meta: {
                requiresAuth: true,
                key: 'Command',
                title: i18n.global.t('terminal.quickCmd'),
                icon: 'reading',
                activeMenu: '/terminals',
            },
        },
    ],
};

export default terminalRouter;
