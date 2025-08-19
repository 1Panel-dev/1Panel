<template>
    <ul class="context-menu" ref="menuRef" :style="{ top: `${adjustedY}px`, left: `${adjustedX}px` }" @click.stop>
        <li
            v-for="(btn, index) in buttons"
            :key="index"
            :class="[{ disabled: isDisabled(btn) }, { divided: btn.divided }]"
            @click="!isDisabled(btn) && onClick(btn)"
        >
            {{ btn.label }}
        </li>
    </ul>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
    x: number;
    y: number;
    row: any;
    buttons: {
        label: string;
        click: (row: any) => void;
        disabled?: boolean | ((row: any) => boolean);
        divided?: boolean;
    }[];
}>();

const emit = defineEmits(['close']);
const menuRef = ref<HTMLElement | null>(null);
const onClick = (btn: any) => {
    btn.click?.(props.row);
    emit('close');
};

const isDisabled = computed(() => {
    return function (btn: any) {
        return typeof btn.disabled === 'function' ? btn.disabled(props.row) : btn.disabled;
    };
});

const adjustedX = ref(props.x);
const adjustedY = ref(props.y);

watch(
    () => [props.x, props.y],
    async () => {
        await nextTick(); // 确保菜单渲染完成
        if (!menuRef.value) return;

        const menuRect = menuRef.value.getBoundingClientRect();
        const windowWidth = window.innerWidth;
        const windowHeight = window.innerHeight;

        // 修正横向
        if (props.x + menuRect.width > windowWidth) {
            adjustedX.value = windowWidth - menuRect.width - 4; // 留点边距
        } else {
            adjustedX.value = props.x;
        }

        // 修正纵向
        if (props.y + menuRect.height > windowHeight) {
            adjustedY.value = windowHeight - menuRect.height - 4;
        } else {
            adjustedY.value = props.y;
        }
    },
    { immediate: true },
);

const handleClickOutside = () => emit('close');

onMounted(() => document.addEventListener('click', () => handleClickOutside));
onUnmounted(() => document.removeEventListener('click', () => handleClickOutside));
</script>
<style scoped>
.context-menu {
    position: fixed;
    background: var(--panel-main-bg-color-9);
    border: 1px solid var(--el-border-color);
    color: var(--el-color-primary);
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
    list-style: none;
    font-size: 14px;
    padding: 4px 0;
    margin: 0;
    z-index: 9999;
    min-width: 120px;
}
.context-menu li {
    padding: 6px 12px;
    cursor: pointer;
}
.context-menu li:hover {
    background-color: var(--panel-menu-bg-color);
}
.context-menu li.disabled {
    color: var(--el-border-color);
    cursor: not-allowed;
}
.context-menu li.divided {
    border-top: 1px solid var(--el-border-color);
}
</style>
