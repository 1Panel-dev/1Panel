<template>
    <div>
        <div class="content-container__search">
            <el-card>
                <div>
                    <el-button
                        v-for="button in buttons"
                        :key="button.value"
                        class="tag-button"
                        :class="currentTab === button.value ? '' : 'no-active'"
                        :type="currentTab === button.value ? 'primary' : ''"
                        @click="handleChange(button.value)"
                    >
                        {{ button.label }}
                    </el-button>
                </div>
            </el-card>
        </div>
        <div class="local-model-content">
            <component :is="currentComponent" />
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import OllamaView from '@/views/ai/model/ollama/index.vue';
import TensorRTView from '@/views/ai/model/tensorrt/index.vue';
import VllmView from '@/xpack/views/vllm/index.vue';

type LocalTab = 'ollama' | 'vllm' | 'tensorrt';

const route = useRoute();
const router = useRouter();

const tabLabels: Record<LocalTab, string> = {
    ollama: 'Ollama',
    vllm: 'vLLM',
    tensorrt: 'TensorRT LLM',
};

const buttons = [
    { label: tabLabels.ollama, value: 'ollama' },
    { label: tabLabels.vllm, value: 'vllm' },
    { label: tabLabels.tensorrt, value: 'tensorrt' },
];

const currentTab = computed<LocalTab>(() => {
    const tab = route.query.tab;
    if (tab === 'vllm' || tab === 'tensorrt') {
        return tab;
    }
    return 'ollama';
});

const currentComponent = computed(() => {
    switch (currentTab.value) {
        case 'vllm':
            return VllmView;
        case 'tensorrt':
            return TensorRTView;
        default:
            return OllamaView;
    }
});

const handleChange = async (target: LocalTab) => {
    await router.replace({
        path: '/ai/model/local',
        query: {
            ...route.query,
            tab: target,
        },
    });
};
</script>

<style lang="scss" scoped>
.local-model-content {
    :deep(.content-container__app) {
        margin-top: 7px;
    }

    :deep(.content-container__main) {
        margin-top: 7px;
    }
}
</style>
