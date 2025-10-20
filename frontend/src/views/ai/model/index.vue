<template>
    <div>
        <RouterButton :buttons="buttons" />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { useGlobalStore } from '@/composables/useGlobalStore';
const { isFxplay } = useGlobalStore();

const buttons = [
    {
        label: 'Ollama',
        path: '/ai/model/ollama',
    },
    {
        label: 'TensorRT LLM',
        path: '/ai/model/tensorrt',
    },
];

onMounted(() => {
    if (!isFxplay.value) {
        buttons.splice(1, 1);
    }
});
</script>
