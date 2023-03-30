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

const baseInfo = ref<Host.FirewallBase>({ status: '', name: '', version: '' });

const acceptParams = (): void => {
    loadBaseInfo();
};
const emit = defineEmits(['search', 'update:status', 'update:loading']);

const loadBaseInfo = async () => {
    await loadFireBaseInfo()
        .then((res) => {
            baseInfo.value = res.data;
            emit('update:status', baseInfo.value.status);
            if (baseInfo.value.status === 'running') {
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
                loadBaseInfo();
            })
            .catch(() => {
                emit('update:loading', false);
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
