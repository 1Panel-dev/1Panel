<template>
    <RouterViewCache :keep-alive="keepAlive" :with-transition="route.meta.disableViewTransition !== true" />
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import type { KeepAlivePattern } from '@/utils/tab-cache';

defineProps<{
    keepAlive?: KeepAlivePattern | null;
}>();

const route = useRoute();
</script>
