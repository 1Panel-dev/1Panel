<template>
    <div v-loading="pageLoading">
        <el-card v-if="isExist && !isActive" class="mask-prompt">
            <span>{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link class="bt" @click="goSetting">【 {{ $t('container.setting') }} 】</el-button>
            <span>{{ $t('container.startIn') }}</span>
        </el-card>
        <NoSuchService v-if="!isExist" name="Docker" />
        <LayoutContent v-if="isExist" backName="Compose" :title="composeTitle" :class="{ mask: !isActive }">
            <template #main>
                <el-empty v-if="!composeName" :description="$t('commons.msg.noneData')" />
                <div v-else class="w-full">
                    <div class="app-status mb-4">
                        <el-card>
                            <div class="flex w-full flex-col gap-4 md:flex-row">
                                <div class="flex items-center gap-1">
                                    <el-button
                                        type="success"
                                        plain
                                        :loading="operateLoading && currentOperation === 'up'"
                                        :disabled="disableOperate"
                                        @click="handleComposeOperate('up')"
                                    >
                                        {{ $t('commons.operate.start') }}
                                    </el-button>
                                    <el-divider direction="vertical" />
                                    <el-button
                                        type="danger"
                                        plain
                                        :loading="operateLoading && currentOperation === 'stop'"
                                        :disabled="disableOperate"
                                        @click="handleComposeOperate('stop')"
                                    >
                                        {{ $t('commons.operate.stop') }}
                                    </el-button>
                                    <el-divider direction="vertical" />
                                    <el-button
                                        type="warning"
                                        plain
                                        :loading="operateLoading && currentOperation === 'restart'"
                                        :disabled="disableOperate"
                                        @click="handleComposeOperate('restart')"
                                    >
                                        {{ $t('commons.operate.restart') }}
                                    </el-button>
                                    <el-divider direction="vertical" />
                                    <el-button type="primary" link :loading="detailLoading" @click="refreshDetail">
                                        {{ $t('commons.button.refresh') }}
                                    </el-button>
                                    <el-divider direction="vertical" />
                                    <el-button
                                        type="primary"
                                        link
                                        :disabled="!composeInfo?.workdir"
                                        @click="openComposeFolder"
                                    >
                                        {{ $t('container.composeDirectory') }}
                                    </el-button>
                                </div>
                            </div>
                            <el-alert
                                v-if="showOperateHelper"
                                type="warning"
                                :closable="false"
                                :title="$t('container.composeDetailHelper')"
                                class="mt-3"
                            />
                        </el-card>
                    </div>

                    <el-card v-if="composeInfo && composeContainers.length > 0" class="mb-4" shadow="never">
                        <template #header>
                            <div class="flex flex-wrap items-center justify-between gap-3">
                                <span class="text-sm font-medium">
                                    {{ $t('container.containerStatus') }}
                                    ( {{ composeInfo?.containerCount || 0 }} / {{ composeInfo?.runningCount || 0 }} )
                                </span>
                            </div>
                        </template>
                        <ComplexTable :data="tableData" :heightDiff="400">
                            <el-table-column
                                :label="$t('commons.table.name')"
                                min-width="150"
                                prop="name"
                                show-overflow-tooltip
                            >
                                <template #default="{ row }">
                                    <el-text type="primary" class="cursor-pointer" @click="onInspectContainer(row)">
                                        {{ row.name }}
                                    </el-text>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('commons.table.status')" min-width="100" prop="state">
                                <template #default="{ row }">
                                    <Status :key="row.state" :status="row.state"></Status>
                                </template>
                            </el-table-column>
                            <el-table-column
                                :label="$t('container.source')"
                                show-overflow-tooltip
                                prop="resource"
                                min-width="120"
                            >
                                <template #default="{ row }">
                                    <div v-if="row.hasLoad" class="flex items-center">
                                        <div class="text-xs">
                                            <div>CPU: {{ row.cpuPercent.toFixed(2) }}%</div>
                                            <div class="text-xs">
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
                                                <el-descriptions direction="vertical" border :column="3" size="small">
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
                                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
                                                            </el-tooltip>
                                                        </template>
                                                        {{ computeSize2(row.sizeRw) }}
                                                    </el-descriptions-item>
                                                    <el-descriptions-item :label="$t('container.sizeRootFs')">
                                                        <template #label>
                                                            {{ $t('container.sizeRootFs') }}
                                                            <el-tooltip :content="$t('container.sizeRootFsHelper')">
                                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
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
                            <el-table-column :label="$t('commons.table.operate')" width="200px" fixed="right">
                                <template #default="{ row }">
                                    <el-button type="primary" link @click="onOpenTerminal(row)">
                                        {{ $t('menu.terminal') }}
                                    </el-button>
                                    <el-divider direction="vertical" />
                                    <el-button type="primary" link @click="onOpenLog(row)">
                                        {{ $t('commons.button.log') }}
                                    </el-button>
                                </template>
                            </el-table-column>
                        </ComplexTable>
                    </el-card>

                    <div class="grid min-h-[600px] grid-cols-1 gap-4 xl:grid-cols-[minmax(0,5fr)_minmax(0,7fr)]">
                        <el-card class="h-full flex flex-col compose-detail-editor" shadow="never">
                            <template #header>
                                <div class="font-medium">
                                    {{ $t('container.composeTemplate') }}
                                </div>
                            </template>
                            <div class="flex-1">
                                <el-form label-position="top" class="flex h-full flex-col" @submit.prevent>
                                    <el-form-item>
                                        <CodemirrorPro
                                            v-model="composeContent"
                                            mode="yaml"
                                            :disabled="disableEdit"
                                            :heightDiff="200"
                                            placeholder="#Define or paste the content of your docker-compose file here"
                                        />
                                    </el-form-item>
                                    <div v-if="showEnvSetting" class="mt-4">
                                        <el-form-item :label="$t('container.env')">
                                            <el-input
                                                v-model="envStr"
                                                type="textarea"
                                                :rows="3"
                                                :placeholder="$t('container.tagHelper')"
                                            />
                                        </el-form-item>
                                        <span class="input-help whitespace-break-spaces">
                                            {{ $t('container.editComposeHelper') }}
                                        </span>
                                        <CodemirrorPro
                                            v-model="envFileContent"
                                            :height="45"
                                            :minHeight="45"
                                            disabled
                                            mode="yaml"
                                        />
                                    </div>
                                </el-form>
                            </div>
                            <div class="mt-4 flex justify-end gap-3">
                                <el-button
                                    type="primary"
                                    :disabled="disableEdit"
                                    :loading="saving"
                                    @click="onSubmitEdit"
                                >
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </div>
                        </el-card>

                        <el-card class="h-full compose-detail-log" shadow="never">
                            <template #header>
                                <div class="flex items-center justify-between gap-3">
                                    <span>{{ $t('commons.button.log') }}</span>
                                </div>
                            </template>
                            <div class="flex flex-1 flex-col">
                                <ContainerLog
                                    v-if="composePath && shouldLoadLog"
                                    :key="logKey"
                                    :compose="composePath"
                                    :resource="composeName"
                                    :highlightDiff="logHeightDiff"
                                />
                                <el-empty v-else :description="$t('commons.msg.noneData')" />
                            </div>
                        </el-card>
                    </div>
                </div>
            </template>
        </LayoutContent>

        <ContainerInspectDialog ref="containerInspectRef" />
        <TerminalDialog ref="terminalDialogRef" />
        <ContainerLogDialog ref="containerLogDialogRef" :highlightDiff="210" />
    </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import NoSuchService from '@/components/layout-content/no-such-service.vue';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import ContainerLog from '@/components/log/container/index.vue';
import ContainerInspectDialog from '@/views/container/container/inspect/index.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import Status from '@/components/status/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import ContainerLogDialog from '@/components/log/container-drawer/index.vue';
import {
    composeOperator,
    composeUpdate,
    containerListStats,
    inspect,
    loadDockerStatus,
    searchCompose,
} from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { routerToFileWithPath, routerToName } from '@/utils/router';
import { MsgError, MsgSuccess } from '@/utils/message';
import { computeCPU, computeSize2, computeSizeForDocker } from '@/utils/util';
import i18n from '@/lang';
import { Histogram } from '@element-plus/icons-vue';

const route = useRoute();

const composeName = ref('');
const composeContent = ref('');
const envStr = ref('');
const envFileContent = ref(`env_file:\n  - 1panel.env`);
const composeInfo = ref<Container.ComposeInfo>();
const isActive = ref(false);
const isExist = ref(false);
const dockerLoading = ref(false);
const detailLoading = ref(false);
const saving = ref(false);
const operateLoading = ref(false);
const currentOperation = ref('');
const logKey = ref(0);
const logHeightDiff = 220;
const containerInspectRef = ref();
const containerStats = ref<any[]>([]);
const terminalDialogRef = ref();
const containerLogDialogRef = ref();
const shouldLoadLog = ref(false);

const pageLoading = computed(() => dockerLoading.value || detailLoading.value);
const composeTitle = computed(() => {
    if (!composeName.value) {
        return i18n.global.t('container.compose');
    }
    return `${i18n.global.t('container.compose')} · ${composeName.value}`;
});
const composePath = computed(() => composeInfo.value?.path || (route.query.path as string) || '');
const composeContainers = computed(() => composeInfo.value?.containers || []);
const disableEdit = computed(() => composeInfo.value?.createdBy === 'Local');
const showEnvSetting = computed(() => composeInfo.value?.createdBy === '1Panel');
const showOperateHelper = computed(() => !composeInfo.value?.createdBy || composeInfo.value?.createdBy === 'Local');
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

const syncRouteParams = () => {
    composeName.value = (route.query.name as string) || (route.params.name as string) || '';
};

const loadStatus = async () => {
    dockerLoading.value = true;
    await loadDockerStatus()
        .then((res) => {
            isActive.value = res.data.isActive;
            isExist.value = res.data.isExist;
            dockerLoading.value = false;
            loadInitialDetail();
        })
        .catch(() => {
            dockerLoading.value = false;
            isActive.value = false;
            isExist.value = false;
        });
};

const loadInitialDetail = async () => {
    if (!composeName.value || !isActive.value || !isExist.value) {
        return;
    }
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
            })
            .finally(() => {
                operateLoading.value = false;
                currentOperation.value = '';
            });
    });
};

const openComposeFolder = () => {
    if (composeInfo.value?.workdir) {
        routerToFileWithPath(composeInfo.value.workdir);
    }
};

const goSetting = async () => {
    routerToName('ContainerSetting');
};

const onInspectContainer = async (item: any) => {
    if (!item.containerID) {
        return;
    }
    const res = await inspect({ id: item.containerID, type: 'container' });
    containerInspectRef.value!.acceptParams({ data: res.data, ports: item.ports || [] });
};

const onOpenTerminal = (row: any) => {
    terminalDialogRef.value?.acceptParams({ container: row.name });
};

const onOpenLog = (row: any) => {
    containerLogDialogRef.value?.acceptParams({ container: row.name });
};

onMounted(() => {
    syncRouteParams();
    loadStatus();
});

watch(
    () => route.fullPath,
    () => {
        syncRouteParams();
        loadInitialDetail();
    },
);
</script>

<style scoped lang="scss">
.app-status {
    :deep(.el-card__body) {
        padding: 16px 20px;
    }
}

.compose-detail-editor {
    :deep(.el-card__body) {
        flex: 1;
        display: flex;
        flex-direction: column;
    }
}

.compose-detail-log {
    :deep(.el-card__body) {
        height: 100%;
        display: flex;
        flex-direction: column;
    }
}
</style>
