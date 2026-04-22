<template>
    <el-steps :active="active" v-bind="stepper" direction="vertical">
        <el-step
            v-for="(step, index) in steps"
            :key="index"
            v-bind="step"
            :class="disable?.(index) && 'fu-step--disable'"
            @click="handleClick(index)"
        >
            <template #description>
                <span>{{ step.description }}</span>
                <el-collapse-transition>
                    <div v-if="index === active" class="fu-steps__container" :style="heightStyle">
                        <slot :step="step" />
                    </div>
                </el-collapse-transition>
            </template>
        </el-step>
    </el-steps>
</template>

<script setup lang="ts">
import { computed, type PropType } from 'vue';

import type { Step, Stepper } from './Stepper';

defineOptions({ name: 'FuVerticalNavigation' });

const props = defineProps({
    stepper: Object as PropType<Stepper>,
    steps: Array as PropType<Step[]>,
    disable: Function as PropType<(index: number) => boolean>,
});

const emit = defineEmits(['active']);

const active = computed(() => props.stepper?.index ?? 0);

const heightStyle = computed(() => {
    const height = Number.parseInt(String(props.stepper?.height ?? ''), 10);
    return Number.isFinite(height) ? { height: `${height}px` } : undefined;
});

const handleClick = (index: number) => {
    if (!props.disable?.(index)) {
        emit('active', index);
    }
};
</script>
