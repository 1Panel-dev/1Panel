<template>
    <div>
        <RouterButton :buttons="buttons" />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import i18n from '@/lang';

const buttons = [
    {
        label: i18n.global.t('aiTools.agents.agent'),
        path: '/ai/agents/agent',
    },
    {
        label: i18n.global.t('aiTools.agents.account'),
        path: '/ai/agents/model',
    },
];
</script>
