<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="$t(engine === 'vllm' ? 'aiTools.model.serviceDetails' : 'database.databaseConnInfo')"
        :size="isMobile ? 'full' : 'large'"
        @close="handleClose"
    >
        <div class="connection-details">
            <header class="service-identity">
                <div class="engine-icon">
                    <img :src="engine === 'ollama' ? ollamaIcon : vllmIcon" :alt="engineName" />
                </div>
                <div class="service-heading">
                    <div class="service-title">
                        <h2>{{ form.name || engineName }}</h2>
                        <Status v-if="form.status" :status="form.status" />
                    </div>
                    <div class="service-version">
                        <span>{{ engineName }}</span>
                        <span v-if="form.version">{{ $t('app.version') }} {{ form.version }}</span>
                    </div>
                </div>
            </header>
            <div class="service-intro">
                <el-icon><InfoFilled /></el-icon>
                <p>{{ $t('aiTools.model.serviceAddressHelper') }}</p>
            </div>
            <section class="detail-card">
                <h3>
                    <el-icon><Connection /></el-icon>
                    {{ $t('aiTools.model.connectionInfo') }}
                </h3>
                <div class="connection-fields">
                    <div class="connection-field primary-field">
                        <span class="field-label">
                            {{ $t(bindDomain.connUrl ? 'setting.proxyUrl' : 'aiTools.model.externalAddress') }}
                        </span>
                        <div v-loading="addressLoading" class="copy-value">
                            <code>{{ externalAddress || '—' }}</code>
                            <CopyButton v-if="externalAddress" :content="externalAddress" :is-icon="false" />
                        </div>
                        <span v-if="addressFailed" class="address-error">{{ $t('aiTools.model.loadFailed') }}</span>
                    </div>
                    <div class="connection-field">
                        <span class="field-label">{{ $t('aiTools.aiProxy.apiType') }}</span>
                        <span class="api-type">{{ engine === 'ollama' ? 'Ollama' : 'openai-completions' }}</span>
                    </div>
                    <div class="connection-field">
                        <span class="field-label">{{ $t('aiTools.model.containerAddress') }}</span>
                        <div class="copy-value">
                            <code>{{ containerAddress || '—' }}</code>
                            <CopyButton v-if="containerAddress" :content="containerAddress" :is-icon="false" />
                        </div>
                        <p class="field-help">{{ $t('aiTools.model.container_conn_helper') }}</p>
                    </div>
                </div>
            </section>
            <section class="detail-card">
                <h3>
                    <el-icon><Box /></el-icon>
                    {{ $t('aiTools.model.deploymentInfo') }}
                </h3>
                <dl class="deployment-fields">
                    <div>
                        <dt>{{ $t('app.containerName') }}</dt>
                        <dd class="copy-detail">
                            <code>{{ form.containerName || '—' }}</code>
                            <CopyButton
                                v-if="form.containerName"
                                :content="form.containerName"
                                :aria-label="$t('commons.button.copy')"
                            />
                        </dd>
                    </div>
                    <div v-if="form.version">
                        <dt>{{ $t('aiTools.model.imageVersion') }}</dt>
                        <dd>{{ form.version }}</dd>
                    </div>
                    <div>
                        <dt>{{ $t('commons.table.port') }}</dt>
                        <dd>
                            <code>{{ form.port || '—' }}</code>
                        </dd>
                    </div>
                    <div v-if="form.modelDir">
                        <dt>{{ $t('aiTools.model.modelDir') }}</dt>
                        <dd class="copy-detail">
                            <code>{{ form.modelDir }}</code>
                            <CopyButton :content="form.modelDir" :aria-label="$t('commons.button.copy')" />
                        </dd>
                    </div>
                </dl>
            </section>
        </div>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.close') }}</el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { computed, reactive, ref, watch } from 'vue';
import { Box, Connection, InfoFilled } from '@element-plus/icons-vue';
import { getAgentSettingInfo } from '@/api/modules/setting';
import { getBindDomain } from '@/api/modules/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';
import ollamaIcon from '@/assets/images/ai-providers/ollama.webp';
import vllmIcon from '@/assets/images/ai-providers/vllm.svg';

const { currentNode, isMobile } = useGlobalStore();
const emit = defineEmits(['close']);
const engine = ref<'ollama' | 'vllm'>('ollama');
const engineName = computed(() => (engine.value === 'ollama' ? 'Ollama' : 'vLLM'));
const drawerVisible = ref(false);
const addressLoading = ref(false);
const addressFailed = ref(false);
const form = reactive({
    name: '',
    status: '',
    version: '',
    modelDir: '',
    systemIP: '',
    containerName: '',
    port: 0,
    containerPort: '11434',
    endpoint: '',
});
const bindDomain = ref({ connUrl: '' });
const externalAddress = computed(() => {
    if (bindDomain.value.connUrl) return bindDomain.value.connUrl;
    if (form.endpoint) return form.endpoint;
    if (!form.port || !form.systemIP) return '';
    const host = form.systemIP.includes(':') && !form.systemIP.startsWith('[') ? `[${form.systemIP}]` : form.systemIP;
    return `http://${host}:${form.port}`;
});
const containerAddress = computed(() => {
    if (!form.containerName) return '';
    return form.containerPort ? `${form.containerName}:${form.containerPort}` : form.containerName;
});

interface DialogProps {
    name?: string;
    status?: string;
    version?: string;
    modelDir?: string;
    port: number;
    containerPort?: string;
    endpoint?: string;
    engine?: 'ollama' | 'vllm';
    containerName: string;
    appinstallID?: number;
}

let request = 0;
const acceptParams = async (param: DialogProps) => {
    const version = ++request;
    engine.value = param.engine || 'ollama';
    Object.assign(form, {
        name: param.name || '',
        status: param.status || '',
        version: param.version || '',
        modelDir: param.modelDir || '',
        systemIP: '',
        containerPort: param.containerPort ?? '11434',
        endpoint: param.endpoint || '',
        containerName: param.containerName,
        port: param.port,
    });
    bindDomain.value.connUrl = '';
    addressFailed.value = false;
    addressLoading.value = true;
    drawerVisible.value = true;
    const [settings, domain] = await Promise.allSettled([
        getAgentSettingInfo(),
        param.appinstallID > 0 ? getBindDomain({ appInstallID: param.appinstallID }) : Promise.resolve(null),
    ]);
    if (version !== request) return;
    if (settings.status === 'fulfilled') form.systemIP = settings.value.data.systemIP || '127.0.0.1';
    if (domain.status === 'fulfilled' && domain.value?.data.websiteID > 0) {
        bindDomain.value.connUrl = domain.value.data.connUrl;
    }
    addressLoading.value = false;
    addressFailed.value = settings.status === 'rejected' && !externalAddress.value;
};
const handleClose = () => {
    request++;
    drawerVisible.value = false;
    emit('close');
};
watch(currentNode, () => {
    if (drawerVisible.value) handleClose();
});
defineExpose({ acceptParams });
</script>

<style lang="scss" scoped>
.connection-details {
    display: grid;
    gap: 20px;
    color: var(--el-text-color-primary);
}
.service-identity,
.service-title,
.service-version {
    display: flex;
    align-items: center;
    gap: 12px;
}
.service-identity {
    padding-top: 4px;
}
.engine-icon {
    display: grid;
    place-items: center;
    flex-shrink: 0;
    width: 52px;
    height: 52px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    background: var(--el-fill-color-extra-light);
    img {
        width: 36px;
        height: 36px;
        object-fit: contain;
    }
}
.service-heading {
    min-width: 0;
}
.service-title {
    flex-wrap: wrap;
    gap: 8px;
    h2 {
        margin: 0;
        font-size: 20px;
        font-weight: 600;
        overflow-wrap: anywhere;
    }
}
.service-version {
    flex-wrap: wrap;
    margin-top: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
}
.service-intro {
    display: flex;
    gap: 10px;
    padding: 14px 16px;
    border-radius: 8px;
    background: var(--el-color-primary-light-9);
    .el-icon {
        flex-shrink: 0;
        margin-top: 3px;
        color: var(--el-color-primary);
    }
    p {
        margin: 0;
        font-size: 12px;
        line-height: 1.8;
    }
}
.detail-card {
    overflow: hidden;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    h3 {
        display: flex;
        align-items: center;
        gap: 8px;
        margin: 0;
        padding: 15px 20px;
        border-bottom: 1px solid var(--el-border-color-lighter);
        font-size: 14px;
        font-weight: 600;
        .el-icon {
            color: var(--el-color-primary);
            font-size: 17px;
        }
    }
}
.connection-fields {
    display: grid;
    gap: 18px;
    padding: 20px;
}
.connection-field {
    display: grid;
    gap: 8px;
    min-width: 0;
}
.field-label,
.field-help,
.deployment-fields dt {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}
.field-help {
    margin: 0;
    line-height: 1.6;
}
.copy-value {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    box-sizing: border-box;
    min-height: 44px;
    padding: 8px 12px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    background: var(--el-fill-color-extra-light);
    :deep(.el-button) {
        flex-shrink: 0;
        height: 26px;
        padding: 0 9px;
        font-size: 12px;
    }
}
.primary-field .copy-value {
    border-color: var(--el-color-primary-light-7);
    background: var(--el-color-primary-light-9);
    code {
        color: var(--el-color-primary);
    }
}
code {
    min-width: 0;
    overflow-wrap: anywhere;
    font-family: var(--el-font-family-monospace, monospace);
    font-size: 13px;
    line-height: 1.6;
}
.api-type {
    font-size: 13px;
}
.address-error {
    color: var(--el-color-danger);
    font-size: 12px;
}
.deployment-fields {
    display: grid;
    gap: 16px;
    margin: 0;
    padding: 20px;
    > div {
        display: grid;
        grid-template-columns: 100px minmax(0, 1fr);
        gap: 16px;
        align-items: baseline;
    }
    dd {
        margin: 0;
        overflow-wrap: anywhere;
        font-size: 13px;
        line-height: 1.6;
    }
}
.copy-detail {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    :deep(.el-button) {
        flex-shrink: 0;
    }
}
@media (max-width: 480px) {
    .detail-card h3,
    .connection-fields,
    .deployment-fields {
        padding: 16px;
    }
    .deployment-fields > div {
        grid-template-columns: minmax(0, 1fr);
        gap: 6px;
    }
}
</style>
