<template>
    <el-steps :active="active" v-bind="stepper">
        <el-step
            v-for="(step, index) in steps"
            :key="index"
            v-bind="step"
            :class="disable?.(index) && 'fu-step--disable'"
            @click="handleClick(index)"
        />
    </el-steps>
</template>

<script setup lang="ts">
import { computed, type PropType } from 'vue';

import type { Step, Stepper } from './Stepper';

defineOptions({ name: 'FuHorizontalNavigation' });

const props = defineProps({
    stepper: Object as PropType<Stepper>,
    steps: Array as PropType<Step[]>,
    disable: Function as PropType<(index: number) => boolean>,
});

const emit = defineEmits(['active']);

const active = computed(() => props.stepper?.index ?? 0);

const handleClick = (index: number) => {
    if (!props.disable?.(index)) {
        emit('active', index);
    }
};
</script>
