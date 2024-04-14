import { ref } from 'vue';
import { defineStore } from 'pinia';

const TabsStore = defineStore(
    'TabsStore',
    () => {
        const isShowTabIcon = ref(true);
        // 缓存的KEY，直接给keepalive使用
        const cachedTabs = ref([]);
        const openedTabs = ref([]);
        const activeTabPath = ref('');

        const getActivePath = (path) => {
            let firstSlashIndex = path.indexOf('/');
            let lastSlashIndex = path.lastIndexOf('/');
            if (firstSlashIndex === -1 || firstSlashIndex === lastSlashIndex) {
                return path;
            }
            return path.substring(firstSlashIndex, lastSlashIndex);
        };

        const getTabIdxByPath = (path) => {
            return openedTabs.value.findIndex((v) => v.path === path);
        };

        const removeAllTabs = () => {
            openedTabs.value = [];
            cachedTabs.value = [];
        };

        const removeUnActiveTabs = () => {
            if (openedTabs.value.length) {
                let idx = getTabIdxByPath(activeTabPath.value);
                idx = idx > -1 ? idx : 0;
                const tab = openedTabs.value[idx];
                removeOtherTabs(tab);
            }
        };

        const findTab = (path) => {
            const idx = getTabIdxByPath(path);
            if (idx > -1) {
                return openedTabs.value[idx];
            }
        };

        const addTab = (tab) => {
            const idx = getTabIdxByPath(tab.path);
            if (idx < 0) {
                openedTabs.value.push(Object.assign({}, tab));
                addCachedTab(tab.name);
            }
        };

        const removeTab = (path) => {
            if (openedTabs.value.length > 1) {
                const idx = getTabIdxByPath(path);
                if (idx > -1) {
                    removeCachedTab(openedTabs.value[idx].name);
                    openedTabs.value.splice(idx, 1);
                }
                return openedTabs.value[openedTabs.value.length - 1].path;
            }
        };

        const removeOtherTabs = (path) => {
            const idx = getTabIdxByPath(path);
            if (idx > -1) {
                const tab = openedTabs.value[idx];
                openedTabs.value = [tab];
                cachedTabs.value = [];
                cachedTabs.value = [tab.name];
            }
        };

        const removeTabs = (path, type) => {
            if (path) {
                const idx = getTabIdxByPath(path);
                let removeTabs = [];
                if (type === 'right') {
                    removeTabs = openedTabs.value.splice(idx + 1);
                } else if (type === 'left') {
                    removeTabs = openedTabs.value.splice(0, idx);
                }
                if (removeTabs.length) {
                    removeTabs.forEach((e) => removeCachedTab(e.name));
                }
            }
        };

        const addCachedTab = (name) => {
            if (name && !cachedTabs.value.includes(name)) {
                cachedTabs.value.push(name);
            }
        };

        const removeCachedTab = (name) => {
            if (name) {
                const idx = cachedTabs.value.findIndex((v) => v === name);
                if (idx > -1) {
                    cachedTabs.value.splice(idx, 1);
                }
            }
        };

        const hasCloseDropdown = (path, type) => {
            const idx = getTabIdxByPath(path);
            switch (type) {
                case 'close':
                case 'other':
                    return openedTabs.value.length > 1;
                case 'left':
                    return idx !== 0;
                case 'right':
                    return idx !== openedTabs.value.length - 1;
            }
        };

        return {
            isShowTabIcon,
            activeTabPath,
            openedTabs,
            cachedTabs,
            addTab,
            findTab,
            addCachedTab,
            removeCachedTab,
            removeTab,
            removeTabs,
            removeOtherTabs,
            removeAllTabs,
            removeUnActiveTabs,
            hasCloseDropdown,
            getActivePath,
        };
    },
    {
        persist: true,
    },
);

export default TabsStore;
