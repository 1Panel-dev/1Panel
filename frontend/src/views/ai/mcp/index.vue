<template>
    <RouterButton :buttons="buttons" />
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { computed } from 'vue';

const buttons = computed<RouterButton[]>(() => [
    {
        label: i18n.global.t('panelMCP.accessTitle'),
        path: '/ai/mcp/1panel',
        permission: 'mcp_access_view',
    },
    {
        label: i18n.global.t('panelMCP.servicesTitle'),
        path: '/ai/mcp/server',
        permission: 'ai_mcp_view',
    },
]);
</script>
