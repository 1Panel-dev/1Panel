<template>
    <div class="app-status card-interval">
        <el-card>
            <div class="flex w-full flex-col gap-4 md:flex-row">
                <div class="flex flex-wrap gap-4 ml-3">
                    <el-tag effect="dark" type="success">{{ base.name }}</el-tag>
                    <Status class="mt-0.5" :status="base.bound ? 'enable' : 'disable'" />
                    <el-tag v-if="!base.initialized" type="info">
                        {{ $t('firewall.notInitialized') }}
                    </el-tag>
                    <el-tooltip :content="familyStatusDescription('IPv4', base.ipv4)" placement="bottom">
                        <el-tag :type="familyStatusType(base.ipv4)">IPv4: {{ familyStatusLabel(base.ipv4) }}</el-tag>
                    </el-tooltip>
                    <el-tooltip :content="familyStatusDescription('IPv6', base.ipv6)" placement="bottom">
                        <el-tag :type="familyStatusType(base.ipv6)">IPv6: {{ familyStatusLabel(base.ipv6) }}</el-tag>
                    </el-tooltip>
                </div>
                <div class="mt-0.5">
                    <el-button
                        v-if="!base.initialized"
                        v-permission
                        v-node-admin
                        type="primary"
                        link
                        @click="emit('operate', 'initialize')"
                    >
                        {{ $t('commons.button.init') }}
                    </el-button>
                    <el-button
                        v-else-if="!base.bound"
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
                </div>
            </div>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import i18n from '@/lang';

defineProps<{ base: Firewall.DockerGuardBase }>();
const emit = defineEmits<{ operate: [operation: 'initialize' | 'bind' | 'unbind'] }>();

const familyStatusLabel = (status: Firewall.DockerGuardFamilyStatus) => {
    if (status.state === 'effective') return i18n.global.t('firewall.effective');
    if (status.state === 'disabled') return i18n.global.t('firewall.notEnabled');
    return i18n.global.t('firewall.notEffective');
};
const familyStatusType = (status: Firewall.DockerGuardFamilyStatus) => {
    if (status.state === 'effective') return 'success';
    return status.state === 'disabled' ? 'info' : 'warning';
};
const familyStatusDescription = (family: string, status: Firewall.DockerGuardFamilyStatus) => {
    if (status.state === 'effective') return i18n.global.t('firewall.dockerGuardStatusEffective', [family]);
    return i18n.global.t(`firewall.dockerGuardStatusReason.${status.reason || 'inspect_failed'}`, [family]);
};
</script>
