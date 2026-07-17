<template>
    <FuTableOperationActions
        v-if="cardRow"
        :buttons="buttons"
        :row="cardRow"
        :ellipsis="ellipsis"
        :extra="2"
        :trigger="trigger"
        :dropdown-style="dropdownStyle"
    />
    <el-table-column v-else v-bind="$attrs" :label="label" :width="resolvedWidth" :align="align" :fixed="resolvedFixed">
        <template #default="{ row }">
            <FuTableOperationActions
                :buttons="buttons"
                :row="row"
                :ellipsis="ellipsis"
                :trigger="trigger"
                :dropdown-style="dropdownStyle"
            />
        </template>
    </el-table-column>
</template>

<script setup lang="ts">
import { computed, type PropType } from 'vue';

import FuTableOperationActions from './TableOperationActions.vue';
import type { FuTableOperationButton } from './shared';

defineOptions({ name: 'FuTableOperations' });

const normalizeWidth = (value?: string | number) => {
    if (value === undefined || value === null || value === '') {
        return undefined;
    }
    if (typeof value === 'number') {
        return value;
    }
    const trimmed = value.trim();
    if (!trimmed || trimmed === 'auto') {
        return undefined;
    }
    return trimmed;
};

const props = defineProps({
    buttons: {
        type: Array as PropType<FuTableOperationButton[]>,
        default: () => [],
    },
    label: {
        type: String,
        default: '',
    },
    width: {
        type: [Number, String],
        default: undefined,
    },
    minWidth: {
        type: [Number, String],
        default: undefined,
    },
    align: {
        type: String,
        default: 'center',
    },
    ellipsis: {
        type: Number,
        default: 2,
    },
    trigger: {
        type: String,
        default: 'hover',
    },
    fixed: {
        type: [Boolean, String],
        default: undefined,
    },
    fix: {
        type: [Boolean, String],
        default: undefined,
    },
    maxHeight: {
        type: [Number, String],
        default: undefined,
    },
    cardRow: {
        type: Object,
        default: undefined,
    },
});

const resolvedFixed = computed(() => {
    if (props.fixed !== undefined) {
        return props.fixed;
    }
    if (typeof props.fix === 'string') {
        return props.fix;
    }
    return props.fix ? 'right' : false;
});

const hasDynamicShow = computed(() => {
    return props.buttons.some((button) => typeof button.show === 'function');
});

const staticVisibleCount = computed(() => {
    return props.buttons.filter((button) => button.show !== false).length;
});

const estimatedVisibleCount = computed(() => {
    if (staticVisibleCount.value === 0) {
        return 0;
    }

    const renderCap = Math.max(props.ellipsis, 0) + 1;
    if (hasDynamicShow.value) {
        return renderCap;
    }

    return Math.min(staticVisibleCount.value, renderCap);
});

const estimatedWidth = computed(() => {
    const buttonsWidth = 35 + estimatedVisibleCount.value * 58 + 58;
    const minWidth = normalizeWidth(props.minWidth);
    if (typeof minWidth === 'number') {
        return Math.max(buttonsWidth, minWidth);
    }
    if (typeof minWidth === 'string') {
        const parsed = Number(minWidth.replace('px', ''));
        if (!Number.isNaN(parsed)) {
            return Math.max(buttonsWidth, parsed);
        }
    }
    return buttonsWidth;
});

const resolvedWidth = computed(() => {
    return normalizeWidth(props.width) ?? estimatedWidth.value;
});

const dropdownStyle = computed(() => {
    if (props.maxHeight === undefined || props.maxHeight === null || props.maxHeight === '') {
        return undefined;
    }
    const maxHeight = typeof props.maxHeight === 'number' ? `${props.maxHeight}px` : props.maxHeight;
    return {
        maxHeight,
        overflowY: 'auto',
    };
});
</script>
