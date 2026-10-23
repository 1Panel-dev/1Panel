<template>
    <DialogPro
        v-model="drawerVisible"
        :title="$t(engine === 'vllm' ? 'aiTools.model.serviceDetails' : 'database.databaseConnInfo')"
        :size="isMobile ? 'full' : 'normal'"
        @close="handleClose"
    >
        <div class="connection-details">
            <div v-if="form.name" class="service-heading">
                <el-tag effect="plain">
                    {{ engine === 'ollama' ? 'Ollama' : engine === 'vllm' ? 'vLLM' : 'TensorRT LLM' }}
                </el-tag>
                <strong>{{ form.name }}</strong>
            </div>
            <section v-if="engine === 'vllm'" class="detail-card">
                <h3 class="detail-title">
                    <el-icon><FolderOpened /></el-icon>
                    {{ $t('aiTools.model.modelDir') }}
                </h3>
                <div class="detail-value model-directory">
                    <code>{{ form.modelDir || '—' }}</code>
                    <CopyButton v-if="form.modelDir" :content="form.modelDir" />
                </div>
            </section>
            <section class="detail-card">
                <h3 class="detail-title">
                    <el-icon><Connection /></el-icon>
                    {{ $t('database.containerConn') }}
                </h3>
                <div class="detail-row">
                    <span class="detail-label">{{ $t('app.containerName') }}</span>
                    <div class="detail-value">
                        <code>{{ form.containerName }}</code>
                        <CopyButton :content="form.containerName" />
                    </div>
                </div>
                <div v-if="form.containerPort" class="detail-row">
                    <span class="detail-label">{{ $t('commons.table.port') }}</span>
                    <div class="detail-value">
                        <code>{{ form.containerPort }}</code>
                        <CopyButton :content="form.containerPort" />
                    </div>
                </div>
                <p class="detail-help">{{ $t('aiTools.model.container_conn_helper') }}</p>
            </section>
            <section class="detail-card">
                <h3 class="detail-title">
                    <el-icon><Link /></el-icon>
                    {{ $t(bindDomain.connUrl ? 'setting.proxyUrl' : 'database.remoteConn') }}
                </h3>
                <div class="detail-row">
                    <span class="detail-label">{{ $t('database.connAddress') }}</span>
                    <div class="detail-value">
                        <code>{{ externalAddress }}</code>
                        <CopyButton :content="externalAddress" />
                    </div>
                </div>
                <div
                    v-if="engine !== 'ollama' && !bindDomain.connUrl && form.port && !form.endpoint"
                    class="detail-row"
                >
                    <span class="detail-label">{{ $t('commons.table.port') }}</span>
                    <div class="detail-value">
                        <code>{{ form.port }}</code>
                        <CopyButton :content="form.port + ''" />
                    </div>
                </div>
                <p class="detail-help">{{ $t('database.remoteConnHelper2') }}</p>
            </section>
            <el-button v-if="engine === 'ollama'" @click="openWebUI">OpenWebUI</el-button>
        </div>
        <template #footer>
            <el-button @click="handleClose">
                {{ $t('commons.button.close') }}
            </el-button>
        </template>
    </DialogPro>
    <PortJumpDialog ref="portRef" />
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import { Connection, FolderOpened, Link } from '@element-plus/icons-vue';
import i18n from '@/lang';
import { ElMessageBox } from 'element-plus';
import { getAppPort } from '@/api/modules/app';
import { useRouter } from 'vue-router';
import PortJumpDialog from '@/components/port-jump/index.vue';
import { getAgentSettingInfo } from '@/api/modules/setting';
import { getBindDomain } from '@/api/modules/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { currentNode, isMobile } = useGlobalStore();
const emit = defineEmits(['close']);
const router = useRouter();
const portRef = ref();
const engine = ref('ollama');

const drawerVisible = ref(false);
const form = reactive({
    name: '',
    modelDir: '',
    systemIP: '',
    containerName: '',
    port: 0 as number | string,
    containerPort: '11434',
    endpoint: '',
});
const bindDomain = ref({
    connUrl: '',
});
const externalAddress = computed(() => {
    if (bindDomain.value.connUrl) return bindDomain.value.connUrl;
    if (form.endpoint) return form.endpoint;
    if (engine.value !== 'ollama' || !form.port || !form.systemIP) return form.systemIP;
    const host = form.systemIP.includes(':') && !form.systemIP.startsWith('[') ? `[${form.systemIP}]` : form.systemIP;
    return `http://${host}:${form.port}`;
});

interface DialogProps {
    name?: string;
    modelDir?: string;
    port: number | string;
    containerPort?: string;
    endpoint?: string;
    engine?: string;
    containerName: string;
    appinstallID?: number;
}

const acceptParams = (param: DialogProps): void => {
    engine.value = param.engine || 'ollama';
    form.name = param.name || '';
    form.modelDir = param.modelDir || '';
    form.containerPort = param.containerPort ?? '11434';
    form.endpoint = param.endpoint || '';
    bindDomain.value.connUrl = '';
    form.containerName = param.containerName;
    form.port = param.port;
    loadSystemIP();
    loadBindDomain(param.appinstallID);
    drawerVisible.value = true;
};

const openWebUI = async () => {
    const node = currentNode.value;
    const { data } = await getAppPort('ollama-webui', '');
    if (node !== currentNode.value || !drawerVisible.value) return;
    if (data) {
        portRef.value.acceptParams({ port: data });
        return;
    }
    await ElMessageBox.confirm(i18n.global.t('app.checkInstalledWarn', ['OpenWebUI']), 'OpenWebUI', {
        confirmButtonText: i18n.global.t('database.goInstall'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    });
    if (node !== currentNode.value || !drawerVisible.value) return;
    await router.push({ name: 'AppAll', query: { install: 'ollama-webui' } });
};

const handleClose = () => {
    drawerVisible.value = false;
    emit('close');
};

const loadSystemIP = async () => {
    if (currentNode.value !== 'local') {
        form.systemIP = currentNode.value || i18n.global.t('database.localIP');
        return;
    }
    const res = await getAgentSettingInfo();
    form.systemIP = res.data.systemIP || i18n.global.t('database.localIP');
};

const loadBindDomain = async (appInstallID: number) => {
    if (appInstallID == undefined || appInstallID <= 0) {
        return;
    }
    try {
        const res = await getBindDomain({
            appInstallID: appInstallID,
        });
        if (res.data.websiteID > 0) {
            bindDomain.value.connUrl = res.data.connUrl;
        }
    } catch (e) {}
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.connection-details {
    display: grid;
    gap: 16px;
}
.service-heading {
    display: flex;
    align-items: center;
    gap: 12px;
    color: var(--el-text-color-primary);
    strong {
        min-width: 0;
        overflow-wrap: anywhere;
        font-size: 16px;
    }
}
.detail-card {
    padding: 18px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    background: var(--el-fill-color-extra-light);
}
.detail-title {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 16px;
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    .el-icon {
        color: var(--el-color-primary);
        font-size: 18px;
    }
}
.detail-row {
    display: grid;
    grid-template-columns: 80px minmax(0, 1fr);
    gap: 12px;
    align-items: baseline;
    & + & {
        margin-top: 12px;
    }
}
.detail-label,
.detail-help {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}
.detail-value {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    min-width: 0;
    code {
        min-width: 0;
        overflow-wrap: anywhere;
        line-height: 1.6;
        font-size: 13px;
        color: var(--el-text-color-primary);
    }
    :deep(.el-button) {
        flex-shrink: 0;
    }
}
.model-directory {
    padding: 12px;
    border-radius: 6px;
    background: var(--el-bg-color);
}
.detail-help {
    margin: 14px 0 0;
    line-height: 1.6;
}
@media (max-width: 768px) {
    .detail-row {
        grid-template-columns: minmax(0, 1fr);
        gap: 4px;
    }
}
</style>
