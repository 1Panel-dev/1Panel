<template>
    <LayoutContent :title="$t('xpack.vm.storagePool')">
        <template #leftToolBar>
            <el-button v-permission type="primary" :disabled="vmUnavailable" @click="openCreate">
                {{ $t('commons.button.create') }}
            </el-button>
        </template>
        <template #rightToolBar>
            <TableSearch v-model:searchName="searchName" @search="search" />
            <TableRefresh @search="search" />
        </template>
        <template #main>
            <div class="storage-card-list" v-if="items.length" v-loading="loading">
                <article v-for="item in items" :key="item.name" class="storage-card">
                    <div class="storage-card__header">
                        <div class="storage-card__title">
                            <span class="storage-card__name" :title="item.name">{{ item.name }}</span>
                            <el-tag round size="small" effect="plain">
                                {{ item.type === 'dir' ? $t('xpack.vm.dir') : '-' }}
                            </el-tag>
                        </div>
                        <el-dropdown v-permission placement="bottom" :disabled="loading || !canToggleStatus(item)">
                            <span class="storage-card__status">
                                <Status
                                    :status="item.status || RESOURCE_STATUS_UNKNOWN"
                                    :msg="item.error"
                                    :operate="canToggleStatus(item)"
                                    :disabled="statusSwitchingIds.includes(item.id)"
                                />
                            </span>
                            <template #dropdown>
                                <el-dropdown-menu>
                                    <el-dropdown-item
                                        :disabled="
                                            statusSwitchingIds.includes(item.id) ||
                                            normalizeStatus(item.status) === 'running'
                                        "
                                        @click="changeStatus(item, 'running')"
                                    >
                                        {{ $t('commons.operate.start') }}
                                    </el-dropdown-item>
                                    <el-dropdown-item
                                        :disabled="
                                            statusSwitchingIds.includes(item.id) ||
                                            normalizeStatus(item.status) === 'inactive'
                                        "
                                        @click="changeStatus(item, 'inactive')"
                                    >
                                        {{ $t('commons.operate.stop') }}
                                    </el-dropdown-item>
                                </el-dropdown-menu>
                            </template>
                        </el-dropdown>
                    </div>
                    <div class="storage-card__sub" :title="item.path">
                        <span>{{ item.path || '-' }}</span>
                    </div>
                    <div class="storage-card__meta">
                        <div class="storage-card__meta-item">
                            <span class="storage-card__meta-label">{{ $t('xpack.vm.resourceAutoStart') }}</span>
                            <el-switch
                                v-permission
                                size="small"
                                v-model="item.autoStart"
                                :disabled="!canToggleStatus(item)"
                                :loading="switchingIds.includes(item.id)"
                                @change="changeAutoStart(item, $event)"
                            />
                        </div>
                        <div class="storage-card__meta-item">
                            <span class="storage-card__meta-label">{{ $t('xpack.vm.vmCount') }}</span>
                            <el-dropdown
                                v-if="item.vmCount"
                                class="storage-vm-dropdown"
                                placement="bottom"
                                trigger="click"
                                @command="goVMList"
                            >
                                <el-button class="storage-card__meta-link" link type="primary">
                                    {{ item.vmCount || 0 }}
                                </el-button>
                                <template #dropdown>
                                    <el-dropdown-menu>
                                        <el-dropdown-item
                                            v-for="vm in item.vms || []"
                                            :key="vm.id"
                                            :command="vm.name"
                                            class="storage-vm-dropdown-item"
                                        >
                                            {{ vm.name }}
                                        </el-dropdown-item>
                                    </el-dropdown-menu>
                                </template>
                            </el-dropdown>
                            <span v-else class="storage-card__meta-value">0</span>
                        </div>
                    </div>
                    <div class="storage-card__detail">
                        <div class="storage-card__detail-item">
                            <span>{{ $t('xpack.vm.capacity') }}</span>
                            <strong>{{ item.capacity ? computeSize(item.capacity) : '-' }}</strong>
                        </div>
                        <div class="storage-card__detail-item">
                            <span>{{ $t('xpack.vm.available') }}</span>
                            <strong>{{ item.available ? computeSize(item.available) : '-' }}</strong>
                        </div>
                        <div class="storage-card__detail-item storage-card__editable">
                            <span>{{ $t('commons.table.description') }}</span>
                            <strong :title="item.description">{{ item.description || '-' }}</strong>
                            <el-button v-permission type="primary" link icon="EditPen" @click="openInlineEdit(item)" />
                        </div>
                    </div>
                    <div class="storage-card__actions">
                        <el-button v-permission link type="primary" @click="openDelete(item)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </div>
                </article>
            </div>
            <el-empty v-else :description="$t('commons.msg.noneData')" />
        </template>
    </LayoutContent>
    <Operate ref="operateRef" @search="search" />
    <ResourceDelete ref="deleteRef" resource-type="storagePool" @deleted="search" />
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { computeSize } from '@/utils/size';
import { VM } from '@/api/interface/vm';
import { listVMStorages, updateVMStorage } from '@/api/modules/vm';
import Operate from '@/views/vm/storage/create/index.vue';
import ResourceDelete from '@/views/vm/components/resource-delete.vue';

const props = defineProps<{
    vmUnavailable: boolean;
    vmEnvironmentLoaded: boolean;
}>();
const loading = ref(false);
const searchName = ref('');
const items = ref<VM.StorageItem[]>([]);
const switchingIds = ref<number[]>([]);
const statusSwitchingIds = ref<number[]>([]);
const operateRef = ref();
const deleteRef = ref<InstanceType<typeof ResourceDelete>>();
const vmUnavailable = computed(() => props.vmUnavailable);
const router = useRouter();
const RESOURCE_STATUS_RUNNING = 'running';
const RESOURCE_STATUS_INACTIVE = 'inactive';
const RESOURCE_STATUS_DELETED = 'deleted';
const RESOURCE_STATUS_UNKNOWN = 'unknown';

const search = async () => {
    if (!props.vmEnvironmentLoaded || vmUnavailable.value) {
        items.value = [];
        return;
    }
    loading.value = true;
    await listVMStorages()
        .then((res) => {
            const keyword = searchName.value.trim().toLowerCase();
            items.value = keyword
                ? (res.data || []).filter((item) => item.name.toLowerCase().includes(keyword))
                : res.data || [];
        })
        .finally(() => {
            loading.value = false;
        });
};

const openCreate = () => {
    if (vmUnavailable.value) {
        return;
    }
    operateRef.value?.acceptParams();
};

const goVMList = (name: string) => {
    router.push({ path: '/vm/list', query: { name } });
};

const changeAutoStart = async (row: VM.StorageItem, value: string | number | boolean) => {
    if (!canToggleStatus(row)) {
        row.autoStart = !Boolean(value);
        return;
    }
    switchingIds.value.push(row.id);
    const nextValue = Boolean(value);
    loading.value = true;
    await updateVMStorage({ id: row.id, autoStart: nextValue })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch((error) => {
            row.autoStart = !nextValue;
            throw error;
        })
        .finally(() => {
            loading.value = false;
            switchingIds.value = switchingIds.value.filter((id) => id !== row.id);
        });
};

const canToggleStatus = (row: VM.StorageItem) => {
    return (
        resourceExists(row.status) &&
        [RESOURCE_STATUS_RUNNING, RESOURCE_STATUS_INACTIVE].includes(normalizeStatus(row.status))
    );
};

const normalizeStatus = (status: string) => status?.toLowerCase() || '';

const resourceExists = (status: string) => normalizeStatus(status) !== RESOURCE_STATUS_DELETED;

const changeStatus = async (row: VM.StorageItem, status: 'running' | 'inactive') => {
    if (!canToggleStatus(row) || normalizeStatus(row.status) === status) {
        return;
    }
    const oldStatus = row.status;
    row.status = status;
    statusSwitchingIds.value.push(row.id);
    loading.value = true;
    await updateVMStorage({ id: row.id, status })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch((error) => {
            row.status = oldStatus;
            throw error;
        })
        .finally(() => {
            loading.value = false;
            statusSwitchingIds.value = statusSwitchingIds.value.filter((id) => id !== row.id);
        });
};

const openInlineEdit = async (row: VM.StorageItem) => {
    let res;
    try {
        res = await ElMessageBox.prompt(
            i18n.global.t('commons.table.description'),
            i18n.global.t('commons.button.edit'),
            {
                inputValue: row.description || '',
                inputType: 'textarea',
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            },
        );
    } catch {
        return;
    }
    const value = String(res.value || '');
    await updateVMStorage({ id: row.id, description: value }).then(() => {
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        search();
    });
};

const openDelete = (row: VM.StorageItem) => {
    deleteRef.value?.open(row);
};

watch(() => [props.vmEnvironmentLoaded, props.vmUnavailable], search, { immediate: true });
</script>

<style scoped lang="scss">
.storage-card-list {
    display: grid;
    grid-template-columns: repeat(3, minmax(360px, 1fr));
    gap: 14px;
}

.storage-card {
    display: flex;
    flex-direction: column;
    min-height: 240px;
    padding: 16px;
    border: 1px solid var(--el-card-border-color, var(--el-border-color-lighter));
    border-radius: 8px;
    background: var(--el-card-bg-color, var(--el-fill-color-blank));
}

.storage-card__header,
.storage-card__title,
.storage-card__meta-item,
.storage-card__detail-item,
.storage-card__actions {
    display: flex;
    align-items: center;
}

.storage-card__header {
    justify-content: space-between;
    gap: 12px;
}

.storage-card__title {
    min-width: 0;
    gap: 8px;
}

.storage-card__status {
    cursor: pointer;
}

.storage-card__name {
    overflow: hidden;
    color: var(--el-text-color-primary);
    font-size: 16px;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.storage-card__sub {
    width: fit-content;
    max-width: 100%;
    min-height: 22px;
    margin-top: 8px;
    color: var(--el-text-color-regular);
    font-size: 13px;
}

.storage-card__meta {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
    margin-top: 14px;
}

.storage-card__meta-item {
    display: block;
    min-width: 0;
    padding: 10px;
    border-radius: 6px;
    background: var(--el-fill-color-extra-light);
}

.storage-card__detail-item span {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.storage-card__meta-label {
    display: block;
    margin-bottom: 4px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.storage-card__meta-value {
    display: block;
    overflow: hidden;
    min-width: 0;
    color: var(--el-text-color-regular);
    font-size: 13px;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.storage-card__meta-value.is-enabled {
    color: var(--el-color-success);
}

.storage-card__meta-link {
    height: auto;
    padding: 0;
    font-size: 13px;
    font-weight: 600;
    line-height: normal;
}

.storage-vm-dropdown {
    line-height: 1;
}

:global(.storage-vm-dropdown-item) {
    max-width: 260px;
}

:global(.storage-vm-dropdown-item span) {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.storage-card__meta-item :deep(.el-switch) {
    height: 18px;
}

.storage-card__detail {
    display: grid;
    gap: 8px;
    margin-top: 16px;
    margin-bottom: 14px;
}

.storage-card__detail-item {
    min-width: 0;
    justify-content: space-between;
    gap: 12px;
}

.storage-card__detail-item strong {
    overflow: hidden;
    min-width: 0;
    color: var(--el-text-color-primary);
    font-size: 13px;
    font-weight: 500;
    text-align: right;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.storage-card__editable {
    display: flex;
    align-items: center;
    min-width: 0;
    gap: 6px;
}

.storage-card__editable .el-button {
    flex: none;
    padding: 0;
}

.storage-card__sub.storage-card__editable span,
.storage-card__detail-item.storage-card__editable strong {
    min-width: 0;
}

.storage-card__sub.storage-card__editable span {
    flex: initial;
    overflow-wrap: anywhere;
    white-space: normal;
}

.storage-card__detail-item.storage-card__editable strong {
    flex: 1;
}

.storage-card__detail-item.storage-card__editable > span {
    flex: none;
}

.storage-card__actions {
    justify-content: flex-end;
    gap: 12px;
    flex-wrap: wrap;
    margin-top: auto;
    padding-top: 14px;
    border-top: 1px solid var(--el-border-color-lighter);
}

.storage-card__actions :deep(.el-button) {
    margin-left: 0;
}

@media (max-width: 1280px) {
    .storage-card-list {
        grid-template-columns: repeat(2, minmax(360px, 1fr));
    }
}

@media (max-width: 768px) {
    .storage-card-list {
        grid-template-columns: 1fr;
    }
}
</style>
