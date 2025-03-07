<template>
    <el-drawer
        v-model="terminalVisible"
        @close="handleClose"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        :size="globalStore.isFullScreen ? '100%' : '50%'"
    >
        <template #header>
            <DrawerHeader :header="$t('container.containerTerminal')" :resource="title" :back="handleClose">
                <template #extra v-if="!mobile">
                    <el-tooltip :content="loadTooltip()" placement="top">
                        <el-button @click="toggleFullscreen" class="fullScreen" icon="FullScreen" plain></el-button>
                    </el-tooltip>
                </template>
            </DrawerHeader>
        </template>
        <Terminal class="mt-2" style="height: calc(100vh - 175px)" ref="terminalRef"></Terminal>

        <template #footer>
            <span class="dialog-footer">
                <el-button type="primary" @click="handleClose">
                    {{ $t('commons.button.disconnect') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref } from 'vue';
import Terminal from '@/components/terminal/index.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { closeOllamaModel } from '@/api/modules/ai';
import { GlobalStore } from '@/store';
import i18n from '@/lang';

const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});

const title = ref();
const terminalVisible = ref(false);
const itemName = ref();
const terminalRef = ref();

interface DialogProps {
    name: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    itemName.value = params.name;
    terminalVisible.value = true;
    initTerm();
};

const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (globalStore.isFullScreen ? 'quitFullscreen' : 'fullscreen'));
};

const initTerm = () => {
    nextTick(() => {
        terminalRef.value.acceptParams({
            endpoint: '/api/v1/containers/exec',
            args: `source=ollama&name=${itemName.value}`,
            error: '',
            initCmd: '',
        });
    });
};

function toggleFullscreen() {
    globalStore.isFullScreen = !globalStore.isFullScreen;
}

const onClose = async () => {
    await closeOllamaModel(itemName.value)
        .then(() => {
            terminalRef.value?.onClose();
        })
        .catch(() => {
            terminalRef.value?.onClose();
        });
};

function handleClose() {
    onClose();
    globalStore.isFullScreen = false;
    terminalVisible.value = false;
}

defineExpose({
    acceptParams,
});
</script>
