<template>
    <DrawerPro
        v-model="terminalVisible"
        :header="$t('menu.terminal')"
        @close="handleClose"
        size="large"
        :close-on-click-modal="false"
    >
        <template #content>
            <Terminal style="height: calc(100vh - 100px)" ref="terminalRef"></Terminal>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref, nextTick } from 'vue';
import Terminal from '@/components/terminal/index.vue';

const terminalVisible = ref(false);
const terminalRef = ref<InstanceType<typeof Terminal> | null>(null);

interface DialogProps {
    cwd: string;
    command: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    terminalVisible.value = true;
    await initTerm(params.cwd);
};

const initTerm = async (cwd: string) => {
    await nextTick();
    terminalRef.value!.acceptParams({
        endpoint: '/api/v2/hosts/terminal',
        args: `command=${encodeURIComponent(`clear && cd ${cwd}`)}`,
        error: '',
        initCmd: '',
    });
};

const onClose = () => {
    terminalRef.value?.onClose();
};

function handleClose() {
    onClose();
    terminalVisible.value = false;
}

defineExpose({
    acceptParams,
});
</script>
