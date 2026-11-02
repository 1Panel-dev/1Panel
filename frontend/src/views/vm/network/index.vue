<template>
    <LayoutContent :title="$t('xpack.vm.network')">
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
            <div class="network-card-list" v-if="items.length" v-loading="loading">
                <article v-for="item in items" :key="item.name" class="network-card">
                    <div class="network-card__header">
                        <div class="network-card__title">
                            <span class="network-card__name" :title="item.name">{{ item.name }}</span>
                            <el-tag round size="small" effect="plain">{{ formatType(item.type) }}</el-tag>
                        </div>
                        <el-dropdown v-permission placement="bottom" :disabled="loading || !canToggleStatus(item)">
                            <span class="network-card__status">
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
                    <div class="network-card__sub">
                        <span v-if="item.type === 'nat'">{{ item.cidr || '-' }}</span>
                        <span v-else-if="item.type === 'macvtap'">{{ item.sourceInterface || '-' }}</span>
                        <span v-else>{{ item.bridgeName || '-' }}</span>
                    </div>
                    <div class="network-card__meta">
                        <div class="network-card__meta-item">
                            <span class="network-card__meta-label">{{ $t('xpack.vm.resourceAutoStart') }}</span>
                            <el-switch
                                v-permission
                                size="small"
                                v-model="item.autoStart"
                                :disabled="!canToggleStatus(item)"
                                :loading="switchingIds.includes(item.id)"
                                @change="changeAutoStart(item, $event)"
                            />
                        </div>
                        <div class="network-card__meta-item">
                            <span class="network-card__meta-label">{{ $t('xpack.vm.vmCount') }}</span>
                            <el-dropdown
                                v-if="item.vmCount"
                                class="network-vm-dropdown"
                                placement="bottom"
                                trigger="click"
                                @command="goVMList"
                            >
                                <el-button class="network-card__meta-link" link type="primary">
                                    {{ item.vmCount || 0 }}
                                </el-button>
                                <template #dropdown>
                                    <el-dropdown-menu>
                                        <el-dropdown-item
                                            v-for="vm in item.vms || []"
                                            :key="vm.id"
                                            :command="vm.name"
                                            class="network-vm-dropdown-item"
                                        >
                                            {{ vm.name }}
                                        </el-dropdown-item>
                                    </el-dropdown-menu>
                                </template>
                            </el-dropdown>
                            <span v-else class="network-card__meta-value">0</span>
                        </div>
                        <div class="network-card__meta-item">
                            <span class="network-card__meta-label">{{ $t('xpack.vm.usedIPs') }}</span>
                            <span
                                v-if="item.type === 'bridge' || item.type === 'macvtap'"
                                class="network-card__meta-value"
                                :title="$t('xpack.vm.usedIPsUnavailable')"
                            >
                                {{ $t('xpack.vm.usedIPsUnavailable') }}
                            </span>
                            <el-button
                                v-else
                                class="network-card__meta-link"
                                link
                                type="primary"
                                @click="openUsedIPs(item)"
                            >
                                {{ item.usedIPs?.length || 0 }}
                            </el-button>
                        </div>
                    </div>
                    <div class="network-card__detail">
                        <div class="network-card__detail-item" v-if="item.type === 'nat'">
                            <span>{{ $t('xpack.vm.gateway') }}</span>
                            <strong>{{ item.gateway || '-' }}</strong>
                        </div>
                        <div class="network-card__detail-item" v-if="item.type === 'nat'">
                            <span>DHCP</span>
                            <strong>{{ formatDHCP(item) }}</strong>
                        </div>
                        <div class="network-card__detail-item" v-if="item.type === 'bridge'">
                            <span>{{ $t('xpack.vm.bridgeName') }}</span>
                            <strong>{{ item.bridgeName || '-' }}</strong>
                        </div>
                        <div class="network-card__detail-item" v-if="item.type === 'macvtap'">
                            <span>{{ $t('xpack.vm.sourceInterface') }}</span>
                            <strong>{{ item.sourceInterface || '-' }}</strong>
                        </div>
                        <div class="network-card__detail-item network-card__editable">
                            <span>{{ $t('commons.table.description') }}</span>
                            <strong :title="item.description">{{ item.description || '-' }}</strong>
                            <el-button
                                v-permission
                                link
                                type="primary"
                                icon="EditPen"
                                @click="openDescriptionEdit(item)"
                            />
                        </div>
                    </div>
                    <div class="network-card__actions">
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
    <ResourceDelete ref="deleteRef" resource-type="network" @deleted="search" />
    <el-dialog
        v-model="usedIPDialogVisible"
        :title="`${selectedNetwork?.name || ''} - ${$t('xpack.vm.usedIPs')}`"
        width="720px"
    >
        <ComplexTable :data="selectedNetwork?.usedIPs || []">
            <el-table-column prop="ip" label="IP" min-width="150" />
            <el-table-column prop="vmName" :label="$t('commons.table.name')" min-width="160" show-overflow-tooltip />
            <el-table-column prop="mac" label="MAC" min-width="170" show-overflow-tooltip />
            <el-table-column prop="source" :label="$t('xpack.vm.network')" min-width="150" show-overflow-tooltip>
                <template #default="{ row }">{{ row.source || '-' }}</template>
            </el-table-column>
        </ComplexTable>
    </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { VM } from '@/api/interface/vm';
import { listVMNetworks, updateVMNetwork } from '@/api/modules/vm';
import Operate from '@/views/vm/network/create/index.vue';
import ResourceDelete from '@/views/vm/components/resource-delete.vue';

const props = defineProps<{
    vmUnavailable: boolean;
    vmEnvironmentLoaded: boolean;
}>();
const loading = ref(false);
const searchName = ref('');
const items = ref<VM.NetworkItem[]>([]);
const switchingIds = ref<number[]>([]);
const statusSwitchingIds = ref<number[]>([]);
const operateRef = ref();
const deleteRef = ref<InstanceType<typeof ResourceDelete>>();
const usedIPDialogVisible = ref(false);
const selectedNetwork = ref<VM.NetworkItem>();
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
    await listVMNetworks()
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

const openUsedIPs = (row: VM.NetworkItem) => {
    selectedNetwork.value = row;
    usedIPDialogVisible.value = true;
};

const goVMList = (name: string) => {
    router.push({ path: '/vm/list', query: { name } });
};

const changeAutoStart = async (row: VM.NetworkItem, value: string | number | boolean) => {
    if (!canToggleStatus(row)) {
        row.autoStart = !Boolean(value);
        return;
    }
    switchingIds.value.push(row.id);
    const nextValue = Boolean(value);
    loading.value = true;
    await updateVMNetwork({ id: row.id, autoStart: nextValue })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch((error) => {
            row.autoStart = !nextValue;
            throw error;
        })
        .finally(() => {
            switchingIds.value = switchingIds.value.filter((id) => id !== row.id);
            loading.value = false;
        });
};

const canToggleStatus = (row: VM.NetworkItem) => {
    return (
        resourceExists(row.status) &&
        [RESOURCE_STATUS_RUNNING, RESOURCE_STATUS_INACTIVE].includes(normalizeStatus(row.status))
    );
};

const normalizeStatus = (status: string) => status?.toLowerCase() || '';

const resourceExists = (status: string) => normalizeStatus(status) !== RESOURCE_STATUS_DELETED;

const changeStatus = async (row: VM.NetworkItem, status: 'running' | 'inactive') => {
    if (!canToggleStatus(row) || normalizeStatus(row.status) === status) {
        return;
    }
    const oldStatus = row.status;
    row.status = status;
    statusSwitchingIds.value.push(row.id);
    loading.value = true;
    await updateVMNetwork({ id: row.id, status })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch((error) => {
            row.status = oldStatus;
            throw error;
        })
        .finally(() => {
            statusSwitchingIds.value = statusSwitchingIds.value.filter((id) => id !== row.id);
            loading.value = false;
        });
};

const openDescriptionEdit = async (row: VM.NetworkItem) => {
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
    loading.value = true;
    await updateVMNetwork({ id: row.id, description: String(res.value || '') })
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const openDelete = (row: VM.NetworkItem) => {
    deleteRef.value?.open(row);
};

const formatType = (type: string) => {
    if (type === 'nat') return 'NAT';
    if (type === 'bridge') return 'Bridge';
    if (type === 'macvtap') return 'Macvtap';
    return type || '-';
};

const formatDHCP = (row: VM.NetworkItem) => {
    if (!row.dhcpStart && !row.dhcpEnd) return '-';
    return `${row.dhcpStart || '-'} - ${row.dhcpEnd || '-'}`;
};

watch(() => [props.vmEnvironmentLoaded, props.vmUnavailable], search, { immediate: true });
</script>

<style scoped lang="scss">
.network-card-list {
    display: grid;
    grid-template-columns: repeat(3, minmax(360px, 1fr));
    gap: 14px;
}

.network-card {
    display: flex;
    flex-direction: column;
    min-height: 240px;
    padding: 16px;
    border: 1px solid var(--el-card-border-color, var(--el-border-color-lighter));
    border-radius: 8px;
    background: var(--el-card-bg-color, var(--el-fill-color-blank));
}

.network-card__header,
.network-card__title,
.network-card__meta-item,
.network-card__detail-item,
.network-card__actions {
    display: flex;
    align-items: center;
}

.network-card__header {
    justify-content: space-between;
    gap: 12px;
}

.network-card__title {
    min-width: 0;
    gap: 8px;
}

.network-card__status {
    cursor: pointer;
}

.network-card__name {
    overflow: hidden;
    color: var(--el-text-color-primary);
    font-size: 16px;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.network-card__sub {
    min-height: 22px;
    margin-top: 8px;
    color: var(--el-text-color-regular);
    font-size: 13px;
}

.network-card__meta {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 8px;
    margin-top: 14px;
}

.network-card__meta-item {
    display: block;
    min-width: 0;
    padding: 10px;
    border-radius: 6px;
    background: var(--el-fill-color-extra-light);
}

.network-card__detail-item span {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.network-card__meta-label {
    display: block;
    margin-bottom: 4px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.network-card__meta-value {
    display: block;
    overflow: hidden;
    min-width: 0;
    color: var(--el-text-color-primary);
    font-size: 13px;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.network-card__meta-value.is-enabled {
    color: var(--el-color-success);
}

.network-card__meta-link {
    height: auto;
    padding: 0;
    font-size: 13px;
    font-weight: 600;
    line-height: normal;
}

.network-card__meta-item :deep(.el-switch) {
    height: 18px;
}

.network-card__detail {
    display: grid;
    gap: 8px;
    margin-top: 16px;
    margin-bottom: 14px;
}

.network-card__detail-item {
    min-width: 0;
    justify-content: space-between;
    gap: 12px;
}

.network-card__detail-item strong {
    overflow: hidden;
    min-width: 0;
    color: var(--el-text-color-primary);
    font-size: 13px;
    font-weight: 500;
    text-align: right;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.network-card__editable .el-button {
    flex: none;
    padding: 0;
}

.network-card__detail-item.network-card__editable strong {
    flex: 1;
}

.network-card__detail-item.network-card__editable > span {
    flex: none;
}

.network-card__actions {
    justify-content: flex-end;
    gap: 12px;
    flex-wrap: wrap;
    margin-top: auto;
    padding-top: 14px;
    border-top: 1px solid var(--el-border-color-lighter);
}

.network-card__actions :deep(.el-button) {
    margin-left: 0;
}

.network-vm-dropdown {
    line-height: 1;
}

:global(.network-vm-dropdown-item) {
    max-width: 260px;
}

:global(.network-vm-dropdown-item span) {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

@media (max-width: 1280px) {
    .network-card-list {
        grid-template-columns: repeat(2, minmax(360px, 1fr));
    }
}

@media (max-width: 768px) {
    .network-card-list {
        grid-template-columns: 1fr;
    }
}
</style>
