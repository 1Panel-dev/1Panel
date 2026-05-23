<template>
    <span class="fu-dropdown-item">
        <el-dropdown-item
            v-bind="$attrs"
            :disabled="computedDisabled"
            :class="{ 'fu-dropdown-item--permission-disabled': computedDisabled }"
        >
            <slot />
        </el-dropdown-item>
    </span>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';

defineOptions({
    name: 'FuDropdownItem',
    inheritAttrs: false,
});

const props = defineProps({
    disabled: {
        type: Boolean,
        default: false,
    },
});

const permissionDisabled = ref(false);

const computedDisabled = computed(() => props.disabled || permissionDisabled.value);

defineExpose({
    setPermissionDisabled: (disabled: boolean) => {
        permissionDisabled.value = disabled;
    },
});
</script>

<style scoped>
.fu-dropdown-item {
    display: contents;
}

:deep(.fu-dropdown-item--permission-disabled),
.fu-dropdown-item.is-disabled :deep(.el-dropdown-menu__item),
.fu-dropdown-item[data-permission-disabled='true'] :deep(.el-dropdown-menu__item) {
    color: var(--el-text-color-disabled) !important;
    opacity: 0.55;
    filter: grayscale(1);
    cursor: not-allowed !important;
    background-color: transparent !important;
}

:deep(.fu-dropdown-item--permission-disabled:hover),
:deep(.fu-dropdown-item--permission-disabled:focus),
.fu-dropdown-item.is-disabled :deep(.el-dropdown-menu__item:hover),
.fu-dropdown-item.is-disabled :deep(.el-dropdown-menu__item:focus),
.fu-dropdown-item[data-permission-disabled='true'] :deep(.el-dropdown-menu__item:hover),
.fu-dropdown-item[data-permission-disabled='true'] :deep(.el-dropdown-menu__item:focus) {
    color: var(--el-text-color-disabled) !important;
    background-color: transparent !important;
}
</style>
