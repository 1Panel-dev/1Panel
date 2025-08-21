<template>
    <DrawerPro
        v-model="logVisible"
        :header="$t('commons.button.log')"
        @close="handleClose"
        :resource="logSearch.container"
        :size="globalStore.isFullScreen ? 'full' : 'large'"
    >
        <template #extra v-if="!mobile">
            <el-tooltip :content="loadTooltip()" placement="top">
                <el-button @click="toggleFullscreen" class="fullScreen" icon="FullScreen" plain></el-button>
            </el-tooltip>
        </template>
        <template #content>
            <ContainerLog :container="config.container" :highlightDiff="highlightDiff" />
        </template>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue';
import screenfull from 'screenfull';
import { GlobalStore } from '@/store';
import ContainerLog from '@/components/log/container/index.vue';

const logVisible = ref(false);
const mobile = computed(() => {
    return globalStore.isMobile();
});
const globalStore = GlobalStore();
const logSearch = reactive({
    isWatch: true,
    container: '',
    containerID: '',
    mode: 'all',
    tail: 100,
});

defineProps({
    highlightDiff: {
        type: Number,
        default: 320,
    },
});

function toggleFullscreen() {
    globalStore.isFullScreen = !globalStore.isFullScreen;
}

const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (globalStore.isFullScreen ? 'quitFullscreen' : 'fullscreen'));
};

const handleClose = async () => {
    logVisible.value = false;
    globalStore.isFullScreen = false;
};

watch(logVisible, (val) => {
    if (screenfull.isEnabled && !val && !mobile.value) screenfull.exit();
});

interface DialogProps {
    container: string;
    containerID: string;
}

const config = ref<DialogProps>({
    container: '',
    containerID: '',
});

const acceptParams = (props: DialogProps): void => {
    config.value.containerID = props.containerID;
    config.value.container = props.container;
    logSearch.container = props.container;
    logVisible.value = true;

    if (!mobile.value) {
        screenfull.on('change', () => {
            globalStore.isFullScreen = screenfull.isFullscreen;
        });
    }
};

onBeforeUnmount(() => {
    handleClose();
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.fullScreen {
    border: none;
}
</style>
