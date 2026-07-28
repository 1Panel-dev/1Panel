<template>
    <el-tag
        size="small"
        effect="dark"
        :type="tagType"
        :class="{ 'api-type-tag--responses': apiType === 'openai-responses' }"
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
    --el-tag-bg-color: #7c3aed;
    --el-tag-border-color: #7c3aed;
    --el-tag-text-color: #ffffff;
}
</style>
