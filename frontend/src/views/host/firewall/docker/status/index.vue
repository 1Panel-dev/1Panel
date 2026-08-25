<template>
    <div class="app-status card-interval">
        <el-card>
            <div class="flex w-full flex-col gap-4 md:flex-row">
                <div class="flex flex-wrap gap-4 ml-3">
                    <el-tag effect="dark" type="success">{{ base.name }}</el-tag>
                    <el-tag>{{ $t('app.version') }}: {{ base.version || '-' }}</el-tag>
                    <el-popover
                        v-if="familyIssues.length"
                        placement="bottom"
                        trigger="hover"
                        :title="$t('commons.msg.infoTitle')"
                        :width="300"
                        popper-class="docker-firewall-family-issue-popper"
                    >
                        <template #reference>
                            <el-icon class="docker-firewall-family-hint-icon" :aria-label="$t('commons.msg.infoTitle')">
                                <WarningFilled />
                            </el-icon>
                        </template>
                        <div class="docker-firewall-family-issue-list">
                            <div
                                v-for="issue in familyIssues"
                                :key="issue.family"
                                class="docker-firewall-family-issue-item"
                            >
                                {{ familyStatusText(issue.family, issue.status) }}
                            </div>
                        </div>
                        <div v-if="retryableFamilyIssues.length" class="docker-firewall-family-issue-footer">
                            <el-button
                                v-permission
                                v-node-admin
                                size="small"
                                type="primary"
                                @click.stop="emit('operate', familyRetryOperation)"
                            >
                                {{ $t('commons.button.retry') }}
                            </el-button>
                        </div>
                    </el-popover>
                </div>
                <div class="mt-0.5 flex items-center">
                    <el-divider v-if="anyFamilyBound" direction="vertical" />
                    <el-button
                        v-if="!anyFamilyInitialized"
                        v-permission
                        v-node-admin
                        type="primary"
                        link
                        @click="emit('operate', 'initialize')"
                    >
                        {{ $t('commons.button.init') }}
                    </el-button>
                    <el-button
                        v-else-if="!anyFamilyBound"
                        v-permission
                        v-node-admin
                        type="primary"
                        link
                        @click="emit('operate', 'bind')"
                    >
                        {{ $t('commons.button.bind') }}
                    </el-button>
                    <el-button v-else v-permission v-node-admin type="primary" link @click="emit('operate', 'unbind')">
                        {{ $t('commons.button.unbind') }}
                    </el-button>
                    <template v-if="anyFamilyInitialized">
                        <el-divider direction="vertical" />
                        <el-button v-permission v-node-admin type="primary" link @click="emit('cleanup')">
                            {{ $t('firewall.cleanupAction') }}
                        </el-button>
                    </template>
                </div>
            </div>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import i18n from '@/lang';
import { WarningFilled } from '@element-plus/icons-vue';
import { computed } from 'vue';

const props = defineProps<{ base: Firewall.DockerGuardBase }>();
const emit = defineEmits<{
    operate: [operation: 'initialize' | 'bind' | 'unbind'];
    cleanup: [];
}>();

const familyStatuses = computed(
    () =>
        [
            { family: 'IPv4', status: props.base.ipv4 },
            { family: 'IPv6', status: props.base.ipv6 },
        ] as const,
);
const availableFamilies = computed(() =>
    familyStatuses.value.filter((item) => item.status.reason !== 'command_missing'),
);
const anyFamilyInitialized = computed(() => availableFamilies.value.some((item) => item.status.initialized));
const anyFamilyBound = computed(() => availableFamilies.value.some((item) => item.status.bound));
const familyIssues = computed(() => {
    if (!anyFamilyBound.value) return [];
    return familyStatuses.value.filter((item) => !item.status.effective);
});
const retryableFamilyIssues = computed(() =>
    familyIssues.value.filter(
        (item) => item.status.reason !== 'command_missing' && item.status.reason !== 'docker_chain_missing',
    ),
);
const familyRetryOperation = computed<'initialize' | 'bind'>(() =>
    retryableFamilyIssues.value.some((item) => !item.status.initialized) ? 'initialize' : 'bind',
);
const dockerChainName = computed(() => (props.base.backend === 'nftables' ? 'NFT_1PANEL_DOCKER' : '1PANEL_DOCKER'));
const familyStatusText = (family: string, status: Firewall.DockerGuardFamilyStatus) => {
    if (
        status.reason === 'command_missing' ||
        status.reason === 'docker_chain_missing' ||
        status.reason === 'inspect_failed'
    ) {
        return i18n.global.t(`firewall.dockerGuardStatusReason.${status.reason}`, [family]);
    }
    const state = i18n.global.t(
        !status.initialized
            ? 'firewall.notInitialized'
            : !status.bound
              ? 'commons.status.unbind'
              : 'firewall.notEffective',
    );
    return i18n.global.t('firewall.familyChainIssue', [family, dockerChainName.value, state]);
};
</script>

<style lang="scss">
.docker-firewall-family-hint-icon {
    align-self: center;
    color: var(--el-color-warning);
    font-size: 16px;
}

.docker-firewall-family-issue-popper.el-popover {
    padding: 12px;
    border-color: var(--el-color-warning-light-7);
    border-radius: 8px;
    box-shadow: var(--el-box-shadow-light);
}

.docker-firewall-family-issue-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.docker-firewall-family-issue-item {
    color: var(--el-text-color-regular);
    font-size: 13px;
    line-height: 20px;
}

.docker-firewall-family-issue-footer {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
    padding-top: 10px;
    border-top: 1px solid var(--el-border-color-lighter);
}
</style>
