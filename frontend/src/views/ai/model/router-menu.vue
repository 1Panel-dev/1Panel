<template>
    <el-card class="router_card p-1 sm:p-0">
        <div class="flex w-full flex-col justify-start items-start sm:flex-row sm:items-center sm:justify-between">
            <el-radio-group v-model="activePath" @change="handleChange">
                <el-radio-button
                    v-for="item in buttons"
                    :key="item.path"
                    class="router_card_button"
                    :value="item.path"
                    size="large"
                >
                    {{ item.label }}
                </el-radio-button>
            </el-radio-group>
        </div>
    </el-card>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

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

const route = useRoute();
const router = useRouter();
const activePath = ref('');

watch(
    () => route.path,
    (path) => {
        activePath.value = path;
    },
    { immediate: true },
);

const handleChange = async (path: string) => {
    if (!path || path === route.path) {
        return;
    }
    await router.push({ path });
};
</script>

<style lang="scss">
.router_card {
    --el-card-padding: 0;

    .el-card__body {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
}

.router_card_button {
    .el-radio-button__inner {
        min-width: 100px;
        height: 100%;
        background-color: var(--panel-button-active) !important;
        box-shadow: none !important;
        border: 2px solid transparent !important;
        color: var(--el-text-color-regular) !important;
    }

    .el-radio-button__original-radio:checked + .el-radio-button__inner {
        color: var(--panel-button-text-color) !important;
        background-color: var(--panel-button-bg-color) !important;
        border-color: var(--panel-color-primary) !important;
        border-radius: 4px;
    }
}
</style>
