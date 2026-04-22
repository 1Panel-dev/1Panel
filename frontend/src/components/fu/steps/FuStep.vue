<template>
    <div class="fu-step" v-loading="loading">
        <slot />
    </div>
</template>

<script setup lang="ts">
import { computed, inject } from 'vue';

import type { Stepper } from './Stepper';

defineOptions({ name: 'FuStep' });

defineProps({
    id: {
        type: String,
        default: '',
    },
    title: {
        type: String,
        default: '',
    },
    description: {
        type: String,
        default: '',
    },
    status: {
        type: String,
        default: '',
    },
    icon: {
        type: String,
        default: '',
    },
});

const stepper = inject<Stepper>('stepper');

const loading = computed(() => {
    return stepper?.isLoading || false;
});
</script>
