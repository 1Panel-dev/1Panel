import { panelCacheKey, writePanelCache, prunePanelCache } from '@/utils/panel-cache';
import router from '@/routers/router';
import NProgress from '@/config/nprogress';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { AxiosCanceler } from '@/api/helper/axios-cancel';
import { hasRouteAccess } from '@/utils/rbac';
import { loadProductProFromDB } from '@/utils/xpack';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
import { TerminalDockSessionStore, TerminalSessionStore } from '@/store';

const axiosCanceler = new AxiosCanceler();

const enterpriseLicenseCheckWhiteList = ['EnterpriseLicenseRequired', 'entrance', 'login', 'Expired'];
const noLoginWhiteList = ['entrance', 'login', 'file-share', '404', 'Expired'];

const clearLicenseStatus = () => {
    const { isEnterpriseLicenseLoaded, isEnterpriseLicensed } = useGlobalStore();
    isEnterpriseLicensed.value = false;
    isEnterpriseLicenseLoaded.value = false;
};

const clearLoginStatus = () => {
    const { globalStore } = useGlobalStore();
    globalStore.setLogStatus(false);
    globalStore.clearAuthInfo();
    clearLicenseStatus();
    TerminalSessionStore().closeAll();
    TerminalDockSessionStore().closeAll();
};

router.beforeEach(async (to) => {
    const { entrance, isEnterprise, isEnterpriseLicenseLoaded, isEnterpriseLicensed, isLogin, isOnRestart } =
        useGlobalStore();
    NProgress.start();
    axiosCanceler.removeAllPending();

    if (!isLogin.value) {
        clearLoginStatus();
    }
    if (!isLogin.value && !noLoginWhiteList.includes(String(to.name))) {
        NProgress.done();
        return entrance.value
            ? {
                  name: 'entrance',
                  params: { code: entrance.value },
              }
            : {
                  name: 'login',
              };
    }
    if (to.name === 'login' && !isLogin.value && entrance.value) {
        NProgress.done();
        return {
            name: 'entrance',
            params: { code: entrance.value },
        };
    }
    if (to.name === 'login' && isLogin.value) {
        NProgress.done();
        return {
            name: 'home',
        };
    }
    if (to.name === 'entrance' && isLogin.value) {
        if (to.params.code === entrance.value) {
            NProgress.done();
            return {
                name: 'home',
            };
        }
        NProgress.done();
        return { name: '404' };
    }
    if (isLogin.value && isEnterprise.value && !enterpriseLicenseCheckWhiteList.includes(String(to.name))) {
        if (!isEnterpriseLicenseLoaded.value) {
            await loadProductProFromDB();
        }
        if (!isEnterpriseLicensed.value) {
            NProgress.done();
            return { name: 'EnterpriseLicenseRequired', query: { code: String(to.params.code || '') } };
        }
    }
    if (to.name === 'EnterpriseLicenseRequired') {
        if (!isLogin.value) {
            NProgress.done();
            return {
                name: 'entrance',
                params: to.params,
            };
        }
        if (!isEnterprise.value || (isEnterpriseLicensed.value && !isOnRestart.value)) {
            NProgress.done();
            return { name: 'home' };
        }
        return true;
    }

    if (to.path === '/apps/all' && to.query.install != undefined) {
        return true;
    }
    if (to.name === 'Expired') {
        return true;
    }
    // Compatibility for old links: explicit navigation no longer needs to clear a cache.
    if (to.query.uncached != undefined) {
        const query = { ...to.query };
        delete query.uncached;
        return { path: to.path, query, hash: to.hash };
    }

    if (!hasRouteAccess(to)) {
        MsgError(i18n.global.t('commons.res.forbidden'));
        NProgress.done();
        return false;
    }
    return true;
});

router.afterEach((to, _from, failure) => {
    prunePanelCache();
    if (failure) {
        NProgress.done();
        return;
    }
    if (to.meta.activeMenu && !to.meta.ignoreTab) {
        let notMathParam = true;
        if (to.matched.some((record) => record.path.includes(':'))) {
            notMathParam = false;
        }
        if (notMathParam) {
            if (to.meta.activeMenu === '/cronjobs' && to.path === '/cronjobs/cronjob/operate') {
                writePanelCache('tabs', panelCacheKey(String(to.meta.activeMenu)), '/cronjobs/cronjob');
            } else if (to.meta.activeMenu === '/containers' && to.path === '/containers/container/operate') {
                writePanelCache('tabs', panelCacheKey(String(to.meta.activeMenu)), '/containers/container');
            } else if (to.meta.activeMenu === '/toolbox' && to.path === '/toolbox/clam/setting') {
                writePanelCache('tabs', panelCacheKey(String(to.meta.activeMenu)), '/toolbox/clam');
            } else {
                writePanelCache('tabs', panelCacheKey(String(to.meta.activeMenu)), to.path);
            }
        }
    }

    NProgress.done();
});

export default router;
