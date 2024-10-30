import router from '@/routers/router';
import NProgress from '@/config/nprogress';
import { GlobalStore } from '@/store';
import { AxiosCanceler } from '@/api/helper/axios-cancel';

const axiosCanceler = new AxiosCanceler();

let isRedirecting = false;

router.beforeEach((to, from, next) => {
    NProgress.start();
    axiosCanceler.removeAllPending();
    const globalStore = GlobalStore();

    if (to.name !== 'entrance' && !globalStore.isLogin) {
        next({
            name: 'entrance',
            params: to.params,
        });
        NProgress.done();
        return;
    }
    if (to.name === 'entrance' && globalStore.isLogin) {
        if (to.params.code === globalStore.entrance) {
            next({
                name: 'home',
            });
            NProgress.done();
            return;
        }
        next({ name: '404' });
        NProgress.done();
        return;
    }

    const activeMenuKey = 'cachedRoute' + (to.meta.activeMenu || '');
    const cachedRoute = localStorage.getItem(activeMenuKey);

    if (
        to.meta.activeMenu &&
        to.meta.activeMenu != from.meta.activeMenu &&
        cachedRoute &&
        cachedRoute !== to.path &&
        !isRedirecting
    ) {
        isRedirecting = true;
        next(cachedRoute);
        NProgress.done();
        return;
    }

    if (!to.matched.some((record) => record.meta.requiresAuth)) return next();

    return next();
});

router.afterEach((to) => {
    if (to.meta.activeMenu && !isRedirecting) {
        localStorage.setItem('cachedRoute' + to.meta.activeMenu, to.path);
    }
    isRedirecting = false;
    NProgress.done();
});

export default router;
