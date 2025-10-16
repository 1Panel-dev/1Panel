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
            <el-alert :closable="false" :title="$t('terminal.localConnJump')" type="info" />
            <Terminal class="mt-2" style="height: calc(100vh - 170px)" ref="terminalRef"></Terminal>
            <div>
                <el-cascader
                    v-model="quickCmd"
                    :options="commandTree"
                    @change="quickInput"
                    :show-all-levels="false"
                    class="w-full -mt-6"
                    placeholder=" "
                    filterable
                >
                    <template #prefix>{{ $t('terminal.quickCommand') }}</template>
                </el-cascader>
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
        args: `command=${encodeURIComponent(`clear && cd ${cwd}`)}`,
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
