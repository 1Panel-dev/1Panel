<template>
    <DrawerPro v-model="drawerVisible" :header="$t('menu.home')" :resource="resource" size="60%" @close="handleClose">
        <div v-loading="loading" class="overview-drawer">
            <div class="toolbar">
                <el-button :loading="loading" @click="loadOverview">
                    <el-icon><Refresh /></el-icon>
                    {{ $t('commons.button.refresh') }}
                </el-button>
            </div>

            <section class="section">
                <span class="section-title">{{ $t('aiTools.agents.overviewSnapshot') }}</span>
                <el-divider class="divider" />
                <el-row :gutter="20" class="metrics-row">
                    <el-col v-for="item in snapshotCards" :key="item.label" :xs="24" :sm="12" :md="12" :lg="6" :xl="6">
                        <el-form-item label-position="top">
                            <template #label>
                                <span class="metric-label">{{ item.label }}</span>
                            </template>
                            <span class="metric-value">{{ item.value }}</span>
                        </el-form-item>
                    </el-col>
                </el-row>
            </section>
        </div>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { Refresh } from '@element-plus/icons-vue';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { getAgentOverview } from '@/api/modules/ai';

const { t } = useI18n();
const drawerVisible = ref(false);
const loading = ref(false);
const resource = ref('');
const agentId = ref(0);
const data = ref<AI.AgentOverview>({
    snapshot: {
        containerStatus: '',
        appVersion: '',
        defaultModel: '',
        channelCount: 0,
        skillCount: 0,
        jobCount: 0,
        sessionCount: 0,
    },
});

const formatNumber = (value: number) => {
    return new Intl.NumberFormat().format(value || 0);
};

const formatText = (value: string) => {
    return value || '-';
};

const snapshotCards = computed(() => [
    { label: t('commons.table.status'), value: formatText(data.value.snapshot.containerStatus) },
    { label: t('aiTools.agents.appVersion'), value: formatText(data.value.snapshot.appVersion) },
    { label: t('aiTools.agents.defaultModel'), value: formatText(data.value.snapshot.defaultModel) },
    { label: t('aiTools.agents.channelCount'), value: formatNumber(data.value.snapshot.channelCount) },
    { label: t('aiTools.agents.skillCount'), value: formatNumber(data.value.snapshot.skillCount) },
    { label: t('aiTools.agents.jobCount'), value: formatNumber(data.value.snapshot.jobCount) },
    { label: t('aiTools.agents.sessionCount'), value: formatNumber(data.value.snapshot.sessionCount) },
]);

const loadOverview = async () => {
    if (!agentId.value) {
        return;
    }
    loading.value = true;
    try {
        const res = await getAgentOverview({ agentId: agentId.value });
        data.value = res.data;
    } finally {
        loading.value = false;
    }
};

const acceptParams = async (row: AI.AgentItem) => {
    agentId.value = row.id;
    resource.value = row.name;
    drawerVisible.value = true;
    await loadOverview();
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.overview-drawer {
    display: flex;
    flex-direction: column;
    gap: 20px;
}

.toolbar {
    display: flex;
    justify-content: flex-end;
}

.section {
    display: flex;
    flex-direction: column;
}

.section-title {
    font-size: 20px;
    font-weight: 600;
}

.divider {
    display: block;
    height: 1px;
    width: 100%;
    margin: 12px 0;
    border-top: 1px var(--el-border-color) var(--el-border-style);
}

.metrics-row {
    padding: 0 16px;
}

.metric-label {
    color: var(--el-text-color-secondary);
}

.metric-value {
    display: inline-block;
    margin-top: 2px;
    font-size: 16px;
    font-weight: 600;
    word-break: break-word;
}

@media (max-width: 1200px) {
    .metrics-row {
        padding: 0;
    }
}
</style>
