<template>
    <el-tabs
        v-model="props.selectTab"
        type="card"
        :closable="props.fileTabs.length > 1"
        class="monaco-editor monaco-editor-background"
        @tab-remove="props.onRemoveTab"
        @tab-change="props.onChangeTab"
    >
        <el-tab-pane v-for="item in props.fileTabs" :key="item.path" :name="item.path">
            <template #label>
                <el-dropdown
                    size="small"
                    :id="item.path"
                    :ref="(el) => setDropdownRef(item.path, el)"
                    trigger="contextmenu"
                    placement="bottom"
                    @visible-change="(visible) => onDropdownVisibleChange(visible, item.path)"
                >
                    <span class="el-dropdown-link">
                        <el-tooltip :content="item.path" placement="bottom-start">
                            {{ item.name }}
                        </el-tooltip>
                    </span>
                    <template #dropdown>
                        <el-dropdown-menu>
                            <el-dropdown-item @click="props.onRemoveTab(item.path)">
                                <el-icon><Close /></el-icon>
                                {{ $t('commons.button.close') }}
                            </el-dropdown-item>
                            <el-dropdown-item @click="props.onRemoveAllTab(item.path, 'left')">
                                <el-icon><DArrowLeft /></el-icon>
                                {{ $t('tabs.closeLeft') }}
                            </el-dropdown-item>
                            <el-dropdown-item @click="props.onRemoveAllTab(item.path, 'right')">
                                <el-icon><DArrowRight /></el-icon>
                                {{ $t('tabs.closeRight') }}
                            </el-dropdown-item>
                            <el-dropdown-item @click="props.onRemoveOtherTab(item.path)">
                                <el-icon><More /></el-icon>
                                {{ $t('tabs.closeOther') }}
                            </el-dropdown-item>
                        </el-dropdown-menu>
                    </template>
                </el-dropdown>
            </template>
        </el-tab-pane>
    </el-tabs>
</template>

<script setup lang="ts">
import { ref } from 'vue';

interface FileTabs {
    path: string;
    name: string;
}

const props = defineProps({
    selectTab: String,
    fileTabs: Array<FileTabs>,
    onRemoveTab: Function,
    onChangeTab: Function,
    onRemoveAllTab: Function,
    onRemoveOtherTab: Function,
});
const dropdownRefs = ref<Record<string, any>>({});

const setDropdownRef = (path: string, el: any) => {
    if (el) {
        dropdownRefs.value[path] = el;
    }
};

const onDropdownVisibleChange = (visible: boolean, currentPath: string) => {
    if (visible) {
        for (const path in dropdownRefs.value) {
            if (path !== currentPath) {
                dropdownRefs.value[path]?.handleClose?.();
            }
        }
    }
};
</script>
