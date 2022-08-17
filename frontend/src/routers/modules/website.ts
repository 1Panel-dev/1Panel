import { Layout } from '@/routers/constant';

const webSiteRouter = {
    sort: 3,
    path: '/websites',
    component: Layout,
    redirect: '/websites',
    meta: {
        icon: 'p-website',
        title: 'menu.website',
    },
    children: [
        {
            path: '/websites',
            name: 'Website',
            component: () => import('@/views/website/project/index.vue'),
            meta: {
                title: 'menu.project',
                keepAlive: true,
            },
        },
        {
            path: '/websites/config',
            name: 'Config',
            component: () => import('@/views/website/config/index.vue'),
            meta: {
                title: 'menu.config',
                keepAlive: true,
            },
        },
        {
            path: '/websites/firewall',
            name: 'Firewall',
            component: () => import('@/views/website/project/index.vue'),
            meta: {
                title: 'menu.firewall',
                keepAlive: true,
            },
        },
    ],
};

export default webSiteRouter;
