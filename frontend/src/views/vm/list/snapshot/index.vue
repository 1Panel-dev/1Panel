<template>
    <DrawerPro
        v-model="open"
        :header="$t('xpack.vm.snapshot')"
        :resource="vmItem?.name || ''"
        size="large"
        @close="handleClose"
    >
        <template #content>
            <div v-if="vmItem" v-loading="loading" class="vm-snapshot">
                <el-alert type="info" :closable="false" :title="$t('xpack.vm.snapshotHelper')" />
                <div class="vm-snapshot-toolbar">
                    <el-button v-permission type="primary" @click="openCreate">
                        {{ $t('commons.button.create') }}
                    </el-button>
                </div>

                <el-timeline v-if="items.length" class="snapshot-timeline">
                    <el-timeline-item
                        v-for="row in items"
                        :key="row.name"
                        :timestamp="formatDate(row.createdAt)"
                        placement="top"
                        :type="row.current ? 'success' : 'primary'"
                        :hollow="!row.current"
                    >
                        <article class="snapshot-card" :class="{ 'is-current': row.current }">
                            <div class="snapshot-card-header">
                                <div class="snapshot-title">
                                    <span class="snapshot-name" :title="row.name">{{ row.name }}</span>
                                    <el-tag v-if="row.current" size="small" type="success">
                                        {{ $t('xpack.vm.currentSnapshot') }}
                                    </el-tag>
                                </div>
                                <div class="snapshot-actions">
                                    <el-button
                                        v-permission
                                        link
                                        type="primary"
                                        :disabled="vmNotShutOff"
                                        @click="recoverSnapshot(row)"
                                    >
                                        {{ $t('commons.button.recover') }}
                                    </el-button>
                                </div>
                            </div>
                            <div class="snapshot-meta">
                                <span>
                                    {{ $t('xpack.vm.snapshotSize') }}：{{ row.size ? computeSize(row.size) : '-' }}
                                </span>
                                <span class="snapshot-description" :title="row.description">
                                    {{ $t('commons.table.description') }}：{{ row.description || '-' }}
                                </span>
                            </div>
                        </article>
                    </el-timeline-item>
                </el-timeline>
                <el-empty v-if="!items.length && !loading" :description="$t('xpack.vm.snapshotEmpty')" />
            </div>
        </template>
    </DrawerPro>

    <DialogPro v-model="createVisible" :title="$t('xpack.vm.createSnapshot')" size="small">
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
            <el-form-item :label="$t('xpack.vm.snapshot')" prop="snapshotName">
                <el-input v-model.trim="form.snapshotName" />
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input v-model="form.description" type="textarea" :rows="3" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="createVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button v-permission type="primary" :loading="submitLoading" @click="submitCreate">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>

    <TaskLog ref="taskLogRef" width="70%" @close="search" />
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { ElMessageBox, type FormInstance } from 'element-plus';
import DrawerPro from '@/components/drawer-pro/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { newUUID } from '@/utils/id';
import { MsgSuccess } from '@/utils/message';
import { computeSize } from '@/utils/size';
import { dateFormat } from '@/utils/date';
import { VM } from '@/api/interface/vm';
import { createVMSnapshot, listVMSnapshots, recoverVMSnapshot } from '@/api/modules/vm';

const open = ref(false);
const createVisible = ref(false);
const loading = ref(false);
const submitLoading = ref(false);
const vmItem = ref<VM.Item>();
const items = ref<VM.SnapshotItem[]>([]);
const formRef = ref<FormInstance>();
const taskLogRef = ref();
const form = reactive({
    snapshotName: '',
    description: '',
});

const rules = reactive({
    snapshotName: [Rules.requiredInput, Rules.vmName],
});

const vmNotShutOff = computed(() => vmItem.value?.status !== 'shut off');

const acceptParams = (row: VM.Item) => {
    vmItem.value = row;
    open.value = true;
    search();
};

const search = async () => {
    if (!vmItem.value?.name) {
        return;
    }
    loading.value = true;
    await listVMSnapshots(vmItem.value.name)
        .then((res) => {
            items.value = res.data || [];
        })
        .finally(() => {
            loading.value = false;
        });
};

const openCreate = () => {
    form.snapshotName = defaultSnapshotName();
    form.description = '';
    createVisible.value = true;
};

const submitCreate = async () => {
    if (!vmItem.value?.name) {
        return;
    }
    await formRef.value?.validate();
    const taskID = newUUID();
    submitLoading.value = true;
    await createVMSnapshot({
        taskID,
        name: vmItem.value.name,
        snapshotName: form.snapshotName,
        description: form.description,
    })
        .then((res) => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            createVisible.value = false;
            openTaskLog(res.data?.taskID || taskID);
        })
        .finally(() => {
            submitLoading.value = false;
        });
};

const recoverSnapshot = async (row: VM.SnapshotItem) => {
    if (!vmItem.value?.name) {
        return;
    }
    await ElMessageBox.confirm(
        i18n.global.t('xpack.vm.snapshotRecoverConfirm', [row.name, vmItem.value.name]),
        i18n.global.t('xpack.vm.recoverSnapshot'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        },
    );
    const taskID = newUUID();
    loading.value = true;
    await recoverVMSnapshot({ taskID, name: vmItem.value.name, snapshotName: row.name })
        .then((res) => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            openTaskLog(res.data?.taskID || taskID);
        })
        .finally(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    items.value = [];
    vmItem.value = undefined;
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value?.openWithTaskID(taskID, true);
};

const formatDate = (value: number) => (value ? dateFormat(null, null, value * 1000) : '-');

const defaultSnapshotName = () => {
    const now = new Date();
    const pad = (value: number) => String(value).padStart(2, '0');
    return `snap-${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}${pad(now.getHours())}${pad(
        now.getMinutes(),
    )}${pad(now.getSeconds())}`;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.vm-snapshot {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.vm-snapshot-toolbar {
    display: flex;
    justify-content: flex-start;
    gap: 8px;
}

.snapshot-timeline {
    margin: 4px 0 0;
    padding-left: 0;

    :deep(.el-timeline-item__timestamp) {
        color: var(--el-text-color-secondary);
    }
}

.snapshot-card {
    min-width: 0;
    flex: 1;
    padding: 14px 16px;
    margin-bottom: 12px;
    background: var(--el-fill-color-blank);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    transition: border-color 0.2s;

    &.is-current {
        border-color: var(--el-color-success-light-5);
        background: var(--el-color-success-light-9);
    }

    &:hover {
        border-color: var(--el-border-color);
    }
}

.snapshot-card-header,
.snapshot-title,
.snapshot-meta {
    display: flex;
    align-items: center;
}

.snapshot-card-header {
    min-width: 0;
    justify-content: space-between;
    gap: 16px;
}

.snapshot-title {
    min-width: 0;
    gap: 8px;
    font-weight: 500;
}

.snapshot-name {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.snapshot-actions {
    display: flex;
    flex-shrink: 0;
}

.snapshot-meta {
    min-width: 0;
    gap: 8px 24px;
    margin-top: 8px;
    overflow: hidden;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.snapshot-description {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

@media (max-width: 768px) {
    .snapshot-card-header,
    .snapshot-meta {
        align-items: flex-start;
        flex-direction: column;
    }

    .snapshot-card-header {
        gap: 8px;
    }

    .snapshot-meta {
        gap: 4px;
    }
}
</style>
