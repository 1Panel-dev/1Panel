<template>
    <div>
        <DrawerPro
            v-model="open"
            :header="$t('monitor.runtimeDiagnostics')"
            size="large"
            :confirm-before-close="captureLoading"
            @before-close="handleBeforeClose"
        >
            <div
                v-loading="captureLoading"
                element-loading-lock
                :element-loading-text="$t('monitor.capturing')"
                class="diagnostics-content"
            >
                <section>
                    <el-row :gutter="12">
                        <el-col v-for="item in summaryCards" :key="item.label" :xs="12" :sm="8" :md="6">
                            <div class="summary-item">
                                <el-card class="summary-title">
                                    <span>{{ item.label }}</span>
                                    <el-tooltip :content="item.description" placement="top" :show-after="200">
                                        <el-icon class="summary-help"><QuestionFilled /></el-icon>
                                    </el-tooltip>
                                    <div class="summary-value">{{ item.value }}</div>
                                </el-card>
                            </div>
                        </el-col>
                    </el-row>
                </section>

                <el-divider />

                <section>
                    <div class="section-title">{{ $t('monitor.manualCapture') }}</div>
                    <div class="profile-toolbar">
                        <el-select
                            v-model="captureForm.type"
                            class="profile-type"
                            popper-class="runtime-profile-select-popper"
                        >
                            <el-option
                                v-for="item in profileTypes"
                                :key="item.value"
                                :label="item.label"
                                :value="item.value"
                            >
                                <div class="runtime-profile-option">
                                    <span class="runtime-profile-option-title">{{ item.label }}</span>
                                    <span class="runtime-profile-option-description">{{ item.description }}</span>
                                </div>
                            </el-option>
                        </el-select>
                        <el-input-number
                            v-if="showDuration"
                            v-model="captureForm.duration"
                            :min="5"
                            :max="30"
                            controls-position="right"
                        />
                        <span v-if="showDuration" class="text-sm">{{ $t('commons.units.secondUnit') }}</span>
                        <el-button type="primary" @click="captureProfile">
                            {{ $t('monitor.startCapture') }}
                        </el-button>
                    </div>
                </section>

                <el-divider />

                <section>
                    <div class="section-header">
                        <div>
                            <div class="section-title">{{ $t('monitor.goroutineSnapshot') }}</div>
                            <span class="goroutine-summary">
                                {{
                                    $t('monitor.goroutineSummary', [
                                        goroutineSnapshot.total,
                                        goroutineSnapshot.groupCount,
                                    ])
                                }}
                            </span>
                        </div>
                        <el-tooltip :content="$t('monitor.captureSnapshot')" placement="top">
                            <el-button
                                icon="Refresh"
                                text
                                circle
                                :loading="goroutineLoading"
                                :aria-label="$t('monitor.captureSnapshot')"
                                @click="loadGoroutines"
                            />
                        </el-tooltip>
                    </div>

                    <el-alert
                        v-if="goroutineSnapshot.truncated"
                        class="snapshot-warning"
                        type="warning"
                        :title="$t('monitor.goroutineTruncated')"
                        :closable="false"
                        show-icon
                    />

                    <ComplexTable :data="goroutineSnapshot.goroutines" v-loading="goroutineLoading" max-height="420">
                        <el-table-column prop="count" :label="$t('monitor.count')" width="90" sortable />
                        <el-table-column prop="state" :label="$t('commons.table.status')" width="180" />
                        <el-table-column
                            prop="top"
                            :label="$t('monitor.topFunction')"
                            min-width="360"
                            show-overflow-tooltip
                        />
                        <el-table-column :label="$t('commons.table.operate')" fixed="right" width="90">
                            <template #default="scope">
                                <el-button link type="primary" @click="showGoroutineStack(scope.row)">
                                    {{ $t('commons.button.view') }}
                                </el-button>
                            </template>
                        </el-table-column>
                    </ComplexTable>
                </section>
            </div>
        </DrawerPro>

        <el-dialog
            v-model="stackDialogVisible"
            :title="$t('monitor.goroutineSnapshot')"
            width="70%"
            append-to-body
            destroy-on-close
        >
            <pre class="stack-view">{{ selectedStack.join('\n') }}</pre>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import {
    createRuntimeProfile,
    loadRuntimeDiagnosticsSummary,
    loadRuntimeGoroutines,
    RuntimeProfileDownloadError,
} from '@/api/modules/host';
import { Host } from '@/api/interface/host';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { computeSize } from '@/utils/size';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';

const { currentNode } = useGlobalStore();
const open = ref(false);
const captureLoading = ref(false);
const goroutineLoading = ref(false);
const stackDialogVisible = ref(false);
const selectedStack = ref<string[]>([]);
const summary = reactive<Host.RuntimeDiagnosticsSummary>({
    rss: 0,
    heapAlloc: 0,
    heapObjects: 0,
    goroutines: 0,
});
const goroutineSnapshot = reactive<Host.RuntimeGoroutineSnapshot>({
    total: 0,
    groupCount: 0,
    truncated: false,
    capturedAt: '',
    goroutines: [],
});
const captureForm = reactive<Host.RuntimeProfileCreate>({ type: 'cpu', duration: 15 });

const profileTypes = computed(() => [
    {
        label: i18n.global.t('monitor.profileCPU'),
        description: i18n.global.t('monitor.profileCPUDesc'),
        value: 'cpu',
    },
    {
        label: i18n.global.t('monitor.profileHeap'),
        description: i18n.global.t('monitor.profileHeapDesc'),
        value: 'heap',
    },
    {
        label: i18n.global.t('monitor.profileGoroutine'),
        description: i18n.global.t('monitor.profileGoroutineDesc'),
        value: 'goroutine',
    },
    {
        label: i18n.global.t('monitor.profileMutex'),
        description: i18n.global.t('monitor.profileMutexDesc'),
        value: 'mutex',
    },
    {
        label: i18n.global.t('monitor.profileBlock'),
        description: i18n.global.t('monitor.profileBlockDesc'),
        value: 'block',
    },
]);

const showDuration = computed(() => captureForm.type !== 'heap' && captureForm.type !== 'goroutine');

const summaryCards = computed(() => [
    {
        label: i18n.global.t('monitor.runtimeRSS'),
        value: computeSize(summary.rss),
        description: i18n.global.t('monitor.runtimeRSSDesc'),
    },
    {
        label: i18n.global.t('monitor.runtimeHeapAlloc'),
        value: computeSize(summary.heapAlloc),
        description: i18n.global.t('monitor.runtimeHeapAllocDesc'),
    },
    {
        label: i18n.global.t('monitor.runtimeHeapObjects'),
        value: summary.heapObjects.toLocaleString(),
        description: i18n.global.t('monitor.runtimeHeapObjectsDesc'),
    },
    {
        label: i18n.global.t('monitor.runtimeGoroutines'),
        value: summary.goroutines.toLocaleString(),
        description: i18n.global.t('monitor.runtimeGoroutineDesc'),
    },
]);

const loadSummary = async () => {
    const res = await loadRuntimeDiagnosticsSummary(currentNode.value);
    Object.assign(summary, res.data);
};

const loadGoroutines = async () => {
    goroutineLoading.value = true;
    try {
        const res = await loadRuntimeGoroutines(currentNode.value);
        Object.assign(goroutineSnapshot, res.data);
    } finally {
        goroutineLoading.value = false;
    }
};

const showGoroutineStack = (row: Host.RuntimeGoroutineGroup) => {
    selectedStack.value = row.stack || [];
    stackDialogVisible.value = true;
};

const captureProfile = async () => {
    captureLoading.value = true;
    try {
        const data = await createRuntimeProfile(captureForm, currentNode.value);
        const url = window.URL.createObjectURL(data);
        const link = document.createElement('a');
        link.href = url;
        link.download = `${captureForm.type}-${Date.now()}.pb.gz`;
        link.click();
        window.URL.revokeObjectURL(url);
        MsgSuccess(i18n.global.t('monitor.captureSuccess'));
        await loadSummary();
    } catch (error) {
        if (error instanceof RuntimeProfileDownloadError) {
            MsgError(error.message || i18n.global.t('commons.res.commonError'));
        }
    } finally {
        captureLoading.value = false;
    }
};

const acceptParams = () => {
    open.value = true;
    Promise.all([loadSummary(), loadGoroutines()]);
};

const handleBeforeClose = (done: () => void) => {
    if (!captureLoading.value) {
        done();
    }
};

defineExpose({ acceptParams });
</script>

<style scoped>
.summary-item {
    padding: 4px 12px 4px 0;
}
.summary-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    color: var(--el-text-color-secondary);
    font-size: 13px;
}
.summary-help {
    flex-shrink: 0;
    color: var(--el-text-color-placeholder);
    cursor: help;
}
.summary-value {
    margin-top: 5px;
    color: var(--el-text-color-primary);
    font-size: 18px;
    font-weight: 600;
}
.diagnostics-content {
    min-height: 520px;
}
.section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    min-height: 32px;
    margin-bottom: 10px;
}
.section-title {
    margin-bottom: 10px;
    color: var(--el-text-color-primary);
    font-size: 15px;
    font-weight: 600;
}
.section-header .section-title {
    margin-bottom: 4px;
}
.profile-toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px;
}
.profile-type {
    width: 160px;
}
.goroutine-summary {
    color: var(--el-text-color-secondary);
    font-size: 13px;
}
.snapshot-warning {
    margin-bottom: 10px;
}
.stack-view {
    max-height: 65vh;
    margin: 0;
    padding: 14px;
    overflow: auto;
    border-radius: 4px;
    background: var(--el-fill-color-light);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 12px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
}
</style>

<style>
.runtime-profile-select-popper {
    min-width: min(560px, calc(100vw - 64px));
}
.runtime-profile-select-popper .el-select-dropdown__item {
    height: auto;
    min-height: 54px;
    padding-top: 6px;
    padding-bottom: 6px;
    line-height: normal;
}
.runtime-profile-option {
    display: flex;
    flex-direction: column;
    gap: 4px;
    white-space: normal;
}
.runtime-profile-option-title {
    color: var(--el-text-color-primary);
    font-weight: 500;
}
.runtime-profile-option-description {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 17px;
}
</style>
