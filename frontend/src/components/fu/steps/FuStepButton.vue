<template>
    <el-button :disabled="disabled" v-bind="$attrs">
        <slot />
    </el-button>
</template>

<script setup lang="ts">
defineOptions({ name: 'FuStepButton' });

defineProps({
    disabled: {
        type: Boolean,
        default: false,
    },
});
</script>
