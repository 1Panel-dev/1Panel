<template>
    <el-tab-pane :name="tabItem.path">
        <template #label>
            <el-dropdown
                size="small"
                :id="tabItem.path"
                ref="dropdownRef"
                trigger="contextmenu"
                @visible-change="$emit('dropdownVisibleChange', $event, tabItem.path)"
            >
                <span class="custom-tabs-label">
                    <span>{{ menuName }}</span>
                </span>
                <template #dropdown>
                    <el-dropdown-menu>
                        <el-dropdown-item
                            v-if="tabsStore.hasCloseDropdown(tabItem.path, 'close')"
                            @click="$emit('closeTab', tabItem.path)"
                        >
                            <el-icon><Close /></el-icon>
                            {{ $t('commons.button.close') }}
                        </el-dropdown-item>
                        <el-dropdown-item
                            v-if="tabsStore.hasCloseDropdown(tabItem.path, 'left')"
                            @click="$emit('closeTabs', tabItem.path, 'left')"
                        >
                            <el-icon><DArrowLeft /></el-icon>
                            {{ $t('tabs.closeLeft') }}
                        </el-dropdown-item>
                        <el-dropdown-item
                            v-if="tabsStore.hasCloseDropdown(tabItem.path, 'right')"
                            @click="$emit('closeTabs', tabItem.path, 'right')"
                        >
                            <el-icon><DArrowRight /></el-icon>
                            {{ $t('tabs.closeRight') }}
                        </el-dropdown-item>
                        <el-dropdown-item
                            v-if="tabsStore.hasCloseDropdown(tabItem.path, 'other')"
                            @click="$emit('closeOtherTabs', tabItem.path)"
                        >
                            <el-icon><More /></el-icon>
                            {{ $t('tabs.closeOther') }}
                        </el-dropdown-item>
                    </el-dropdown-menu>
                </template>
            </el-dropdown>
        </template>
    </el-tab-pane>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { TabsStore } from '@/store';
import i18n from '@/lang';
import { Close, DArrowLeft, DArrowRight, More } from '@element-plus/icons-vue';

const tabsStore = TabsStore();

const props = defineProps({
    tabItem: {
        type: Object,
        required: true,
    },
});

defineEmits(['closeTab', 'closeOtherTabs', 'closeTabs', 'dropdownVisibleChange']);

const menuName = computed(() => {
    let title;
    if (props.tabItem.meta.parent) {
        title = i18n.global.t(props.tabItem.meta.parent) + '-' + i18n.global.t(props.tabItem.meta.title);
    } else {
        title = i18n.global.t(props.tabItem.meta.title);
    }
    if (props.tabItem.meta.detail) {
        title = title + '-' + i18n.global.t(props.tabItem.meta.detail);
    }
    return title;
});

const dropdownRef = ref();

defineExpose({
    dropdownRef,
});
</script>

<style scoped>
.common-tabs .custom-tabs-label span {
    vertical-align: middle;
    margin-left: 4px;
}
</style>
