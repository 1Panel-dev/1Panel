<template>
    <div>
        <RouterButton :buttons="[{ label: $t('aiTools.model.ollamaManagement'), path: '/ai/ollama' }]" />
        <OllamaModels :key="currentNode" />
    </div>
</template>

<script setup lang="ts">
import { useGlobalStore } from '@/composables/useGlobalStore';
import OllamaModels from './models/index.vue';

const { currentNode } = useGlobalStore();
</script>
