<template>
    <DrawerPro
        v-model="visible"
        :header="$t(enginesOnly ? 'aiTools.model.addEngine' : 'aiTools.model.addService')"
        :size="isMobile ? 'full' : '50%'"
    >
        <div v-if="selected !== 'account' && !vllmOnly" class="engine-options">
            <el-button
                class="engine-option"
                :class="{ 'is-selected': selected === 'vllm' }"
                :aria-pressed="selected === 'vllm'"
                @click="selected = 'vllm'"
            >
                <ProviderLogo provider="vllm" :size="18" />
                <span class="engine-content">
                    <span class="engine-name">vLLM</span>
                    <span class="engine-description">{{ $t('aiTools.model.instanceDeployment') }}</span>
                </span>
            </el-button>
            <el-button
                v-if="!vllmOnly"
                class="engine-option"
                :class="{ 'is-selected': selected === 'ollama' }"
                :aria-pressed="selected === 'ollama'"
                @click="selected = 'ollama'"
            >
                <ProviderLogo provider="ollama" :size="18" />
                <span class="engine-content">
                    <span class="engine-name">Ollama</span>
                    <span class="engine-description" :class="{ 'engine-running': ollama?.status === 'Running' }">
                        {{
                            ollama?.isExist
                                ? $t('commons.status.' + ollama.status.toLowerCase())
                                : $t('aiTools.model.modelManagement')
                        }}
                    </span>
                </span>
            </el-button>
            <el-button
                v-if="isFxplay && !vllmOnly"
                class="engine-option"
                :class="{ 'is-selected': selected === 'tensorrt' }"
                :aria-pressed="selected === 'tensorrt'"
                @click="selected = 'tensorrt'"
            >
                <ProviderLogo provider="nvidia" :size="18" />
                <span class="engine-content">
                    <span class="engine-name">TensorRT LLM</span>
                    <span class="engine-description">{{ $t('aiTools.model.nvidiaDeployment') }}</span>
                </span>
            </el-button>
        </div>
        <el-divider v-if="selected !== 'account' && !vllmOnly" />
        <DockerStatus
            v-if="visible && selected !== 'account'"
            v-model:is-active="dockerActive"
            v-model:is-exist="dockerExists"
        />
        <div v-loading="initializing" class="service-form">
            <el-alert v-if="loadFailed" :title="$t('aiTools.model.loadFailed')" type="error" :closable="false" />
            <AccountForm
                v-if="activeForm === 'account'"
                ref="accountRef"
                inline
                @search="emit('search')"
                @close="visible = false"
            />
            <component
                :is="vllmForm"
                v-if="activeForm === 'vllm' && vllmForm"
                ref="vllmRef"
                inline
                @search="emit('search')"
                @close="visible = false"
                @task="openTask"
            />
            <OllamaForm
                v-if="activeForm === 'ollama'"
                ref="ollamaRef"
                inline
                @search="emit('search')"
                @close="visible = false"
                @task="openTask"
            />
            <TensorForm
                v-if="activeForm === 'tensorrt'"
                ref="tensorRef"
                inline
                @search="emit('search')"
                @close="visible = false"
                @task="openTask"
            />
            <template v-if="selected === 'vllm' && (!isProductPro || !vllmAvailable)">
                <el-alert :title="$t('license.vllm')" type="info" :closable="false" />
                <el-button class="mt-4" type="primary" @click="openVllmLicense">
                    {{ $t('license.levelUpPro') }}
                </el-button>
            </template>
            <section v-if="selected === 'ollama' && ollama?.isExist" class="installed-engine">
                <div
                    class="installed-engine-overview"
                    role="link"
                    tabindex="0"
                    :aria-label="$t('aiTools.model.modelManagement')"
                    @click="openOllama"
                    @keydown.enter.self.prevent="openOllama"
                    @keydown.space.self.prevent="openOllama"
                >
                    <div class="installed-engine-heading">
                        <ProviderLogo provider="ollama" :size="40" />
                        <div class="installed-engine-identity">
                            <h3>{{ ollama.name }}</h3>
                            <span>Ollama · {{ ollama.version }}</span>
                        </div>
                        <Status :status="ollama.status" />
                    </div>
                    <div class="installed-engine-actions" @click.stop>
                        <el-button type="primary" plain @click="openOllamaConnection">
                            {{ $t('database.databaseConnInfo') }}
                        </el-button>
                        <el-button :loading="openingWebUI" @click="openWebUI">OpenWebUI</el-button>
                        <el-button @click="openOllama">
                            {{ $t('aiTools.model.modelManagement') }}
                        </el-button>
                    </div>
                </div>
                <h4 class="installed-engine-title">{{ $t('aiTools.model.serviceDetails') }}</h4>
                <el-descriptions :column="1" border>
                    <el-descriptions-item :label="$t('app.containerName')">
                        <div class="installed-engine-value">
                            <code>{{ ollama.containerName }}</code>
                            <CopyButton :content="ollama.containerName" />
                        </div>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('commons.table.port')">
                        {{ ollama.httpPort || '—' }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('aiTools.model.endpoint')">
                        <div class="installed-engine-value">
                            <code>{{ ollama.modelAccountBaseURL || '—' }}</code>
                            <CopyButton v-if="ollama.modelAccountBaseURL" :content="ollama.modelAccountBaseURL" />
                        </div>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('xpack.node.baseDir')">
                        <div class="installed-engine-value">
                            <code>{{ ollama.installPath || '—' }}</code>
                            <CopyButton v-if="ollama.installPath" :content="ollama.installPath" />
                        </div>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('commons.table.createdAt')">
                        {{ ollama.createdAt ? dateFormatSimple(ollama.createdAt) : '—' }}
                    </el-descriptions-item>
                </el-descriptions>
            </section>
        </div>
        <template v-if="activeForm" #footer>
            <el-button :disabled="currentForm?.loading" @click="visible = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button
                v-permission
                type="primary"
                :loading="initializing || currentForm?.loading"
                :disabled="initializing || loadFailed || !currentForm || currentForm.disabled"
                @click="currentForm.submit()"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
        <template v-else-if="selected === 'ollama' && ollama?.isExist" #footer>
            <el-button @click="visible = false">{{ $t('commons.button.close') }}</el-button>
        </template>
    </DrawerPro>
    <EngineConnection ref="connectionRef" />
    <PortJumpDialog ref="portRef" />
    <AccountForm ref="editAccountRef" @search="emit('search')" />
    <TaskLog ref="taskRef" @close="emit('search')" />
</template>

<script setup lang="ts">
import { computed, nextTick, ref, shallowRef, watch } from 'vue';
import { ElMessageBox } from 'element-plus';
import { useRouter } from 'vue-router';
import ProviderLogo from '@/components/agent-provider-logo/index.vue';
import type { AI } from '@/api/interface/ai';
import type { App } from '@/api/interface/app';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { loadOptionalComponent } from '@/extensions/optional';
import AccountForm from '../account-form/index.vue';
import TensorForm from '@/views/ai/model/tensorrt/operate/index.vue';
import OllamaForm from '@/views/ai/model/ollama/operate/index.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import { checkAppInstalled, getAppPort } from '@/api/modules/app';
import { dateFormatSimple } from '@/utils/date';
import i18n from '@/lang';
import EngineConnection from '../connection/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import { vllmAvailable } from '../../deployment';

const props = defineProps({ enginesOnly: Boolean, vllmOnly: Boolean });
const emit = defineEmits(['search', 'manage']);
const { isProductPro, isFxplay, isMobile, currentNode } = useGlobalStore();
const router = useRouter();
const visible = ref(false);
const selected = ref<'account' | 'vllm' | 'ollama' | 'tensorrt'>('account');
const modelDir = ref('');
const dockerActive = ref(false);
const dockerExists = ref(true);
const ollama = ref<App.CheckInstalled>();
const accountRef = ref();
const editAccountRef = ref();
const tensorRef = ref();
const ollamaRef = ref();
const vllmRef = ref();
const taskRef = ref();
const connectionRef = ref();
const portRef = ref();
const openingWebUI = ref(false);
const vllmForm = shallowRef();
const initializing = ref(false);
const loadFailed = ref(false);
const activeForm = computed(() => {
    if (!visible.value) return '';
    if (selected.value === 'account') return 'account';
    if (!dockerActive.value || !dockerExists.value) return '';
    if (selected.value === 'vllm' && isProductPro.value && vllmAvailable) return 'vllm';
    if (selected.value === 'ollama' && ollama.value && !ollama.value.isExist) return 'ollama';
    if (selected.value === 'tensorrt' && isFxplay.value) return 'tensorrt';
    return '';
});

const currentForm = computed(() => {
    if (activeForm.value === 'account') return accountRef.value;
    if (activeForm.value === 'vllm') return vllmRef.value;
    if (activeForm.value === 'ollama') return ollamaRef.value;
    if (activeForm.value === 'tensorrt') return tensorRef.value;
    return undefined;
});

watch(
    activeForm,
    async (form, _, onCleanup) => {
        let cancelled = false;
        onCleanup(() => (cancelled = true));
        loadFailed.value = false;
        initializing.value = Boolean(form);
        if (!form) return;
        try {
            if (form === 'vllm' && !vllmForm.value) {
                const component = await loadOptionalComponent('/src/xpack/views/vllm/operate/index.vue');
                if (cancelled) return;
                vllmForm.value = component;
            }
            await nextTick();
            if (cancelled) return;
            if (form === 'account') await accountRef.value.open();
            if (form === 'vllm') await vllmRef.value.openCreate(modelDir.value);
            if (form === 'ollama') await ollamaRef.value.openCreate();
            if (form === 'tensorrt') tensorRef.value.openCreate();
        } catch {
            if (!cancelled) loadFailed.value = true;
        } finally {
            if (!cancelled) initializing.value = false;
        }
    },
    { flush: 'post' },
);

const open = async (engine?: 'ollama' | 'vllm' | 'tensorrt', defaultModelDir = '') => {
    loadFailed.value = false;
    modelDir.value = defaultModelDir;
    selected.value = props.vllmOnly ? 'vllm' : engine || (props.enginesOnly ? 'vllm' : 'account');
    dockerActive.value = false;
    dockerExists.value = true;
    ollama.value = undefined;
    visible.value = true;
    if (selected.value === 'account' || props.vllmOnly) return;
    const { data } = await checkAppInstalled('ollama', '');
    if (visible.value) ollama.value = data;
};

const openAccount = (account: AI.AgentAccountItem) => {
    editAccountRef.value.open({ ...account, baseURL: account.baseUrl });
};

const openOllama = async () => {
    if (!ollama.value) return;
    visible.value = false;
    emit('manage', { engine: 'ollama', id: ollama.value.appInstallId, name: ollama.value.name });
};

const openOllamaConnection = () => {
    if (!ollama.value?.isExist) return;
    connectionRef.value.acceptParams({
        name: ollama.value.name,
        engine: 'ollama',
        port: ollama.value.httpPort,
        containerName: ollama.value.containerName,
        appinstallID: ollama.value.appInstallId,
    });
};

const openWebUI = async () => {
    const node = currentNode.value;
    openingWebUI.value = true;
    try {
        const { data } = await getAppPort('ollama-webui', '');
        if (node !== currentNode.value || !visible.value || selected.value !== 'ollama') return;
        if (data) {
            await portRef.value.acceptParams({ port: data });
            return;
        }
        const confirmed = await ElMessageBox.confirm(
            i18n.global.t('app.checkInstalledWarn', ['OpenWebUI']),
            'OpenWebUI',
            {
                confirmButtonText: i18n.global.t('database.goInstall'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            },
        ).catch(() => false);
        if (!confirmed || node !== currentNode.value || !visible.value || selected.value !== 'ollama') return;
        visible.value = false;
        await router.push({ name: 'AppAll', query: { install: 'ollama-webui' } });
    } finally {
        openingWebUI.value = false;
    }
};

const openVllmLicense = () => {
    visible.value = false;
    emit('manage', { engine: 'vllm', id: 0, name: '' });
};
const openTask = (taskID: string) => taskRef.value.openWithTaskID(taskID);

defineExpose({ open, openAccount });
</script>

<style scoped lang="scss">
.engine-options {
    display: grid;
    grid-auto-flow: column;
    grid-auto-columns: minmax(0, 1fr);
    gap: 6px;
}
.engine-option {
    min-width: 0;
    height: auto;
    margin: 0;
    padding: 10px 8px;
    border-radius: 6px;
    text-align: left;
    white-space: normal;
    :deep(> span) {
        width: 100%;
        display: flex;
        align-items: flex-start;
        gap: 6px;
        min-width: 0;
    }
    .el-icon {
        flex-shrink: 0;
        color: var(--el-color-primary);
        margin-top: 2px;
    }
    &.is-selected {
        border-color: var(--el-color-primary);
        background: var(--el-color-primary-light-9);
    }
}
.engine-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
}
.engine-name {
    font-size: 12px;
    font-weight: 600;
}
.engine-description {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.4;
}
.engine-name,
.engine-description {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.engine-running {
    color: var(--el-color-success);
}
.service-form {
    min-height: 100px;
}
.installed-engine-overview {
    padding: 20px;
    border: 1px solid var(--el-border-color-light);
    border-radius: 8px;
    background: var(--el-fill-color-extra-light);
    cursor: pointer;
    &:hover {
        border-color: var(--el-color-primary);
    }
    &:focus-visible {
        outline: 2px solid var(--el-color-primary);
        outline-offset: 2px;
    }
}
.installed-engine-heading {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 14px;
}
.installed-engine-identity {
    flex: 1;
    min-width: 0;
    h3 {
        margin: 0 0 6px;
        font-size: 16px;
        overflow-wrap: anywhere;
    }
    span {
        color: var(--el-text-color-secondary);
        font-size: 12px;
    }
}
.installed-engine-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 20px;
    .el-button {
        margin-left: 0;
    }
}
.installed-engine-title {
    margin: 24px 0 12px;
    font-size: 14px;
    font-weight: 600;
}
.installed-engine {
    :deep(.el-descriptions__table) {
        table-layout: fixed;
    }
    :deep(.el-descriptions__label) {
        width: 110px;
    }
}
.installed-engine-value {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    code {
        min-width: 0;
        overflow-wrap: anywhere;
    }
    :deep(.el-button) {
        flex-shrink: 0;
    }
}
@media (max-width: 600px) {
    .engine-options {
        grid-auto-flow: row;
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }
}
</style>
