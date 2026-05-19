<template>
    <li
        ref="itemRef"
        v-bind="itemAttrs"
        data-el-collection-item
        :aria-disabled="computedDisabled"
        :class="itemClass"
        :style="attrs.style"
        :tabindex="tabIndex"
        :role="role"
        @click="handleClick"
        @focus="handleFocus"
        @keydown.self="handleKeydown"
        @mousedown="handleMousedown"
        @pointermove="handlePointerMove"
        @pointerleave="handlePointerLeave"
    >
        <el-icon v-if="icon || $slots.icon">
            <slot name="icon">
                <component :is="icon" />
            </slot>
        </el-icon>
        <slot />
    </li>
</template>

<script setup lang="ts">
import { DROPDOWN_INJECTION_KEY, EVENT_CODE } from 'element-plus';
import { COLLECTION_ITEM_SIGN } from 'element-plus/es/components/collection/index.mjs';
import {
    ROVING_FOCUS_COLLECTION_INJECTION_KEY,
    ROVING_FOCUS_GROUP_INJECTION_KEY,
} from 'element-plus/es/components/roving-focus-group/index.mjs';
import { computed, getCurrentInstance, inject, onBeforeUnmount, onMounted, ref, useAttrs } from 'vue';

defineOptions({
    name: 'FuDropdownItem',
    inheritAttrs: false,
});

const props = defineProps({
    command: {
        type: [Object, String, Number],
        default: () => ({}),
    },
    disabled: {
        type: Boolean,
        default: false,
    },
    divided: {
        type: Boolean,
        default: false,
    },
    textValue: String,
    icon: {
        type: [String, Object, Function],
    },
});

const emit = defineEmits(['click', 'pointermove', 'pointerleave']);

const attrs = useAttrs();
const itemRef = ref<HTMLElement>();
const permissionDisabled = ref(false);
const instance = getCurrentInstance();
const itemId = `fu-dropdown-item-${instance?.uid ?? Math.random().toString(36).slice(2)}`;

const dropdownContext = inject<any>(DROPDOWN_INJECTION_KEY, undefined);
const collectionContext = inject<any>(ROVING_FOCUS_COLLECTION_INJECTION_KEY, undefined);
const rovingFocusContext = inject<any>(ROVING_FOCUS_GROUP_INJECTION_KEY, undefined);

const computedDisabled = computed(() => props.disabled || permissionDisabled.value);

const itemAttrs = computed(() => {
    return Object.fromEntries(Object.entries(attrs).filter(([key]) => key !== 'class' && key !== 'style'));
});

const itemClass = computed(() => [
    attrs.class,
    'el-dropdown-menu__item',
    {
        'el-dropdown-menu__item--divided': props.divided,
        'is-disabled': computedDisabled.value,
        'fu-dropdown-item--permission-disabled': computedDisabled.value,
    },
]);

const tabIndex = computed(() => (rovingFocusContext?.currentTabbedId?.value === itemId ? 0 : -1));

const role = computed(() => {
    const menuRole = dropdownContext?.role?.value;
    if (menuRole === 'menu') {
        return 'menuitem';
    }
    if (menuRole === 'navigation') {
        return 'link';
    }
    return 'button';
});

const handleMousedown = (event: MouseEvent) => {
    if (computedDisabled.value) {
        event.preventDefault();
        return;
    }
    rovingFocusContext?.onItemFocus?.(itemId);
};

const handleFocus = () => {
    rovingFocusContext?.onItemFocus?.(itemId);
};

const handleKeydown = (event: KeyboardEvent) => {
    if ([EVENT_CODE.enter, EVENT_CODE.numpadEnter, EVENT_CODE.space].includes(event.code)) {
        event.preventDefault();
        event.stopImmediatePropagation();
        handleClick(event);
        return;
    }

    if (event.code === EVENT_CODE.tab && event.shiftKey) {
        rovingFocusContext?.onItemShiftTab?.();
        return;
    }
    rovingFocusContext?.onKeydown?.(event);
};

const handleClick = (event: Event) => {
    if (computedDisabled.value) {
        event.stopImmediatePropagation();
        return;
    }

    emit('click', event);
    if (event.type !== 'keydown' && event.defaultPrevented) {
        return;
    }

    if (dropdownContext?.hideOnClick?.value) {
        dropdownContext.handleClick?.();
    }
    dropdownContext?.commandHandler?.(props.command, instance, event);
};

const handlePointerMove = (event: PointerEvent) => {
    emit('pointermove', event);
    if (event.pointerType !== 'mouse') {
        return;
    }
    if (computedDisabled.value) {
        dropdownContext?.onItemLeave?.(event);
        return;
    }

    const target = event.currentTarget as HTMLElement;
    if (target === document.activeElement || target.contains(document.activeElement)) {
        return;
    }

    dropdownContext?.onItemEnter?.(event);
    if (!event.defaultPrevented) {
        target.focus({ preventScroll: true });
    }
};

const handlePointerLeave = (event: PointerEvent) => {
    emit('pointerleave', event);
    if (event.pointerType === 'mouse') {
        dropdownContext?.onItemLeave?.(event);
    }
};

onMounted(() => {
    if (!itemRef.value) {
        return;
    }
    collectionContext?.itemMap?.set(itemRef.value, {
        ref: itemRef.value,
        ...attrs,
        [COLLECTION_ITEM_SIGN]: '',
    });
});

onBeforeUnmount(() => {
    if (itemRef.value) {
        collectionContext?.itemMap?.delete(itemRef.value);
    }
});

defineExpose({
    setPermissionDisabled: (disabled: boolean) => {
        permissionDisabled.value = disabled;
    },
});
</script>

<style scoped>
.fu-dropdown-item--permission-disabled {
    opacity: 0.45;
    cursor: not-allowed;
}
</style>
