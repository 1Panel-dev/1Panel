<template>
    <DialogPro v-model="open" :title="$t('menu.terminal')" size="w-70" @close="handleClose">
        <Terminal class="mt-2" style="height: 55vh" ref="terminalRef"></Terminal>

        <template #footer>
            <span class="dialog-footer">
                <el-button type="primary" @click="handleClose">
                    {{ $t('commons.button.disConn') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { nextTick, onBeforeUnmount, ref } from 'vue';
import Terminal from '@/components/terminal/index.vue';
import { closeOllamaModel } from '@/api/modules/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { currentNode } = useGlobalStore();
const open = ref(false);
const itemName = ref();
const terminalRef = ref();
let operateNode = '';

interface DialogProps {
    name: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    operateNode = currentNode.value;
    itemName.value = params.name;
    open.value = true;
    initTerm();
};

const initTerm = () => {
    nextTick(() => {
        if (!open.value) return;
        terminalRef.value?.acceptParams({
            endpoint: '/api/v2/hosts/terminal/container',
            args: `source=ollama&name=${itemName.value}`,
            error: '',
            initCmd: '',
        });
    });
};

const onClose = async () => {
    const name = itemName.value;
    if (!name) return;
    itemName.value = '';
    await closeOllamaModel(name, operateNode)
        .then(() => {
            terminalRef.value?.onClose();
        })
        .catch(() => {
            terminalRef.value?.onClose();
        });
};

function handleClose() {
    open.value = false;
    onClose();
}
onBeforeUnmount(handleClose);

defineExpose({
    acceptParams,
});
</script>
