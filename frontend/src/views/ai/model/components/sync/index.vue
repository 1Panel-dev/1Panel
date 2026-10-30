<template>
    <DialogPro
        v-model="visible"
        :title="$t('commons.button.sync') + ' - ' + form.name"
        :size="isMobile ? 'full' : 'normal'"
        :show-close="!saving"
        @close="close"
    >
        <div v-loading="loading">
            <el-form v-if="form.id" ref="formRef" :model="form" label-position="top" :disabled="saving">
                <AccountAddress
                    ref="addressRef"
                    :form="form"
                    :host-port="form.hostPort"
                    :container-port="form.containerPort"
                    :host-mode="form.hostMode"
                />
            </el-form>
            <el-alert v-else-if="!loading" type="error" :closable="false" :title="$t('aiTools.model.loadFailed')" />
        </div>
        <template #footer>
            <el-button :disabled="saving" @click="close">{{ $t('commons.button.cancel') }}</el-button>
            <el-button v-permission type="primary" :loading="saving" :disabled="loading || !form.id" @click="submit">
                {{ $t('commons.button.sync') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { onBeforeUnmount, reactive, ref, watch } from 'vue';
import type { FormInstance } from 'element-plus';
import type { AI } from '@/api/interface/ai';
import { syncOllamaModelAccount } from '@/api/modules/ai';
import { checkAppInstalled } from '@/api/modules/app';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { useRoute } from 'vue-router';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import AccountAddress from '../account-address/index.vue';
import { loadVllmAPI, readPages, type Deployment } from '../../deployment';

const emit = defineEmits(['search']);
const { currentNode, isMobile } = useGlobalStore();
const route = useRoute();
const visible = ref(false);
const loading = ref(false);
const saving = ref(false);
const formRef = ref<FormInstance>();
const addressRef = ref<InstanceType<typeof AccountAddress>>();
const form = reactive({
    id: 0,
    engine: 'ollama' as 'ollama' | 'vllm',
    name: '',
    containerName: '',
    hostPort: 0,
    containerPort: '',
    hostMode: false,
    modelAccountBaseURLType: 'container' as AI.ModelAccountSync['modelAccountBaseURLType'],
    modelAccountBaseURL: '',
});
let request = 0;

const close = () => {
    request++;
    visible.value = false;
    saving.value = false;
};

const open = async (row: Deployment) => {
    if (row.engine !== 'ollama' && row.engine !== 'vllm') return;
    const version = ++request;
    form.id = 0;
    form.name = row.name;
    form.engine = row.engine;
    visible.value = true;
    loading.value = true;
    try {
        if (row.engine === 'ollama') {
            const { data: item } = await checkAppInstalled('ollama', row.name);
            if (version !== request) return;
            if (!item.isExist || item.appInstallId !== row.id) {
                MsgError(i18n.global.t('commons.res.notFound'));
                return;
            }
            Object.assign(form, {
                id: item.appInstallId,
                containerName: item.containerName,
                hostPort: item.httpPort,
                containerPort: '11434',
                hostMode: false,
                modelAccountBaseURLType: item.modelAccountBaseURLType || 'container',
                modelAccountBaseURL: item.modelAccountBaseURL || '',
            });
        } else {
            const api = await loadVllmAPI();
            if (version !== request) return;
            const items = await readPages((page) => api.pageVllm({ ...page, info: row.name }));
            if (version !== request) return;
            const item = items.find((item) => item.id === row.id);
            if (!item) {
                MsgError(i18n.global.t('commons.res.notFound'));
                return;
            }
            const hostMode = item.appVersion.trim().toLowerCase().startsWith('nvidia-gb10-dspark-');
            Object.assign(form, {
                id: item.id,
                containerName: item.containerName,
                hostPort: item.port,
                containerPort:
                    item.command.match(/--port\s+(\d+)/)?.[1] || (item.imageType === 'intel' ? '8001' : '8000'),
                hostMode,
                modelAccountBaseURLType:
                    hostMode && item.modelAccountBaseURLType === 'container'
                        ? 'systemIP'
                        : item.modelAccountBaseURLType || (hostMode ? 'systemIP' : 'container'),
                modelAccountBaseURL: item.modelAccountBaseURL || '',
            });
        }
    } finally {
        if (version === request) loading.value = false;
    }
};

const submit = async () => {
    if (loading.value || saving.value || !form.id || !formRef.value) return;
    const version = request;
    saving.value = true;
    try {
        const valid = await formRef.value.validate().catch(() => false);
        if (!valid || version !== request || !visible.value) return;
        const req: AI.ModelAccountSync = {
            id: form.id,
            modelAccountBaseURLType: form.modelAccountBaseURLType,
            modelAccountBaseURL: addressRef.value.baseURL,
        };
        if (form.engine === 'ollama') {
            await syncOllamaModelAccount(req);
        } else {
            const api = await loadVllmAPI();
            if (version !== request || !visible.value) return;
            await api.operateVllm({ ...req, operate: 'sync', taskID: '' });
        }
        if (version !== request) return;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('search');
        close();
    } finally {
        if (version === request) saving.value = false;
    }
};

watch([currentNode, () => route.fullPath], close);
onBeforeUnmount(close);
defineExpose({ open });
</script>
