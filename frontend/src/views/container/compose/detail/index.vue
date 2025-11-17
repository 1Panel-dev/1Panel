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
                <div v-else class="compose-detail-container">
                    <!-- Status Bar - Firewall Style -->
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

                    <!-- Container Status List -->
                    <el-card v-if="composeInfo && composeContainers.length > 0" class="mb-4" shadow="never">
                        <template #header>
                            <div class="flex flex-wrap items-center justify-between gap-3">
                                <span class="text-sm font-medium">
                                    {{ $t('container.containerStatus') }}
                                    ( {{ composeInfo?.containerCount || 0 }} / {{ composeInfo?.runningCount || 0 }} )
                                </span>
                            </div>
                        </template>
                        <div class="compose-container-list-horizontal">
                            <div
                                v-for="item in composeContainers"
                                :key="item.containerID || item.name"
                                class="compose-container-item-horizontal"
                                @click="onInspectContainer(item)"
                                :title="$t('commons.button.view')"
                            >
                                <div class="container-name">
                                    {{ item.name }}
                                </div>
                                <Status :status="item.state" />
                            </div>
                        </div>
                    </el-card>

                    <div class="compose-detail-grid">
                        <el-card class="compose-detail-card compose-detail-editor" shadow="never">
                            <template #header>
                                <div class="font-medium">
                                    {{ $t('container.composeTemplate') }}
                                </div>
                            </template>
                            <div class="compose-editor-body">
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

                        <!-- Log Card -->
                        <el-card class="compose-detail-card compose-detail-log" shadow="never">
                            <template #header>
                                <div class="flex items-center justify-between gap-3">
                                    <span>{{ $t('commons.button.log') }}</span>
                                </div>
                            </template>
                            <div class="compose-log-body">
                                <ContainerLog
                                    v-if="composePath"
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
    </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import NoSuchService from '@/components/layout-content/no-such-service.vue';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import ContainerLog from '@/components/log/container/index.vue';
import ContainerInspectDialog from '@/views/container/container/inspect/index.vue';
import { composeOperator, composeUpdate, inspect, loadDockerStatus, searchCompose } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { routerToFileWithPath, routerToName } from '@/utils/router';
import { MsgError, MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

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
            refreshDetail();
        })
        .catch(() => {
            dockerLoading.value = false;
            isActive.value = false;
            isExist.value = false;
        });
};

const refreshDetail = async () => {
    if (!composeName.value || !isActive.value || !isExist.value) {
        return;
    }
    detailLoading.value = true;
    try {
        await loadComposeInfo();
        await loadComposeContent();
        logKey.value++;
    } finally {
        detailLoading.value = false;
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
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
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

onMounted(() => {
    syncRouteParams();
    loadStatus();
});

watch(
    () => route.fullPath,
    () => {
        syncRouteParams();
        refreshDetail();
    },
);

watch([isActive, isExist], () => {
    refreshDetail();
});
</script>

<style scoped lang="scss">
.compose-detail-container {
    width: 100%;
}

.app-status {
    :deep(.el-card__body) {
        padding: 16px 20px;
    }
}

.compose-detail-grid {
    display: grid;
    gap: 16px;
    grid-template-columns: minmax(0, 5fr) minmax(0, 7fr);
    min-height: 600px;
}

.compose-detail-card {
    height: 100%;
}

.compose-detail-editor {
    display: flex;
    flex-direction: column;

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

.compose-editor-body {
    flex: 1;
}

.compose-container-summary {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: 12px;
    padding: 0 12px 12px;
}

.summary-card {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    padding: 12px;
    background: var(--el-fill-color-light);
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.summary-label {
    font-size: 12px;
    color: var(--el-text-color-secondary);
}

.summary-value {
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color-primary);
}

.compose-container-list-horizontal {
    display: flex;
    gap: 12px;
    overflow-x: auto;
    padding: 12px;

    &::-webkit-scrollbar {
        height: 6px;
    }

    &::-webkit-scrollbar-thumb {
        background-color: var(--el-border-color);
        border-radius: 3px;

        &:hover {
            background-color: var(--el-border-color-dark);
        }
    }

    &::-webkit-scrollbar-track {
        background-color: var(--el-fill-color-lighter);
        border-radius: 3px;
    }
}

.compose-container-item-horizontal {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 10px 14px;
    background-color: var(--el-fill-color-light);
    border-radius: 6px;
    border: 1px solid var(--el-border-color-lighter);
    min-width: 140px;
    flex-shrink: 0;
    transition: all 0.2s ease;
    cursor: pointer;

    &:hover {
        border-color: var(--el-color-primary-light-5);
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.12);
        transform: translateY(-2px);
        background-color: var(--el-color-primary-light-9);
    }

    &:active {
        transform: translateY(0);
    }

    .container-name {
        font-size: 13px;
        font-weight: 500;
        color: var(--el-text-color-primary);
        text-align: center;
        word-break: break-word;
        max-width: 120px;
        transition: color 0.2s ease;
    }

    &:hover .container-name {
        color: var(--el-color-primary);
    }
}

.compose-log-body {
    flex: 1;
    display: flex;
    flex-direction: column;
}

@media (max-width: 1280px) {
    .compose-detail-grid {
        grid-template-columns: 1fr;
        min-height: auto;
    }
}
</style>
