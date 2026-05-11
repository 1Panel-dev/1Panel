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
import { computed, defineAsyncComponent } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import OllamaView from '@/views/ai/model/ollama/index.vue';
import TensorRTView from '@/views/ai/model/tensorrt/index.vue';
import { loadOptionalComponent } from '@/extensions/optional';
import i18n from '@/lang';

const VllmView = defineAsyncComponent(() => loadOptionalComponent('/src/xpack/views/vllm/index.vue'));
const ModelDownloaderView = defineAsyncComponent(() =>
    loadOptionalComponent('/src/xpack/views/model-downloader/index.vue'),
);

type LocalTab = 'ollama' | 'vllm' | 'tensorrt' | 'downloader';

const route = useRoute();
const router = useRouter();

const tabLabels: Record<LocalTab, string> = {
    ollama: 'Ollama',
    vllm: 'vLLM',
    tensorrt: 'TensorRT LLM',
    downloader: i18n.global.t('aiTools.model.downloader'),
};

const buttons: Array<{ label: string; value: LocalTab }> = [
    { label: tabLabels.ollama, value: 'ollama' },
    { label: tabLabels.vllm, value: 'vllm' },
    { label: tabLabels.tensorrt, value: 'tensorrt' },
    { label: tabLabels.downloader, value: 'downloader' },
];

const currentTab = computed<LocalTab>(() => {
    const tab = route.query.tab;
    if (tab === 'vllm' || tab === 'tensorrt' || tab === 'downloader') {
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
        case 'downloader':
            return ModelDownloaderView;
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
.content-container__search {
    margin-top: 7px;

    :deep(.el-card) {
        --el-card-padding: 12px;
    }
}

.local-model-content {
    :deep(.content-container__app) {
        margin-top: 7px;
    }

    :deep(.content-container__main) {
        margin-top: 7px;
    }
}
</style>
