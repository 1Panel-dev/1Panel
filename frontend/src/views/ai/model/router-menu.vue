<template>
    <RouterButton :buttons="buttons" />
</template>

<script lang="ts" setup>
const buttons = [
    {
        label: 'Ollama',
        path: '/ai/model/ollama',
    },
    {
        label: 'vLLM',
        path: '/ai/model/vllm',
    },
    {
        label: 'TensorRT LLM',
        path: '/ai/model/tensorrt',
    },
];
</script>
