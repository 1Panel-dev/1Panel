<template>
    <el-tabs
        v-bind="$attrs"
        v-model="tabsStore.activeTabPath"
        class="common-tabs"
        type="card"
        :closable="tabsStore.openedTabs.length > 1"
        @tab-change="tabChange"
        @tab-remove="closeTab"
    >
        <tabs-view-item
            v-for="item in tabsStore.openedTabs"
            ref="tabItems"
            :key="item.path"
            :tab-item="item"
            @close-tab="closeTab"
            @close-other-tabs="closeOtherTabs"
            @close-tabs="closeTabs"
            @dropdown-visible-change="dropdownVisibleChange"
        />
    </el-tabs>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { TabsStore } from '@/store';
import { useRoute, useRouter } from 'vue-router';
import TabsViewItem from './components/TabItem.vue';

const router = useRouter();
const route = useRoute();
const tabsStore = TabsStore();
const tabItems = ref();

onMounted(() => {
    if (!tabsStore.openedTabs.length) {
        tabsStore.addTab(route);
    }
    tabsStore.activeTabPath = route.path;
});

const tabChange = (tabPath) => {
    const tab = tabsStore.findTab(tabPath);
    if (tab) {
        router.push(tab);
        tabsStore.activeTabPath = tab.path;
    }
};

const closeTab = (tabPath) => {
    const lastTabPath = tabsStore.removeTab(tabPath);
    if (lastTabPath) {
        tabChange(lastTabPath);
    }
};

const closeOtherTabs = (tabPath) => {
    tabsStore.removeOtherTabs(tabPath);
    tabChange(tabPath);
};

const closeTabs = (tabPath, type) => {
    tabsStore.removeTabs(tabPath, type);
    tabChange(tabPath);
};

const dropdownVisibleChange = (visible, tabPath) => {
    if (visible) {
        // 关闭其他下拉菜单
        tabItems.value.forEach(({ dropdownRef }) => {
            if (dropdownRef.id !== tabPath) {
                dropdownRef.handleClose();
            }
        });
    }
};
</script>

<style scoped>
:deep(.el-tabs__header) {
    margin: 0;
}
</style>
