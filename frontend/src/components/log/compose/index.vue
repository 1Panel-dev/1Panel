<template>
    <DrawerPro
        v-model="open"
        :header="resource"
        @close="handleClose"
        :size="isFullScreen ? 'full' : '60%'"
        :resource="container"
    >
        <template #extra v-if="!isMobile">
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
import { onBeforeUnmount, ref, watch } from 'vue';
import screenfull from 'screenfull';
import ContainerLog from '@/components/log/container/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isMobile, isFullScreen } = useGlobalStore();
const open = ref(false);
const resource = ref('');
const container = ref('');
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

const handleClose = () => {
    open.value = false;
    isFullScreen.value = false;
};

function toggleFullscreen() {
    isFullScreen.value = !isFullScreen.value;
}
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (isFullScreen.value ? 'quitFullscreen' : 'fullscreen'));
};

watch(logVisible, (val) => {
    if (screenfull.isEnabled && !val && !isMobile.value) screenfull.exit();
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
