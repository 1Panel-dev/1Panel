<template>
    <div>
        <RouterButton :buttons="tabs" />
        <OllamaModels :key="currentNode" />
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import i18n from '@/lang';
import OllamaModels from './models/index.vue';

const { currentNode } = useGlobalStore();
const tabs = computed(() => [
    { label: i18n.global.t('aiTools.model.files'), path: '/ai/model/files' },
    { label: i18n.global.t('aiTools.model.services'), path: '/ai/model/account' },
    { label: i18n.global.t('aiTools.model.engines'), path: '/ai/model/local' },
]);
</script>
