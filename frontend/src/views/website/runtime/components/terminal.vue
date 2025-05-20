<template>
    <DrawerPro
        v-model="terminalVisible"
        :header="$t('menu.terminal')"
        @close="handleClose"
        :resource="title"
        size="large"
    >
        <template #content>
            <el-form ref="formRef" :model="form" label-position="top">
                <el-button v-if="!terminalOpen" @click="initTerm(formRef)">
                    {{ $t('commons.button.conn') }}
                </el-button>
                <el-button v-else @click="onClose()">{{ $t('commons.button.disConn') }}</el-button>
                <Terminal
                    style="height: calc(100vh - 200px); margin-top: 18px"
                    ref="terminalRef"
                    v-if="terminalOpen"
                ></Terminal>
            </el-form>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref, nextTick } from 'vue';
import { ElForm, FormInstance } from 'element-plus';
import Terminal from '@/components/terminal/index.vue';

const title = ref();
const terminalVisible = ref(false);
const terminalOpen = ref(false);
const form = reactive({
    isCustom: false,
    command: '',
    user: '',
    containerID: '',
});
const formRef = ref();
const terminalRef = ref<InstanceType<typeof Terminal> | null>(null);

interface DialogProps {
    containerID: string;
    container: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    terminalVisible.value = true;
    form.containerID = params.containerID;
    title.value = params.container;
    form.isCustom = false;
    form.user = '';
    form.command = '/bin/bash';
    terminalOpen.value = false;
    initTerm(formRef.value);
};

const initTerm = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        terminalOpen.value = true;
        await nextTick();
        terminalRef.value!.acceptParams({
            endpoint: '/api/v2/containers/exec',
            args: `source=container&containerid=${form.containerID}&user=${form.user}&command=${form.command}`,
            error: '',
            initCmd: '',
        });
    });
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
