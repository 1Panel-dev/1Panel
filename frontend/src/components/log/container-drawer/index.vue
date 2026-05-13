<template>
    <DrawerPro
        v-model="logVisible"
        :header="$t('commons.button.log')"
        @close="handleClose"
        :resource="logSearch.container"
        :size="isFullScreen ? 'full' : '60%'"
    >
        <template #extra v-if="!isMobile">
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
import { onBeforeUnmount, reactive, ref, watch } from 'vue';
import screenfull from 'screenfull';
import ContainerLog from '@/components/log/container/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isMobile, isFullScreen } = useGlobalStore();

const logVisible = ref(false);
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
    isFullScreen.value = !isFullScreen.value;
}

const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (isFullScreen.value ? 'quitFullscreen' : 'fullscreen'));
};

const handleClose = async () => {
    logVisible.value = false;
    isFullScreen.value = false;
};

watch(logVisible, (val) => {
    if (screenfull.isEnabled && !val && !isMobile.value) screenfull.exit();
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

    if (!isMobile.value) {
        screenfull.on('change', () => {
            isFullScreen.value = screenfull.isFullscreen;
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
