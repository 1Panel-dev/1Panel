<template>
    <el-popover
        v-if="issues.length"
        placement="bottom-start"
        trigger="click"
        :width="360"
        :popper-options="{ modifiers: [{ name: 'preventOverflow', options: { padding: 16 } }] }"
        popper-class="firewall-family-popover"
    >
        <template #reference>
            <el-button class="family-warning" text :aria-label="$t('commons.msg.infoTitle')">
                <el-icon><WarningFilled /></el-icon>
            </el-button>
        </template>
        <div v-for="issue in issues" :key="issue.family" class="family-row">
            <div class="family-row__body">
                <el-tag size="small" type="info" effect="plain">{{ issue.family === 'ipv6' ? 'IPv6' : 'IPv4' }}</el-tag>
                <p>{{ issue.message }}</p>
            </div>
            <el-button
                v-if="issue.action"
                v-permission
                v-node-admin
                size="small"
                type="primary"
                plain
                :loading="running === issue.family"
                :disabled="disabled || (!!running && running !== issue.family)"
                @click="operate(issue)"
            >
                {{
                    $t(
                        issue.action === 'bind'
                            ? 'commons.button.bind'
                            : issue.action === 'repair'
                              ? 'firewall.familyRepair'
                              : 'commons.button.init',
                    )
                }}
            </el-button>
        </div>
    </el-popover>
    <TaskLog ref="taskLogRef" @close="complete" />
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { WarningFilled } from '@element-plus/icons-vue';
import { Firewall } from '@/api/interface/firewall';
import { operateFirewallFamily } from '@/api/modules/firewall';
import TaskLog from '@/components/log/task/index.vue';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

interface FamilyIssue {
    family: 'ipv4' | 'ipv6';
    message: string;
    action?: 'initialize' | 'repair' | 'bind';
}
const props = defineProps<{
    subsystem: Firewall.BackendSubsystem;
    backend: Firewall.Provider;
    issues: FamilyIssue[];
    disabled?: boolean;
}>();
const emit = defineEmits<{ complete: [] }>();
const running = ref('');
const taskLogRef = ref<InstanceType<typeof TaskLog>>();
const complete = () => {
    running.value = '';
    emit('complete');
};
const operate = async (issue: FamilyIssue) => {
    if (running.value || props.disabled || !issue.action) return;
    running.value = issue.family;
    try {
        const result = (
            await operateFirewallFamily({
                subsystem: props.subsystem,
                backend: props.backend,
                family: issue.family,
                operation: issue.action,
            })
        ).data;
        if (result.queued && result.taskID) {
            taskLogRef.value?.openWithTaskID(result.taskID, true);
        } else {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            complete();
        }
    } catch (error) {
        complete();
        throw error;
    }
};
</script>

<style scoped lang="scss">
.family-warning {
    color: var(--el-color-warning);
    padding: 6px;
    height: 28px;
    align-self: center;
    border-radius: 6px;
    font-size: 17px;
}
.family-row {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 14px 0;
    & + & {
        border-top: 1px solid var(--el-border-color-lighter);
    }
    &:first-child {
        padding-top: 0;
    }
    &:last-child {
        padding-bottom: 0;
    }
    &__body {
        flex: 1;
        min-width: 0;
        p {
            margin: 7px 0 0;
            color: var(--el-text-color-regular);
            font-size: 12px;
            line-height: 1.6;
            overflow-wrap: anywhere;
        }
    }
}
</style>
<style lang="scss">
.el-popover.firewall-family-popover {
    padding: 16px;
    max-width: calc(100vw - 32px);
    box-sizing: border-box;
    border-radius: 8px;
}
</style>
