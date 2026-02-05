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
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="searchName" />
                <TableRefresh @search="search()" />
                <TableSetting title="container-refresh" @search="refresh()" />
            </template>
            <template #main>
                <el-row v-if="data.length > 0 || isOnCreate" :gutter="20" class="row-box">
                    <el-col :xs="24" :sm="24" :md="8" :lg="8" :xl="6">
                        <el-card>
                            <el-table
                                :max-height="loadTableHeight()"
                                :show-header="false"
                                @row-click="(row, column, event) => loadDetail(row, true)"
                                :data="data"
                            >
                                <el-table-column prop="name">
                                    <template #default="{ row }">
                                        <div class="cursor-pointer">
                                            <div class="font-medium text-base">
                                                {{ row.name }}
                                            </div>
                                            <div class="mb-1">
                                                <el-text class="w-12" link size="small" type="info">
                                                    {{ loadFrom(row) }}
                                                </el-text>
                                                <el-divider direction="vertical" />
                                                <el-text link size="small" type="info" class="ml-2">
                                                    {{ row.createdAt }}
                                                </el-text>
                                                <el-divider direction="vertical" />
                                                <el-text
                                                    link
                                                    v-if="row.containerCount === 0"
                                                    type="danger"
                                                    size="small"
                                                >
                                                    {{ $t('container.exited') }}
                                                </el-text>
                                                <el-text
                                                    link
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
                                            </div>
                                            <el-button
                                                plain
                                                round
                                                size="small"
                                                :disabled="!row?.workdir"
                                                @click="openComposeFolder(row)"
                                            >
                                                {{ $t('home.dir') }}
                                            </el-button>
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
                                            <el-button plain round size="small" @click="onDelete(row)">
                                                {{ $t('commons.operate.delete') }}
                                            </el-button>
                                        </div>
                                    </template>
                                </el-table-column>
                            </el-table>
                        </el-card>
                    </el-col>
                    <el-col :xs="24" :sm="24" :md="16" :lg="16" :xl="18">
                        <el-card v-if="currentCompose && !isOnCreate" v-loading="detailLoading">
                            <el-table
                                v-if="composeContainers.length > 0"
                                :data="tableData"
                                size="small"
                                max-height="250"
                            >
                                <el-table-column
                                    :label="$t('commons.table.name')"
                                    prop="name"
                                    show-overflow-tooltip
                                    fixed="left"
                                >
                                    <template #default="{ row }">
                                        <el-text type="primary" class="cursor-pointer" @click="onInspectContainer(row)">
                                            {{ row.name }}
                                        </el-text>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('commons.table.status')" prop="state">
                                    <template #default="{ row }">
                                        <Status :key="row.state" :status="row.state"></Status>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('container.source')" show-overflow-tooltip prop="resource">
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
                                                        <el-descriptions-item v-if="row.hasLoadSize">
                                                            <template #label>
                                                                {{ $t('container.sizeRw') }}
                                                                <el-tooltip :content="$t('container.sizeRwHelper')">
                                                                    <el-icon class="icon-item"><InfoFilled /></el-icon>
                                                                </el-tooltip>
                                                            </template>
                                                            {{ computeSize2(row.sizeRw) }}
                                                        </el-descriptions-item>
                                                        <el-descriptions-item
                                                            :label="$t('container.sizeRootFs')"
                                                            v-if="row.hasLoadSize"
                                                        >
                                                            <template #label>
                                                                {{ $t('container.sizeRootFs') }}
                                                                <el-tooltip :content="$t('container.sizeRootFsHelper')">
                                                                    <el-icon class="icon-item"><InfoFilled /></el-icon>
                                                                </el-tooltip>
                                                            </template>
                                                            {{ computeSize2(row.sizeRootFs) }}
                                                        </el-descriptions-item>
                                                    </el-descriptions>

                                                    <el-button
                                                        class="mt-2"
                                                        v-if="!row.hasLoadSize"
                                                        size="small"
                                                        link
                                                        type="primary"
                                                        @click="loadSize(row)"
                                                    >
                                                        {{ $t('container.loadSize') }}
                                                    </el-button>
                                                </template>
                                            </el-popover>
                                        </div>
                                        <div v-if="!row.hasLoad">
                                            <el-button link loading></el-button>
                                        </div>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('commons.table.operate')">
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

                            <el-radio-group class="mt-1 mb-1" v-model="showType">
                                <el-radio-button value="compose">{{ $t('container.compose') }}</el-radio-button>
                                <el-radio-button value="log">{{ $t('commons.button.log') }}</el-radio-button>
                            </el-radio-group>
                            <el-select
                                class="p-w-300 mt-2 ml-2"
                                v-model="currentYamlPath"
                                @change="inspectCompose(currentCompose.name, currentYamlPath)"
                                v-if="currentCompose.path.indexOf(',') !== -1"
                            >
                                <template #prefix>{{ $t('container.composeFile') }}</template>
                                <el-option
                                    v-for="item in currentCompose.path.split(',')"
                                    :key="item"
                                    :value="item"
                                    :label="item.split('/').pop()"
                                />
                            </el-select>
                            <div v-show="showType === 'compose'">
                                <CodemirrorPro
                                    v-model="composeContent"
                                    mode="yaml"
                                    :heightDiff="475"
                                    placeholder="#Define or paste the content of your docker-compose file here"
                                />
                                <span class="envTitle">{{ $t('container.env') }}</span>
                                <el-input
                                    placeholder="key=value"
                                    type="textarea"
                                    :rows="3"
                                    :disabled="currentCompose.createdBy === 'Apps'"
                                    v-model="env"
                                />
                                <span v-if="currentCompose.createdBy === 'Apps'" class="input-help">
                                    {{ $t('container.composeEnvHelper2') }}
                                </span>
                                <div class="mt-2">
                                    <el-checkbox v-model="form.forcePull" :label="$t('container.forcePull')" />
                                    <span class="input-help">{{ $t('container.forcePullHelper') }}</span>
                                </div>

                                <el-button type="primary" class="mt-2" @click="onSubmitEdit">
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </div>

                            <div v-show="showType === 'log'">
                                <ContainerLog
                                    v-model:loading="detailLoading"
                                    :key="currentCompose.path"
                                    :compose="currentCompose.path"
                                    :resource="currentCompose.name"
                                    :highlightDiff="450"
                                    :defaultFollow="true"
                                />
                            </div>
                        </el-card>
                        <el-card v-else>
                            <el-form
                                ref="formRef"
                                @submit.prevent
                                label-position="top"
                                :model="form"
                                :rules="rules"
                                v-loading="detailLoading"
                            >
                                <el-form-item :label="$t('app.source')">
                                    <el-radio-group v-model="form.from" @change="onEdit('form')">
                                        <el-radio value="edit">{{ $t('commons.button.edit') }}</el-radio>
                                        <el-radio value="path">{{ $t('container.pathSelect') }}</el-radio>
                                        <el-radio value="template">{{ $t('container.composeTemplate') }}</el-radio>
                                    </el-radio-group>
                                </el-form-item>
                                <el-form-item v-if="form.from === 'path'" prop="path">
                                    <el-input
                                        @change="onEdit('')"
                                        :placeholder="$t('commons.example') + '/tmp/docker-compose.yml'"
                                        v-model="form.path"
                                    >
                                        <template #prepend>
                                            <el-button icon="Folder" @click="fileRef.acceptParams({ dir: false })" />
                                        </template>
                                    </el-input>
                                </el-form-item>
                                <el-form-item v-if="form.from === 'template'" prop="template">
                                    <el-select v-model="form.template" @change="onEdit('template')">
                                        <template #prefix>{{ $t('container.template') }}</template>
                                        <el-option
                                            v-for="item in templateOptions"
                                            :key="item.id"
                                            :value="item.id"
                                            :label="item.name"
                                        />
                                    </el-select>
                                </el-form-item>
                                <el-form-item v-if="form.from === 'edit' || form.from === 'template'" prop="name">
                                    <el-input @input="changePath" @change="onEdit('')" v-model.trim="form.name">
                                        <template #prefix>
                                            <span style="margin-right: 8px">{{ $t('file.dir') }}</span>
                                        </template>
                                    </el-input>
                                    <span class="input-help">
                                        {{ $t('container.composePathHelper', [composeFile]) }}
                                    </span>
                                </el-form-item>
                                <el-form-item>
                                    <div v-if="form.from === 'edit' || form.from === 'template'" class="w-full">
                                        <CodemirrorPro
                                            v-model="form.file"
                                            placeholder="#Define or paste the content of your docker-compose file here"
                                            mode="yaml"
                                            :heightDiff="400"
                                        ></CodemirrorPro>
                                    </div>
                                </el-form-item>
                                <span class="envTitle">{{ $t('container.env') }}</span>
                                <el-input placeholder="key=value" type="textarea" :rows="3" v-model="form.env" />
                                <span class="envTitle">{{ $t('commons.button.set') }}</span>
                                <el-form-item>
                                    <el-checkbox v-model="form.forcePull" :label="$t('container.forcePull')" />
                                    <span class="input-help">{{ $t('container.forcePullHelper') }}</span>
                                </el-form-item>
                            </el-form>

                            <el-button type="primary" class="mt-2" @click="onSubmit(formRef)">
                                {{ $t('commons.button.save') }}
                            </el-button>
                        </el-card>
                    </el-col>
                </el-row>
                <el-empty v-else :description="$t('commons.msg.noneData')" />
            </template>
        </LayoutContent>

        <TaskLog ref="taskLogRef" width="70%">
            <template #task-footer>
                <el-button @click="handleClose">{{ $t('commons.table.backToList') }}</el-button>
                <el-button type="primary" @click="closeTask">{{ $t('commons.table.keepEdit') }}</el-button>
            </template>
        </TaskLog>
        <FileList ref="fileRef" @choose="loadDir" />
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
import TaskLog from '@/components/log/task/index.vue';
import FileList from '@/components/file-list/index.vue';
import ContainerInspectDialog from '@/views/container/container/inspect/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import ContainerLogDialog from '@/components/log/container-drawer/index.vue';
import DeleteDialog from '@/views/container/compose/delete/index.vue';
import {
    composeOperate,
    composeUpdate,
    containerItemStats,
    containerListStats,
    inspect,
    listComposeTemplate,
    loadComposeEnv,
    searchCompose,
    testCompose,
    upCompose,
} from '@/api/modules/container';
import DockerStatus from '@/views/container/docker-status/index.vue';
import i18n from '@/lang';
import { Container } from '@/api/interface/container';
import { routerToFileWithPath } from '@/utils/router';
import { MsgError, MsgSuccess } from '@/utils/message';
import { computeCPU, computeSize2, computeSizeForDocker, newUUID } from '@/utils/util';
import { Rules } from '@/global/form-rules';
import { loadBaseDir } from '@/api/modules/setting';
import { ElForm } from 'element-plus';

const data = ref<any[]>([]);
const loading = ref(false);
const detailLoading = ref(false);
const currentCompose = ref<Container.ComposeInfo | null>(null);
const currentYamlPath = ref('');
const composeContainers = ref([]);
const composeContent = ref('');

const dialogDelRef = ref();
const containerInspectRef = ref();
const terminalDialogRef = ref();
const containerLogDialogRef = ref();

const searchName = ref('');
const showType = ref('compose');
const containerStats = ref<any[]>([]);
const env = ref();

const isOnCreate = ref();
const oldFrom = ref('edit');
const templateOptions = ref();
const baseDir = ref();
const composeFile = ref();
const taskLogRef = ref();
const fileRef = ref();
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const form = reactive({
    taskID: '',
    name: '',
    from: 'edit',
    path: '',
    file: '',
    template: null as number,
    env: '',
    forcePull: false,
});
const rules = reactive({
    name: [Rules.requiredInput, Rules.composeName],
    path: [Rules.requiredInput],
    template: [Rules.requiredSelect],
});

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

const loadTableHeight = () => {
    if (currentCompose.value?.createdBy === '1Panel') {
        return `calc(100vh - 120px)`;
    } else {
        return `calc(100vh - 240px)`;
    }
};

const refresh = async () => {
    if (!isActive.value || !isExist.value) {
        return;
    }
    let params = {
        info: searchName.value,
        page: 1,
        pageSize: 100,
    };
    await searchCompose(params).then((res) => {
        data.value = res.data.items || [];
    });
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
    form.forcePull = false;
    isOnCreate.value = false;
    detailLoading.value = true;
    currentCompose.value = row;
    currentYamlPath.value = row.path.indexOf(',') !== -1 ? row.path.split(',')[0] : row.path;
    env.value = row.env || '';
    composeContainers.value = row.containers || [];
    inspectCompose(row.name, currentYamlPath.value);
};

const inspectCompose = async (name: string, detailPath: string) => {
    await inspect({ id: name, type: 'compose', detail: detailPath })
        .then((res) => {
            composeContent.value = res.data;
            detailLoading.value = false;
        })
        .finally(() => {
            loadContainerStats();
            detailLoading.value = false;
        });
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
    isOnCreate.value = true;
    loadTemplates();
    form.name = '';
    form.from = 'edit';
    form.path = '';
    form.file = '';
    form.template = null;
    form.env = '';
    form.forcePull = false;
    loadPath();
    loadTemplates();
};
const onEdit = (item: string) => {
    if (item === 'template') {
        changeTemplate();
    }
    if (item === 'form') {
        changeFrom();
    }
};
const changeTemplate = () => {
    for (const item of templateOptions.value) {
        if (form.template === item.id) {
            form.file = item.content;
            break;
        }
    }
};
const changeFrom = () => {
    if ((oldFrom.value === 'edit' || oldFrom.value === 'template') && form.file) {
        ElMessageBox.confirm(i18n.global.t('container.fromChangeHelper'), i18n.global.t('app.source'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        })
            .then(() => {
                if (oldFrom.value === 'template') {
                    form.template = null;
                    form.file = '';
                }
                if (oldFrom.value === 'edit') {
                    form.file = '';
                }
                oldFrom.value = form.from;
            })
            .catch(() => {
                form.from = oldFrom.value;
            });
    } else {
        oldFrom.value = form.from;
    }
};
const loadTemplates = async () => {
    const res = await listComposeTemplate();
    templateOptions.value = res.data;
};
const loadPath = async () => {
    const pathRes = await loadBaseDir();
    baseDir.value = pathRes.data;
    changePath();
};
const changePath = async () => {
    composeFile.value = baseDir.value + '/docker/compose/' + form.name;
};
const loadDir = async (path: string) => {
    form.path = path;
    await loadComposeEnv(path).then((res) => {
        form.env = res.data || '';
    });
};
const handleClose = () => {
    search(true);
    taskLogRef.value?.handleClose();
};
const closeTask = () => {
    taskLogRef.value?.handleClose();
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
            path: row.path,
            operation: operation,
            withFile: false,
            force: false,
        };
        await composeOperate(params)
            .then(async () => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                await search();
                if (currentCompose.value) {
                    const updated = data.value.find((item) => item.name === currentCompose.value.name);
                    if (updated) {
                        await loadDetail(updated, true);
                    }
                }
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const loadSize = async (row: any) => {
    containerItemStats(row.containerID).then((res) => {
        row.sizeRw = res.data.sizeRw || 0;
        row.sizeRootFs = res.data.sizeRootFs || 0;
        row.hasLoadSize = true;
    });
};

const onSubmitEdit = async () => {
    const taskID = newUUID();
    const param = {
        taskID: taskID,
        name: currentCompose.value.name,
        path: currentCompose.value.path,
        detailPath: currentYamlPath.value,
        content: composeContent.value,
        createdBy: currentCompose.value.createdBy,
        env: env.value || '',
        forcePull: form.forcePull,
    };
    loading.value = true;
    await composeUpdate(param)
        .then(async () => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            openTaskLog(taskID);
            await search();
            if (currentCompose.value) {
                const updated = data.value.find((item) => item.name === currentCompose.value.name);
                if (updated) {
                    await loadDetail(updated, true);
                }
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if ((form.from === 'edit' || form.from === 'template') && form.file.length === 0) {
            MsgError(i18n.global.t('container.contentEmpty'));
            return;
        }
        loading.value = true;
        await testCompose(form)
            .then(async (res) => {
                loading.value = false;
                if (res.data) {
                    form.taskID = newUUID();
                    await upCompose(form);
                    openTaskLog(form.taskID);
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                }
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const openComposeFolder = (row: any) => {
    if (row?.workdir) {
        routerToFileWithPath(row.workdir);
    }
};
const onInspectContainer = async (item: any) => {
    if (!item.containerID) {
        return;
    }
    const res = await inspect({ id: item.containerID, type: 'container', detail: '' });
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
.envTitle {
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-primary);
    margin-top: 12px;
    margin-bottom: 4px;
    display: block;
}
</style>
