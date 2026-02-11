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
                        <template v-if="baseInfo.name !== 'iptables'">
                            <el-button type="primary" v-if="baseInfo.isActive" @click="onOperate('stop')" link>
                                {{ $t('commons.button.stop') }}
                            </el-button>
                            <el-button type="primary" v-if="!baseInfo.isActive" @click="onOperate('start')" link>
                                {{ $t('commons.button.start') }}
                            </el-button>
                            <el-divider direction="vertical" />
                            <el-button type="primary" @click="onOperate('restart')" link>
                                {{ $t('commons.button.restart') }}
                            </el-button>
                        </template>
                        <template v-if="!baseInfo.isInit || (props.currentTab === 'forward' && !baseInfo.isBind)">
                            <el-divider direction="vertical" />
                            <el-button type="primary" link @click="onInit">
                                {{ $t('commons.button.init') }}
                            </el-button>
                        </template>
                        <template v-if="baseInfo.name === 'iptables' && baseInfo.isInit && props.currentTab == 'base'">
                            <el-divider direction="vertical" />
                            <el-button v-if="baseInfo.isBind" type="primary" link @click="onUnBind">
                                {{ $t('commons.button.unbind') }}
                            </el-button>
                            <el-button v-if="!baseInfo.isBind" type="primary" link @click="onBind">
                                {{ $t('commons.button.bind') }}
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
        <NoSuchService v-else name="Firewalld / Ufw / iptables" />

        <LayoutContent :divider="true" v-if="baseInfo.isExist && baseInfo.isActive && !baseInfo.isInit">
            <template #main>
                <div class="app-warn">
                    <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                        <span>{{ loadInitMsg() }}</span>
                    </div>
                    <div>
                        <img src="@/assets/images/no_app.svg" />
                    </div>
                </div>
            </template>
        </LayoutContent>

        <DockerRestart
            ref="dockerRef"
            v-model:withDockerRestart="withDockerRestart"
            @submit="onSubmit"
            :title="$t('firewall.firewallHelper', [$t('commons.button.' + operation)])"
        >
            <template #helper>
                <span>{{ $t('firewall.' + operation + 'FirewallHelper') }}</span>
            </template>
        </DockerRestart>
    </div>
</template>

<script lang="ts" setup>
import { Host } from '@/api/interface/host';
import { loadFireBaseInfo, operateFilterChain, operateFire } from '@/api/modules/host';
import i18n from '@/lang';
import NoSuchService from '@/components/layout-content/no-such-service.vue';
import DockerRestart from '@/components/docker-proxy/docker-restart.vue';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { ref } from 'vue';
import { loadDockerStatus } from '@/api/modules/container';

const props = defineProps({
    currentTab: String,
});

const baseInfo = ref<Host.FirewallBase>({
    isActive: false,
    isExist: true,
    isInit: false,
    isBind: false,
    name: '',
    version: '',
    pingStatus: '',
});
const onPing = ref('Disable');
const oldStatus = ref();
const dockerRef = ref();
const operation = ref('restart');
const dockerStatus = ref();
const withDockerRestart = ref(false);

const acceptParams = (): void => {
    loadBaseInfo(true);
    loadDocker();
};
const emit = defineEmits([
    'search',
    'update:is-active',
    'update:is-bind',
    'update:loading',
    'update:maskShow',
    'update:name',
]);

const loadBaseInfo = async (search: boolean) => {
    await loadFireBaseInfo(props.currentTab)
        .then(async (res) => {
            baseInfo.value = res.data;
            onPing.value = baseInfo.value.pingStatus;
            oldStatus.value = onPing.value;
            if (baseInfo.value.isInit) {
                emit('update:name', baseInfo.value.name);
            } else {
                emit('update:name', '-');
            }
            emit('update:is-active', baseInfo.value.isActive);
            emit('update:is-bind', baseInfo.value.isBind);

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

const loadDocker = async () => {
    const res = await loadDockerStatus();
    dockerStatus.value = res.data.isExist;
};

const loadInitMsg = () => {
    switch (props.currentTab) {
        case 'base':
            return i18n.global.t('firewall.initHelper', [i18n.global.t('firewall.baseIptables')]);
        case 'forward':
            return i18n.global.t('firewall.initHelper', [i18n.global.t('firewall.forwardIptables')]);
        case 'advance':
            return i18n.global.t('firewall.initHelper', [i18n.global.t('firewall.advanceIptables')]);
    }
};

const onInit = async () => {
    let chainName = '';
    let msg = '';
    switch (props.currentTab) {
        case 'base':
            chainName = '1PANEL_BASIC';
            msg = i18n.global.t('firewall.initMsg', [i18n.global.t('firewall.baseIptables')]);
        case 'forward':
            chainName = '1PANEL_FORWARD';
            msg = i18n.global.t('firewall.initMsg', [i18n.global.t('firewall.forwardIptables')]);
        case 'advance':
            chainName = '1PANEL_INPUT';
            msg = i18n.global.t('firewall.initMsg', [i18n.global.t('firewall.advanceIptables')]);
    }
    ElMessageBox.confirm(msg, i18n.global.t('commons.button.init'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        await operateFilterChain(chainName, 'init-' + props.currentTab).then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            loadBaseInfo(true);
        });
    });
};

const onBind = async () => {
    ElMessageBox.confirm(i18n.global.t('firewall.bindHelper'), i18n.global.t('commons.button.bind'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        await operateFilterChain('1PANEL_BASIC', 'bind-base').then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            loadBaseInfo(true);
        });
    });
};
const onUnBind = async () => {
    ElMessageBox.confirm(i18n.global.t('firewall.unbindHelper'), i18n.global.t('commons.button.unbind'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        await operateFilterChain('1PANEL_BASIC', 'unbind-base').then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            loadBaseInfo(true);
        });
    });
};

const onOperate = async (op: string) => {
    operation.value = op;
    if (baseInfo.value.name === 'iptables' || !dockerStatus.value) {
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
            operation = operation === 'Disable' ? 'disableBanPing' : 'enableBanPing';
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
