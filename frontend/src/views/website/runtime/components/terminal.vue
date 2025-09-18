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
            <el-form ref="formRef" :model="form" label-position="top" @submit.prevent>
                <el-form-item :label="$t('commons.table.user')" prop="user">
                    <el-input placeholder="root" clearable v-model="form.user" @keyup.enter="reConnect">
                        <template #append>
                            <el-button @click="reConnect">{{ $t('commons.button.conn') }}</el-button>
                        </template>
                    </el-input>
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
import { reactive, ref, nextTick } from 'vue';
import Terminal from '@/components/terminal/index.vue';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const title = ref();
const terminalVisible = ref(false);
const terminalOpen = ref(false);
const form = reactive({
    isCustom: false,
    command: '',
    user: '',
    containerID: '',
});
const terminalRef = ref<InstanceType<typeof Terminal> | null>(null);

interface DialogProps {
    containerID: string;
    container: string;
    user: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    terminalVisible.value = true;
    form.containerID = params.containerID;
    title.value = params.container;
    form.isCustom = false;
    if (params.user && params.user != '') {
        form.user = params.user;
    }
    form.command = '/bin/bash';
    initTerm();
};

const initTerm = async () => {
    terminalOpen.value = true;
    await nextTick();
    terminalRef.value!.acceptParams({
        endpoint: '/api/v2/containers/exec',
        args: `source=container&containerid=${form.containerID}&user=${form.user}&command=${form.command}`,
        error: '',
        initCmd: '',
    });
};

const reConnect = async () => {
    if (terminalRef.value) {
        terminalRef.value.onClose();
    }
    terminalOpen.value = false;
    await nextTick();
    initTerm();
};

const onClose = () => {
    terminalRef.value?.onClose();
    terminalOpen.value = false;
};

function handleClose() {
    onClose();
    terminalVisible.value = false;
}

defineExpose({
    acceptParams,
});
</script>
<style lang="scss" scoped>
.terminal {
    height: calc(100vh - 180px);
}
</style>
