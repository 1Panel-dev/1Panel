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
            <div class="terminal-container">
                <el-alert :closable="false" :title="$t('terminal.localConnJump')" type="info" />
                <Terminal class="terminal-content" ref="terminalRef"></Terminal>
                <div class="quick-command">
                    <el-cascader
                        v-model="quickCmd"
                        :options="commandTree"
                        @change="quickInput"
                        :show-all-levels="false"
                        class="w-full"
                        placeholder=" "
                        filterable
                    >
                        <template #prefix>{{ $t('terminal.quickCommand') }}</template>
                    </el-cascader>
                </div>
            </div>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref, nextTick } from 'vue';
import Terminal from '@/components/terminal/index.vue';
import { getCommandTree } from '@/api/modules/command';
import i18n from '@/lang';

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
        args: `command=${encodeURIComponent(`clear && cd "${cwd}"`)}`,
        error: '',
        initCmd: '',
    });
};

const loadCommandTree = async () => {
    const res = await getCommandTree('command');
    commandTree.value = res.data || [];
    for (const item of commandTree.value) {
        if (item.label === 'Default') {
            item.label = i18n.global.t('commons.table.default');
        }
    }
};

function quickInput(val: Array<string>) {
    if (val.length < 1) {
        return;
    }
    quickCmd.value = val[val.length - 1];
    terminalRef.value?.sendMsg(quickCmd.value + '\n');
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

<style scoped>
.terminal-container {
    display: flex;
    flex-direction: column;
    height: calc(100vh - 140px);
}

.terminal-content {
    flex: 1;
    overflow: hidden;
    margin-top: 8px;
}

.quick-command {
    flex-shrink: 0;
    margin-top: 1px;
}
</style>
