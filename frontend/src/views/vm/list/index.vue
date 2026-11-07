<template>
    <div>
        <div v-loading="fullPageLoading">
            <LayoutContent v-loading="loading" :title="$t('xpack.vm.title')">
                <template #leftToolBar>
                    <el-button
                        v-permission
                        type="primary"
                        :disabled="vmUnavailable || vmCountLimitReached || vmCountLoading"
                        @click="openCreate"
                    >
                        {{ $t('commons.button.create') }}
                    </el-button>
                    <el-button v-permission :disabled="vmUnavailable || loading" @click="syncVMs">
                        {{ $t('commons.button.sync') }}
                    </el-button>
                </template>
                <template #rightToolBar>
                    <el-radio-group v-model="viewMode">
                        <el-tooltip :content="$t('commons.table.card')" placement="top">
                            <el-radio-button value="card">
                                <el-icon><Grid /></el-icon>
                            </el-radio-button>
                        </el-tooltip>
                        <el-tooltip :content="$t('commons.table.table')" placement="top">
                            <el-radio-button value="table">
                                <el-icon><List /></el-icon>
                            </el-radio-button>
                        </el-tooltip>
                    </el-radio-group>
                    <TableSearch v-model:searchName="searchName" @search="search" />
                    <TableRefresh @search="refreshPage" />
                </template>
                <template #main>
                    <el-alert
                        v-if="vmCountLimit !== undefined && vmCountLimit > 0"
                        class="mb-4"
                        :title="$t('xpack.vm.countLimitHelper', [vmCountLimit])"
                        :type="vmCountLimitReached ? 'warning' : 'info'"
                        :closable="false"
                        show-icon
                    />
                    <template v-if="viewMode === 'card'">
                        <div v-if="items.length" class="vm-card-grid">
                            <el-card v-for="row in items" :key="row.name" class="vm-card" shadow="hover">
                                <div class="vm-card-header">
                                    <div class="vm-card-title">
                                        <el-button
                                            link
                                            type="primary"
                                            class="vm-card-name"
                                            :disabled="isFailedCreate(row) || isDeleted(row)"
                                            @click="openDetail(row)"
                                        >
                                            {{ row.name }}
                                        </el-button>
                                    </div>
                                    <div class="vm-card-status-actions">
                                        <el-tooltip
                                            v-if="!isFailedCreate(row)"
                                            :content="$t('xpack.vm.healthCheck')"
                                            placement="top"
                                        >
                                            <el-button
                                                class="vm-health-check-button"
                                                link
                                                type="primary"
                                                :disabled="!canHealthCheck(row)"
                                                :icon="FirstAidKit"
                                                @click="openHealthCheck(row)"
                                            />
                                        </el-tooltip>
                                        <div class="vm-status-with-log">
                                            <el-button
                                                v-if="canOpenCreatingTaskLog(row)"
                                                size="small"
                                                type="primary"
                                                plain
                                                round
                                                @click.stop="openTaskLog(row.createTaskID)"
                                            >
                                                {{ $t('commons.status.creating') }}
                                                <el-icon class="is-loading">
                                                    <Loading />
                                                </el-icon>
                                            </el-button>
                                            <el-dropdown
                                                v-else
                                                v-permission
                                                placement="bottom"
                                                :disabled="
                                                    loading || isCreating(row) || isFailedCreate(row) || isDeleted(row)
                                                "
                                            >
                                                <span class="vm-status-operate vm-card-status-trigger">
                                                    <Status :status="vmStatus(row.status)" :operate="true" />
                                                </span>
                                                <template #dropdown>
                                                    <el-dropdown-menu>
                                                        <el-dropdown-item
                                                            :disabled="checkStatus('start', row)"
                                                            @click="operate(row, 'start')"
                                                        >
                                                            {{ $t('commons.operate.start') }}
                                                        </el-dropdown-item>
                                                        <el-dropdown-item
                                                            :disabled="checkStatus('shutdown', row)"
                                                            @click="operate(row, 'shutdown')"
                                                        >
                                                            {{ $t('commons.operate.stop') }}
                                                        </el-dropdown-item>
                                                        <el-dropdown-item
                                                            :disabled="checkStatus('reboot', row)"
                                                            @click="operate(row, 'reboot')"
                                                        >
                                                            {{ $t('commons.button.restart') }}
                                                        </el-dropdown-item>
                                                        <el-dropdown-item
                                                            :disabled="checkStatus('suspend', row)"
                                                            @click="operate(row, 'suspend')"
                                                        >
                                                            {{ $t('xpack.vm.suspend') }}
                                                        </el-dropdown-item>
                                                        <el-dropdown-item
                                                            :disabled="checkStatus('resume', row)"
                                                            @click="operate(row, 'resume')"
                                                        >
                                                            {{ $t('xpack.vm.resume') }}
                                                        </el-dropdown-item>
                                                        <el-dropdown-item
                                                            :disabled="checkStatus('destroy', row)"
                                                            @click="operate(row, 'destroy')"
                                                        >
                                                            {{ $t('xpack.vm.destroy') }}
                                                        </el-dropdown-item>
                                                    </el-dropdown-menu>
                                                </template>
                                            </el-dropdown>
                                            <el-tooltip
                                                v-if="canOpenCreateTaskLog(row)"
                                                :content="$t('commons.button.log')"
                                                placement="top"
                                            >
                                                <el-icon
                                                    class="vm-create-failed-log"
                                                    @click.stop="openTaskLog(row.createTaskID)"
                                                >
                                                    <WarningFilled />
                                                </el-icon>
                                            </el-tooltip>
                                        </div>
                                    </div>
                                </div>

                                <div class="vm-card-resource">
                                    <div class="vm-card-resource-item">
                                        <span>{{ $t('xpack.vm.cpu') }}</span>
                                        <strong>{{ row.cpu }} {{ $t('commons.units.core') }}</strong>
                                    </div>
                                    <div class="vm-card-resource-item">
                                        <span>{{ $t('xpack.vm.memory') }}</span>
                                        <strong>{{ computeSizeFromMB(row.memory) }}</strong>
                                    </div>
                                    <div class="vm-card-resource-item">
                                        <span>{{ $t('xpack.vm.disk') }}</span>
                                        <strong>{{ formatDiskSize(row.diskSize) }}</strong>
                                    </div>
                                </div>

                                <div class="vm-card-info">
                                    <div>
                                        <span>{{ $t('xpack.vm.network') }}</span>
                                        <strong>{{ row.network || '-' }}</strong>
                                    </div>
                                    <div>
                                        <span>{{ $t('xpack.vm.storagePool') }}</span>
                                        <strong>{{ row.storagePool || '-' }}</strong>
                                    </div>
                                    <div>
                                        <span>{{ $t('xpack.vm.diskPath') }}</span>
                                        <strong>{{ row.diskPath || '-' }}</strong>
                                    </div>
                                    <div>
                                        <span>{{ $t('commons.table.description') }}</span>
                                        <strong>{{ row.description || '-' }}</strong>
                                    </div>
                                </div>

                                <div class="vm-card-actions">
                                    <el-button
                                        v-if="!isFailedCreate(row)"
                                        class="vm-console-button"
                                        type="primary"
                                        plain
                                        size="small"
                                        :disabled="!canOpenConsole(row)"
                                        icon="Monitor"
                                        @click="openConsole(row)"
                                    >
                                        VNC {{ $t('xpack.vm.console') }}
                                    </el-button>
                                    <div class="vm-card-secondary-actions">
                                        <el-button
                                            v-if="!isFailedCreate(row)"
                                            link
                                            type="primary"
                                            :disabled="!isRunning(row)"
                                            @click="openMonitor(row)"
                                        >
                                            {{ $t('xpack.vm.monitor') }}
                                        </el-button>
                                        <el-button
                                            v-if="!isFailedCreate(row)"
                                            v-permission
                                            link
                                            type="primary"
                                            :disabled="!canRename(row)"
                                            @click="openRename(row)"
                                        >
                                            {{ $t('xpack.vm.rename') }}
                                        </el-button>
                                        <el-button
                                            v-permission
                                            link
                                            type="primary"
                                            :disabled="!canEdit(row)"
                                            @click="openEdit(row)"
                                        >
                                            {{ $t('commons.button.edit') }}
                                        </el-button>
                                        <el-button
                                            v-if="!isFailedCreate(row)"
                                            link
                                            type="primary"
                                            :disabled="loading || isCreating(row) || isDeleted(row)"
                                            @click="disksRef?.acceptParams(row)"
                                        >
                                            {{ $t('xpack.vm.disks') }}
                                        </el-button>
                                        <el-button
                                            v-if="!isFailedCreate(row)"
                                            link
                                            type="primary"
                                            :disabled="!canOpenSnapshot(row)"
                                            @click="openSnapshot(row)"
                                        >
                                            {{ $t('xpack.vm.snapshot') }}
                                        </el-button>
                                        <el-button
                                            v-if="!isFailedCreate(row)"
                                            v-permission
                                            link
                                            type="primary"
                                            :disabled="loading || !canOpenCreateTemplate(row)"
                                            @click="openCreateTemplate(row)"
                                        >
                                            {{ $t('xpack.vm.convertToTemplate') }}
                                        </el-button>
                                        <el-button
                                            v-permission
                                            link
                                            type="primary"
                                            :disabled="loading"
                                            @click="openDelete(row)"
                                        >
                                            {{ $t('commons.button.delete') }}
                                        </el-button>
                                    </div>
                                </div>
                            </el-card>
                        </div>
                        <el-empty v-else />
                        <div v-if="paginationConfig.total > 0" class="vm-card-pagination">
                            <el-pagination
                                v-model:current-page="paginationConfig.currentPage"
                                v-model:page-size="paginationConfig.pageSize"
                                :total="paginationConfig.total"
                                :page-sizes="[10, 20, 50, 100]"
                                layout="total, sizes, prev, pager, next, jumper"
                                @size-change="handleSizeChange"
                                @current-change="search"
                            />
                        </div>
                    </template>
                    <ComplexTable v-else :data="items" :pagination-config="paginationConfig" @search="search">
                        <el-table-column
                            :label="$t('commons.table.name')"
                            min-width="180"
                            fixed="left"
                            show-overflow-tooltip
                        >
                            <template #default="{ row }">
                                <el-button
                                    link
                                    type="primary"
                                    :disabled="isFailedCreate(row) || isDeleted(row)"
                                    @click="openDetail(row)"
                                >
                                    {{ row.name }}
                                </el-button>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('commons.table.status')" width="130">
                            <template #default="{ row }">
                                <div class="vm-status-with-log">
                                    <el-button
                                        v-if="canOpenCreatingTaskLog(row)"
                                        size="small"
                                        type="primary"
                                        plain
                                        round
                                        @click.stop="openTaskLog(row.createTaskID)"
                                    >
                                        {{ $t('commons.status.creating') }}
                                        <el-icon class="is-loading">
                                            <Loading />
                                        </el-icon>
                                    </el-button>
                                    <el-dropdown
                                        v-else
                                        v-permission
                                        placement="bottom"
                                        :disabled="loading || isCreating(row) || isFailedCreate(row) || isDeleted(row)"
                                    >
                                        <span class="vm-status-operate vm-card-status-trigger">
                                            <Status :status="vmStatus(row.status)" :operate="true" />
                                        </span>
                                        <template #dropdown>
                                            <el-dropdown-menu>
                                                <el-dropdown-item
                                                    :disabled="checkStatus('start', row)"
                                                    @click="operate(row, 'start')"
                                                >
                                                    {{ $t('commons.operate.start') }}
                                                </el-dropdown-item>
                                                <el-dropdown-item
                                                    :disabled="checkStatus('shutdown', row)"
                                                    @click="operate(row, 'shutdown')"
                                                >
                                                    {{ $t('commons.operate.stop') }}
                                                </el-dropdown-item>
                                                <el-dropdown-item
                                                    :disabled="checkStatus('reboot', row)"
                                                    @click="operate(row, 'reboot')"
                                                >
                                                    {{ $t('commons.button.restart') }}
                                                </el-dropdown-item>
                                                <el-dropdown-item
                                                    :disabled="checkStatus('suspend', row)"
                                                    @click="operate(row, 'suspend')"
                                                >
                                                    {{ $t('xpack.vm.suspend') }}
                                                </el-dropdown-item>
                                                <el-dropdown-item
                                                    :disabled="checkStatus('resume', row)"
                                                    @click="operate(row, 'resume')"
                                                >
                                                    {{ $t('xpack.vm.resume') }}
                                                </el-dropdown-item>
                                                <el-dropdown-item
                                                    :disabled="checkStatus('destroy', row)"
                                                    @click="operate(row, 'destroy')"
                                                >
                                                    {{ $t('xpack.vm.destroy') }}
                                                </el-dropdown-item>
                                            </el-dropdown-menu>
                                        </template>
                                    </el-dropdown>
                                    <el-tooltip
                                        v-if="canOpenCreateTaskLog(row)"
                                        :content="$t('commons.button.log')"
                                        v-bind="TABLE_TOOLTIP_OPTIONS"
                                    >
                                        <el-icon
                                            class="vm-create-failed-log"
                                            @click.stop="openTaskLog(row.createTaskID)"
                                        >
                                            <WarningFilled />
                                        </el-icon>
                                    </el-tooltip>
                                </div>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('xpack.vm.cpu')" width="100">
                            <template #default="{ row }">{{ row.cpu }} {{ $t('commons.units.core') }}</template>
                        </el-table-column>
                        <el-table-column :label="$t('xpack.vm.memory')" width="120">
                            <template #default="{ row }">{{ computeSizeFromMB(row.memory) }}</template>
                        </el-table-column>
                        <el-table-column :label="$t('xpack.vm.disk')" width="110">
                            <template #default="{ row }">{{ formatDiskSize(row.diskSize) }}</template>
                        </el-table-column>
                        <el-table-column
                            :label="$t('xpack.vm.network')"
                            prop="network"
                            min-width="140"
                            show-overflow-tooltip
                        >
                            <template #default="{ row }">{{ row.network || '-' }}</template>
                        </el-table-column>
                        <el-table-column
                            :label="$t('xpack.vm.storagePool')"
                            prop="storagePool"
                            min-width="140"
                            show-overflow-tooltip
                        >
                            <template #default="{ row }">{{ row.storagePool || '-' }}</template>
                        </el-table-column>
                        <el-table-column
                            :label="$t('xpack.vm.diskPath')"
                            prop="diskPath"
                            min-width="220"
                            show-overflow-tooltip
                        >
                            <template #default="{ row }">{{ row.diskPath || '-' }}</template>
                        </el-table-column>
                        <el-table-column
                            :label="$t('commons.table.description')"
                            prop="description"
                            min-width="180"
                            show-overflow-tooltip
                        >
                            <template #default="{ row }">{{ row.description || '-' }}</template>
                        </el-table-column>
                        <fu-table-operations
                            :buttons="tableOperationButtons"
                            :ellipsis="2"
                            :label="$t('commons.table.operate')"
                            width="220"
                            fixed="right"
                        />
                    </ComplexTable>
                </template>
            </LayoutContent>

            <Operate
                ref="operateRef"
                :resource-limit="resourceLimit"
                :resource-limit-loading="resourceLimitLoading"
                @open-log="openTaskLog"
            />
            <Detail ref="detailRef" />
            <Snapshot ref="snapshotRef" />
            <Disks ref="disksRef" @refresh="search" />
            <DrawerPro
                v-model="consoleVisible"
                :header="$t('xpack.vm.console')"
                :resource="consoleTitle"
                size="large"
                :full-screen="true"
                @close="closeConsole"
            >
                <template #content>
                    <NoVNC
                        v-if="consoleVisible && consolePath"
                        :key="consolePath"
                        :path="consolePath"
                        class="vm-console-drawer"
                    />
                </template>
            </DrawerPro>

            <Monitor ref="monitorRef" />
            <HealthCheck ref="healthCheckRef" />

            <TemplateDialog
                ref="templateRef"
                :storage-items="storageItems"
                :storage-loading="storageItemsLoading"
                @task-success="openTaskLog"
            />
            <RenameDialog ref="renameRef" @task-success="openTaskLog" />
            <DeleteDialog ref="deleteRef" @task-success="openTaskLog" />

            <TaskLog ref="taskLogRef" width="70%" @close="refreshAfterTask" />
        </div>
    </div>
</template>

<script setup lang="ts">
import { TABLE_TOOLTIP_OPTIONS } from '@/components/table/shared';
import { computed, reactive, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import TaskLog from '@/components/log/task/index.vue';
import LayoutContent from '@/components/layout-content/index.vue';
import DrawerPro from '@/components/drawer-pro/index.vue';
import Operate from './operate/index.vue';
import Detail from './detail/index.vue';
import Monitor from './monitor/index.vue';
import HealthCheck from './health-check/index.vue';
import Snapshot from './snapshot/index.vue';
import Disks from './disks/index.vue';
import TemplateDialog from './template/index.vue';
import RenameDialog from './rename/index.vue';
import DeleteDialog from './delete/index.vue';
import NoVNC from '@/views/vm/console/novnc.vue';
import { VM } from '@/api/interface/vm';
import type { Dashboard } from '@/api/interface/dashboard';
import { FirstAidKit, Grid, List, Loading, WarningFilled } from '@element-plus/icons-vue';
import { loadCurrentInfo } from '@/api/modules/dashboard';
import { getVMConsole, listVMStorages, operateVM, searchVMs, syncVMs as syncVMResources } from '@/api/modules/vm';
import { computeSizeFromMB } from '@/utils/size.js';
import { GlobalStore } from '@/store';
import { isVMMemoryAllocated } from '@/utils/vm';

const BYTE_PER_MIB = 1024 * 1024;
const MEMORY_RESERVE_MIN = 1024 * BYTE_PER_MIB;
const VM_LIST_VIEW_MODE_KEY = 'vm-list-view-mode';
const vmCountLimit = ref<number>();
type VMListViewMode = 'card' | 'table';
const props = defineProps<{
    vmUnavailable: boolean;
    vmEnvironmentLoaded: boolean;
}>();

const getStoredViewMode = (): VMListViewMode => {
    const mode = localStorage.getItem(VM_LIST_VIEW_MODE_KEY);
    return mode === 'card' || mode === 'table' ? mode : 'card';
};

const loading = ref(false);
const fullPageLoading = ref(false);
const searchName = ref('');
const viewMode = ref<VMListViewMode>(getStoredViewMode());
const items = ref<VM.Item[]>([]);
const resourceItems = ref<VM.Item[]>([]);
const vmTotal = ref<number>();
const storageItems = ref<VM.StorageItem[]>([]);
const hostCurrentInfoLoading = ref(false);
const vmResourceItemsLoading = ref(false);
const storageItemsLoading = ref(false);
const hostCurrentInfo = ref<Dashboard.CurrentInfo>();
const taskLogRef = ref();
const operateRef = ref();
const detailRef = ref();
const snapshotRef = ref();
const disksRef = ref();
const globalStore = GlobalStore();
const route = useRoute();
const consoleVisible = ref(false);
const consoleTitle = ref('');
const consolePath = ref('');
const monitorRef = ref();
const healthCheckRef = ref();
const templateRef = ref();
const renameRef = ref();
const deleteRef = ref();

const getQueryName = () => {
    const name = route.query.name;
    return Array.isArray(name) ? name[0] || '' : String(name || '');
};

searchName.value = getQueryName();

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});

const tableOperationButtons = [
    {
        label: `VNC ${i18n.global.t('xpack.vm.console')}`,
        disabled: (row: VM.Item) => !canOpenConsole(row),
        click: (row: VM.Item) => openConsole(row),
    },
    {
        label: i18n.global.t('xpack.vm.healthCheck'),
        show: (row: VM.Item) => !isFailedCreate(row),
        disabled: (row: VM.Item) => !canHealthCheck(row),
        click: (row: VM.Item) => openHealthCheck(row),
    },
    {
        label: i18n.global.t('xpack.vm.monitor'),
        show: (row: VM.Item) => !isFailedCreate(row),
        disabled: (row: VM.Item) => !isRunning(row),
        click: (row: VM.Item) => openMonitor(row),
    },
    {
        label: i18n.global.t('xpack.vm.rename'),
        permission: true,
        show: (row: VM.Item) => !isFailedCreate(row),
        disabled: (row: VM.Item) => !canRename(row),
        click: (row: VM.Item) => openRename(row),
    },
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        disabled: (row: VM.Item) => !canEdit(row),
        click: (row: VM.Item) => openEdit(row),
    },
    {
        label: i18n.global.t('xpack.vm.disks'),
        show: (row: VM.Item) => !isFailedCreate(row),
        disabled: (row: VM.Item) => loading.value || isCreating(row) || isDeleted(row),
        click: (row: VM.Item) => disksRef.value?.acceptParams(row),
    },
    {
        label: i18n.global.t('xpack.vm.snapshot'),
        show: (row: VM.Item) => !isFailedCreate(row),
        disabled: (row: VM.Item) => !canOpenSnapshot(row),
        click: (row: VM.Item) => openSnapshot(row),
    },
    {
        label: i18n.global.t('xpack.vm.convertToTemplate'),
        permission: true,
        show: (row: VM.Item) => !isFailedCreate(row),
        disabled: (row: VM.Item) => loading.value || !canOpenCreateTemplate(row),
        click: (row: VM.Item) => openCreateTemplate(row),
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        disabled: () => loading.value,
        click: (row: VM.Item) => openDelete(row),
    },
];

const vmUnavailable = computed(() => props.vmUnavailable);
const vmCountLimitReached = computed(
    () => (vmCountLimit.value ?? 0) > 0 && (vmTotal.value || 0) >= (vmCountLimit.value ?? 0),
);
const vmCountLoading = computed(
    () =>
        vmCountLimit.value === undefined ||
        (vmCountLimit.value > 0 && (vmTotal.value === undefined || vmResourceItemsLoading.value)),
);
const vmAllocatedMemoryBytes = computed(() =>
    resourceItems.value
        .filter((item) => isVMMemoryAllocated(item.status))
        .reduce((total, item) => total + (Number(item.memory) || 0) * BYTE_PER_MIB, 0),
);
const memoryReservedBytes = computed(() => {
    const total = hostCurrentInfo.value?.memoryTotal || 0;
    if (!total) {
        return 0;
    }
    return MEMORY_RESERVE_MIN;
});

const resourceLimit = computed(() => ({
    cpuTotal: hostCurrentInfo.value?.cpuTotal || 0,
    memoryTotal: hostCurrentInfo.value?.memoryTotal || 0,
    memoryAllocated: vmAllocatedMemoryBytes.value,
    memoryReserved: memoryReservedBytes.value,
}));
const resourceLimitLoading = computed(() => hostCurrentInfoLoading.value || vmResourceItemsLoading.value);

const vmStatus = (status: string) => {
    if (status === 'shut off') {
        return 'stopped';
    }
    return status;
};

const checkStatus = (action: string, row: VM.Item) => {
    return !canOperate(action, row);
};

const canOperate = (action: string, row: VM.Item) => {
    if (loading.value || isCreating(row) || isFailedCreate(row) || isDeleted(row)) {
        return false;
    }
    switch (action) {
        case 'start':
            return isShutOff(row);
        case 'shutdown':
        case 'reboot':
        case 'suspend':
            return isRunning(row) || row.status === 'blocked';
        case 'resume':
            return isPaused(row) || row.status === 'pmsuspended';
        case 'destroy':
            return isVMMemoryAllocated(row.status) || row.status === 'crashed';
        default:
            return false;
    }
};

const storageUsedPercent = (item: VM.StorageItem) => {
    if (item.usedPercent !== undefined) {
        return Number(item.usedPercent) || 0;
    }
    if (!item.capacity) {
        return 0;
    }
    return ((item.capacity - (item.available || 0)) / item.capacity) * 100;
};

const search = async () => {
    if (vmUnavailable.value) {
        items.value = [];
        return;
    }
    loading.value = true;
    await searchVMs({
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        info: searchName.value,
    })
        .then((res) => {
            items.value = res.data.items || [];
            paginationConfig.total = res.data.total || 0;
            vmCountLimit.value = res.data.limit;
        })
        .finally(() => {
            loading.value = false;
        });
};

const handleSizeChange = () => {
    paginationConfig.currentPage = 1;
    search();
};

const refreshPage = async () => {
    fullPageLoading.value = true;
    try {
        await search();
        void loadHostResources();
    } finally {
        fullPageLoading.value = false;
    }
};

const syncVMs = async () => {
    if (vmUnavailable.value || loading.value) {
        return;
    }
    fullPageLoading.value = true;
    try {
        await syncVMResources();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await search();
        void loadHostResources();
    } finally {
        fullPageLoading.value = false;
    }
};

const loadHostResources = async () => {
    if (vmUnavailable.value) {
        return;
    }
    await Promise.allSettled([loadHostCurrentInfo(), loadVMResourceItems(), loadStorageItems()]);
};

const loadHostCurrentInfo = async () => {
    hostCurrentInfoLoading.value = true;
    try {
        const res = await loadCurrentInfo('all', 'all', globalStore.currentNode || 'local');
        hostCurrentInfo.value = res.data;
    } finally {
        hostCurrentInfoLoading.value = false;
    }
};

const loadVMResourceItems = async () => {
    vmResourceItemsLoading.value = true;
    try {
        const res = await searchVMs({ page: 1, pageSize: 10000, info: '' });
        resourceItems.value = res.data.items || [];
        vmTotal.value = res.data.total || 0;
        vmCountLimit.value = res.data.limit;
    } finally {
        vmResourceItemsLoading.value = false;
    }
};

const loadStorageItems = async () => {
    storageItemsLoading.value = true;
    try {
        const res = await listVMStorages();
        storageItems.value = (res.data || []).map((item) => ({
            ...item,
            usedPercent: storageUsedPercent(item),
        }));
    } finally {
        storageItemsLoading.value = false;
    }
};

const openCreate = () => {
    if (vmUnavailable.value || vmCountLimitReached.value || vmCountLoading.value) {
        return;
    }
    operateRef.value?.openCreate();
};

const openEdit = (row: VM.Item) => {
    if (!canEdit(row)) {
        return;
    }
    operateRef.value?.openEdit(row);
};

const openDetail = async (row: VM.Item) => {
    if (isFailedCreate(row) || isDeleted(row)) {
        return;
    }
    detailRef.value?.acceptParams(row.name);
};

const openSnapshot = (row: VM.Item) => {
    if (!canOpenSnapshot(row)) {
        return;
    }
    snapshotRef.value?.acceptParams(row);
};

const openCreateTemplate = (row: VM.Item) => {
    if (!canOpenCreateTemplate(row)) {
        return;
    }
    templateRef.value?.open(row);
};

const openRename = (row: VM.Item) => {
    if (!canRename(row)) {
        MsgError(i18n.global.t('xpack.vm.renameRunningHelper'));
        return;
    }
    renameRef.value?.open(row);
};

const canOpenConsole = (row: VM.Item) => {
    return isRunning(row) && row.graphics?.type?.toLowerCase() === 'vnc';
};

const buildConsolePath = (row: VM.Item, graphics: VM.Graphics) => {
    if (graphics.url) {
        return graphics.url;
    }
    const apiBase = (import.meta.env.VITE_API_URL as string).replace(/\/$/, '');
    const params = new URLSearchParams({
        name: row.name,
        operateNode: globalStore.currentNode || 'local',
    });
    return `${apiBase}/vms/console/ws?${params.toString()}`.replace(/^\//, '');
};

const openConsole = async (row: VM.Item) => {
    if (!canOpenConsole(row)) {
        MsgError(i18n.global.t('xpack.vm.consoleNotReady'));
        return;
    }
    const res = await getVMConsole(row.name);
    const graphics = res.data || row.graphics;
    if (!graphics || graphics.type?.toLowerCase() !== 'vnc') {
        MsgError(i18n.global.t('xpack.vm.consoleNotReady'));
        return;
    }
    consoleTitle.value = row.name;
    consolePath.value = buildConsolePath(row, graphics);
    consoleVisible.value = true;
};

const closeConsole = () => {
    consoleVisible.value = false;
    consoleTitle.value = '';
    consolePath.value = '';
};

const openMonitor = async (row: VM.Item) => {
    monitorRef.value?.open(row);
};

const openHealthCheck = (row: VM.Item) => {
    if (!canHealthCheck(row)) {
        return;
    }
    healthCheckRef.value?.open(row.name);
};

const openDelete = async (row: VM.Item) => {
    deleteRef.value?.open(row);
};

const operate = async (row: VM.Item, action: string) => {
    if (!canOperate(action, row)) {
        return;
    }
    await ElMessageBox.confirm(
        i18n.global.t('xpack.vm.operateConfirm', [i18n.global.t(`xpack.vm.${action}`), row.name]),
        i18n.global.t(`xpack.vm.${action}`),
    );
    loading.value = true;
    await operateVM(row.name, action)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value?.openWithTaskID(taskID, true);
};

const refreshAfterTask = () => {
    search();
    void loadHostResources();
};

const isRunning = (row: VM.Item) => row.status === 'running';
const isPaused = (row: VM.Item) => row.status === 'paused';
const isShutOff = (row: VM.Item) => row.status === 'shut off';
const isDeleted = (row: VM.Item) => row.status?.toLowerCase() === 'deleted';
const isCreating = (row: VM.Item) => row.status?.toLowerCase() === 'creating';
const isFailedCreate = (row: VM.Item) => row.status === 'failed' && !!row.createTaskID;
const canOpenSnapshot = (row: VM.Item) => !loading.value && !isCreating(row) && !isFailedCreate(row) && !isDeleted(row);
const canHealthCheck = (row: VM.Item) => !loading.value && !isCreating(row) && !isFailedCreate(row) && !isDeleted(row);
const canCreateTemplate = (row: VM.Item) => isShutOff(row);
const canOpenCreateTemplate = (row: VM.Item) => canCreateTemplate(row) && !storageItemsLoading.value;
const canRename = (row: VM.Item) => isShutOff(row);
const canEdit = (row: VM.Item) => isShutOff(row);
const canOpenCreateTaskLog = (row: VM.Item) => isFailedCreate(row);
const canOpenCreatingTaskLog = (row: VM.Item) => isCreating(row) && !!row.createTaskID;

const formatDiskSize = (size: number) => {
    if (!size) {
        return '-';
    }
    return `${size} GiB`;
};

watch(viewMode, (mode) => {
    localStorage.setItem(VM_LIST_VIEW_MODE_KEY, mode);
});

watch(
    () => [props.vmEnvironmentLoaded, props.vmUnavailable],
    async ([loaded, unavailable]) => {
        if (!loaded) {
            return;
        }
        if (unavailable) {
            items.value = [];
            resourceItems.value = [];
            storageItems.value = [];
            return;
        }
        fullPageLoading.value = true;
        try {
            await search();
            await loadHostResources();
            if (route.query.create === '1') {
                openCreate();
            }
        } finally {
            fullPageLoading.value = false;
        }
    },
    { immediate: true },
);

watch(
    () => route.query.name,
    () => {
        searchName.value = getQueryName();
        paginationConfig.currentPage = 1;
        if (props.vmEnvironmentLoaded && !vmUnavailable.value) {
            search();
        }
    },
);
</script>

<style scoped lang="scss">
.vm-status-operate {
    display: inline-flex;
    align-items: center;
    gap: 4px;
}

.vm-status-with-log {
    display: inline-flex;
    align-items: center;
    gap: 6px;
}

.vm-create-failed-log {
    cursor: pointer;
    color: var(--el-color-danger);
    font-size: 15px;
}

.vm-card-status-trigger {
    cursor: pointer;
}

.vm-status-loading {
    color: var(--el-color-primary);
}

.vm-card-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 14px;
}

.vm-card {
    min-width: 0;

    :deep(.el-card__body) {
        display: flex;
        min-height: 248px;
        flex-direction: column;
        padding: 16px;
    }
}

.vm-card-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
}

.vm-card-title {
    min-width: 0;
}

.vm-card-status-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 10px;
    flex-wrap: wrap;
    flex-shrink: 0;
    max-width: 58%;
}

.vm-card-status-actions :deep(.el-button) {
    margin-left: 0;
}

.vm-health-check-button {
    font-size: 18px;
}

.vm-card-name {
    max-width: 100%;
    height: auto;
    padding: 0;
    font-size: 15px;
    font-weight: 600;
    line-height: 22px;
    white-space: normal;
    text-align: left;
}

.vm-card-uuid {
    display: block;
    margin-top: 2px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.vm-card-resource {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 8px;
    margin-top: 14px;
}

.vm-card-resource-item {
    min-width: 0;
    padding: 10px;
    border-radius: 6px;
    background: var(--el-fill-color-extra-light);
}

.vm-card-resource-item span,
.vm-card-info span {
    display: block;
    margin-bottom: 4px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.vm-card-resource-item strong,
.vm-card-info strong {
    display: block;
    min-width: 0;
    overflow: hidden;
    color: var(--el-text-color-primary);
    font-size: 13px;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.vm-card-info {
    display: grid;
    gap: 8px;
    margin-top: 14px;
}

.vm-card-info > div {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    min-width: 0;
}

.vm-card-info span {
    flex: none;
    margin-bottom: 0;
}

.vm-card-info strong {
    font-weight: 500;
    text-align: right;
}

.vm-card-actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
    margin-top: auto;
    padding-top: 14px;
    border-top: 1px solid var(--el-border-color-lighter);
}

.vm-card-actions :deep(.el-button) {
    margin-left: 0;
}

.vm-console-button {
    min-width: 108px;
    font-weight: 600;
}

.vm-card-secondary-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 14px;
    flex-wrap: wrap;
}

.vm-card-secondary-actions :deep(.el-button) {
    margin-left: 0;
}

.vm-card-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
}

.vm-host-resource-card {
    :deep(.el-card__body) {
        padding: 16px;
    }
}

.vm-host-resource-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 14px;
}

.vm-host-resource-title {
    margin-right: 10px;
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
}

.vm-host-resource-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
}

.vm-resource-item {
    min-width: 0;
    padding: 12px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    background: var(--el-fill-color-extra-light);
}

.vm-resource-item.warning {
    border-color: var(--el-color-warning-light-5);
    background: var(--el-color-warning-light-9);
}

.vm-resource-item.danger {
    border-color: var(--el-color-danger-light-5);
    background: var(--el-color-danger-light-9);
}

.vm-resource-item-head,
.vm-resource-meta {
    display: flex;
    align-items: center;
    gap: 12px;
}

.vm-resource-item-head {
    justify-content: space-between;
    margin-bottom: 10px;
    font-weight: 600;
    color: var(--el-text-color-primary);
}

.vm-resource-meta {
    justify-content: flex-start;
    flex-wrap: wrap;
    margin-top: 8px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.vm-resource-meta span {
    min-width: 0;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.vm-helper-prompt {
    z-index: 10;
    top: 300px;
}

.vm-console-drawer {
    height: calc(100vh - 96px);
}

.secondary-text {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.input-help.is-warning {
    color: var(--el-color-warning);
}

.vm-detail {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.section-title {
    margin-top: 8px;
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
}

@media (max-width: 1200px) {
    .vm-card-grid {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .vm-host-resource-grid {
        grid-template-columns: 1fr;
    }
}

@media (max-width: 768px) {
    .vm-card-grid,
    .vm-card-resource {
        grid-template-columns: 1fr;
    }

    .vm-card-header {
        flex-direction: column;
    }

    .vm-card-status-actions {
        justify-content: flex-start;
        max-width: 100%;
    }

    .vm-card-actions,
    .vm-card-secondary-actions {
        width: 100%;
    }

    .vm-console-button {
        flex: 1;
    }

    .vm-card-pagination {
        justify-content: flex-start;
        overflow-x: auto;
    }
}
</style>
