<template>
    <DrawerPro
        v-model="open"
        :header="$t('commons.button.log')"
        @close="handleClose"
        :size="globalStore.isFullScreen ? 'full' : 'large'"
    >
        <template #extra v-if="!mobile">
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
import { GlobalStore } from '@/store';
import i18n from '@/lang';
import screenfull from 'screenfull';

const globalStore = GlobalStore();
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
    globalStore.isFullScreen = false;
    em('close', false);
};

const mobile = computed(() => {
    return globalStore.isMobile();
});

function toggleFullscreen() {
    globalStore.isFullScreen = !globalStore.isFullScreen;
}
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (globalStore.isFullScreen ? 'quitFullscreen' : 'fullscreen'));
};

watch(open, (val) => {
    if (screenfull.isEnabled && !val && !mobile.value) screenfull.exit();
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

<style lang="scss">
.fullScreen {
    border: none;
}
</style>
