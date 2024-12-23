import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';
import { Layout } from '@/routers/constant';

let modules: Record<string, RouteRecordRaw> = import.meta.glob('./modules/*.ts', { eager: true });
const xpackModules: Record<string, RouteRecordRaw> = import.meta.glob('../xpack/routers/*.ts', { eager: true });
modules = { ...modules, ...xpackModules };

const homeRouter: RouteRecordRaw = {
    path: '/',
    component: Layout,
    redirect: '/',
    meta: {
        keepAlive: true,
        title: 'menu.home',
        icon: 'p-home',
    },
    children: [
        {
            path: '/',
            name: 'home',
            component: () => import('@/views/home/index.vue'),
            meta: {
                requiresAuth: true,
            },
        },
    ],
};

export const routerArray: RouteRecordRaw[] = [];

export const rolesRoutes = [
    ...Object.keys(modules)
        .map((key) => modules[key]['default'])
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

export const menuList: RouteRecordRaw[] = [];
rolesRoutes.forEach((item) => {
    let menuItem = JSON.parse(JSON.stringify(item));
    let menuChildren: RouteRecordRaw[] = [];
    if (menuItem.children == undefined) {
        return;
    }
    menuItem.children.forEach((child: any) => {
        if (child.hidden == undefined || child.hidden == false) {
            menuChildren.push(child);
        }
    });
    menuItem.children = menuChildren as RouteRecordRaw[];
    menuList.push(menuItem);
});
menuList.unshift(homeRouter);

export const routes: RouteRecordRaw[] = [
    homeRouter,
    {
        path: '/login',
        name: 'login',
        props: true,
        component: () => import('@/views/login/index.vue'),
        meta: {
            requiresAuth: false,
            key: 'login',
        },
    },
    {
        path: '/:code?',
        name: 'entrance',
        component: () => import('@/views/login/entrance/index.vue'),
        props: true,
    },
    ...routerArray,
    {
        path: '/:pathMatch(.*)',
        redirect: { name: '404' },
    },
];
const router = createRouter({
    history: createWebHistory('/'),
    routes: routes as RouteRecordRaw[],
    strict: false,
    scrollBehavior: () => ({ left: 0, top: 0 }),
});

export default router;
