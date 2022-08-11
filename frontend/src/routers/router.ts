import { createRouter, createWebHashHistory, RouteRecordRaw } from 'vue-router';
import { Layout } from '@/routers/constant';

const modules = import.meta.globEager('./modules/*.ts');

export const routerArray: RouteRecordRaw[] = [];

export const rolesRoutes = [
    ...Object.keys(modules)
        .map((key) => modules[key].default)
        .sort((r1, r2) => {
            r1.sort ??= Number.MAX_VALUE;
            r2.sort ??= Number.MAX_VALUE;
            return r1.sort - r2.sort;
        }),
];

rolesRoutes.forEach((item) => {
    const menu = item as RouteRecordRaw;
    routerArray.push(menu);
});
export const routes: RouteRecordRaw[] = [
    {
        path: '/',
        component: Layout,
        redirect: '/home/index',
        meta: {
            keepAlive: true,
            requiresAuth: true,
            title: 'menu.home',
            key: 'home',
            icon: 'home-filled',
        },
        children: [
            {
                path: '/home/index',
                name: 'home',
                component: () => import('@/views/home/index.vue'),
            },
        ],
    },
    {
        path: '/login',
        name: 'login',
        component: () => import('@/views/login/index.vue'),
        meta: {
            requiresAuth: false,
            key: 'login',
        },
    },
    ...routerArray,
    {
        path: '/:pathMatch(.*)',
        redirect: { name: '404' },
    },
];
const router = createRouter({
    history: createWebHashHistory(),
    routes: routes as RouteRecordRaw[],
    strict: false,
    scrollBehavior: () => ({ left: 0, top: 0 }),
});

export default router;
