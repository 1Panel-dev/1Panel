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
                    <el-icon v-if="tabsStore.isShowTabIcon && menuIcon">
                        <el-icon>
                            <SvgIcon :iconName="menuIcon" />
                        </el-icon>
                    </el-icon>
                    <span>{{ menuName }}</span>
                </span>
                <template #dropdown>
                    <el-dropdown-menu>
                        <el-dropdown-item
                            v-if="tabsStore.hasCloseDropdown(tabItem.path, 'close')"
                            @click="$emit('closeTab', tabItem.path)"
                        >
                            <el-icon><Close /></el-icon>
                            {{ $t('tabs.close') }}
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
import { useI18n } from 'vue-i18n';
import { Close, DArrowLeft, DArrowRight, More } from '@element-plus/icons-vue';
import SvgIcon from '@/components/svg-icon/svg-icon.vue';

const i18n = useI18n();
const tabsStore = TabsStore();

const props = defineProps({
    tabItem: {
        type: Object,
        required: true,
    },
});

defineEmits(['closeTab', 'closeOtherTabs', 'closeTabs', 'dropdownVisibleChange']);

const menuName = computed(() => {
    return i18n.t(props.tabItem.meta.title);
});

const menuIcon = computed(() => {
    return props.tabItem.meta.icon;
});
const dropdownRef = ref();

defineExpose({
    dropdownRef,
});
</script>

<style scoped>
.common-tabs .custom-tabs-label .el-icon {
    vertical-align: middle;
}
.common-tabs .custom-tabs-label span {
    vertical-align: middle;
    margin-left: 4px;
}
</style>
