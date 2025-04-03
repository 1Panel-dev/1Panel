<template>
    <div class="app-status" style="margin-top: 20px" v-if="baseInfo.isExist">
        <el-card>
            <div class="flex w-full flex-col gap-4 md:flex-row">
                <div class="flex flex-wrap gap-4">
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
                    <el-divider direction="vertical" />
                    <el-button type="primary" @click="onOperate('restart')" link>
                        {{ $t('commons.button.restart') }}
                    </el-button>
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
    <div v-if="!baseInfo.isExist">
        <LayoutContent :title="$t('menu.firewall')" :divider="true">
            <template #main>
                <div class="app-warn">
                    <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                        <span>{{ $t('firewall.notSupport') }}</span>
                        <span @click="toDoc" class="flex items-center justify-center gap-0.5">
                            <el-icon><Position /></el-icon>
                            {{ $t('firewall.quickJump') }}
                        </span>
                    </div>
                    <div>
                        <img src="@/assets/images/no_app.svg" />
                    </div>
                </div>
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { Host } from '@/api/interface/host';
import { loadFireBaseInfo, operateFire } from '@/api/modules/host';
import i18n from '@/lang';
import router from '@/routers';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { ref } from 'vue';

const baseInfo = ref<Host.FirewallBase>({ isActive: false, isExist: false, name: '', version: '', pingStatus: '' });
const onPing = ref('Disable');
const oldStatus = ref();

const acceptParams = (): void => {
    loadBaseInfo(true);
};
const emit = defineEmits(['search', 'update:is-active', 'update:loading', 'update:maskShow', 'update:name']);

const toDoc = () => {
    router.push({ name: 'Library' });
};

const loadBaseInfo = async (search: boolean) => {
    await loadFireBaseInfo()
        .then((res) => {
            baseInfo.value = res.data;
            onPing.value = baseInfo.value.pingStatus;
            oldStatus.value = onPing.value;
            emit('update:name', baseInfo.value.name);
            emit('update:is-active', baseInfo.value.isActive);
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

const onOperate = async (operation: string) => {
    emit('update:maskShow', false);
    let operationHelper = i18n.global.t('firewall.' + operation + 'FirewallHelper');
    let title = i18n.global.t('firewall.firewallHelper', [i18n.global.t('commons.button.' + operation)]);
    ElMessageBox.confirm(operationHelper, title, {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    })
        .then(async () => {
            emit('update:loading', true);
            emit('update:maskShow', true);
            await operateFire(operation)
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    loadBaseInfo(true);
                })
                .catch(() => {
                    loadBaseInfo(true);
                });
        })
        .catch(() => {
            emit('update:maskShow', true);
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
            await operateFire(operation)
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
