<template>
    <component
        :is="inline ? 'div' : DrawerPro"
        v-bind="inline ? {} : { modelValue: open, header: $t('commons.button.install'), size: 'large' }"
        @update:model-value="open = $event"
        @close="handleClose"
    >
        <AppInstallForm ref="formRef" v-model="form" :loading="loading" sectioned hide-gpu-config>
            <template #after-params>
                <GPUConfig ref="gpuRef" v-model="form.gpu" />
                <details ref="syncSectionRef" open class="form-section">
                    <summary class="form-section-header">
                        <span>{{ $t('aiTools.model.syncModelAccount') }}</span>
                        <el-switch
                            @click.stop
                            @keydown.stop
                            v-model="form.syncModelAccount"
                            :aria-label="$t('aiTools.model.syncModelAccount')"
                            @change="(enabled) => enabled && syncSectionRef && (syncSectionRef.open = true)"
                        />
                    </summary>
                    <div v-show="form.syncModelAccount" class="form-section-body">
                        <AccountAddress
                            ref="accountRef"
                            show-sync
                            hide-sync-toggle
                            :form="form"
                            :host-port="Number(form.params.PANEL_APP_PORT_HTTP) || 11434"
                            container-port="11434"
                        />
                    </div>
                </details>
            </template>
        </AppInstallForm>
        <template v-if="!inline" #footer>
            <el-button :disabled="loading" @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button
                v-permission
                type="primary"
                :loading="loading"
                :disabled="loading || gpuRef?.loading || !form.appDetailId"
                @click="submit"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </component>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import { installApp } from '@/api/modules/app';
import { newUUID } from '@/utils/id';
import { MsgError } from '@/utils/message';
import i18n from '@/lang';
import DrawerPro from '@/components/drawer-pro/index.vue';
import AppInstallForm from '@/views/app-store/detail/form/index.vue';
import AccountAddress from '../../components/account-address/index.vue';
import { AI } from '@/api/interface/ai';
import GPUConfig from '../../components/gpu-config/index.vue';

const props = defineProps({ inline: Boolean });
const emit = defineEmits(['search', 'task', 'close']);
const open = ref(false);
const loading = ref(false);
const formRef = ref<InstanceType<typeof AppInstallForm>>();
const accountRef = ref();
const syncSectionRef = ref<HTMLDetailsElement>();
const gpuRef = ref<InstanceType<typeof GPUConfig>>();
const form = ref({
    gpu: {
        enabled: true,
        mode: 'all',
        deviceIDs: [],
        capabilities: ['gpu', 'utility', 'compute'],
    } as AI.ModelGPUConfig,
    appDetailId: 0,
    name: '',
    params: {} as Record<string, any>,
    containerName: '',
    syncModelAccount: true,
    modelAccountBaseURLType: 'container',
    modelAccountBaseURL: '',
});
const handleClose = () => {
    open.value = false;
    formRef.value?.resetForm();
    emit('close');
};
const openCreate = async () => {
    open.value = true;
    loading.value = true;
    try {
        await nextTick();
        if (!open.value) return;
        formRef.value.resetForm();
        formRef.value.setFormData({
            gpu: { enabled: true, mode: 'all', deviceIDs: [], capabilities: ['gpu', 'utility', 'compute'] },
            restartPolicy: 'unless-stopped',
            syncModelAccount: true,
            modelAccountBaseURLType: 'container',
            modelAccountBaseURL: '',
        });
        await formRef.value.initForm('ollama', true);
        if (!open.value) return;
        await gpuRef.value?.init(true);
        await nextTick();
        formRef.value?.clearValidate();
    } finally {
        loading.value = false;
    }
};
const submit = async () => {
    if (loading.value || gpuRef.value?.loading || !form.value.appDetailId || !(await formRef.value?.validate())) return;
    const data = formRef.value.getFormData();
    if (data.editCompose && !data.dockerCompose.trim()) {
        MsgError(i18n.global.t('app.composeNullErr'));
        return;
    }
    loading.value = true;
    const taskID = newUUID();
    try {
        await installApp({
            ...data,
            modelAccountBaseURL: accountRef.value.baseURL,
            taskID,
        });
        emit('search');
        emit('task', taskID);
        open.value = false;
    } finally {
        loading.value = false;
    }
};
watch(
    () => form.value.name,
    (name) => {
        form.value.containerName = name;
    },
);
watch(open, (value) => {
    if (props.inline && !value) handleClose();
});
defineExpose({
    openCreate,
    submit,
    loading,
    disabled: computed(() => !form.value.appDetailId || gpuRef.value?.loading),
});
</script>

<style lang="scss" scoped>
@use '@/views/ai/model/components/form-section.scss';
</style>
