<template>
    <div>
        <div class="app-status" style="margin-top: 20px">
            <el-card>
                <div>
                    <el-tag style="float: left" effect="dark" type="success">{{ baseInfo.name }}</el-tag>
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
                        <el-divider direction="vertical" />
                        <el-button type="primary" link>{{ $t('firewall.noPing') }}</el-button>
                        <el-switch
                            style="margin-left: 10px"
                            inactive-value="Disable"
                            active-value="Enable"
                            @change="onPingOperate(baseInfo.pingStatus)"
                            v-model="onPing"
                        />
                    </span>

                    <span v-if="baseInfo.status === 'not running'" class="buttons">
                        <el-button type="primary" @click="onOperate('start')" link>
                            {{ $t('commons.button.start') }}
                        </el-button>
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
import { ElMessageBox } from 'element-plus';
import { ref } from 'vue';

const baseInfo = ref<Host.FirewallBase>({ status: '', name: '', version: '', pingStatus: '' });
const onPing = ref();

const acceptParams = (): void => {
    loadBaseInfo(true);
};
const emit = defineEmits(['search', 'update:status', 'update:loading']);

const loadBaseInfo = async (search: boolean) => {
    await loadFireBaseInfo()
        .then((res) => {
            baseInfo.value = res.data;
            onPing.value = baseInfo.value.pingStatus;
            emit('update:status', baseInfo.value.status);
            if (baseInfo.value.status === 'running' && search) {
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
    let operationHelper = i18n.global.t('firewall.' + operation + 'FirewallHelper');
    let title = i18n.global.t('firewall.firewallHelper', [i18n.global.t('commons.button.' + operation)]);
    ElMessageBox.confirm(operationHelper, title, {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        emit('update:loading', true);
        emit('update:status', 'running');
        await operateFire(operation)
            .then(() => {
                loadBaseInfo(true);
            })
            .catch(() => {
                emit('update:loading', false);
            });
    });
};

const onPingOperate = async (operation: string) => {
    let operationHelper =
        operation === 'Enabel' ? i18n.global.t('firewall.noPingHelper') : i18n.global.t('firewall.onPingHelper');
    ElMessageBox.confirm(operationHelper, i18n.global.t('firewall.noPing'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    })
        .then(async () => {
            emit('update:loading', true);
            emit('update:status', 'running');
            operation = operation === 'Disable' ? 'enablePing' : 'disablePing';
            await operateFire(operation)
                .then(() => {
                    loadBaseInfo(false);
                })
                .catch(() => {
                    loadBaseInfo(false);
                });
        })
        .catch(() => {
            emit('update:loading', true);
            loadBaseInfo(false);
        });
};

defineExpose({
    acceptParams,
});
</script>
