import router from '@/routers/router';
import NProgress from '@/config/nprogress';
import { GlobalStore } from '@/store';
import { AxiosCanceler } from '@/api/helper/axios-cancel';
import { hasRouteAccess } from '@/utils/rbac';
import { loadProductProFromDB } from '@/utils/xpack';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';

const axiosCanceler = new AxiosCanceler();

let isRedirecting = false;
const xpackEELicenseCheckWhiteList = ['XpackEELicenseRequired', 'entrance', 'login', 'Expired'];

const clearLicenseStatus = () => {
    const globalStore = GlobalStore();
    globalStore.isXpackEELicensed = false;
    globalStore.isXpackEELicenseLoaded = false;
};

const clearLoginStatus = () => {
    const globalStore = GlobalStore();
    globalStore.setLogStatus(false);
    globalStore.clearAuthInfo();
    clearLicenseStatus();
};

router.beforeEach(async (to, from, next) => {
    NProgress.start();
    axiosCanceler.removeAllPending();
    const globalStore = GlobalStore();
    if (!globalStore.isLogin) {
        clearLoginStatus();
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
    if (globalStore.isLogin && globalStore.isXpackEE && !xpackEELicenseCheckWhiteList.includes(String(to.name))) {
        if (!globalStore.isXpackEELicenseLoaded) {
            await loadProductProFromDB();
        }
        if (!globalStore.isXpackEELicensed) {
            next({ name: 'XpackEELicenseRequired', query: { code: String(to.params.code || '') } });
            NProgress.done();
            return;
        }
    }
    if (to.name === 'XpackEELicenseRequired') {
        if (!globalStore.isLogin) {
            next({
                name: 'entrance',
                params: to.params,
            });
            NProgress.done();
            return;
        }
        if (!globalStore.isXpackEE || globalStore.isXpackEELicensed) {
            next({ name: 'home' });
            NProgress.done();
            return;
        }
        return next();
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

    if (!hasRouteAccess(to)) {
        MsgError(i18n.global.t('commons.res.forbidden'));
        next(false);
        NProgress.done();
        return;
    }
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
