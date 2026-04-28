import router from '@/routers/router';
import NProgress from '@/config/nprogress';
import { GlobalStore } from '@/store';
import { AxiosCanceler } from '@/api/helper/axios-cancel';
import { hasRouteAccess } from '@/utils/rbac';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';

const axiosCanceler = new AxiosCanceler();

let isRedirecting = false;
let loginStatusLoading: Promise<boolean> | null = null;
let xpackEELoading: Promise<boolean> | null = null;
let licenseStatusLoading: Promise<boolean> | null = null;
let loginStatusLoaded = false;
let xpackEEStatusLoaded = false;
let licenseStatusLoaded = false;
const xpackEELicenseCheckWhiteList = ['XpackEELicenseRequired', 'entrance', 'login', 'Expired'];

const clearLoginStatus = () => {
    const globalStore = GlobalStore();
    globalStore.setLogStatus(false);
    globalStore.clearAuthInfo();
    globalStore.isXpackEELicensed = false;
    loginStatusLoaded = false;
    licenseStatusLoaded = false;
};

const loadLoginStatus = async () => {
    const globalStore = GlobalStore();
    if (!globalStore.isLogin) {
        return false;
    }
    if (loginStatusLoaded) {
        return true;
    }
    if (!loginStatusLoading) {
        loginStatusLoading = fetch('/api/v2/core/auth/current', {
            credentials: 'include',
            headers: {
                CurrentNode: encodeURIComponent(globalStore.currentNode),
            },
        })
            .then((res) => res.json())
            .then((res) => {
                const loggedIn = res?.code === 200;
                if (!loggedIn) {
                    clearLoginStatus();
                    return false;
                }
                loginStatusLoaded = true;
                return true;
            })
            .catch(() => {
                clearLoginStatus();
                return false;
            })
            .finally(() => {
                loginStatusLoading = null;
            });
    }
    return loginStatusLoading;
};

const loadXpackEEStatus = async () => {
    if (xpackEEStatusLoaded) {
        return GlobalStore().isXpackEE;
    }
    if (!xpackEELoading) {
        xpackEELoading = fetch('/api/v2/core/auth/setting', {
            credentials: 'include',
        })
            .then((res) => res.json())
            .then((res) => {
                const globalStore = GlobalStore();
                globalStore.isXpackEE = !!res?.data?.isXpackEE;
                xpackEEStatusLoaded = true;
                return globalStore.isXpackEE;
            })
            .catch(() => GlobalStore().isXpackEE)
            .finally(() => {
                xpackEELoading = null;
            });
    }
    return xpackEELoading;
};

const loadXpackEELicenseStatus = async () => {
    if (licenseStatusLoaded) {
        return GlobalStore().isXpackEELicensed;
    }
    if (!licenseStatusLoading) {
        licenseStatusLoading = fetch('/api/v2/core/xpackee/licenses/status', {
            credentials: 'include',
            headers: {
                CurrentNode: encodeURIComponent(GlobalStore().currentNode),
            },
        })
            .then((res) => res.json())
            .then((res) => {
                const globalStore = GlobalStore();
                const licensed = res?.data?.status === 'Bound';
                globalStore.isXpackEELicensed = licensed;
                licenseStatusLoaded = true;
                return licensed;
            })
            .catch(() => false)
            .finally(() => {
                licenseStatusLoading = null;
            });
    }
    return licenseStatusLoading;
};

router.beforeEach(async (to, from, next) => {
    NProgress.start();
    axiosCanceler.removeAllPending();
    const globalStore = GlobalStore();
    if (globalStore.isLogin) {
        const loggedIn = await loadLoginStatus();
        if (!loggedIn) {
            if (to.name !== 'entrance') {
                next({
                    name: 'entrance',
                    params: to.params,
                });
                NProgress.done();
                return;
            }
            return next();
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
    if (globalStore.isLogin && to.name !== 'login') {
        await loadXpackEEStatus();
    }
    if (globalStore.isLogin && globalStore.isXpackEE && !xpackEELicenseCheckWhiteList.includes(String(to.name))) {
        const licensed = await loadXpackEELicenseStatus();
        if (!licensed) {
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
        const licensed = await loadXpackEELicenseStatus();
        if (licensed) {
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
