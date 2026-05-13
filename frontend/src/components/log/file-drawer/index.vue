<template>
    <DrawerPro
        v-model="open"
        :header="$t('commons.button.log')"
        @close="handleClose"
        :size="isFullScreen ? 'full' : 'large'"
    >
        <template #extra v-if="!isMobile">
            <el-tooltip :content="loadTooltip()" placement="top">
                <el-button @click="toggleFullscreen" class="fullScreen" icon="FullScreen" plain></el-button>
            </el-tooltip>
        </template>
        <template #content>
            <LogFile :config="config" :height-diff="props.heightDiff"></LogFile>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import LogFile from '@/components/log/file/index.vue';
import i18n from '@/lang';
import screenfull from 'screenfull';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isMobile, isFullScreen } = useGlobalStore();

interface LogProps {
    id: number;
    type: string;
    name: string;
    tail: boolean;
}

const props = defineProps({
    heightDiff: {
        type: Number,
        default: 0,
    },
    style: {
        type: Object,
        default: () => ({}),
    },
});

const open = ref(false);
const config = ref();
const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    isFullScreen.value = false;
    em('close', false);
};

function toggleFullscreen() {
    isFullScreen.value = !isFullScreen.value;
}
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (isFullScreen.value ? 'quitFullscreen' : 'fullscreen'));
};

watch(open, (val) => {
    if (screenfull.isEnabled && !val && !isMobile.value) screenfull.exit();
});

const acceptParams = (logProps: LogProps) => {
    config.value = logProps;
    open.value = true;
};

onBeforeUnmount(() => {
    handleClose();
});

defineExpose({ acceptParams });
</script>

<style lang="scss" scoped>
.fullScreen {
    border: none;
}
</style>
