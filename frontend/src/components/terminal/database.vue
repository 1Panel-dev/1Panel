<template>
    <DrawerPro
        v-model="open"
        :header="$t('menu.terminal')"
        @close="handleClose"
        :resource="resourceName"
        :autoClose="!open"
        size="large"
        :fullScreen="true"
    >
        <template #content>
            <Terminal style="height: calc(100vh - 100px)" ref="terminalRef"></Terminal>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref, nextTick } from 'vue';
import Terminal from '@/components/terminal/index.vue';

const open = ref(false);
const terminalRef = ref<InstanceType<typeof Terminal> | null>(null);
const database = ref();
const databaseType = ref();
const resourceName = ref();
const command = ref('/bin/sh');
const user = ref('');
const containerID = ref('');
const initCmd = ref('');
const waitForPrompt = ref('');

interface DialogProps {
    databaseType?: string;
    database?: string;
    resourceName?: string;
    command?: string;
    user?: string;
    containerID?: string;
    initCmd?: string;
    waitForPrompt?: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    database.value = params.database || '';
    databaseType.value = params.databaseType || '';
    resourceName.value = params.resourceName || params.database || params.containerID || '';
    command.value = params.command || '/bin/sh';
    user.value = params.user || '';
    containerID.value = params.containerID || '';
    initCmd.value = params.initCmd || '';
    waitForPrompt.value = params.waitForPrompt || '';
    open.value = false;
    await initTerm();
};

const initTerm = async () => {
    open.value = true;
    await nextTick();
    const args = containerID.value
        ? `source=container&containerid=${containerID.value}&user=${user.value}&command=${command.value}`
        : `source=database&databaseType=${databaseType.value}&database=${database.value}`;
    terminalRef.value!.acceptParams({
        endpoint: '/api/v2/hosts/terminal/container',
        args,
        error: '',
        initCmd: initCmd.value,
        waitForPrompt: waitForPrompt.value,
    });
};

function handleClose() {
    terminalRef.value?.onClose();
    open.value = false;
}

defineExpose({
    acceptParams,
});
</script>
