<template>
    <div>
        <RouterButton :buttons="buttons" />
        <LayoutContent>
            <RouterViewCache />
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import i18n from '@/lang';

const buttons = [
    {
        label: i18n.global.t('aiTools.gpu.current'),
        path: '/ai/gpu/current',
    },
    {
        label: i18n.global.t('aiTools.gpu.history'),
        path: '/ai/gpu/history',
    },
];
</script>
