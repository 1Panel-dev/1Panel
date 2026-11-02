<template>
    <LayoutContent :title="$t('xpack.vm.template')">
        <template #rightToolBar>
            <TableSearch v-model:searchName="searchName" @search="search" />
            <TableRefresh @search="search" />
        </template>
        <template #main>
            <div v-loading="loading" class="template-card-content">
                <div v-if="filteredItems.length" class="template-card-grid">
                    <el-card v-for="item in filteredItems" :key="item.id" class="template-card" shadow="hover">
                        <div class="template-card-header">
                            <div class="template-card-title">
                                <span class="template-card-name" :title="item.name">{{ item.name }}</span>
                            </div>
                            <el-tag :type="item.isExist ? 'success' : 'danger'" round size="small" effect="plain">
                                {{ item.isExist ? $t('commons.status.ready') : $t('commons.status.deleted') }}
                            </el-tag>
                        </div>

                        <div class="template-card-resource">
                            <div class="template-card-resource-item">
                                <span>{{ $t('xpack.vm.cpu') }}</span>
                                <strong>{{ item.cpu }} {{ $t('commons.units.core') }}</strong>
                            </div>
                            <div class="template-card-resource-item">
                                <span>{{ $t('xpack.vm.memory') }}</span>
                                <strong>{{ computeSizeFromMB(item.memory) }}</strong>
                            </div>
                            <div class="template-card-resource-item">
                                <span>{{ $t('xpack.vm.disk') }}</span>
                                <strong>{{ formatDiskSize(item.diskSize) }}</strong>
                            </div>
                        </div>

                        <div class="template-card-info">
                            <div>
                                <span>{{ $t('xpack.vm.sourceVM') }}</span>
                                <strong>{{ item.sourceVM || '-' }}</strong>
                            </div>
                            <div>
                                <span>{{ $t('xpack.vm.diskBus') }}</span>
                                <strong>{{ item.diskBus?.toUpperCase() || '-' }}</strong>
                            </div>
                            <div>
                                <span>{{ $t('xpack.vm.storagePool') }}</span>
                                <strong>{{ item.storagePool || '-' }}</strong>
                            </div>
                            <div>
                                <span>{{ $t('xpack.vm.diskPath') }}</span>
                                <strong :title="item.diskPath">{{ item.diskPath || '-' }}</strong>
                            </div>
                            <div>
                                <span>{{ $t('xpack.vm.templateSize') }}</span>
                                <strong>{{ item.isExist ? computeSize(item.size) : '-' }}</strong>
                            </div>
                            <div>
                                <span>{{ $t('commons.table.date') }}</span>
                                <strong>{{ formatDate(item.createdAt) }}</strong>
                            </div>
                            <div>
                                <span>{{ $t('commons.table.description') }}</span>
                                <strong :title="item.description">{{ item.description || '-' }}</strong>
                            </div>
                        </div>

                        <div class="template-card-actions">
                            <el-button
                                v-permission
                                link
                                type="primary"
                                :disabled="vmUnavailable"
                                :loading="deleteLoading"
                                @click="openDelete(item)"
                            >
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </div>
                    </el-card>
                </div>
                <el-empty v-else-if="!loading" :description="$t('commons.msg.noneData')" />
            </div>
        </template>
    </LayoutContent>

    <TaskLog ref="taskLogRef" width="70%" @close="search" />
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import TaskLog from '@/components/log/task/index.vue';
import { VM } from '@/api/interface/vm';
import { deleteVMTemplate, listVMTemplates } from '@/api/modules/vm';
import { computeSize, computeSizeFromMB } from '@/utils/size';
import { dateFormat } from '@/utils/date';
import { newUUID } from '@/utils/id';
import { MsgSuccess } from '@/utils/message';

const props = defineProps<{
    vmUnavailable: boolean;
    vmEnvironmentLoaded: boolean;
}>();
const vmUnavailable = computed(() => props.vmUnavailable);
const loading = ref(false);
const deleteLoading = ref(false);
const searchName = ref('');
const items = ref<VM.TemplateItem[]>([]);
const taskLogRef = ref();

const filteredItems = computed(() => {
    const keyword = searchName.value.trim().toLowerCase();
    if (!keyword) {
        return items.value;
    }
    return items.value.filter((item) =>
        [item.name, item.sourceVM, item.diskPath, item.storagePool, item.description]
            .filter(Boolean)
            .some((value) => value.toLowerCase().includes(keyword)),
    );
});

const search = async () => {
    if (!props.vmEnvironmentLoaded || vmUnavailable.value) {
        items.value = [];
        return;
    }
    loading.value = true;
    await listVMTemplates()
        .then((res) => {
            items.value = res.data || [];
        })
        .finally(() => {
            loading.value = false;
        });
};

const openDelete = async (row: VM.TemplateItem) => {
    if (vmUnavailable.value) {
        return;
    }
    const templateName = row.diskPath ? `${row.name}（${row.diskPath}）` : row.name;
    const confirmMessage = i18n.global.t('xpack.vm.deleteTemplateConfirm', [templateName]);
    try {
        await ElMessageBox.confirm(confirmMessage, i18n.global.t('commons.button.delete'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'warning',
        });
    } catch {
        return;
    }
    const taskID = newUUID();
    deleteLoading.value = true;
    await deleteVMTemplate({ taskID, id: row.id, deleteDisk: row.isExist })
        .then((res) => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            taskLogRef.value?.openWithTaskID(res.data?.taskID || taskID, true);
        })
        .finally(() => {
            deleteLoading.value = false;
        });
};

const formatDiskSize = (size: number) => {
    if (!size) {
        return '-';
    }
    return `${size} GiB`;
};

const formatDate = (value: string) => (value ? dateFormat(0, 0, value) : '-');

watch(() => [props.vmEnvironmentLoaded, props.vmUnavailable], search, { immediate: true });
</script>

<style scoped lang="scss">
.template-card-content {
    min-height: 260px;
}

.template-card-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 14px;
}

.template-card {
    min-width: 0;

    :deep(.el-card__body) {
        display: flex;
        min-height: 248px;
        flex-direction: column;
        padding: 16px;
    }
}

.template-card-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
}

.template-card-title {
    min-width: 0;
}

.template-card-name {
    display: block;
    overflow: hidden;
    color: var(--el-text-color-primary);
    font-size: 15px;
    font-weight: 600;
    line-height: 22px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.template-card-resource {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 8px;
    margin-top: 14px;
}

.template-card-resource-item {
    min-width: 0;
    padding: 10px;
    border-radius: 6px;
    background: var(--el-fill-color-extra-light);
}

.template-card-resource-item span,
.template-card-info span {
    display: block;
    margin-bottom: 4px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.template-card-resource-item strong,
.template-card-info strong {
    display: block;
    min-width: 0;
    overflow: hidden;
    color: var(--el-text-color-primary);
    font-size: 13px;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.template-card-info {
    display: grid;
    gap: 8px;
    margin-top: 14px;
    margin-bottom: 14px;
}

.template-card-info > div {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    min-width: 0;
}

.template-card-info span {
    flex: none;
    margin-bottom: 0;
}

.template-card-info strong {
    font-weight: 500;
    text-align: right;
}

.template-card-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 12px;
    flex-wrap: wrap;
    margin-top: auto;
    padding-top: 14px;
    border-top: 1px solid var(--el-border-color-lighter);
}

.template-card-actions :deep(.el-button) {
    margin-left: 0;
}

@media (max-width: 1280px) {
    .template-card-grid {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }
}

@media (max-width: 768px) {
    .template-card-grid,
    .template-card-resource {
        grid-template-columns: 1fr;
    }

    .template-card-header {
        flex-direction: column;
        align-items: stretch;
    }
}
</style>
