<template>
    <div>
        <div class="app-status card-interval" v-if="baseInfo.isExist">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4 ml-3">
                        <el-tag effect="dark" type="success">{{ baseInfo.name }}</el-tag>
                        <Status class="mt-0.5" :status="baseInfo.isActive ? 'enable' : 'disable'" />
                        <el-tag>{{ $t('app.version') }}: {{ baseInfo.version }}</el-tag>
                    </div>
                    <div class="mt-0.5">
                        <el-button type="primary" v-if="baseInfo.isActive" @click="onOperate('stop')" link>
                            {{ $t('commons.button.stop') }}
                        </el-button>
                        <el-button type="primary" v-if="!baseInfo.isActive" @click="onOperate('start')" link>
                            {{ $t('commons.button.start') }}
                        </el-button>
                        <template v-if="baseInfo.name !== 'iptables'">
                            <el-divider direction="vertical" />
                            <el-button type="primary" @click="onOperate('restart')" link>
                                {{ $t('commons.button.restart') }}
                            </el-button>
                        </template>
                        <span v-if="onPing !== 'None'">
                            <el-divider direction="vertical" />
                            <el-button type="primary" link>{{ $t('firewall.noPing') }}</el-button>
                            <el-switch
                                size="small"
                                class="ml-2"
                                inactive-value="Disable"
                                active-value="Enable"
                                @change="onPingOperate"
                                v-model="onPing"
                            />
                        </span>
                    </div>
                </div>
            </el-card>
        </div>
        <NoSuchService v-if="!baseInfo.isExist" name="Firewalld / Ufw" />
        <DockerRestart
            ref="dockerRef"
            v-model:withDockerRestart="withDockerRestart"
            @submit="onSubmit"
            :title="$t('firewall.firewallHelper', [i18n.global.t('commons.button.' + operation)])"
        >
            <template #helper>
                <span>{{ $t('firewall.' + operation + 'FirewallHelper') }}</span>
            </template>
        </DockerRestart>
    </div>
</template>

<script lang="ts" setup>
import { Host } from '@/api/interface/host';
import { loadFireBaseInfo, operateFire, getFilterRules } from '@/api/modules/host';
import i18n from '@/lang';
import NoSuchService from '@/components/layout-content/no-such-service.vue';
import DockerRestart from '@/components/docker-proxy/docker-restart.vue';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { ref } from 'vue';

defineProps({
    showAdvancedControls: {
        type: Boolean,
        default: false,
    },
});

const baseInfo = ref<Host.FirewallBase>({ isActive: false, isExist: true, name: '', version: '', pingStatus: '' });
const onPing = ref('Disable');
const oldStatus = ref();
const dockerRef = ref();
const operation = ref('restart');
const withDockerRestart = ref(false);

// Iptables specific state
const chainInfoMap = ref<Map<string, Host.IptablesChainInfo>>(new Map());

const acceptParams = (): void => {
    loadBaseInfo(true);
};
const emit = defineEmits(['search', 'update:is-active', 'update:loading', 'update:maskShow', 'update:name']);

const loadBaseInfo = async (search: boolean) => {
    await loadFireBaseInfo()
        .then(async (res) => {
            baseInfo.value = res.data;
            onPing.value = baseInfo.value.pingStatus;
            oldStatus.value = onPing.value;
            emit('update:name', baseInfo.value.name);
            emit('update:is-active', baseInfo.value.isActive);

            // Load iptables chain info if firewall is iptables
            if (baseInfo.value.name === 'iptables' && baseInfo.value.isActive) {
                await loadIptablesChainInfo();
            } else {
                chainInfoMap.value.clear();
            }

            if (search) {
                emit('search');
            } else {
                emit('update:loading', false);
            }
        })
        .catch(() => {
            emit('update:loading', false);
            emit('update:maskShow', true);
            emit('update:name', '-');
        });
};

const loadIptablesChainInfo = async () => {
    const params: Host.IptablesFilterRuleSearch = {
        chains: ['INPUT', 'OUTPUT', '1PANEL_INPUT', '1PANEL_OUTPUT'],
    };
    await getFilterRules(params)
        .then((res) => {
            chainInfoMap.value.clear();
            res.data.forEach((chainInfo: Host.IptablesChainInfo) => {
                chainInfoMap.value.set(chainInfo.name, chainInfo);
            });
        })
        .catch(() => {
            chainInfoMap.value.clear();
        });
};

const onOperate = async (op: string) => {
    operation.value = op;
    // For iptables, skip docker restart warning and execute directly
    if (baseInfo.value.name === 'iptables') {
        emit('update:loading', true);
        emit('update:maskShow', true);
        await operateFire(operation.value, false)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                loadBaseInfo(true);
            })
            .catch(() => {
                loadBaseInfo(true);
            });
    } else {
        dockerRef.value.acceptParams({ title: i18n.global.t('firewall.dockerRestart') });
    }
};

const onSubmit = async () => {
    emit('update:loading', true);
    emit('update:maskShow', true);
    await operateFire(operation.value, withDockerRestart.value)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            loadBaseInfo(true);
        })
        .catch(() => {
            loadBaseInfo(true);
        });
};

const onPingOperate = async (operation: string) => {
    emit('update:maskShow', false);
    let operationHelper =
        operation === 'Enable' ? i18n.global.t('firewall.noPingHelper') : i18n.global.t('firewall.onPingHelper');
    ElMessageBox.confirm(operationHelper, i18n.global.t('firewall.noPingTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    })
        .then(async () => {
            emit('update:loading', true);
            operation = operation === 'Disable' ? 'disablePing' : 'enablePing';
            emit('update:maskShow', true);
            await operateFire(operation, false)
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    loadBaseInfo(false);
                })
                .catch(() => {
                    loadBaseInfo(false);
                });
        })
        .catch(() => {
            emit('update:maskShow', true);
            onPing.value = oldStatus.value;
        });
};

defineExpose({
    acceptParams,
});
</script>
