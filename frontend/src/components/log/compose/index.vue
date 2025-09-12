<template>
    <DrawerPro
        v-model="open"
        :header="resource"
        @close="handleClose"
        :size="globalStore.isFullScreen ? 'full' : 'large'"
        :resource="container"
    >
        <template #extra v-if="!mobile">
            <el-tooltip :content="loadTooltip()" placement="top">
                <el-button @click="toggleFullscreen" class="fullScreen" icon="FullScreen" plain></el-button>
            </el-tooltip>
        </template>
        <template #content>
            <ContainerLog
                :compose="compose"
                :resource="resource"
                :container="container"
                :node="node"
                :highlightDiff="highlightDiff"
            />
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { GlobalStore } from '@/store';
import screenfull from 'screenfull';
import ContainerLog from '@/components/log/container/index.vue';

const open = ref(false);
const resource = ref('');
const container = ref('');
const globalStore = GlobalStore();
const logVisible = ref(false);
const compose = ref('');
const highlightDiff = ref(150);
const node = ref('');

interface DialogProps {
    compose: string;
    resource: string;
    container: string;
    node: string;
}

const defaultProps = defineProps({
    highlightDiff: {
        type: Number,
        default: 150,
    },
});

const mobile = computed(() => {
    return globalStore.isMobile();
});

const handleClose = () => {
    open.value = false;
    globalStore.isFullScreen = false;
};

function toggleFullscreen() {
    globalStore.isFullScreen = !globalStore.isFullScreen;
}
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (globalStore.isFullScreen ? 'quitFullscreen' : 'fullscreen'));
};

watch(logVisible, (val) => {
    if (screenfull.isEnabled && !val && !mobile.value) screenfull.exit();
});

const acceptParams = (props: DialogProps): void => {
    highlightDiff.value = defaultProps.highlightDiff;
    compose.value = props.compose;
    resource.value = props.resource;
    container.value = props.container;
    node.value = props.node;
    open.value = true;
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
