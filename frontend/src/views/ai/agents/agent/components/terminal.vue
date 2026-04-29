<template>
    <DrawerPro
        v-model="terminalVisible"
        :header="$t('menu.terminal')"
        @close="handleClose"
        :resource="title"
        fullScreen
        :size="globalStore.isFullScreen ? 'full' : 'large'"
        :autoClose="false"
    >
        <template #content>
            <el-form label-position="top" @submit.prevent>
                <el-form-item :label="$t('commons.table.user')" prop="user">
                    <el-select v-model="form.user" :disabled="form.users.length <= 1" @change="reConnect">
                        <el-option v-for="item in form.users" :key="item" :label="item" :value="item" />
                    </el-select>
                </el-form-item>
            </el-form>
            <Terminal class="terminal" ref="terminalRef" v-if="terminalOpen"></Terminal>
        </template>
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
import { nextTick, reactive, ref } from 'vue';
import Terminal from '@/components/terminal/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { globalStore, currentNode } = useGlobalStore();

const title = ref('');
const terminalVisible = ref(false);
const terminalOpen = ref(false);
const terminalRef = ref<InstanceType<typeof Terminal> | null>(null);
const form = reactive({
    containerID: '',
    users: [] as string[],
    user: '',
    shell: '',
    initCmd: '',
    node: '',
});

interface DialogProps {
    containerID: string;
    title: string;
    users: string[];
    shell: string;
    initCmd?: string;
    node?: string;
}

const acceptParams = async (params: DialogProps): Promise<void> => {
    terminalVisible.value = true;
    form.containerID = params.containerID;
    form.users = [...params.users];
    form.user = form.users[0];
    form.shell = params.shell;
    form.initCmd = params.initCmd || '';
    form.node = params.node || currentNode.value;
    title.value = params.title;
    await reConnect();
};

const initTerm = async () => {
    terminalOpen.value = true;
    await nextTick();
    let args = `source=container&containerid=${form.containerID}&user=${form.user}&command=${form.shell}`;
    if (form.node) {
        args += `&operateNode=${form.node}`;
    }
    terminalRef.value?.acceptParams({
        endpoint: '/api/v2/containers/exec',
        args,
        error: '',
        initCmd: form.initCmd,
    });
};

const reConnect = async () => {
    terminalRef.value?.onClose();
    terminalOpen.value = false;
    await nextTick();
    await initTerm();
};

const onClose = () => {
    terminalRef.value?.onClose();
    terminalOpen.value = false;
};

const handleClose = () => {
    onClose();
    terminalVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.terminal {
    height: calc(100vh - 237px);
}
</style>
