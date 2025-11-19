<template>
    <div v-loading="loading">
        <docker-status
            v-model:isActive="isActive"
            v-model:isExist="isExist"
            v-model:loading="loading"
            @search="search"
        />

        <LayoutContent v-if="isExist" :title="$t('container.compose', 2)" :class="{ mask: !isActive }">
            <template #leftToolBar>
                <el-button type="primary" @click="onOpenDialog()">
                    {{ $t('container.createCompose') }}
                </el-button>
            </template>
            <template #main>
                <div class="flex gap-4 h-full">
                    <div class="w-45 flex-shrink-0 flex flex-col border-r pr-4">
                        <el-input
                            v-model="searchName"
                            :placeholder="$t('commons.button.search')"
                            clearable
                            class="mb-4"
                            @clear="search"
                            @keyup.enter="search"
                        >
                            <template #prefix>
                                <el-icon><Search /></el-icon>
                            </template>
                        </el-input>

                        <div class="flex-1 overflow-auto">
                            <div
                                v-for="item in data"
                                :key="item.name"
                                class="p-3 mb-2 rounded cursor-pointer border transition-colors duration-200"
                                :class="[
                                    selectedCompose?.name === item.name
                                        ? 'bg-blue-50 border-blue-500 dark:bg-blue-500/30 dark:border-blue-500/80'
                                        : 'border-transparent hover:bg-slate-50 dark:hover:bg-white/[0.08]',
                                ]"
                                @click="loadDetail(item)"
                            >
                                <div class="min-w-0">
                                    <div class="font-medium truncate">{{ item.name }}</div>
                                    <div class="text-xs text-gray-500 mt-1">
                                        <el-text v-if="item.containerCount === 0" type="danger" size="small">
                                            {{ $t('container.exited') }}
                                        </el-text>
                                        <el-text
                                            v-else
                                            :type="item.containerCount === item.runningCount ? 'success' : 'warning'"
                                            size="small"
                                        >
                                            {{ $t('container.running', [item.runningCount, item.containerCount]) }}
                                        </el-text>
                                    </div>
                                </div>
                            </div>
                            <el-empty v-if="data.length === 0" :description="$t('commons.msg.noneData')" />
                        </div>
                    </div>

                    <div v-if="selectedCompose" class="flex-1 min-w-0">
                        <div v-loading="detailLoading" class="h-full flex flex-col gap-4">
                            <el-card shadow="never">
                                <template #header>
                                    <div class="flex items-center justify-between">
                                        <span class="font-medium">
                                            {{ composeInfo?.name }} - {{ $t('container.compose') }}
                                        </span>
                                        <span class="text-sm text-gray-500 truncate max-w-[50%]">
                                            <div class="text-sm">
                                                <span class="text-gray-500">{{ $t('app.source') }}:</span>
                                                <el-text v-if="composeInfo?.createdBy === ''" size="small">
                                                    {{ $t('commons.table.local') }}
                                                </el-text>
                                                <el-text
                                                    v-else-if="composeInfo?.createdBy === 'Apps'"
                                                    type="success"
                                                    size="small"
                                                >
                                                    {{ $t('menu.apps') }}
                                                </el-text>
                                                <el-text
                                                    v-else-if="composeInfo?.createdBy === '1Panel'"
                                                    type="primary"
                                                    size="small"
                                                >
                                                    1Panel
                                                </el-text>
                                                <el-divider direction="vertical" />
                                                {{ $t('commons.table.createdAt') }}: {{ composeInfo?.createdAt }}
                                            </div>
                                        </span>
                                    </div>
                                </template>
                                <div class="flex items-center gap-3">
                                    <el-button-group>
                                        <el-button
                                            type="success"
                                            :loading="operateLoading && currentOperation === 'up'"
                                            :disabled="disableOperate"
                                            @click="handleComposeOperate('up')"
                                        >
                                            <el-icon class="mr-1"><VideoPlay /></el-icon>
                                            {{ $t('commons.operate.start') }}
                                        </el-button>
                                        <el-button
                                            type="warning"
                                            :loading="operateLoading && currentOperation === 'restart'"
                                            :disabled="disableOperate"
                                            @click="handleComposeOperate('restart')"
                                        >
                                            <el-icon class="mr-1"><RefreshRight /></el-icon>
                                            {{ $t('commons.operate.restart') }}
                                        </el-button>
                                        <el-button
                                            type="danger"
                                            :loading="operateLoading && currentOperation === 'stop'"
                                            :disabled="disableOperate"
                                            @click="handleComposeOperate('stop')"
                                        >
                                            <el-icon class="mr-1"><VideoPause /></el-icon>
                                            {{ $t('commons.operate.stop') }}
                                        </el-button>
                                    </el-button-group>
                                    <el-button :loading="detailLoading" @click="refreshDetail">
                                        <el-icon class="mr-1"><Refresh /></el-icon>
                                        {{ $t('commons.button.refresh') }}
                                    </el-button>
                                    <el-button :disabled="!composeInfo?.workdir" @click="openComposeFolder">
                                        <el-icon class="mr-1"><Folder /></el-icon>
                                        {{ $t('container.composeDirectory') }}
                                    </el-button>
                                    <el-button type="danger" plain @click="onDeleteCompose">
                                        <el-icon class="mr-1"><Delete /></el-icon>
                                        {{ $t('commons.button.delete') }}
                                    </el-button>
                                </div>
                            </el-card>

                            <el-card v-if="composeInfo && composeContainers.length > 0" shadow="never">
                                <template #header>
                                    <span class="text-sm font-medium">
                                        {{ $t('container.containerStatus') }}
                                        ({{ composeInfo?.runningCount || 0 }}/{{ composeInfo?.containerCount || 0 }})
                                    </span>
                                </template>
                                <el-table :data="tableData" size="small" max-height="200">
                                    <el-table-column
                                        :label="$t('commons.table.name')"
                                        prop="name"
                                        show-overflow-tooltip
                                    >
                                        <template #default="{ row }">
                                            <el-text
                                                type="primary"
                                                class="cursor-pointer"
                                                @click="onInspectContainer(row)"
                                            >
                                                {{ row.name }}
                                            </el-text>
                                        </template>
                                    </el-table-column>
                                    <el-table-column :label="$t('commons.table.status')" prop="state" width="100">
                                        <template #default="{ row }">
                                            <Status :key="row.state" :status="row.state"></Status>
                                        </template>
                                    </el-table-column>
                                    <el-table-column
                                        :label="$t('container.source')"
                                        show-overflow-tooltip
                                        prop="resource"
                                    >
                                        <template #default="{ row }">
                                            <div v-if="row.hasLoad" class="flex items-center">
                                                <div class="text-xs">
                                                    <div>CPU: {{ row.cpuPercent.toFixed(2) }}%</div>
                                                    <div>
                                                        {{ $t('monitor.memory') }}: {{ row.memoryPercent.toFixed(2) }}%
                                                    </div>
                                                </div>
                                                <el-popover placement="right" width="500px" trigger="hover">
                                                    <template #reference>
                                                        <el-icon
                                                            class="cursor-pointer text-gray-500 hover:text-primary ml-1"
                                                            :size="16"
                                                        >
                                                            <Histogram />
                                                        </el-icon>
                                                    </template>
                                                    <template #default>
                                                        <el-descriptions
                                                            direction="vertical"
                                                            border
                                                            :column="3"
                                                            size="small"
                                                        >
                                                            <el-descriptions-item :label="$t('container.cpuUsage')">
                                                                {{ computeCPU(row.cpuTotalUsage) }}
                                                            </el-descriptions-item>
                                                            <el-descriptions-item :label="$t('container.cpuTotal')">
                                                                {{ computeCPU(row.systemUsage) }}
                                                            </el-descriptions-item>
                                                            <el-descriptions-item :label="$t('container.core')">
                                                                {{ row.percpuUsage }}
                                                            </el-descriptions-item>
                                                            <el-descriptions-item :label="$t('container.memUsage')">
                                                                {{ computeSizeForDocker(row.memoryUsage) }}
                                                            </el-descriptions-item>
                                                            <el-descriptions-item :label="$t('container.memCache')">
                                                                {{ computeSizeForDocker(row.memoryCache) }}
                                                            </el-descriptions-item>
                                                            <el-descriptions-item :label="$t('container.memTotal')">
                                                                {{ computeSizeForDocker(row.memoryLimit) }}
                                                            </el-descriptions-item>
                                                            <el-descriptions-item>
                                                                <template #label>
                                                                    {{ $t('container.sizeRw') }}
                                                                    <el-tooltip :content="$t('container.sizeRwHelper')">
                                                                        <el-icon class="icon-item">
                                                                            <InfoFilled />
                                                                        </el-icon>
                                                                    </el-tooltip>
                                                                </template>
                                                                {{ computeSize2(row.sizeRw) }}
                                                            </el-descriptions-item>
                                                            <el-descriptions-item :label="$t('container.sizeRootFs')">
                                                                <template #label>
                                                                    {{ $t('container.sizeRootFs') }}
                                                                    <el-tooltip
                                                                        :content="$t('container.sizeRootFsHelper')"
                                                                    >
                                                                        <el-icon class="icon-item">
                                                                            <InfoFilled />
                                                                        </el-icon>
                                                                    </el-tooltip>
                                                                </template>
                                                                {{ computeSize2(row.sizeRootFs) }}
                                                            </el-descriptions-item>
                                                        </el-descriptions>
                                                    </template>
                                                </el-popover>
                                            </div>
                                            <div v-if="!row.hasLoad">
                                                <el-button link loading></el-button>
                                            </div>
                                        </template>
                                    </el-table-column>
                                    <el-table-column :label="$t('commons.table.operate')" width="180" fixed="right">
                                        <template #default="{ row }">
                                            <el-button type="primary" link size="small" @click="onOpenTerminal(row)">
                                                {{ $t('menu.terminal') }}
                                            </el-button>
                                            <el-divider direction="vertical" />
                                            <el-button type="primary" link size="small" @click="onOpenLog(row)">
                                                {{ $t('commons.button.log') }}
                                            </el-button>
                                        </template>
                                    </el-table-column>
                                </el-table>
                            </el-card>

                            <div class="flex-1 flex gap-4 min-h-0">
                                <div class="flex flex-col gap-4 min-h-0 min-w-0 flex-[1]">
                                    <el-card shadow="never" class="flex flex-col">
                                        <template #header>
                                            <div class="flex items-center justify-between">
                                                <span class="font-medium">{{ $t('container.compose') }}</span>
                                                <el-button
                                                    type="primary"
                                                    size="small"
                                                    :disabled="disableEdit"
                                                    :loading="saving"
                                                    @click="onSubmitEdit"
                                                >
                                                    {{ $t('commons.button.save') }}
                                                </el-button>
                                            </div>
                                        </template>
                                        <div class="flex-1 overflow-hidden">
                                            <CodemirrorPro
                                                v-model="composeContent"
                                                mode="yaml"
                                                :disabled="disableEdit"
                                                :heightDiff="100"
                                                placeholder="#Define or paste the content of your docker-compose file here"
                                            />
                                        </div>
                                    </el-card>

                                    <el-card v-if="showEnvSetting" shadow="never">
                                        <template #header>
                                            <div class="flex items-center justify-between">
                                                <span class="font-medium">.env</span>
                                                <el-button
                                                    type="primary"
                                                    size="small"
                                                    :disabled="disableEdit"
                                                    :loading="saving"
                                                    @click="onSubmitEdit"
                                                >
                                                    {{ $t('commons.button.save') }}
                                                </el-button>
                                            </div>
                                        </template>
                                        <el-form-item :label="$t('container.env')">
                                            <el-input
                                                v-model="envStr"
                                                type="textarea"
                                                :rows="3"
                                                :placeholder="$t('container.tagHelper')"
                                            />
                                        </el-form-item>
                                    </el-card>
                                </div>

                                <el-card shadow="never" class="flex flex-col min-h-0 min-w-0 flex-[2]">
                                    <template #header>
                                        <div class="flex items-center justify-between">
                                            <span class="font-medium">{{ $t('commons.button.log') }}</span>
                                            <el-button type="primary" size="small" @click="openComposeLogDrawer">
                                                {{ $t('commons.button.view') }}
                                            </el-button>
                                        </div>
                                    </template>
                                    <div class="flex-1 overflow-auto">
                                        <ContainerLog
                                            v-if="composePath && shouldLoadLog"
                                            :key="logKey"
                                            :compose="composePath"
                                            :resource="composeName"
                                            :highlightDiff="200"
                                            :showControl="false"
                                            :defaultFollow="true"
                                        />
                                        <el-empty v-else :description="$t('commons.msg.noneData')" />
                                    </div>
                                </el-card>
                            </div>
                        </div>
                    </div>
                </div>
            </template>
        </LayoutContent>

        <CreateDialog @search="search" ref="dialogRef" />
        <DeleteDialog @search="search" ref="dialogDelRef" />
        <ContainerInspectDialog ref="containerInspectRef" />
        <TerminalDialog ref="terminalDialogRef" />
        <ContainerLogDialog ref="containerLogDialogRef" :highlightDiff="210" />
        <ComposeLogs ref="composeLogRef" />
    </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import ContainerLog from '@/components/log/container/index.vue';
import ContainerInspectDialog from '@/views/container/container/inspect/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import ContainerLogDialog from '@/components/log/container-drawer/index.vue';
import CreateDialog from '@/views/container/compose/create/index.vue';
import DeleteDialog from '@/views/container/compose/delete/index.vue';
import ComposeLogs from '@/components/log/compose/index.vue';
import { composeOperator, composeUpdate, containerListStats, inspect, searchCompose } from '@/api/modules/container';
import DockerStatus from '@/views/container/docker-status/index.vue';
import i18n from '@/lang';
import { Container } from '@/api/interface/container';
import { routerToFileWithPath } from '@/utils/router';
import { MsgError, MsgSuccess } from '@/utils/message';
import { computeCPU, computeSize2, computeSizeForDocker } from '@/utils/util';
import {
    Delete,
    Folder,
    Histogram,
    Refresh,
    RefreshRight,
    Search,
    VideoPause,
    VideoPlay,
} from '@element-plus/icons-vue';

const data = ref<any[]>([]);
const loading = ref(false);
const selectedCompose = ref<Container.ComposeInfo | null>(null);
const detailLoading = ref(false);
const operateLoading = ref(false);
const currentOperation = ref('');
const saving = ref(false);
const composeName = ref('');
const composeContent = ref('');
const envStr = ref('');
const composeInfo = ref<Container.ComposeInfo>();
const containerStats = ref<any[]>([]);
const logKey = ref(0);
const shouldLoadLog = ref(false);
const containerInspectRef = ref();
const terminalDialogRef = ref();
const containerLogDialogRef = ref();
const composeLogRef = ref();
const searchName = ref('');

const isActive = ref(false);
const isExist = ref(false);

const composePath = computed(() => composeInfo.value?.path || selectedCompose.value?.path || '');
const composeContainers = computed(() => composeInfo.value?.containers || []);
const disableEdit = computed(() => composeInfo.value?.createdBy === 'Local');
const showEnvSetting = computed(() => composeInfo.value?.createdBy === '1Panel');
const disableOperate = computed(
    () => !composeInfo.value || !composePath.value || !isActive.value || !isExist.value || operateLoading.value,
);

const tableData = computed(() => {
    return composeContainers.value.map((container) => {
        const stats = containerStats.value.find((s) => s.containerID === container.containerID);
        return {
            ...container,
            hasLoad: !!stats,
            cpuPercent: stats?.cpuPercent || 0,
            memoryPercent: stats?.memoryPercent || 0,
            cpuTotalUsage: stats?.cpuTotalUsage || 0,
            systemUsage: stats?.systemUsage || 0,
            percpuUsage: stats?.percpuUsage || 0,
            memoryCache: stats?.memoryCache || 0,
            memoryUsage: stats?.memoryUsage || 0,
            memoryLimit: stats?.memoryLimit || 0,
            sizeRw: stats?.sizeRw || 0,
            sizeRootFs: stats?.sizeRootFs || 0,
        };
    });
});

const closeDetail = () => {
    selectedCompose.value = null;
    composeName.value = '';
    composeInfo.value = undefined;
    composeContent.value = '';
    envStr.value = '';
    shouldLoadLog.value = false;
};

const openComposeFolder = () => {
    if (composeInfo.value?.workdir) {
        routerToFileWithPath(composeInfo.value.workdir);
    }
};

const search = async () => {
    if (!isActive.value || !isExist.value) {
        return;
    }
    let params = {
        info: searchName.value,
        page: 1,
        pageSize: 100,
    };
    loading.value = true;
    await searchCompose(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
        })
        .finally(() => {
            loading.value = false;
        });
};

const loadDetail = async (row: Container.ComposeInfo) => {
    if (selectedCompose.value?.name === row.name) {
        closeDetail();
        return;
    }
    selectedCompose.value = row;
    composeName.value = row.name;
    detailLoading.value = true;
    shouldLoadLog.value = false;
    try {
        await loadComposeContent();
        await new Promise((resolve) => setTimeout(resolve, 100));
        await loadComposeInfo();
        detailLoading.value = false;
        await new Promise((resolve) => setTimeout(resolve, 100));
        shouldLoadLog.value = true;
        logKey.value++;
        await new Promise((resolve) => setTimeout(resolve, 100));
        await loadContainerStats();
    } catch (error) {
        detailLoading.value = false;
        throw error;
    }
};

const loadComposeInfo = async () => {
    const params = {
        info: composeName.value,
        page: 1,
        pageSize: 1,
    };
    const res = await searchCompose(params);
    const items = res.data?.items || [];
    const target = items.find((item) => item.name === composeName.value) || items[0];
    if (!target) {
        composeInfo.value = undefined;
        envStr.value = '';
        MsgError(i18n.global.t('commons.msg.noneData'));
        return;
    }
    composeInfo.value = target;
    envStr.value = (target.env || []).join('\n');
};

const loadComposeContent = async () => {
    const res = await inspect({ id: composeName.value, type: 'compose' });
    composeContent.value = res.data;
};

const loadContainerStats = async () => {
    try {
        const res = await containerListStats();
        containerStats.value = res.data || [];
    } catch (error) {
        containerStats.value = [];
    }
};

const refreshDetail = async () => {
    if (!composeName.value || !isActive.value || !isExist.value) {
        return;
    }
    detailLoading.value = true;
    try {
        await loadComposeInfo();
        detailLoading.value = false;
        await new Promise((resolve) => setTimeout(resolve, 300));
        await loadContainerStats();
    } catch (error) {
        detailLoading.value = false;
        throw error;
    }
};

const dialogRef = ref();
const onOpenDialog = async () => {
    dialogRef.value!.acceptParams();
};

const onDeleteCompose = () => {
    if (!selectedCompose.value) return;
    dialogDelRef.value.acceptParams({
        name: selectedCompose.value.name,
        path: selectedCompose.value.path,
    });
};

const dialogDelRef = ref();

const handleComposeOperate = async (operation: 'up' | 'stop' | 'restart') => {
    if (!composeInfo.value || !composePath.value) {
        return;
    }
    const mes = i18n.global.t('container.composeOperatorHelper', [
        composeInfo.value.name,
        i18n.global.t('commons.operate.' + operation),
    ]);
    ElMessageBox.confirm(mes, i18n.global.t('commons.operate.' + operation), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        currentOperation.value = operation;
        operateLoading.value = true;
        const params = {
            name: composeInfo.value!.name,
            path: composePath.value,
            operation: operation,
            withFile: false,
            force: false,
        };
        await composeOperator(params)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                refreshDetail();
                search();
            })
            .finally(() => {
                operateLoading.value = false;
                currentOperation.value = '';
            });
    });
};

const onSubmitEdit = async () => {
    if (!composeInfo.value || !composePath.value || disableEdit.value) {
        return;
    }
    const param = {
        name: composeName.value,
        path: composePath.value,
        content: composeContent.value,
        createdBy: composeInfo.value.createdBy,
        env: envStr.value ? envStr.value.split('\n') : [],
    };
    saving.value = true;
    await composeUpdate(param)
        .then(async () => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            await loadComposeContent();
            refreshDetail();
        })
        .finally(() => {
            saving.value = false;
        });
};

const onInspectContainer = async (item: any) => {
    if (!item.containerID) {
        return;
    }
    const res = await inspect({ id: item.containerID, type: 'container' });
    containerInspectRef.value!.acceptParams({ data: res.data, ports: item.ports || [] });
};

const onOpenTerminal = (row: any) => {
    if (!row.containerID) {
        return;
    }
    const title = i18n.global.t('menu.container') + ' ' + row.name;
    terminalDialogRef.value?.acceptParams({ containerID: row.containerID, title });
};

const onOpenLog = (row: any) => {
    containerLogDialogRef.value?.acceptParams({ container: row.name });
};

const openComposeLogDrawer = () => {
    if (!composePath.value || !composeName.value) return;
    composeLogRef.value?.acceptParams({
        compose: composePath.value,
        resource: composeName.value,
        container: '',
    });
};
</script>
