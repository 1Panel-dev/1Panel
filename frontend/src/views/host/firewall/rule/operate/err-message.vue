<template>
    <DialogPro v-model="visible" :title="$t('commons.msg.operationFailed')" size="large" @close="handleDialogClose">
        <div class="error-list">
            <div
                v-for="failure in result.errors || []"
                :key="failure.index"
                class="error-item"
                :class="failure.status === 'skipped' ? 'is-skipped' : 'is-failed'"
            >
                <div class="error-item-header">
                    <span class="error-rule">{{ ruleSummary(failure) }}</span>
                    <span class="error-status">
                        {{
                            failure.status === 'skipped'
                                ? $t('commons.status.unexecuted')
                                : $t('commons.msg.operationFailed')
                        }}
                    </span>
                </div>
                <pre v-if="failureReason(failure.error)" class="error-reason">{{ failureReason(failure.error) }}</pre>
            </div>
        </div>

        <template #footer>
            <el-button type="primary" @click="close">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import { reactive, ref } from 'vue';

const emit = defineEmits<{ (event: 'close'): void }>();

const visible = ref(false);
const result = reactive<Firewall.CreateResponse>({
    succeeded: 0,
    failed: 0,
    skipped: 0,
    errors: [],
});

const wildcardAddress = (family: Firewall.Family) => {
    if (family === 'ipv6') return '::/0';
    if (family === 'inet') return '0.0.0.0/0, ::/0';
    return '0.0.0.0/0';
};

const familyLabel = (family: Firewall.Family) => (family === 'inet' ? 'IPv4/IPv6' : family.toUpperCase());

const ruleSummary = (failure: Firewall.CreateFailure) => {
    const rule = failure.rule;
    const address = rule.sourceAddress || wildcardAddress(rule.scope.family);
    return `#${failure.index + 1} · ${familyLabel(rule.scope.family)} · ${rule.protocol.toUpperCase()} · ${address} → ${rule.destinationPort || '*'}`;
};

const failureReason = (error?: string) =>
    (error || '')
        .replace(/^execute UFW rule:\s*/i, '')
        .replace(/^stderr:\s*/i, '')
        .replace(/\s*,\s*err:\s*exit status \d+\s*$/i, '')
        .trim();

const acceptParams = (data: Firewall.CreateResponse) => {
    Object.assign(result, data);
    result.errors = data.errors || [];
    visible.value = true;
};

const close = () => {
    visible.value = false;
    emit('close');
};

const handleDialogClose = () => {
    visible.value = false;
    emit('close');
};

defineExpose({ acceptParams });
</script>

<style lang="scss" scoped>
.error-list {
    max-height: 56vh;
    padding-right: 6px;
    overflow-y: auto;
}

.error-item {
    padding: 14px 4px;

    &.is-skipped {
        opacity: 0.72;
    }
}

.error-item + .error-item {
    border-top: 1px solid var(--el-border-color-lighter);
}

.error-item-header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
}

.error-rule {
    min-width: 0;
    flex: 1;
    color: var(--el-text-color-primary);
    font-weight: 500;
    line-height: 24px;
    word-break: break-word;
}

.error-status {
    padding: 3px 10px;
    flex: none;
    border: 1px solid var(--el-color-danger-light-7);
    border-radius: 12px;
    background: var(--el-color-danger-light-9);
    color: var(--el-color-danger);
    font-size: 12px;
}

.error-item.is-skipped .error-status {
    border-color: var(--el-border-color-light);
    background: var(--el-fill-color-light);
    color: var(--el-text-color-secondary);
}

.error-reason {
    padding: 0;
    margin: 8px 0 0;
    color: var(--el-color-danger);
    font-family: inherit;
    font-size: 13px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-word;
}
</style>
