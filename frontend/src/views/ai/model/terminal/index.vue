<template>
    <DrawerPro
        v-model="open"
        :title="$t('menu.terminal')"
        @close="handleClose"
        :resource="title"
        :size="globalStore.isFullScreen ? 'full' : 'large'"
    >
        <el-alert type="error" :closable="false">
            <template #title>
                <span>{{ $t('commons.msg.disConn', ['/bye exit']) }}</span>
            </template>
        </el-alert>
        <Terminal class="mt-2" style="height: calc(100vh - 225px)" ref="terminalRef"></Terminal>

        <template #footer>
            <span class="dialog-footer">
                <el-button type="primary" @click="handleClose">
                    {{ $t('commons.button.disConn') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { nextTick, ref } from 'vue';
import Terminal from '@/components/terminal/index.vue';
import { closeOllamaModel } from '@/api/modules/ai';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const title = ref();
const open = ref(false);
const itemName = ref();
const terminalRef = ref();

interface DialogProps {
    name: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    itemName.value = params.name;
    open.value = true;
    initTerm();
};

const initTerm = () => {
    nextTick(() => {
        terminalRef.value.acceptParams({
            endpoint: '/api/v2/ai/ollama/exec',
            args: `name=${itemName.value}`,
            error: '',
            initCmd: '',
        });
    });
};

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
    open.value = false;
}

defineExpose({
    acceptParams,
});
</script>
