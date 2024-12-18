<template>
    <div v-loading="loading">
        <el-drawer v-model="drawerVisible" :close-on-click-modal="false" :close-on-press-escape="false">
            <template #header>
                <DrawerHeader :header="$t('setting.recoverDetail')" :back="handleClose" />
            </template>

            <el-form label-position="top">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <span class="card-title">{{ $t('setting.recover') }}</span>
                        <el-divider class="divider" />
                        <div v-if="!snapInfo.recoverStatus && !snapInfo.lastRecoveredAt">
                            <el-alert center class="alert" style="height: 257px" :closable="false">
                                <el-button size="large" round plain type="primary" @click="recoverSnapshot(true)">
                                    {{ $t('setting.recover') }}
                                </el-button>
                            </el-alert>
                        </div>
                        <el-card v-else class="mini-border-card">
                            <div v-if="!snapInfo.recoverStatus" class="mini-border-card">
                                <div v-if="snapInfo.lastRecoveredAt">
                                    <el-form-item :label="$t('commons.table.status')">
                                        <el-tag type="success">
                                            {{ $t('commons.table.statusSuccess') }}
                                        </el-tag>
                                        <el-button
                                            @click="recoverSnapshot(true)"
                                            style="margin-left: 10px"
                                            type="primary"
                                        >
                                            {{ $t('setting.recover') }}
                                        </el-button>
                                    </el-form-item>
                                    <el-form-item :label="$t('setting.lastRecoverAt')">
                                        {{ snapInfo.lastRecoveredAt }}
                                    </el-form-item>
                                </div>
                            </div>
                            <div v-else>
                                <el-form-item :label="$t('commons.table.status')">
                                    <el-tag type="danger" v-if="snapInfo.recoverStatus === 'Failed'">
                                        {{ $t('commons.table.statusFailed') }}
                                    </el-tag>
                                    <el-tag type="success" v-if="snapInfo.recoverStatus === 'Success'">
                                        {{ $t('commons.table.statusSuccess') }}
                                    </el-tag>
                                    <el-tag type="info" v-if="snapInfo.recoverStatus === 'Waiting'">
                                        {{ $t('commons.table.statusWaiting') }}
                                    </el-tag>
                                </el-form-item>
                                <el-form-item
                                    :label="$t('setting.lastRecoverAt')"
                                    v-if="snapInfo.recoverStatus !== 'Waiting'"
                                >
                                    {{ snapInfo.lastRecoveredAt }}
                                </el-form-item>
                                <div v-if="snapInfo.recoverStatus === 'Failed'">
                                    <el-form-item :label="$t('commons.button.log')">
                                        <span style="word-break: break-all; flex-wrap: wrap; word-wrap: break-word">
                                            {{ snapInfo.recoverMessage }}
                                        </span>
                                    </el-form-item>
                                    <el-form-item>
                                        <el-button @click="recoverSnapshot(false)" type="primary">
                                            {{ $t('commons.button.retry') }}
                                        </el-button>
                                    </el-form-item>
                                </div>
                            </div>
                        </el-card>

                        <div v-if="snapInfo.recoverStatus === 'Failed'">
                            <span class="card-title">{{ $t('setting.rollback') }}</span>
                            <el-divider class="divider" />
                            <div v-if="!snapInfo.rollbackStatus && !snapInfo.lastRollbackedAt">
                                <el-alert center class="alert" style="height: 257px" :closable="false">
                                    <el-button size="large" round plain type="primary" @click="rollbackSnapshot()">
                                        {{ $t('setting.rollback') }}
                                    </el-button>
                                </el-alert>
                            </div>
                            <div v-if="!snapInfo.rollbackStatus">
                                <div v-if="snapInfo.lastRollbackedAt">
                                    <el-form-item :label="$t('commons.table.status')">
                                        <el-tag type="success">
                                            {{ $t('commons.table.statusSuccess') }}
                                        </el-tag>
                                        <el-button @click="rollbackSnapshot" style="margin-left: 10px" type="primary">
                                            {{ $t('setting.rollback') }}
                                        </el-button>
                                    </el-form-item>
                                    <el-form-item :label="$t('setting.lastRollbackAt')">
                                        {{ snapInfo.lastRollbackedAt }}
                                    </el-form-item>
                                </div>
                            </div>
                            <div v-else>
                                <el-form-item :label="$t('commons.table.status')">
                                    <el-tag type="success" v-if="snapInfo.rollbackStatus === 'Success'">
                                        {{ $t('commons.table.statusSuccess') }}
                                    </el-tag>
                                    <el-tag type="danger" v-if="snapInfo.rollbackStatus === 'Failed'">
                                        {{ $t('commons.table.statusFailed') }}
                                    </el-tag>
                                    <el-tag type="info" v-if="snapInfo.rollbackStatus === 'Waiting'">
                                        {{ $t('commons.table.statusWaiting') }}
                                    </el-tag>
                                    <el-button
                                        style="margin-left: 15px"
                                        :disabled="snapInfo.rollbackStatus !== 'Success'"
                                        @click="rollbackSnapshot"
                                    >
                                        {{ $t('setting.rollback') }}
                                    </el-button>
                                </el-form-item>
                                <el-form-item
                                    :label="$t('setting.lastRollbackAt')"
                                    v-if="snapInfo.rollbackStatus !== 'Waiting'"
                                >
                                    {{ snapInfo.lastRollbackedAt }}
                                </el-form-item>
                                <div v-if="snapInfo.rollbackStatus === 'Failed'">
                                    <el-form-item :label="$t('commons.button.log')">
                                        <span style="word-break: break-all; flex-wrap: wrap; word-wrap: break-word">
                                            {{ snapInfo.rollbackMessage }}
                                        </span>
                                    </el-form-item>
                                    <el-form-item>
                                        <el-button @click="rollbackSnapshot()" type="primary">
                                            {{ $t('commons.button.retry') }}
                                        </el-button>
                                    </el-form-item>
                                </div>
                            </div>
                        </div>
                    </el-col>
                </el-row>
            </el-form>
        </el-drawer>
    </div>

    <SnapRecover ref="recoverRef" @close="handleClose" />
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { Setting } from '@/api/interface/setting';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { snapshotRollback } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';
import { loadOsInfo } from '@/api/modules/dashboard';
import SnapRecover from '@/views/setting/snapshot/recover/index.vue';

const drawerVisible = ref(false);
const snapInfo = ref();
const loading = ref();

const recoverRef = ref();

interface DialogProps {
    snapInfo: Setting.SnapshotInfo;
}
const acceptParams = (params: DialogProps): void => {
    snapInfo.value = params.snapInfo;
    drawerVisible.value = true;
};
const emit = defineEmits(['search']);

const handleClose = () => {
    drawerVisible.value = false;
};

const recoverSnapshot = async (isNew: boolean) => {
    loading.value = true;
    await loadOsInfo()
        .then((res) => {
            loading.value = false;
            let params = {
                id: snapInfo.value.id,
                isNew: isNew,
                name: snapInfo.value.name,
                reDownload: false,
                secret: snapInfo.value.secret,

                arch: res.data.kernelArch,
                size: snapInfo.value.size,
                freeSize: res.data.diskSize,
            };
            recoverRef.value.acceptParams(params);
        })
        .catch(() => {
            loading.value = false;
        });
};

const rollbackSnapshot = async () => {
    ElMessageBox.confirm(i18n.global.t('setting.rollbackHelper'), i18n.global.t('setting.rollback'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await snapshotRollback({ id: snapInfo.value.id, isNew: false, reDownload: false, secret: '' })
            .then(() => {
                emit('search');
                loading.value = false;
                drawerVisible.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.divider {
    display: block;
    height: 1px;
    width: 100%;
    margin: 12px 0;
    border-top: 1px var(--el-border-color) var(--el-border-style);
}
.alert {
    background-color: var(--panel-alert-bg-color);
}

.card-title {
    font-size: 14px;
    font-weight: 500;
    line-height: 25px;
    color: var(--el-button-text-color, var(--el-text-color-regular));
}
</style>
