<template>
    <div v-loading="loading">
        <el-drawer v-model="drawerVisible">
            <template #header>
                <DrawerHeader :header="$t('setting.recoverDetail')" :back="handleClose" />
            </template>
            <el-form label-width="120px">
                <el-card>
                    <template #header>
                        <div class="card-header">
                            <span>{{ $t('setting.recover') }}</span>
                        </div>
                    </template>
                    <div v-if="!snapInfo.recoverStatus">
                        <div v-if="snapInfo.lastRecoveredAt">
                            <el-form-item :label="$t('commons.table.status')">
                                <el-tag type="success">
                                    {{ $t('commons.table.statusSuccess') }}
                                </el-tag>
                                <el-button @click="recoverSnapshot(true)" style="margin-left: 10px" type="primary">
                                    {{ $t('setting.recover') }}
                                </el-button>
                            </el-form-item>
                            <el-form-item :label="$t('setting.lastRecoverAt')">
                                {{ snapInfo.lastRecoveredAt }}
                            </el-form-item>
                        </div>
                        <div v-else>
                            <el-form-item>
                                <el-tag type="info">
                                    {{ $t('setting.noRecoverRecord') }}
                                </el-tag>
                            </el-form-item>
                            <el-form-item>
                                <el-button @click="recoverSnapshot(true)" type="primary">
                                    {{ $t('setting.recover') }}
                                </el-button>
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
                            <el-button
                                style="margin-left: 15px"
                                @click="recoverSnapshot(true)"
                                :disabled="snapInfo.recoverStatus !== 'Success'"
                            >
                                {{ $t('setting.recover') }}
                            </el-button>
                        </el-form-item>
                        <el-form-item :label="$t('setting.lastRecoverAt')" v-if="snapInfo.recoverStatus !== 'Waiting'">
                            {{ snapInfo.lastRecoveredAt }}
                        </el-form-item>
                        <div v-if="snapInfo.recoverStatus === 'Failed'">
                            <el-form-item :label="$t('commons.button.log')">
                                <span style="word-break: break-all; flex-wrap: wrap; word-wrap: break-word">
                                    {{ snapInfo.recoverMessage }}
                                </span>
                            </el-form-item>
                            <el-form-item>
                                <el-button @click="dialogVisible = true" type="primary">
                                    {{ $t('commons.button.retry') }}
                                </el-button>
                            </el-form-item>
                        </div>
                    </div>
                </el-card>

                <el-card style="margin-top: 20px" v-if="snapInfo.recoverStatus === 'Failed'">
                    <template #header>
                        <div class="card-header">
                            <span>{{ $t('setting.rollback') }}</span>
                        </div>
                    </template>
                    <div v-if="!snapInfo.rollbackStatus">
                        <div v-if="snapInfo.lastRollbackedAt">
                            <el-form-item :label="$t('commons.table.status')">
                                <el-tag type="success">
                                    {{ $t('commons.table.statusSuccess') }}
                                </el-tag>
                                <el-button @click="rollbackSnapshot" style="margin-left: 10px" type="primary">
                                    {{ $t('setting.recover') }}
                                </el-button>
                            </el-form-item>
                            <el-form-item :label="$t('setting.lastRollbackAt')">
                                {{ snapInfo.lastRollbackedAt }}
                            </el-form-item>
                        </div>
                        <div v-else>
                            <el-form-item>
                                <el-tag type="info">
                                    {{ $t('setting.noRollbackRecord') }}
                                </el-tag>
                            </el-form-item>
                            <el-form-item>
                                <el-button @click="rollbackSnapshot" type="primary">
                                    {{ $t('setting.rollback') }}
                                </el-button>
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
                </el-card>
            </el-form>
        </el-drawer>
        <el-dialog v-model="dialogVisible" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('commons.button.retry') }}</span>
                </div>
            </template>
            <div>
                <span>{{ $t('setting.reDownload') }}</span>
                <el-switch style="margin-left: 15px" v-model="reDownload" />
            </div>
            <div style="margin-top: 15px">
                <span>{{ $t('setting.recoverHelper', [snapInfo.name]) }}</span>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="dialogVisible = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="doRecover(false)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { Setting } from '@/api/interface/setting';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { snapshotRecover, snapshotRollback } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';

const drawerVisible = ref(false);
const snapInfo = ref();
const loading = ref();

const dialogVisible = ref();
const reDownload = ref();

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

const doRecover = async (isNew: boolean) => {
    loading.value = true;
    await snapshotRecover({ id: snapInfo.value.id, isNew: isNew, reDownload: reDownload.value })
        .then(() => {
            emit('search');
            loading.value = false;
            dialogVisible.value = false;
            drawerVisible.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const recoverSnapshot = async (isNew: boolean) => {
    ElMessageBox.confirm(i18n.global.t('setting.recoverHelper', [snapInfo.value.name]), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        doRecover(isNew);
    });
};

const rollbackSnapshot = async () => {
    ElMessageBox.confirm(i18n.global.t('setting.rollbackHelper'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await snapshotRollback({ id: snapInfo.value.id, isNew: false, reDownload: false })
            .then(() => {
                emit('search');
                loading.value = false;
                dialogVisible.value = false;
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
