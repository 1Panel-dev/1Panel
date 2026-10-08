<template>
    <div v-if="base.isExist" class="app-status card-interval">
        <el-card>
            <div class="flex w-full flex-col gap-4 md:flex-row">
                <div class="flex flex-wrap gap-4 ml-3">
                    <el-tag effect="dark" type="success">{{ base.name }}</el-tag>
                    <el-tag>{{ $t('app.version') }}: {{ base.version || '-' }}</el-tag>
                    <FamilyIssues
                        subsystem="docker"
                        :backend="backendName as Firewall.Provider"
                        :issues="familyActions"
                        @complete="emit('refresh')"
                    />
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
                            {{ $t('commons.button.reset') }}
                        </el-button>
                    </template>
                </div>
            </div>
        </el-card>
    </div>
    <NoSuchService v-else :name="backendName">
        <i18n-t keypath="firewall.selectedBackendNotInstalled" tag="span">
            <template #backend>{{ backendName }}</template>
            <template #library>
                <button type="button" class="firewall-backend-link" @click="goToScriptLibrary">
                    {{ $t('cronjob.library.library') }}
                </button>
            </template>
            <template #settings>
                <button type="button" class="firewall-backend-link" @click="goToFirewallSetting">
                    {{ $t('commons.button.set') }}
                </button>
            </template>
        </i18n-t>
    </NoSuchService>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import NoSuchService from '@/components/layout-content/no-such-service.vue';
import FamilyIssues from '@/views/host/firewall/components/family-issues.vue';
import i18n from '@/lang';
import { computed } from 'vue';
import { routerToName, routerToNameWithQuery } from '@/utils/router';
import { dockerGuardFamilyStatusMessage } from '@/views/host/firewall/docker/model';

const props = defineProps<{ base: Firewall.DockerGuardBase }>();
const emit = defineEmits<{
    operate: [operation: 'initialize' | 'bind' | 'unbind'];
    cleanup: [];
    refresh: [];
}>();
const backendName = computed(() => props.base.backend || props.base.name);
const goToFirewallSetting = () => routerToName('FirewallSetting');
const goToScriptLibrary = () => routerToNameWithQuery('Library', { uncached: 'true' });

const familyStatuses = computed(() =>
    (
        [
            { family: 'IPv4', status: props.base.ipv4 },
            { family: 'IPv6', status: props.base.ipv6 },
        ] as const
    ).filter((item) => item.family === 'IPv4' || props.base.ipv6Enabled !== false),
);
const availableFamilies = computed(() =>
    familyStatuses.value.filter((item) => item.status.reason !== 'command_missing'),
);
const anyFamilyInitialized = computed(() => availableFamilies.value.some((item) => item.status.initialized));
const anyFamilyBound = computed(() => availableFamilies.value.some((item) => item.status.bound));
const familyIssues = computed(() => {
    if (!anyFamilyInitialized.value && !familyStatuses.value.some((item) => item.status.partial)) return [];
    return familyStatuses.value.filter((item) => !item.status.effective);
});
const familyActions = computed(() =>
    familyIssues.value.map((issue) => {
        const status = issue.status;
        let action: 'initialize' | 'repair' | 'bind' | undefined;
        if (!['command_missing', 'docker_chain_missing', 'inspect_failed'].includes(status.reason || '')) {
            action = status.partial
                ? 'repair'
                : !status.initialized
                  ? 'initialize'
                  : status.reason === 'jump_missing'
                    ? 'bind'
                    : 'repair';
        }
        return {
            family: issue.family === 'IPv6' ? ('ipv6' as const) : ('ipv4' as const),
            message: status.partial
                ? i18n.global.t('firewall.familyIncomplete', [issue.family])
                : dockerGuardFamilyStatusMessage(props.base, issue.family, status),
            action,
        };
    }),
);
</script>
