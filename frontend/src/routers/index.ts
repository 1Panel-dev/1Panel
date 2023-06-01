import router from '@/routers/router';
import NProgress from '@/config/nprogress';
import { GlobalStore } from '@/store';
import { AxiosCanceler } from '@/api/helper/axios-cancel';

const axiosCanceler = new AxiosCanceler();

/**
 * @description 路由拦截 beforeEach（路由配置无数种方法，个人觉得最简便）
 * */
router.beforeEach((to, from, next) => {
    NProgress.start();
    axiosCanceler.removeAllPending();
    const globalStore = GlobalStore();

    if (to.name === 'entrance' && globalStore.isLogin) {
        if (to.params.code === globalStore.entrance) {
            globalStore.setLogStatus(false);
            next({
                name: 'entrance',
                params: { code: globalStore.entrance },
            });
            NProgress.done();
            return;
        }
        next({ name: '404' });
        NProgress.done();
        return;
    }

    if (!to.matched.some((record) => record.meta.requiresAuth)) return next();
    return next();
});

router.afterEach(() => {
    NProgress.done();
});

export default router;
