<template>
    <DrawerPro
        v-model="open"
        :header="resource"
        @close="handleClose"
        :size="globalStore.isFullScreen ? 'full' : 'large'"
        :resource="resource"
    >
        <template #extra v-if="!mobile">
            <el-tooltip :content="loadTooltip()" placement="top">
                <el-button @click="toggleFullscreen" class="fullScreen" icon="FullScreen" plain></el-button>
            </el-tooltip>
        </template>
        <template #content>
            <ContainerLog :compose="compose" />
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { GlobalStore } from '@/store';
import screenfull from 'screenfull';
import ContainerLog from '@/components/container-log/index.vue';

const open = ref(false);
const resource = ref('');
const globalStore = GlobalStore();
const logVisible = ref(false);
const compose = ref('');

interface DialogProps {
    compose: string;
    resource: string;
}

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
    compose.value = props.compose;
    resource.value = props.resource;
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
