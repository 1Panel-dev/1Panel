import router from '@/routers/router';
import NProgress from '@/config/nprogress';
import { GlobalStore } from '@/store';
import { AxiosCanceler } from '@/api/helper/axios-cancel';
import { hasPermission } from '@/utils/rbac';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';

const axiosCanceler = new AxiosCanceler();

let isRedirecting = false;

router.beforeEach((to, from, next) => {
    NProgress.start();
    axiosCanceler.removeAllPending();
    const globalStore = GlobalStore();
    if (globalStore.isXpackEE && !globalStore.isAdmin) {
        if (xpackEEJumper(to, next)) {
            return;
        }
    }
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

    if (to.path === '/apps/all' && to.query.install != undefined) {
        return next();
    }
    const activeMenuKey = 'cachedRoute' + (to.meta.activeMenu || '');
    if (to.query.uncached != undefined) {
        const query = { ...to.query };
        delete query.uncached;
        localStorage.removeItem(activeMenuKey);
        return next({ path: to.path, query });
    }

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

    const requiredPermission = [...to.matched].reverse().find((record) => record.meta?.permission)?.meta?.permission as
        | string
        | undefined;
    if (requiredPermission && !hasPermission(requiredPermission)) {
        MsgError(i18n.global.t('commons.res.forbidden'));
        next(false);
        NProgress.done();
        return;
    }

    if (!to.matched.some((record) => record.meta.requiresAuth)) return next();

    return next();
});

router.afterEach((to) => {
    if (to.meta.activeMenu && !to.meta.ignoreTab && !isRedirecting) {
        let notMathParam = true;
        if (to.matched.some((record) => record.path.includes(':'))) {
            notMathParam = false;
        }
        if (notMathParam) {
            if (to.meta.activeMenu === '/cronjobs' && to.path === '/cronjobs/cronjob/operate') {
                localStorage.setItem('cachedRoute' + to.meta.activeMenu, '/cronjobs/cronjob');
            } else if (to.meta.activeMenu === '/containers' && to.path === '/containers/container/operate') {
                localStorage.setItem('cachedRoute' + to.meta.activeMenu, '/containers/container');
            } else if (to.meta.activeMenu === '/toolbox' && to.path === '/toolbox/clam/setting') {
                localStorage.setItem('cachedRoute' + to.meta.activeMenu, '/toolbox/clam');
            } else {
                localStorage.setItem('cachedRoute' + to.meta.activeMenu, to.path);
            }
        }
    }

    isRedirecting = false;
    NProgress.done();
});

export default router;

const xpackEEJumper = (to: any, next: any) => {
    switch (to.name) {
        case 'Panel':
        case 'Safe':
        case 'Alert':
            if (hasPermission('setting_view')) {
                return false;
            }
            MsgError(i18n.global.t('commons.res.forbidden'));
            next(false);
            NProgress.done();
            return true;
        case 'License':
            if (hasPermission('setting_view')) {
                return false;
            }
            MsgError(i18n.global.t('commons.res.forbidden'));
            next(false);
            NProgress.done();
            return true;
        case 'Node':
        case 'SimpleNode':
        case 'NodeAppUpgrade':
            if (hasPermission('node_view')) {
                return false;
            }
            MsgError(i18n.global.t('commons.res.forbidden'));
            next(false);
            NProgress.done();
            return true;
        case 'UserXpackEEUser':
            if (hasPermission('setting_view')) {
                return false;
            }
            MsgError(i18n.global.t('commons.res.forbidden'));
            next(false);
            NProgress.done();
            return true;
        case 'XpackEERole':
            if (hasPermission('setting_view')) {
                return false;
            }
            MsgError(i18n.global.t('commons.res.forbidden'));
            next(false);
            NProgress.done();
            return true;
    }
};
