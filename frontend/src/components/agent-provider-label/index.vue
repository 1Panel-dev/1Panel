<template>
    <span class="agent-provider-label">
        <ProviderLogo :provider="provider" :display-name="name" :size="size" />
        <span class="agent-provider-label__name" :title="name">{{ name }}</span>
    </span>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import ProviderLogo from '@/components/agent-provider-logo/index.vue';
import { getAgentProviderDisplayName } from '@/utils/agent';

const props = withDefaults(
    defineProps<{
        provider: string;
        displayName?: string;
        size?: number;
    }>(),
    {
        displayName: '',
        size: 16,
    },
);

const name = computed(() => getAgentProviderDisplayName(props.provider, props.displayName));
</script>

<style scoped lang="scss">
.agent-provider-label {
    display: inline-flex;
    align-items: center;
    max-width: 100%;
    min-width: 0;
    gap: 6px;
    white-space: nowrap;
}

.agent-provider-label__name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
