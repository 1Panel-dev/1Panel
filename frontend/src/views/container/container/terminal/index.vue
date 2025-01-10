<template>
    <el-drawer
        v-model="terminalVisible"
        @close="handleClose"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.containerTerminal')" :resource="title" :back="handleClose" />
        </template>
        <el-form ref="formRef" :model="form" label-position="top">
            <el-form-item :label="$t('commons.table.user')" prop="user">
                <el-input placeholder="root" clearable v-model="form.user" />
            </el-form-item>
            <el-form-item
                v-if="form.isCustom"
                :label="$t('container.command')"
                prop="command"
                :rules="Rules.requiredInput"
            >
                <el-checkbox style="width: 180px" border v-model="form.isCustom" @change="onChangeCommand">
                    {{ $t('container.custom') }}
                </el-checkbox>
                <el-input style="width: calc(100% - 180px)" clearable v-model="form.command" />
            </el-form-item>
            <el-form-item
                v-if="!form.isCustom"
                :label="$t('container.command')"
                prop="command"
                :rules="Rules.requiredSelect"
            >
                <el-checkbox style="width: 180px" border v-model="form.isCustom" @change="onChangeCommand">
                    {{ $t('container.custom') }}
                </el-checkbox>
                <el-select style="width: calc(100% - 180px)" filterable clearable v-model="form.command">
                    <el-option value="/bin/ash" label="/bin/ash" />
                    <el-option value="/bin/bash" label="/bin/bash" />
                    <el-option value="/bin/sh" label="/bin/sh" />
                </el-select>
            </el-form-item>

            <el-button v-if="!terminalOpen" @click="initTerm(formRef)">
                {{ $t('commons.button.conn') }}
            </el-button>
            <el-button v-else @click="onClose()">{{ $t('commons.button.disconnect') }}</el-button>
            <Terminal
                style="height: calc(100vh - 302px); margin-top: 18px"
                ref="terminalRef"
                v-if="terminalOpen"
            ></Terminal>
        </el-form>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref, nextTick } from 'vue';
import { ElForm, FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import Terminal from '@/components/terminal/index.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';

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
    form.command = '/bin/sh';
    terminalOpen.value = false;
};

const onChangeCommand = async () => {
    form.command = '';
};

const initTerm = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        terminalOpen.value = true;
        await nextTick();
        terminalRef.value!.acceptParams({
            endpoint: '/api/v1/containers/exec',
            args: `containerid=${form.containerID}&user=${form.user}&command=${form.command}`,
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
