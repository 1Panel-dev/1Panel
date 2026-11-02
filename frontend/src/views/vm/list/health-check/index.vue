<template>
    <DrawerPro v-model="visible" :header="$t('xpack.vm.healthCheck')" :resource="vmName" size="large">
        <div v-loading="loading" class="vm-health-check">
            <template v-if="result">
                <el-alert :title="summaryTitle" :type="summaryType" :closable="false" show-icon />

                <div class="vm-health-summary">
                    <el-tag type="success" effect="plain">
                        {{ $t('xpack.vm.healthCheckResult.pass') }} {{ result.pass }}
                    </el-tag>
                    <el-tag type="warning" effect="plain">
                        {{ $t('xpack.vm.healthCheckResult.warn') }} {{ result.warn }}
                    </el-tag>
                    <el-tag type="danger" effect="plain">
                        {{ $t('xpack.vm.healthCheckResult.fail') }} {{ result.fail }}
                    </el-tag>
                </div>

                <div class="vm-health-items">
                    <div v-for="(item, index) in result.items" :key="`${item.check}-${index}`" class="vm-health-item">
                        <el-tag :type="levelType(item.level)" effect="plain" class="vm-health-level">
                            {{ levelLabel(item.level) }}
                        </el-tag>
                        <div class="vm-health-content">
                            <div class="vm-health-message">
                                <strong>{{ checkTitle(item) }}</strong>
                                <span>{{ itemMessage(item) }}</span>
                            </div>
                            <div v-if="item.fixCommand" class="vm-health-extra">
                                <span class="vm-health-extra__label">
                                    {{ $t('xpack.vm.healthCheckResult.fixCommand') }}
                                </span>
                                <div class="vm-health-command">
                                    <code>{{ item.fixCommand }}</code>
                                    <CopyButton :content="item.fixCommand" />
                                </div>
                            </div>
                            <div v-if="item.detail" class="vm-health-extra">
                                <span class="vm-health-extra__label">
                                    {{ $t('xpack.vm.healthCheckResult.technicalDetail') }}
                                </span>
                                <pre class="vm-health-detail">{{ item.detail }}</pre>
                            </div>
                        </div>
                    </div>
                </div>
            </template>
        </div>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import i18n from '@/lang';
import { VM } from '@/api/interface/vm';
import { healthCheckVM } from '@/api/modules/vm';

const visible = ref(false);
const loading = ref(false);
const vmName = ref('');
const result = ref<VM.HealthCheckResult>();
let requestID = 0;
let abortController: AbortController | undefined;

const summaryType = computed<'success' | 'warning' | 'error'>(() => {
    if (result.value?.fail) {
        return 'error';
    }
    if (result.value?.warn) {
        return 'warning';
    }
    return 'success';
});

const summaryTitle = computed(() => {
    if (result.value?.fail) {
        return i18n.global.t('commons.status.failed');
    }
    if (result.value?.warn) {
        return i18n.global.t('commons.status.unhealthy');
    }
    return i18n.global.t('commons.status.healthy');
});

const levelLabel = (level: VM.HealthCheckItem['level']) => {
    return i18n.global.t(`xpack.vm.healthCheckResult.${level.toLowerCase()}`);
};

const checkTitle = (item: VM.HealthCheckItem) => {
    const key = `xpack.vm.healthCheckResult.check.${item.check}`;
    const title = i18n.global.t(key);
    return item.params?.subject ? `${title} · ${item.params.subject}` : title;
};

const itemMessage = (item: VM.HealthCheckItem) => {
    if (item.level === 'PASS' || !item.reason) {
        return i18n.global.t('xpack.vm.healthCheckResult.normal');
    }
    const key = `xpack.vm.healthCheckResult.reason.${item.reason}`;
    return i18n.global.t(key, item.params || {});
};

const levelType = (level: VM.HealthCheckItem['level']): 'success' | 'warning' | 'danger' | 'info' => {
    switch (level) {
        case 'PASS':
            return 'success';
        case 'WARN':
            return 'warning';
        case 'FAIL':
            return 'danger';
        default:
            return 'info';
    }
};

const open = async (name: string) => {
    abortController?.abort();
    const currentRequestID = ++requestID;
    const controller = new AbortController();
    abortController = controller;
    vmName.value = name;
    result.value = undefined;
    visible.value = true;
    loading.value = true;
    try {
        const res = await healthCheckVM(name, controller.signal);
        if (currentRequestID !== requestID) {
            return;
        }
        result.value = res.data;
    } catch {
        if (currentRequestID === requestID) {
            result.value = undefined;
        }
    } finally {
        if (currentRequestID === requestID) {
            loading.value = false;
            abortController = undefined;
        }
    }
};

watch(visible, (value) => {
    if (!value) {
        requestID++;
        abortController?.abort();
        abortController = undefined;
        loading.value = false;
    }
});

onBeforeUnmount(() => {
    requestID++;
    abortController?.abort();
});

defineExpose({ open });
</script>

<style scoped lang="scss">
.vm-health-check {
    min-height: 240px;
    color: var(--el-text-color-primary);
    font-size: 13px;
}

.vm-health-summary {
    display: flex;
    gap: 8px;
    margin: 16px 0;
}

.vm-health-items {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.vm-health-item {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 8px 12px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    background-color: var(--el-fill-color-extra-light);
}

.vm-health-level {
    align-self: center;
    flex-shrink: 0;
}

.vm-health-content {
    flex: 1;
    min-width: 0;
}

.vm-health-message {
    display: flex;
    flex-direction: column;
    gap: 2px;
    font-size: 13px;
    line-height: 20px;
}

.vm-health-message span {
    word-break: break-all;
}

.vm-health-extra {
    margin-top: 10px;
}

.vm-health-extra__label {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.vm-health-command {
    display: flex;
    align-items: center;
    margin-top: 4px;
    padding: 8px 10px;
    border-radius: 4px;
    background-color: var(--el-fill-color);
}

.vm-health-command code {
    flex: 1;
    overflow-wrap: anywhere;
}

.vm-health-detail {
    max-height: 240px;
    margin: 4px 0 0;
    padding: 10px 12px;
    overflow: auto;
    border: 1px solid var(--el-border-color);
    border-radius: 4px;
    background-color: var(--el-fill-color);
    color: var(--el-text-color-secondary);
    font-family: var(--el-font-family);
    font-size: 11px;
    line-height: 17px;
    white-space: pre-wrap;
    word-break: break-all;
}
</style>
