<template>
    <el-tag
        size="small"
        effect="light"
        :type="tagType"
        :class="{
            'api-type-tag--responses': apiType === 'openai-responses',
            'api-type-tag--embedding': apiType === 'openai-embeddings',
        }"
    >
        {{ apiType || '-' }}
    </el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue';

type TagType = 'primary' | 'success' | 'warning' | 'info';

const props = defineProps<{
    apiType: string;
}>();

const tagType = computed<TagType>(() => {
    if (props.apiType === 'openai-completions') {
        return 'primary';
    }
    if (props.apiType === 'openai-responses') {
        return 'info';
    }
    if (props.apiType === 'anthropic-messages') {
        return 'warning';
    }
    if (props.apiType.endsWith('-images')) {
        return 'success';
    }
    return 'info';
});
</script>

<style scoped lang="scss">
.api-type-tag--responses {
    --el-tag-bg-color: #f3e8ff;
    --el-tag-border-color: #d8b4fe;
    --el-tag-text-color: #7e22ce;
}

.api-type-tag--embedding {
    --el-tag-bg-color: #e6f4f1;
    --el-tag-border-color: #a7d9d1;
    --el-tag-text-color: #0f766e;
}
</style>
