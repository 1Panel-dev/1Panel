<template>
    <DrawerPro
        v-model="terminalVisible"
        :header="$t('menu.terminal')"
        @close="handleClose"
        size="large"
        :autoClose="false"
        :fullScreen="true"
    >
        <template #content>
            <Terminal style="height: calc(100vh - 120px)" ref="terminalRef"></Terminal>
            <div>
                <el-select v-model="quickCmd" clearable filterable @change="quickInput" class="w-full -mt-6">
                    <template #prefix>{{ $t('terminal.quickCommand') }}</template>
                    <el-option-group v-for="group in commandTree" :key="group.label" :label="group.label">
                        <el-option
                            v-for="(cmd, index) in group.children"
                            :key="index"
                            :label="cmd.name"
                            :value="cmd.command"
                        />
                    </el-option-group>
                </el-select>
            </div>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref, nextTick } from 'vue';
import Terminal from '@/components/terminal/index.vue';
import { getCommandTree } from '@/api/modules/command';

const terminalVisible = ref(false);
const terminalRef = ref<InstanceType<typeof Terminal> | null>(null);

let quickCmd = ref();
const commandTree = ref();

interface DialogProps {
    cwd: string;
    command: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    terminalVisible.value = true;
    loadCommandTree();
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

const loadCommandTree = async () => {
    const res = await getCommandTree('command');
    commandTree.value = res.data || [];
};

function quickInput(val: any) {
    terminalRef.value?.sendMsg(val + '\n');
    quickCmd.value = '';
}

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
