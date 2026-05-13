<template>
    <span class="agent-provider-logo" :style="logoStyle" :aria-label="`${displayName || provider} logo`" role="img">
        <img v-if="logo.src" class="agent-provider-logo__image" :src="logo.src" :alt="displayName || provider" />
        <span v-else>{{ logo.mark }}</span>
    </span>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { getAgentProviderLogo } from '@/utils/agent-provider-logo';

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

const logo = computed(() => getAgentProviderLogo(props.provider, props.displayName));
const logoStyle = computed(() => ({
    width: `${props.size}px`,
    height: `${props.size}px`,
    minWidth: `${props.size}px`,
    background: logo.value.background,
    color: logo.value.color,
    fontSize: `${Math.max(10, Math.floor(props.size * 0.48))}px`,
}));
</script>

<style scoped lang="scss">
.agent-provider-logo {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    line-height: 1;
    letter-spacing: 0;
    user-select: none;
    overflow: hidden;
    box-sizing: border-box;
    vertical-align: middle;
    flex: 0 0 auto;
}

.agent-provider-logo__image {
    display: block;
    max-width: 100%;
    max-height: 100%;
    object-fit: contain;
}
</style>
