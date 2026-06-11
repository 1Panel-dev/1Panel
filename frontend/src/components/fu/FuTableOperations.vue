<template>
    <el-table-column v-bind="$attrs" :label="label" :width="resolvedWidth" :align="align" :fixed="resolvedFixed">
        <template #default="{ row }">
            <div class="fu-table-operations" @click.stop>
                <span
                    v-for="button in getInlineButtons(row)"
                    :key="buildButtonKey(button)"
                    class="fu-table-operations__action"
                >
                    <el-button
                        class="fu-table-operations__button"
                        link
                        :type="button.type || 'primary'"
                        :disabled="isButtonDisabled(button, row)"
                        @click.stop="handleButtonClick(button, row)"
                    >
                        <el-icon v-if="button.icon" class="mr-1">
                            <component :is="button.icon" />
                        </el-icon>
                        <span>{{ button.label }}</span>
                    </el-button>
                </span>
                <span v-if="getMoreButtons(row).length > 0" class="fu-table-operations__action">
                    <el-dropdown
                        class="fu-table-more-button"
                        :trigger="trigger"
                        @command="(button) => handleButtonClick(button, row)"
                    >
                        <span class="fu-table-operations__dropdown-trigger">
                            <el-button class="fu-table-operations__button" link type="primary" @click.stop>
                                {{ t('fu.table.more') }}
                            </el-button>
                        </span>
                        <template #dropdown>
                            <el-dropdown-menu :style="dropdownStyle">
                                <el-dropdown-item
                                    v-for="button in getMoreButtons(row)"
                                    :key="buildButtonKey(button)"
                                    :command="button"
                                    :disabled="isButtonDisabled(button, row)"
                                    :divided="button.divided"
                                >
                                    <el-icon v-if="button.icon" class="mr-1">
                                        <component :is="button.icon" />
                                    </el-icon>
                                    <span>{{ button.label }}</span>
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                </span>
            </div>
        </template>
    </el-table-column>
</template>

<script setup lang="ts">
import { computed, type PropType } from 'vue';
import { useI18n } from 'vue-i18n';

import { resolveMaybeFn, type FuTableOperationButton } from './shared';
import { hasManagePermissionAccess, hasPermissionAccess } from '@/utils/permission';

defineOptions({ name: 'FuTableOperations' });

const { t } = useI18n();

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

const getVisibleButtons = (row: any) => {
    return props.buttons.filter((button) => resolveMaybeFn(button.show ?? true, row));
};

const hasMore = (row: any) => {
    return getVisibleButtons(row).length > props.ellipsis + 1;
};

const getInlineButtons = (row: any) => {
    const visibleButtons = getVisibleButtons(row);
    if (props.ellipsis <= 0) {
        return [];
    }
    return hasMore(row) ? visibleButtons.slice(0, props.ellipsis) : visibleButtons;
};

const getMoreButtons = (row: any) => {
    const visibleButtons = getVisibleButtons(row);
    return hasMore(row) ? visibleButtons.slice(props.ellipsis) : [];
};

const isButtonDisabled = (button: FuTableOperationButton, row: any) => {
    let permissionDisabled = false;
    const permissionOptions = { nodeAdmin: button.nodeAdmin === true };
    if (button.permission === true) {
        permissionDisabled = !hasManagePermissionAccess(undefined, permissionOptions);
    } else if (button.permission !== undefined) {
        permissionDisabled = !hasPermissionAccess(button.permission, permissionOptions);
    }
    return permissionDisabled || Boolean(resolveMaybeFn(button.disabled ?? false, row));
};

const handleButtonClick = (button: FuTableOperationButton, row: any) => {
    if (isButtonDisabled(button, row)) {
        return;
    }
    button.click?.(row);
};

const buildButtonKey = (button: FuTableOperationButton) => {
    return String(button.label || button.command || button.type || button.icon?.name || 'button');
};
</script>
