<template>
    <div>
        <el-button class="mt-4" icon="Back" @click="router.push('/ai/model/local')">
            {{ $t('aiTools.model.engines') }}
        </el-button>
        <component :is="currentComponent" />
    </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import TensorRTView from '@/views/ai/model/tensorrt/index.vue';
import VllmPlaceholder from '@/views/ai/model/vllm/index.vue';
import { loadOptionalComponent } from '@/extensions/optional';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { vllmAvailable } from '../deployment';

const VllmView = defineAsyncComponent(() => loadOptionalComponent('/src/xpack/views/vllm/index.vue'));
const route = useRoute();
const router = useRouter();
const { isFxplay, isProductPro } = useGlobalStore();
const currentComponent = computed(() => {
    if (route.query.tab === 'vllm') return isProductPro.value && vllmAvailable ? VllmView : VllmPlaceholder;
    if (route.query.tab === 'tensorrt' && isFxplay.value) return TensorRTView;
    return VllmPlaceholder;
});
watch(
    [() => route.query.tab, isFxplay],
    async () => {
        if (route.query.tab === 'tensorrt' && !isFxplay.value) await router.replace('/ai/model/local');
    },
    { immediate: true },
);
</script>
