import { Layout } from '@/routers/constant';

const toolboxRouter = {
    sort: 7,
    path: '/toolbox',
    component: Layout,
    redirect: '/toolbox/supervisor',
    meta: {
        title: 'menu.toolbox',
        icon: 'p-toolbox',
    },
    children: [
        {
            path: '/toolbox',
            name: 'Toolbox',
            redirect: '/toolbox/supervisor',
            component: () => import('@/views/toolbox/index.vue'),
            meta: {},
            children: [
                {
                    path: 'supervisor',
                    name: 'Supervisor',
                    component: () => import('@/views/toolbox/supervisor/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/toolbox',
                        requiresAuth: false,
                    },
                },
                {
                    path: 'fail2ban',
                    name: 'Fail2ban',
                    component: () => import('@/views/toolbox/fail2ban/index.vue'),
                    hidden: true,
                    meta: {
                        activeMenu: '/toolbox',
                        requiresAuth: false,
                    },
                },
            ],
        },
    ],
};

export default toolboxRouter;
