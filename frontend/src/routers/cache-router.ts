import { RouteRecordRaw, RouteRecordName } from 'vue-router';
import { routerArray } from '@/routers/router';

/**
 * @description 使用递归，过滤需要缓存的路由
 * @param {Array} _route 所有路由表
 * @param {Array} _cache 缓存的路由表
 * @return void
 * */
let cacheRouter: RouteRecordName[] = [];
const filterKeepAlive = (_route: RouteRecordRaw[], _cache: RouteRecordName[]): void => {
    _route.forEach((item) => {
        item.meta?.keepAlive && item.name && _cache.push(item.name);
        item?.children?.length && filterKeepAlive(item.children, _cache);
    });
};

filterKeepAlive(routerArray, cacheRouter);

export default cacheRouter;
