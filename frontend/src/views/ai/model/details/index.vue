<template>
    <DrawerPro v-model="visible" :header="$t('aiTools.model.serviceDetails')" :size="isMobile ? 'full' : 'large'">
        <div class="service-details">
            <header class="service-overview">
                <div class="service-identity">
                    <div class="provider-icon">
                        <ProviderLogo :provider="account.provider" :display-name="providerName" :size="32" />
                    </div>
                    <div class="service-heading">
                        <div class="service-title">
                            <h2>{{ account.name }}</h2>
                            <el-tag :type="status.type" size="small">{{ status.label }}</el-tag>
                        </div>
                        <div class="service-source">
                            <span>{{ providerName }}</span>
                            <span>·</span>
                            <span>
                                {{
                                    $t(
                                        account.appInstallId
                                            ? 'aiTools.model.localService'
                                            : 'aiTools.model.cloudService',
                                    )
                                }}
                            </span>
                        </div>
                    </div>
                </div>
                <div class="service-actions">
                    <el-button v-permission :icon="Edit" :disabled="verifying" @click="emit('edit')">
                        {{ $t('commons.button.edit') }}
                    </el-button>
                    <el-button :icon="Box" :disabled="account.provider === 'vllm' || verifying" @click="emit('models')">
                        {{ $t('aiTools.agents.modelPool') }}
                    </el-button>
                    <el-button
                        v-if="account.validateAvailability"
                        v-permission
                        :icon="RefreshRight"
                        :loading="verifying"
                        @click="emit('verify')"
                    >
                        {{ $t('aiTools.model.reverify') }}
                    </el-button>
                </div>
            </header>
            <section class="detail-card">
                <h3>
                    <el-icon><Connection /></el-icon>
                    {{ $t('aiTools.model.connectionInfo') }}
                </h3>
                <div class="connection-fields">
                    <div class="connection-field">
                        <span class="field-label">Base URL</span>
                        <div class="copy-value primary-value">
                            <code>{{ account.baseUrl || '—' }}</code>
                            <CopyButton v-if="account.baseUrl" :content="account.baseUrl" :is-icon="false" />
                        </div>
                    </div>
                    <div v-if="account.rememberApiKey && account.apiKey" class="connection-field">
                        <span class="field-label">API Key</span>
                        <el-input
                            :key="account.id"
                            :model-value="account.apiKey"
                            type="password"
                            readonly
                            show-password
                            autocomplete="off"
                            aria-label="API Key"
                            class="[&_input]:!font-mono"
                        >
                            <template #append>
                                <CopyButton :content="account.apiKey" :is-icon="false" />
                            </template>
                        </el-input>
                    </div>
                    <dl class="connection-options">
                        <div>
                            <dt>{{ $t('aiTools.aiProxy.apiType') }}</dt>
                            <dd>{{ account.apiType || '—' }}</dd>
                        </div>
                        <div>
                            <dt>{{ $t('terminal.authMode') }}</dt>
                            <dd>{{ authMode }}</dd>
                        </div>
                        <div v-if="deploymentRoute">
                            <dt>{{ $t('aiTools.model.deploymentInfo') }}</dt>
                            <dd>
                                <el-button link type="primary" @click="router.push(deploymentRoute)">
                                    {{
                                        $t(
                                            account.provider === 'ollama'
                                                ? 'aiTools.model.ollamaManagement'
                                                : 'aiTools.model.deployments',
                                        )
                                    }}
                                    <el-icon class="ml-1"><ArrowRight /></el-icon>
                                </el-button>
                            </dd>
                        </div>
                    </dl>
                </div>
            </section>
            <section class="detail-card">
                <h3>
                    <el-icon><Box /></el-icon>
                    {{ $t('aiTools.agents.modelPool') }}
                    <span class="model-count">{{ account.models?.length || 0 }}</span>
                </h3>
                <div v-if="account.models?.length" class="model-list">
                    <div v-for="model in account.models" :key="model.recordId || model.id" class="model-item">
                        <div class="model-description">
                            <code>{{ model.id }}</code>
                            <span v-if="model.name && model.name !== model.id" class="model-display-name">
                                {{ model.name }}
                            </span>
                        </div>
                        <el-tag
                            v-if="account.validateAvailability && model.id === verifyModel"
                            size="small"
                            type="info"
                        >
                            {{ $t('aiTools.agents.verifyModel') }}
                        </el-tag>
                        <CopyButton :content="model.id" :aria-label="$t('commons.button.copy')" />
                    </div>
                </div>
                <el-empty v-else :image-size="48" :description="$t('commons.msg.noneData')" />
            </section>
            <section class="detail-card">
                <h3>
                    <el-icon><CircleCheck /></el-icon>
                    {{ $t('aiTools.model.verificationInfo') }}
                </h3>
                <div class="verification-content">
                    <dl class="detail-fields">
                        <div>
                            <dt>{{ $t('aiTools.model.connectionStatus') }}</dt>
                            <dd>
                                <el-tag :type="status.type" size="small">{{ status.label }}</el-tag>
                            </dd>
                        </div>
                        <div>
                            <dt>{{ $t('aiTools.agents.verifyModel') }}</dt>
                            <dd>
                                <code>{{ verifyModel || '—' }}</code>
                            </dd>
                        </div>
                        <div>
                            <dt>{{ $t('aiTools.model.verifiedAt') }}</dt>
                            <dd>{{ account.verifiedAt ? dateFormatSimpleWithSecond(account.verifiedAt) : '—' }}</dd>
                        </div>
                    </dl>
                    <div v-if="account.validateAvailability && account.verifyError" class="verify-error" role="status">
                        <strong>{{ $t('aiTools.model.verifyFailed') }}</strong>
                        <p>{{ account.verifyError }}</p>
                    </div>
                </div>
            </section>
            <dl class="detail-fields service-meta">
                <div>
                    <dt>{{ $t('commons.table.createdAt') }}</dt>
                    <dd>{{ account.createdAt ? dateFormatSimpleWithSecond(account.createdAt) : '—' }}</dd>
                </div>
                <div v-if="account.remark">
                    <dt>{{ $t('commons.table.description') }}</dt>
                    <dd class="remark">{{ account.remark }}</dd>
                </div>
            </dl>
        </div>
        <template #footer>
            <el-button @click="visible = false">{{ $t('commons.button.close') }}</el-button>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import type { TagProps } from 'element-plus';
import { ArrowRight, Box, CircleCheck, Connection, Edit, RefreshRight } from '@element-plus/icons-vue';
import type { AI } from '@/api/interface/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { getAgentProviderDisplayName } from '@/utils/agent';
import { dateFormatSimpleWithSecond } from '@/utils/date';
import { hasRouteAccess } from '@/utils/rbac';
import ProviderLogo from '@/components/agent-provider-logo/index.vue';

const props = defineProps<{
    account: AI.AgentAccountItem;
    status: { type: TagProps['type']; label: string };
    verifying: boolean;
}>();
const visible = defineModel<boolean>({ default: false });
const emit = defineEmits<{ edit: []; models: []; verify: [] }>();
const router = useRouter();
const { isMobile } = useGlobalStore();
const providerName = computed(() => getAgentProviderDisplayName(props.account.provider, props.account.providerName));
const authMode = computed(() => {
    if (props.account.provider === 'gemini') return 'x-goog-api-key';
    if (props.account.apiType === 'anthropic-messages') {
        return props.account.authMode === 'bearer' ? 'Bearer Token' : props.account.authMode || 'x-api-key';
    }
    return 'Bearer Token';
});
const verifyModel = computed(() => props.account.verifyModel || props.account.models?.[0]?.id || '');
const deploymentRoute = computed(() => {
    if (!props.account.appInstallId) return;
    const name = { ollama: 'OllamaModels', vllm: 'LocalModel' }[props.account.provider];
    if (!name || !router.hasRoute(name)) return;
    const target = { name, query: { uncached: 'true' } };
    return hasRouteAccess(router.resolve(target)) ? target : undefined;
});
</script>

<style scoped lang="scss">
.service-details {
    display: grid;
    gap: 20px;
    color: var(--el-text-color-primary);
}
.service-overview {
    padding-top: 4px;
}
.service-identity,
.service-title,
.service-source,
.service-actions {
    display: flex;
    align-items: center;
    gap: 12px;
}
.provider-icon {
    display: grid;
    place-items: center;
    flex-shrink: 0;
    width: 52px;
    height: 52px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    background: var(--el-fill-color-extra-light);
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
.service-source {
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
}
.service-actions {
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 20px;
    .el-button {
        margin: 0;
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
.detail-fields dt,
.connection-options dt {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}
.copy-value {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    min-height: 44px;
    box-sizing: border-box;
    padding: 8px 12px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    :deep(.el-button) {
        flex-shrink: 0;
        height: 26px;
        padding: 0 9px;
        font-size: 12px;
    }
}
.primary-value {
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
.connection-options {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 18px;
    margin: 0;
    dd {
        margin: 8px 0 0;
        font-size: 13px;
        overflow-wrap: anywhere;
    }
}
.model-count {
    margin-left: auto;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    font-weight: 400;
}
.model-list {
    max-height: 280px;
    overflow: auto;
}
.model-item {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 12px 20px;
    & + .model-item {
        border-top: 1px solid var(--el-border-color-lighter);
    }
    .el-tag,
    :deep(.el-button) {
        flex-shrink: 0;
    }
}
.model-description {
    display: grid;
    gap: 4px;
    flex: 1;
    min-width: 0;
}
.model-display-name {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    overflow-wrap: anywhere;
}
.verification-content {
    display: grid;
    gap: 16px;
    padding: 20px;
}
.detail-fields {
    display: grid;
    gap: 16px;
    margin: 0;
    > div {
        display: grid;
        grid-template-columns: 100px minmax(0, 1fr);
        gap: 16px;
        align-items: baseline;
    }
    dd {
        margin: 0;
        font-size: 13px;
        line-height: 1.6;
        overflow-wrap: anywhere;
    }
}
.verify-error {
    padding: 12px;
    border-radius: 6px;
    background: var(--el-color-danger-light-9);
    color: var(--el-color-danger);
    font-size: 12px;
    p {
        margin: 8px 0 0;
        line-height: 1.6;
        white-space: pre-wrap;
        overflow-wrap: anywhere;
    }
}
.service-meta {
    padding: 0 4px;
}
.remark {
    white-space: pre-wrap;
}
@media (max-width: 480px) {
    .connection-options {
        grid-template-columns: minmax(0, 1fr);
    }
    .detail-card h3,
    .connection-fields,
    .verification-content {
        padding: 16px;
    }
    .model-item {
        padding: 12px 16px;
        flex-wrap: wrap;
    }
    .model-description {
        flex-basis: calc(100% - 40px);
    }
    .detail-fields > div {
        grid-template-columns: minmax(0, 1fr);
        gap: 6px;
    }
}
</style>
