<template>
    <el-drawer
        v-model="open"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        :before-close="handleClose"
        :size="globalStore.isFullScreen ? '100%' : '50%'"
    >
        <template #header>
            <DrawerHeader :header="$t('website.log')" :back="handleClose">
                <template #extra v-if="!mobile">
                    <el-tooltip :content="loadTooltip()" placement="top">
                        <el-button @click="toggleFullscreen" class="fullScreen" icon="FullScreen" plain></el-button>
                    </el-tooltip>
                </template>
            </DrawerHeader>
        </template>
        <div>
            <LogFile :config="config"></LogFile>
        </div>
    </el-drawer>
</template>
<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import LogFile from '@/components/log-file/index.vue';
import screenfull from 'screenfull';
import { GlobalStore } from '@/store';
import i18n from '@/lang';

const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

interface LogProps {
    id: number;
    type: string;
    style: string;
    name: string;
    tail: boolean;
}

const open = ref(false);
const config = ref();
const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    em('close', false);
    globalStore.isFullScreen = false;
};

watch(open, (val) => {
    if (screenfull.isEnabled && !val && !mobile.value) screenfull.exit();
});

function toggleFullscreen() {
    globalStore.isFullScreen = !globalStore.isFullScreen;
}

const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (globalStore.isFullScreen ? 'quitFullscreen' : 'fullscreen'));
};

const acceptParams = (props: LogProps) => {
    config.value = props;
    open.value = true;

    if (!mobile.value) {
        screenfull.on('change', () => {
            globalStore.isFullScreen = screenfull.isFullscreen;
        });
    }
};

defineExpose({ acceptParams });
</script>

<style lang="scss">
.fullScreen {
    border: none;
}
</style>
