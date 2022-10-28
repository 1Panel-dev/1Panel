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
            },
        },
        {
            path: '/websites/projects/config',
            name: 'WebsiteConfig',
            component: () => import('@/views/website/project/config/index.vue'),
            hidden: true,
            meta: {
                activeMenu: '/websites',
            },
        },
        {
            path: '/websites/config',
            name: 'Config',
            component: () => import('@/views/website/config/index.vue'),
            meta: {
                title: 'menu.config',
            },
        },
        {
            path: '/websites/firewall',
            name: 'Firewall',
            component: () => import('@/views/website/project/index.vue'),
            meta: {
                title: 'menu.firewall',
            },
        },
    ],
};

export default webSiteRouter;
