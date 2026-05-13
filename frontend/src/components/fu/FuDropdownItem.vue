<template>
    <el-dropdown-item v-bind="$attrs" :disabled="computedDisabled">
        <slot></slot>
    </el-dropdown-item>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { hasManagePermissionAccess, type PermissionBindingValue } from '@/utils/permission';

defineOptions({
    name: 'FuDropdownItem',
    inheritAttrs: false,
});

const props = defineProps({
    disabled: {
        type: Boolean,
        default: false,
    },
    permission: {
        type: [String, Array],
        default: undefined,
    },
});

const permissionDisabled = ref(false);

const hasPermission = computed(() => hasManagePermissionAccess(props.permission as PermissionBindingValue));

const computedDisabled = computed(() => props.disabled || permissionDisabled.value || !hasPermission.value);

defineExpose({
    setPermissionDisabled: (disabled: boolean) => {
        permissionDisabled.value = disabled;
    },
});
</script>
