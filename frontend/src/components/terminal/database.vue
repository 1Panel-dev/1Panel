<template>
    <DrawerPro
        v-model="open"
        :header="$t('menu.terminal')"
        @close="handleClose"
        :resource="database"
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

interface DialogProps {
    databaseType: string;
    database: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    database.value = params.database;
    databaseType.value = params.databaseType;
    open.value = false;
    await initTerm();
};

const initTerm = async () => {
    open.value = true;
    await nextTick();
    terminalRef.value!.acceptParams({
        endpoint: '/api/v2/containers/exec',
        args: `source=database&databaseType=${databaseType.value}&database=${database.value}`,
        error: '',
        initCmd: '',
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
