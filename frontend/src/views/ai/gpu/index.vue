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
import { computed } from 'vue';

const props = defineProps({
    historyDisabled: {
        type: Boolean,
        default: false,
    },
});

const buttons = computed(() => [
    {
        label: i18n.global.t('aiTools.gpuGroup.overview'),
        path: '/ai/gpu/current',
    },
    { label: i18n.global.t('aiTools.gpuGroup.groups'), path: '/ai/gpu/groups' },
    {
        label: i18n.global.t('aiTools.gpuGroup.history'),
        path: '/ai/gpu/history',
        disabled: props.historyDisabled,
    },
]);
</script>
