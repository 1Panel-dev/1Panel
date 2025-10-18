<template>
    <div>
        <RouterButton :buttons="buttons" />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
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
</script>
