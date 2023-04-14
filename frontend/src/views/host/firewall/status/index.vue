<template>
    <div>
        <div class="app-status" style="margin-top: 20px">
            <el-card>
                <div>
                    <el-tag effect="dark" type="success">{{ baseInfo.name }}</el-tag>
                    <el-tag round class="status-content" v-if="baseInfo.status === 'running'" type="success">
                        {{ $t('commons.status.running') }}
                    </el-tag>
                    <el-tag round class="status-content" v-if="baseInfo.status === 'not running'" type="info">
                        {{ $t('commons.status.stopped') }}
                    </el-tag>
                    <el-tag class="status-content">{{ $t('app.version') }}: {{ baseInfo.version }}</el-tag>

                    <span v-if="baseInfo.status === 'running'" class="buttons">
                        <el-button type="primary" @click="onOperate('stop')" link>
                            {{ $t('commons.button.stop') }}
                        </el-button>
                    </span>
                    <span v-if="baseInfo.status === 'not running'" class="buttons">
                        <el-button type="primary" @click="onOperate('start')" link>
                            {{ $t('commons.button.start') }}
                        </el-button>
                    </span>
                    <span v-if="onPing !== 'None'">
                        <el-divider direction="vertical" />
                        <el-button type="primary" link>{{ $t('firewall.noPing') }}</el-button>
                        <el-switch
                            style="margin-left: 10px"
                            inactive-value="Disable"
                            active-value="Enable"
                            @change="onPingOperate"
                            v-model="onPing"
                        />
                    </span>
                </div>
            </el-card>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { Host } from '@/api/interface/host';
import { loadFireBaseInfo, operateFire } from '@/api/modules/host';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { ref } from 'vue';

const baseInfo = ref<Host.FirewallBase>({ status: '', name: '', version: '', pingStatus: '' });
const onPing = ref('Disable');
const oldStatus = ref();

const acceptParams = (): void => {
    loadBaseInfo(true);
};
const emit = defineEmits(['search', 'update:status', 'update:loading', 'update:maskShow', 'update:name']);

const loadBaseInfo = async (search: boolean) => {
    await loadFireBaseInfo()
        .then((res) => {
            baseInfo.value = res.data;
            onPing.value = baseInfo.value.pingStatus;
            oldStatus.value = onPing.value;
            emit('update:name', baseInfo.value.name);
            emit('update:status', baseInfo.value.status);
            if (search) {
                emit('search');
            } else {
                emit('update:loading', false);
            }
        })
        .catch(() => {
            emit('update:loading', false);
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
            emit('update:status', 'running');
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
            emit('update:status', 'running');
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
