import type { RouteRecordRaw } from 'vue-router';
import { getSettingBaseInfo } from '@/api/modules/setting';
import { menuList } from '@/routers/router';
import { GlobalStore, MenuStore } from '@/store';
import { applyPanelCacheSettings } from '@/utils/panel-cache';
import { hasPermissionMetaAccess, hasRouteRoleAccess } from '@/utils/rbac';

export async function loadMenuList() {
    const menuStore = MenuStore();
    const requestId = ++menuStore.menuRequestId;
    let hideMenu: string | undefined;
    try {
        const { data } = await getSettingBaseInfo();
        if (requestId !== menuStore.menuRequestId) {
            return;
        }
        menuStore.systemVersion = data.systemVersion;
        GlobalStore().menuAccordion = data.menuAccordion === 'Enable';
        applyPanelCacheSettings(data.panelCache);
        hideMenu = data.hideMenu;
    } catch {
        if (requestId !== menuStore.menuRequestId) {
            return;
        }
        menuStore.systemVersion = '';
    }

    let menus: RouteRecordRaw[];
    try {
        menus = hideMenu ? buildMenuListFromSettings(hideMenu) : buildAuthVisibleMenuList(menuList);
    } catch {
        menus = buildAuthVisibleMenuList(menuList);
    }
    if (JSON.stringify(menuStore.menuList) !== JSON.stringify(menus)) {
        menuStore.setMenuList(menus);
    }
    menuStore.isMenuLoaded = true;
}

function getCheckedLabels(menu: any, showSet: Set<string>) {
    for (const item of menu) {
        if (item.isShow) {
            showSet.add(item.label);
        }
        if (item.children) {
            getCheckedLabels(item.children, showSet);
        }
    }
}

function allowMenuItem(item: RouteRecordRaw) {
    if (!hasRouteRoleAccess(item.meta)) {
        return false;
    }
    return hasPermissionMetaAccess(item.meta?.permission as string | string[] | undefined);
}

function buildMenuListFromSettings(hideMenuValue?: string) {
    const hideMenu = JSON.parse(hideMenuValue || '[]');
    const showSet = new Set<string>();
    getCheckedLabels(hideMenu, showSet);
    const rstMenuList: RouteRecordRaw[] = [];
    const resMenuList = adjustAndCleanMenu(hideMenu, menuList);
    for (const menu of resMenuList) {
        const menuItem = buildVisibleMenu(menu, showSet);
        if (menuItem) {
            rstMenuList.push(menuItem);
        }
    }
    return rstMenuList;
}

function buildAuthVisibleMenuList(source: RouteRecordRaw[]) {
    return source
        .map((item) => {
            if (!allowMenuItem(item)) {
                return null;
            }
            const menuItem = JSON.parse(JSON.stringify(item));
            const children = Array.isArray(menuItem.children) ? menuItem.children : [];
            if (children.length === 0) {
                return menuItem;
            }
            menuItem.children = buildAuthVisibleMenuList(children).filter(Boolean);
            if (menuItem.children.length === 0) {
                return null;
            }
            if (menuItem.children.length === 1) {
                const onlyChild = menuItem.children[0];
                if (onlyChild.meta?.icon) {
                    menuItem.meta.icon = onlyChild.meta.icon;
                }
                if (onlyChild.meta?.title) {
                    menuItem.meta.title = onlyChild.meta.title;
                }
            }
            if (menuItem.name === 'Xpack-Menu') {
                menuItem.meta.hideInSidebar = false;
            }
            return menuItem;
        })
        .filter(Boolean) as RouteRecordRaw[];
}

function buildVisibleMenu(menu: RouteRecordRaw, showSet: Set<string>): RouteRecordRaw | null {
    const menuItem = JSON.parse(JSON.stringify(menu));
    if (!menuItem?.name || !showSet.has(menuItem.name as string)) {
        return null;
    }
    if (!allowMenuItem(menuItem)) {
        return null;
    }

    const children = Array.isArray(menuItem.children) ? menuItem.children : [];
    if (children.length === 0) {
        return menuItem;
    }

    const visibleChildren = children.map((item) => buildVisibleMenu(item, showSet)).filter(Boolean) as RouteRecordRaw[];

    menuItem.children = visibleChildren;
    if (menuItem.children.length === 0) {
        return null;
    }

    if (menuItem.children.length === 1) {
        const onlyChild = menuItem.children[0];
        if (onlyChild.meta?.icon) {
            menuItem.meta.icon = onlyChild.meta.icon;
        }
        if (onlyChild.meta?.title) {
            menuItem.meta.title = onlyChild.meta.title;
        }
    }
    if (menuItem.name === 'Xpack-Menu') {
        menuItem.meta.hideInSidebar = false;
    }
    return menuItem;
}

function adjustAndCleanMenu(menuItem, list) {
    const menuList = JSON.parse(JSON.stringify(list));
    const itemMap = new Map();
    for (const parent of menuList) {
        itemMap.set(parent.name, parent);
        if (Array.isArray(parent.children)) {
            for (const child of parent.children) {
                itemMap.set(child.name, child);
            }
        }
    }

    function buildTree(refList) {
        const result = [];

        for (const ref of refList) {
            const refName = ref.label;
            const matched = itemMap.get(refName);

            if (!matched) continue;

            if (Array.isArray(ref.children) && ref.children.length > 0) {
                matched.children = buildTree(ref.children);
            } else {
                delete matched.children;
            }

            result.push(matched);
        }

        return result;
    }

    const newMenu = buildTree(menuItem);
    for (const menu of newMenu) {
        if (menu.children?.length === 1) {
            const onlyChild = menu.children[0];
            if (onlyChild.meta?.icon) {
                menu.meta.icon = onlyChild.meta.icon;
            }
            if (onlyChild.meta?.title) {
                menu.meta.title = onlyChild.meta.title;
            }
        }
    }

    return newMenu;
}
