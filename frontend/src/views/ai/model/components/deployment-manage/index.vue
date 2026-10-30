<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="header"
        :resource="instance?.name"
        :size="isMobile ? 'full' : 'large'"
        @close="emit('search')"
    >
        <template #content>
            <div class="deployment-action" v-loading="loading">
                <el-alert v-if="failed" :title="$t('aiTools.model.loadFailed')" type="error" :closable="false" />
                <el-empty v-else-if="missing" :description="$t('commons.res.notFound')" />
                <template v-else-if="visible && instance && !loading && view !== VllmPlaceholder">
                    <div v-if="instance.engine !== 'ollama'" class="service-card">
                        <div class="flex flex-wrap gap-3 items-center mb-4">
                            <strong>
                                {{ engineLabels[instance.engine] }}
                                {{
                                    instance.engine === 'vllm' ? formatVllmVersion(instance.version) : instance.version
                                }}
                            </strong>
                            <Status :status="instance.status" />
                        </div>
                        <div class="service-actions">
                            <el-button
                                v-for="button in visibleButtons"
                                :key="button.label"
                                :disabled="
                                    deleteLoading ||
                                    (button.disabled?.(instance) ?? false) ||
                                    (button.permission && !hasManagePermissionAccess())
                                "
                                @click="button.click(instance)"
                            >
                                {{ button.label }}
                            </el-button>
                        </div>
                    </div>
                    <el-descriptions v-if="instance.engine !== 'ollama'" :column="1" border>
                        <el-descriptions-item :label="$t('commons.table.name')">
                            {{ instance.name }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('menu.container')">
                            {{ instance.containerName }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('aiTools.model.modelDir')">
                            {{ instance.modelDir || '—' }}
                        </el-descriptions-item>
                    </el-descriptions>
                </template>
                <OllamaModels
                    v-if="visible && !standalone && instance?.engine === 'ollama' && !failed && !missing"
                    :key="instance.id"
                    :instance-name="instance.name"
                />
                <VllmPlaceholder v-if="visible && view === VllmPlaceholder" />
            </div>
        </template>
    </DrawerPro>
    <component
        :is="view"
        v-if="visible && view && view !== VllmPlaceholder && !failed && !missing"
        ref="contentRef"
        v-bind="
            view === AppParams
                ? { modelGPU: true }
                : view === EngineConnection
                  ? {}
                  : view === ContainerLogs
                    ? { highlightDiff: 210 }
                    : view === TaskLog
                      ? { width: '70%' }
                      : ['config', 'monitor'].includes(actionName)
                        ? {}
                        : { dialog: true }
        "
        @close="returnToManagement"
        @search="emit('search')"
        @task="openTask"
    />
    <DialogPro
        v-model="deleteVisible"
        :title="$t('commons.button.delete') + ' - ' + (deleting?.name || '')"
        size="small"
    >
        <p>{{ $t('commons.msg.delete', [deleting?.name || '']) }}</p>
        <el-checkbox v-if="deleting?.engine === 'vllm'" v-model="forceDelete" :label="$t('website.forceDelete')" />
        <template #footer>
            <el-button :disabled="deleteLoading" @click="deleteVisible = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button v-permission type="primary" :loading="deleteLoading" @click="deleteDeployment">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { computed, markRaw, nextTick, ref, shallowRef, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { loadOptionalComponent } from '@/extensions/optional';
import { checkAppInstalled, searchAppInstalled } from '@/api/modules/app';
import { pageTensorRTLLM, deleteTensorRTLLM } from '@/api/modules/ai';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
import { hasPermission } from '@/utils/rbac';
import { hasManagePermissionAccess } from '@/utils/permission';
import { newUUID } from '@/utils/id';
import OllamaModels from '../../ollama/models/index.vue';
import AppParams from '@/views/app-store/installed/detail/index.vue';
import EngineConnection from '../connection/index.vue';
import TensorForm from '../../tensorrt/operate/index.vue';
import VllmPlaceholder from '../../vllm/index.vue';
import ContainerLogs from '@/components/log/container-drawer/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import { formatVllmVersion, readPages, loadVllmAPI, vllmAvailable, type Deployment } from '../../deployment';

const emit = defineEmits(['search', 'upgrade']);
const { isMobile, currentNode, isProductPro, isFxplay } = useGlobalStore();
const route = useRoute();
const visible = ref(false);
const standalone = ref(false);
const drawerVisible = computed({
    get: () => visible.value && !standalone.value,
    set: (value) => {
        visible.value = value;
    },
});
const loading = ref(false);
const failed = ref(false);
const missing = ref(false);
const header = ref('');
const instance = ref<Deployment>();
const actionName = ref('manage');
const deleteVisible = ref(false);
const deleting = ref<Deployment>();
const forceDelete = ref(false);
const deleteLoading = ref(false);
const deleted = ref(false);
const engineLabels = { vllm: 'vLLM', ollama: 'Ollama', tensorrt: 'TensorRT LLM' };
const view = shallowRef();
const contentRef = ref();
let request = 0;
const openTask = async (taskID: string) => {
    emit('search');
    visible.value = true;
    actionName.value = 'task';
    view.value = markRaw(TaskLog);
    await nextTick();
    contentRef.value?.openWithTaskID(taskID);
};
const openUpgrade = async (installId: number, version: number) => {
    const installed = await readPages((page) =>
        searchAppInstalled({ ...page, name: instance.value.name, update: true }),
    );
    if (version !== request || !visible.value) return;
    const app = installed.find((item) => item.id === installId);
    if (!app) {
        missing.value = true;
        if (standalone.value) MsgError(i18n.global.t('commons.res.notFound'));
        return;
    }
    if (standalone.value) visible.value = false;
    emit('upgrade', app);
};
const open = async (
    row: Pick<Deployment, 'engine' | 'id' | 'name'> & Partial<Deployment>,
    action = 'manage',
    outside = false,
) => {
    const version = ++request;
    standalone.value = outside;
    if (row.engine === 'tensorrt' && !isFxplay.value) return;
    instance.value = { version: '', containerName: '', status: '', modelDir: '', models: [], ...row };
    actionName.value = action;
    deleted.value = false;
    header.value = `${engineLabels[row.engine]} ${i18n.global.t('aiTools.model.manage')}`;
    view.value = undefined;
    failed.value = false;
    missing.value = false;
    visible.value = true;
    loading.value = true;
    try {
        if (row.engine === 'vllm' && (!isProductPro.value || !vllmAvailable)) {
            view.value = markRaw(VllmPlaceholder);
            return;
        }
        if (row.engine === 'ollama') {
            const { data } = await checkAppInstalled('ollama', row.name);
            if (version !== request || !visible.value) return;
            if (!data.isExist) {
                missing.value = true;
                if (standalone.value) MsgError(i18n.global.t('commons.res.notFound'));
                return;
            }
            instance.value = {
                ...instance.value,
                id: data.appInstallId,
                version: data.version,
                status: data.status,
                containerName: data.containerName,
            };
            if (action === 'upgrade') {
                await openUpgrade(data.appInstallId, version);
                return;
            }
            const component = { config: AppParams, connection: EngineConnection, log: ContainerLogs }[action];
            view.value = component ? markRaw(component) : undefined;
            await nextTick();
            if (version !== request || !visible.value) return;
            if (action === 'config') await contentRef.value?.acceptParams({ id: data.appInstallId, edit: true });
            if (action === 'connection')
                contentRef.value?.acceptParams({
                    name: data.name,
                    engine: 'ollama',
                    port: data.httpPort,
                    containerName: data.containerName,
                    appinstallID: data.appInstallId,
                });
            if (action === 'log')
                contentRef.value?.acceptParams({ containerID: data.containerName, container: data.containerName });
            return;
        }
        const items =
            row.engine === 'vllm'
                ? await readPages((page) => loadVllmAPI().then((api) => api.pageVllm({ ...page, info: row.name })))
                : await readPages((page) => pageTensorRTLLM({ ...page, name: row.name }));
        const detail = items.find((item) => item.id === row.id);
        if (version !== request || !visible.value) return;
        if (!detail) {
            missing.value = true;
            if (standalone.value) MsgError(i18n.global.t('commons.res.notFound'));
            return;
        }
        instance.value = {
            ...instance.value,
            status: detail.status,
            containerName: detail.containerName,
            modelDir: detail.modelDir,
            version: 'appVersion' in detail ? detail.appVersion : detail.version,
            upgradable: 'upgradable' in detail && detail.upgradable,
        };
        if (action === 'connection') {
            view.value = markRaw(EngineConnection);
            await nextTick();
            contentRef.value?.acceptParams({
                containerName: detail.containerName,
                port:
                    'port' in detail
                        ? detail.port
                        : (detail.exposedPorts || []).map((item) => item.hostPort).join(', '),
                containerPort:
                    'port' in detail ? '' : (detail.exposedPorts || []).map((item) => item.containerPort).join(', '),
                endpoint: 'modelAccountBaseURL' in detail ? detail.modelAccountBaseURL : '',
                engine: row.engine,
                name: row.name,
                modelDir: detail.modelDir,
            });
        } else if (action === 'delete') {
            deleting.value = instance.value;
            forceDelete.value = false;
            deleteVisible.value = true;
        } else if (action === 'config') {
            const component =
                row.engine === 'vllm'
                    ? await loadOptionalComponent('/src/xpack/views/vllm/operate/index.vue')
                    : TensorForm;
            if (version !== request || !visible.value) return;
            view.value = markRaw(component);
            await nextTick();
            await contentRef.value?.openEdit(detail);
        } else if (action === 'monitor') {
            const component = await loadOptionalComponent('/src/xpack/views/vllm/monitor/index.vue');
            if (version !== request || !visible.value) return;
            view.value = markRaw(component);
            await nextTick();
            contentRef.value?.acceptParams(detail);
        } else if (action === 'upgrade' && 'appInstallId' in detail) {
            await openUpgrade(detail.appInstallId, version);
        } else if (action === 'log') {
            const task = 'appInstallId' in detail && ['Installing', 'Uninstalling'].includes(detail.status);
            view.value = markRaw(task ? TaskLog : ContainerLogs);
            await nextTick();
            if (task) {
                contentRef.value?.openWithResourceID(
                    'App',
                    detail.status === 'Installing' ? 'TaskInstall' : 'TaskUninstall',
                    detail.appInstallId,
                );
            } else {
                contentRef.value?.acceptParams({
                    containerID: detail.containerName,
                    container: detail.containerName,
                });
            }
        }
    } catch {
        if (version === request) {
            failed.value = true;
            if (standalone.value) MsgError(i18n.global.t('aiTools.model.loadFailed'));
        }
    } finally {
        if (version === request) loading.value = false;
    }
};
watch(visible, (value) => {
    if (!value) request++;
});
watch([currentNode, () => route.fullPath, isProductPro, isFxplay], () => {
    request++;
    visible.value = false;
    deleteVisible.value = false;
    deleting.value = undefined;
});
const operationDisabled = (row: Deployment) =>
    deleteLoading.value ||
    ['Installing', 'Starting', 'ReStarting', 'Restarting', 'Rebuilding', 'Waiting', 'Uninstalling'].includes(
        row.status,
    );
const deleteDeployment = async () => {
    const version = request;
    const row = deleting.value;
    if (!row) return;
    deleteLoading.value = true;
    try {
        if (row.engine === 'tensorrt') {
            await deleteTensorRTLLM({ id: row.id });
        } else {
            const taskID = newUUID();
            await (
                await loadVllmAPI()
            ).operateVllm({ id: row.id, operate: 'delete', taskID, forceDelete: forceDelete.value });
            if (version !== request || !visible.value) return;
            await openTask(taskID);
        }
        if (version !== request || !visible.value) return;
        deleted.value = true;
        deleteVisible.value = false;
        emit('search');
        if (row.engine === 'tensorrt') visible.value = false;
    } finally {
        deleteLoading.value = false;
    }
};
const deploymentButtons = [
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        disabled: operationDisabled,
        click: (row: Deployment) => open(row, 'config'),
    },
    {
        label: i18n.global.t('menu.monitor'),
        show: (row: Deployment) => row.engine === 'vllm',
        click: (row: Deployment) => open(row, 'monitor'),
    },
    {
        label: i18n.global.t('commons.button.upgrade'),
        permission: true,
        show: (row: Deployment) => row.upgradable && hasPermission('app_manage'),
        disabled: operationDisabled,
        click: (row: Deployment) => open(row, 'upgrade'),
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        disabled: (row: Deployment) =>
            deleteLoading.value ||
            (operationDisabled(row) && !['Uninstalling', 'ReStarting', 'Restarting'].includes(row.status)),
        click: (row: Deployment) => {
            deleting.value = row;
            forceDelete.value = false;
            deleteVisible.value = true;
        },
    },
];
const visibleButtons = computed(() =>
    instance.value ? deploymentButtons.filter((button) => !button.show || button.show(instance.value)) : [],
);
const returnToManagement = async () => {
    const previousView = view.value;
    await nextTick();
    if (view.value !== previousView) return;
    if (['task', 'upgrade', 'config'].includes(actionName.value)) emit('search');
    if (deleted.value || standalone.value) visible.value = false;
    else if (instance.value) open(instance.value);
};
defineExpose({ open });
</script>

<style scoped lang="scss">
.service-card {
    border: 1px solid var(--el-border-color-light);
    border-radius: 8px;
    padding: 16px;
    margin-bottom: 20px;
    background: var(--el-fill-color-extra-light);
}
.service-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    .el-button {
        margin-left: 0;
    }
}
.deployment-action {
    min-height: 180px;
    overflow-wrap: anywhere;
    :deep(.content-container__main > .el-card) {
        border: none;
        box-shadow: none;
        --el-card-padding: 0px;
    }
}
</style>
