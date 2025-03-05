<template>
    <DrawerPro
        v-model="terminalVisible"
        :header="$t('menu.terminal')"
        @close="handleClose"
        :resource="scriptName"
        size="large"
    >
        <template #content>
            <el-alert type="error" :closable="false">
                <template #title>
                    <span>{{ $t('commons.msg.disConn', ['exit']) }}</span>
                </template>
            </el-alert>
            <Terminal style="height: calc(100vh - 235px); margin-top: 18px" ref="terminalRef"></Terminal>
        </template>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="onClose()">{{ $t('commons.button.disConn') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref, nextTick } from 'vue';
import Terminal from '@/components/terminal/index.vue';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const terminalVisible = ref(false);
const terminalRef = ref<InstanceType<typeof Terminal> | null>(null);
const scriptID = ref();
const scriptName = ref();

interface DialogProps {
    scriptID: number;
    scriptName: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    terminalVisible.value = true;
    scriptID.value = params.scriptID;
    scriptName.value = params.scriptName;
    initTerm();
};

const initTerm = async () => {
    await nextTick();
    terminalRef.value!.acceptParams({
        endpoint: '/api/v2/core/script/run',
        args: `script_id=${scriptID.value}&current_node=${globalStore.currentNode}`,
        error: '',
        initCmd: '',
    });
};

const onClose = () => {
    terminalRef.value?.onClose();
    terminalVisible.value = false;
};

function handleClose() {
    onClose();
    terminalVisible.value = false;
}

defineExpose({
    acceptParams,
});
</script>
