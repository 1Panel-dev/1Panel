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
            component: () => import('@/views/website/website/index.vue'),
            meta: {
                title: 'menu.website',
                requiresAuth: false,
            },
        },
        {
            path: '/websites/:id/config/:tab',
            name: 'WebsiteConfig',
            component: () => import('@/views/website/website/config/index.vue'),
            hidden: true,
            props: true,
            meta: {
                activeMenu: '/websites',
                requiresAuth: false,
            },
        },
        {
            path: '/websites/ssl',
            name: 'SSL',
            component: () => import('@/views/website/ssl/index.vue'),
            meta: {
                title: 'menu.ssl',
                requiresAuth: false,
            },
        },
        {
            path: '/websites/runtimes/php',
            name: 'PHP',
            component: () => import('@/views/website/runtime/php/index.vue'),
            meta: {
                title: 'menu.runtime',
                requiresAuth: false,
            },
        },
        {
            path: '/websites/runtimes/node',
            name: 'Node',
            hidden: true,
            component: () => import('@/views/website/runtime/node/index.vue'),
            meta: {
                activeMenu: '/websites/runtimes/php',
                requiresAuth: false,
            },
        },
        {
            path: '/websites/runtimes/java',
            name: 'java',
            hidden: true,
            component: () => import('@/views/website/runtime/java/index.vue'),
            meta: {
                activeMenu: '/websites/runtimes/java',
                requiresAuth: false,
            },
        },
        {
            path: '/websites/runtimes/go',
            name: 'go',
            hidden: true,
            component: () => import('@/views/website/runtime/go/index.vue'),
            meta: {
                activeMenu: '/websites/runtimes/go',
                requiresAuth: false,
            },
        },
    ],
};

export default webSiteRouter;
