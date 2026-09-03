import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import { getTabCacheName } from '@/utils/tab-cache';

const TabsStore = defineStore(
    'TabsStore',
    () => {
        const isShowTabIcon = ref(true);
        const openedTabs = ref([]);
        const activeTabPath = ref('');
        const keepAliveTabs = computed(() => {
            return openedTabs.value.filter((tab) => tab.keepAlive).map((tab) => getTabCacheName(tab.path));
        });

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
        };

        const removeUnActiveTabs = () => {
            if (openedTabs.value.length) {
                let idx = getTabIdxByPath(activeTabPath.value);
                idx = idx > -1 ? idx : 0;
                const tab = openedTabs.value[idx];
                removeOtherTabs(tab.path);
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
                openedTabs.value.push(Object.assign({}, tab, { keepAlive: false }));
            }
        };

        const toggleKeepAlive = (path) => {
            const tab = findTab(path);
            if (tab) {
                tab.keepAlive = !tab.keepAlive;
            }
        };

        const removeTab = (path) => {
            if (openedTabs.value.length > 1) {
                const idx = getTabIdxByPath(path);
                if (idx > -1) {
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
            }
        };

        const removeTabs = (path, type) => {
            if (path) {
                const idx = getTabIdxByPath(path);
                if (type === 'right') {
                    openedTabs.value.splice(idx + 1);
                } else if (type === 'left') {
                    openedTabs.value.splice(0, idx);
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
            keepAliveTabs,
            addTab,
            findTab,
            toggleKeepAlive,
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
