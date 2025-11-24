<template>
    <div v-loading="loading">
        <docker-status
            v-model:isActive="isActive"
            v-model:isExist="isExist"
            v-model:loading="loading"
            @search="search(true)"
        />

        <LayoutContent v-if="isExist" :title="$t('container.compose', 2)" :class="{ mask: !isActive }">
            <template #leftToolBar>
                <el-button type="primary" @click="onOpenDialog()">
                    {{ $t('container.createCompose') }}
                </el-button>
            </template>
            <template #main>
                <el-row v-if="data.length > 0" :gutter="20" class="row-box">
                    <el-col :span="6">
                        <el-card>
                            <el-input
                                v-model="searchName"
                                :placeholder="$t('commons.button.search')"
                                clearable
                                @clear="search"
                                @keyup.enter="search"
                            >
                                <template #prefix>
                                    <el-icon><Search /></el-icon>
                                </template>
                            </el-input>

                            <ComplexTable :show-header="false" @row-click="loadDetail" :data="data">
                                <el-table-column prop="name">
                                    <template #default="{ row }">
                                        <div class="cursor-pointer">
                                            <div class="font-medium truncate">
                                                {{ row.name }}
                                                <el-divider direction="vertical" />
                                                <el-text v-if="row.containerCount === 0" type="danger" size="small">
                                                    {{ $t('container.exited') }}
                                                </el-text>
                                                <el-text
                                                    v-else
                                                    :type="
                                                        row.containerCount === row.runningCount ? 'success' : 'warning'
                                                    "
                                                    size="small"
                                                >
                                                    {{
                                                        $t('container.running', [row.runningCount, row.containerCount])
                                                    }}
                                                </el-text>
                                                <el-divider direction="vertical" />
                                                <el-button
                                                    link
                                                    type="primary"
                                                    icon="Folder"
                                                    :disabled="!currentCompose?.workdir"
                                                    @click="openComposeFolder"
                                                />
                                            </div>
                                            <div class="mt-1 mb-2">
                                                <el-tag size="small" type="info">{{ loadFrom(row) }}</el-tag>
                                                <el-tag size="small" type="info" class="ml-2">
                                                    {{ row.createdAt }}
                                                </el-tag>
                                            </div>
                                            <el-button
                                                plain
                                                round
                                                size="small"
                                                @click="handleComposeOperate('up', row)"
                                            >
                                                {{ $t('commons.operate.start') }}
                                            </el-button>
                                            <el-button
                                                plain
                                                round
                                                size="small"
                                                @click="handleComposeOperate('stop', row)"
                                            >
                                                {{ $t('commons.operate.stop') }}
                                            </el-button>
                                            <el-button
                                                plain
                                                round
                                                size="small"
                                                @click="handleComposeOperate('restart', row)"
                                            >
                                                {{ $t('commons.operate.restart') }}
                                            </el-button>
                                            <el-button
                                                :disabled="row.createdBy !== '1Panel'"
                                                plain
                                                round
                                                size="small"
                                                @click="onDelete(row)"
                                            >
                                                {{ $t('commons.operate.delete') }}
                                            </el-button>
                                        </div>
                                    </template>
                                </el-table-column>
                            </ComplexTable>
                        </el-card>
                    </el-col>
                    <el-col :span="18">
                        <el-card v-if="currentCompose" v-loading="detailLoading">
                            <el-table
                                v-if="composeContainers.length > 0"
                                :data="tableData"
                                size="small"
                                max-height="250"
                            >
                                <el-table-column :label="$t('commons.table.name')" prop="name" show-overflow-tooltip>
                                    <template #default="{ row }">
                                        <el-text type="primary" class="cursor-pointer" @click="onInspectContainer(row)">
                                            {{ row.name }}
                                        </el-text>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('commons.table.status')" prop="state" width="150">
                                    <template #default="{ row }">
                                        <Status :key="row.state" :status="row.state"></Status>
                                    </template>
                                </el-table-column>
                                <el-table-column
                                    :label="$t('container.source')"
                                    show-overflow-tooltip
                                    prop="resource"
                                    min-width="150"
                                >
                                    <template #default="{ row }">
                                        <div v-if="row.hasLoad">
                                            <div class="source-font">CPU: {{ row.cpuPercent.toFixed(2) }}%</div>
                                            <div class="float-left source-font">
                                                {{ $t('monitor.memory') }}: {{ row.memoryPercent.toFixed(2) }}%
                                            </div>
                                            <el-popover placement="right" width="500px" class="float-right">
                                                <template #reference>
                                                    <svg-icon iconName="p-xiangqing" class="svg-icon"></svg-icon>
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
                                <el-table-column :label="$t('commons.table.operate')" width="180" fixed="right">
                                    <template #default="{ row }">
                                        <el-button type="primary" link @click="onOpenTerminal(row)">
                                            {{ $t('menu.terminal') }}
                                        </el-button>
                                        <el-button type="primary" link @click="onOpenLog(row)">
                                            {{ $t('commons.button.log') }}
                                        </el-button>
                                    </template>
                                </el-table-column>
                            </el-table>

                            <el-radio-group size="small" class="mt-1 mb-1" v-model="showType">
                                <el-radio-button value="compose">{{ $t('container.compose') }}</el-radio-button>
                                <el-radio-button value="log">{{ $t('commons.button.log') }}</el-radio-button>
                            </el-radio-group>
                            <el-form ref="formRef" v-show="showType === 'compose'" @submit.prevent label-position="top">
                                <CodemirrorPro
                                    v-model="composeContent"
                                    mode="yaml"
                                    :heightDiff="475"
                                    placeholder="#Define or paste the content of your docker-compose file here"
                                />
                                <div v-if="currentCompose.createdBy === '1Panel'">
                                    <el-form-item :label="$t('container.env')" prop="environmentStr">
                                        <el-input
                                            type="textarea"
                                            :placeholder="$t('container.tagHelper')"
                                            :rows="3"
                                            v-model="envStr"
                                        />
                                    </el-form-item>
                                    <span class="input-help whitespace-break-spaces">
                                        {{ $t('container.editComposeHelper') }}
                                    </span>
                                    <CodemirrorPro
                                        v-model="env"
                                        :height="45"
                                        :minHeight="45"
                                        disabled
                                        mode="yaml"
                                    ></CodemirrorPro>
                                </div>
                                <el-button type="primary" class="float-right" @click="onSubmitEdit">
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </el-form>

                            <div v-show="showType === 'log'">
                                <ContainerLog
                                    :key="currentCompose.path"
                                    :compose="currentCompose.path"
                                    :highlightDiff="450"
                                    :defaultFollow="true"
                                />
                            </div>
                        </el-card>
                    </el-col>
                </el-row>
                <el-empty v-else :description="$t('commons.msg.noneData')" />
            </template>
        </LayoutContent>

        <CreateDialog @search="search" ref="dialogCreateRef" />
        <DeleteDialog @search="search(true)" ref="dialogDelRef" />
        <ContainerInspectDialog ref="containerInspectRef" />
        <TerminalDialog ref="terminalDialogRef" />
        <ContainerLogDialog ref="containerLogDialogRef" :highlightDiff="210" />
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
import { composeOperator, composeUpdate, containerListStats, inspect, searchCompose } from '@/api/modules/container';
import DockerStatus from '@/views/container/docker-status/index.vue';
import i18n from '@/lang';
import { Container } from '@/api/interface/container';
import { routerToFileWithPath } from '@/utils/router';
import { MsgSuccess } from '@/utils/message';
import { computeCPU, computeSize2, computeSizeForDocker } from '@/utils/util';
import { Search } from '@element-plus/icons-vue';

const data = ref<any[]>([]);
const loading = ref(false);
const detailLoading = ref(false);
const currentCompose = ref<Container.ComposeInfo | null>(null);
const composeContainers = ref([]);
const composeContent = ref('');
const envStr = ref('');
const env = ref('env_file:\n  - 1panel.env');

const dialogCreateRef = ref();
const dialogDelRef = ref();
const containerInspectRef = ref();
const terminalDialogRef = ref();
const containerLogDialogRef = ref();

const searchName = ref('');
const showType = ref('compose');
const containerStats = ref<any[]>([]);

const isActive = ref(false);
const isExist = ref(false);

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

const loadFrom = (row: any) => {
    switch (row.createdBy) {
        case '1Panel':
            return '1Panel';
        case 'Apps':
            return i18n.global.t('menu.apps');
        default:
            return i18n.global.t('commons.table.local');
    }
};

const search = async (withRefreshDetail?: boolean) => {
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
            if (data.value.length > 0 && withRefreshDetail) {
                loadDetail(data.value[0], true);
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

const loadDetail = async (row: Container.ComposeInfo, withRefresh: boolean) => {
    if (currentCompose.value?.name === row.name && withRefresh !== true) {
        return;
    }
    detailLoading.value = true;
    currentCompose.value = row;
    composeContainers.value = row.containers || [];
    await inspect({ id: currentCompose.value.name, type: 'compose' }).then((res) => {
        composeContent.value = res.data;
        detailLoading.value = false;
    });
    loadContainerStats();
};

const loadContainerStats = async () => {
    try {
        const res = await containerListStats();
        containerStats.value = res.data || [];
    } catch (error) {
        containerStats.value = [];
    }
};

const onOpenDialog = async () => {
    dialogCreateRef.value!.acceptParams();
};

const onDelete = (row: any) => {
    dialogDelRef.value.acceptParams({
        name: row.name,
        path: row.path,
    });
};

const handleComposeOperate = async (operation: 'up' | 'stop' | 'restart', row: any) => {
    const mes = i18n.global.t('container.composeOperatorHelper', [
        row.name,
        i18n.global.t('commons.operate.' + operation),
    ]);
    ElMessageBox.confirm(mes, i18n.global.t('commons.operate.' + operation), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        const params = {
            name: row.name,
            path: currentCompose.value.path,
            operation: operation,
            withFile: false,
            force: false,
        };
        await composeOperator(params)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
                if (row.name === currentCompose.value?.name) {
                    loadDetail(currentCompose.value, true);
                }
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const onSubmitEdit = async () => {
    const param = {
        name: currentCompose.value.name,
        path: currentCompose.value.path,
        content: composeContent.value,
        createdBy: currentCompose.value.createdBy,
        env: envStr.value ? envStr.value.split('\n') : [],
    };
    loading.value = true;
    await composeUpdate(param)
        .then(async () => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            await loadDetail(currentCompose.value, true);
        })
        .finally(() => {
            loading.value = false;
        });
};

const openComposeFolder = () => {
    if (currentCompose.value?.workdir) {
        routerToFileWithPath(currentCompose.value.workdir);
    }
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
</script>

<style scoped lang="scss">
.svg-icon {
    margin-top: -3px;
    font-size: 6px;
    cursor: pointer;
}
</style>
