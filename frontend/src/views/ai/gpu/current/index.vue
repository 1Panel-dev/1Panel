<template>
    <div>
        <RouterMenu :history-disabled="historyDisabled" />
        <div>
            <LayoutContent v-loading="loading" :title="$t('aiTools.gpuGroup.overview')">
                <template #leftToolBar>
                    <div class="monitor-heading">
                        <h3 class="environment-heading">{{ $t('aiTools.gpuGroup.overview') }}</h3>
                        <template v-if="loadFailed || gpuInfo.environment || gpuInfo.warnings?.length">
                            <el-tag :type="environmentPassed ? 'success' : loadFailed ? 'danger' : 'warning'">
                                {{
                                    $t(`aiTools.gpu.environment.${environmentStatus}Label`, {
                                        count: environmentIssueCount,
                                    })
                                }}
                            </el-tag>
                            <el-button class="environment-detail-link" link @click="environmentDialogVisible = true">
                                {{ $t('aiTools.gpu.environment.details') }}
                            </el-button>
                        </template>
                    </div>
                </template>
                <template #rightToolBar>
                    <span v-if="hasAccelerators && gpuInfo.collectedAt" class="collection-time">
                        {{ $t('aiTools.gpu.collectedAt') }}:
                        {{ new Date(gpuInfo.collectedAt).toLocaleString() }}
                    </span>
                    <TableSetting title="gpu-refresh" @search="search()" />
                    <el-tooltip :content="$t('aiTools.gpu.environment.recheck')">
                        <TableRefresh :disabled="loading" @search="refresh()" />
                    </el-tooltip>
                </template>
                <template #main>
                    <section class="topology-overview mb-4" v-loading="topologyLoading">
                        <h3 class="environment-heading mb-3">{{ $t('aiTools.gpuGroup.topology') }}</h3>
                        <el-alert
                            v-if="topologyFailed"
                            class="mb-3"
                            :closable="false"
                            type="warning"
                            :title="$t('aiTools.model.loadFailed')"
                        />
                        <GPUTopology v-if="!topologyFailed" :devices="topology.devices" :nodes="topology.numaNodes" />
                        <div v-for="warning in topology.warnings" :key="warning" class="input-help">{{ warning }}</div>
                    </section>

                    <div v-for="summary in overviewRows" :key="summary.vendor" class="device-overview mb-2">
                        <div class="overview-item overview-model">
                            <strong>{{ commonDeviceModels[summary.vendor] || vendorTitle(summary.vendor) }}</strong>
                        </div>
                        <div class="overview-item">
                            <span>{{ $t('aiTools.gpu.driverVersion') }}</span>
                            <strong>{{ displayValue(summary.driverVersion) }}</strong>
                        </div>
                        <div class="overview-item">
                            <span>{{ $t('aiTools.gpu.cudaVersion') }}</span>
                            <el-tooltip
                                v-if="summary.vendor === 'nvidia' || summary.cudaVersion"
                                :content="$t('aiTools.gpu.environment.cudaHelper')"
                            >
                                <strong>{{ displayValue(summary.cudaVersion) }}</strong>
                            </el-tooltip>
                            <strong v-else>{{ $t('aiTools.gpu.environment.notApplicable') }}</strong>
                        </div>
                        <div class="overview-item">
                            <span>
                                {{
                                    $t(
                                        summary.vendor === 'intel' || summary.vendor === 'xpu'
                                            ? 'aiTools.gpu.environment.gpuCount'
                                            : 'aiTools.gpu.environment.cardCount',
                                    )
                                }}
                            </span>
                            <strong>{{ summary.cardCountText }}</strong>
                        </div>
                        <div class="overview-item">
                            <span>
                                {{
                                    $t(
                                        summary.sharedMemoryOnly
                                            ? 'aiTools.gpu.systemMemory'
                                            : 'aiTools.gpu.memoryTotal',
                                    )
                                }}
                            </span>
                            <el-tooltip
                                :disabled="!summary.sharedMemoryOnly && !summary.memoryPartial"
                                :content="summaryMemoryTooltip(summary)"
                            >
                                <strong v-if="!summary.sharedMemoryOnly && summary.memoryUsage">
                                    {{ $t('home.used') }} {{ formatSummaryMemory(summary.memoryUsage.used) }} /
                                    {{ formatSummaryMemory(summary.memoryUsage.total) }}
                                    <template v-if="summary.memoryUsage.percentage !== null">
                                        ({{ summary.memoryUsage.percentage }}%)
                                    </template>
                                </strong>
                                <strong v-else>
                                    {{
                                        formatSummaryMemory(
                                            summary.sharedMemoryOnly ? summary.sharedMemoryTotal : summary.memoryTotal,
                                        )
                                    }}
                                </strong>
                            </el-tooltip>
                            <el-tooltip
                                v-if="!summary.sharedMemoryOnly && summary.sharedMemoryTotal !== null"
                                :content="$t('aiTools.gpu.systemMemoryHelper')"
                            >
                                <span class="whitespace-normal">
                                    {{
                                        $t('aiTools.gpu.environment.sharedMemory', {
                                            memory: formatSummaryMemory(summary.sharedMemoryTotal),
                                        })
                                    }}
                                </span>
                            </el-tooltip>
                        </div>
                    </div>
                    <el-empty
                        v-if="!hasAccelerators && !loading && !loadFailed && !showEnvironment"
                        :description="emptyDescription"
                    />
                    <div v-if="hasAccelerators" class="gpu-group-list">
                        <section v-for="group in gpuGroups" :key="group.key" class="gpu-group">
                            <div v-if="group.title" class="group-header">
                                <div class="group-identity">
                                    <strong>{{ group.title }}</strong>
                                </div>
                            </div>

                            <div class="device-card-grid">
                                <el-card
                                    v-for="item in group.devices"
                                    :key="deviceKey(item)"
                                    class="device-card"
                                    shadow="never"
                                >
                                    <template #header>
                                        <div class="device-card-header">
                                            <div class="device-title">
                                                <span
                                                    v-if="item.type === 'ascend'"
                                                    class="status-dot"
                                                    :class="
                                                        item.health === 'OK'
                                                            ? 'status-ok'
                                                            : isAvailable(item.health)
                                                              ? 'status-error'
                                                              : ''
                                                    "
                                                ></span>
                                                <strong>{{ deviceTitle(item) }}</strong>
                                                <span
                                                    v-if="!commonDeviceModels[item.type]"
                                                    class="card-product-name"
                                                    :title="item.productName"
                                                >
                                                    · {{ item.productName }}
                                                </span>
                                            </div>
                                            <div class="device-actions">
                                                <el-button
                                                    type="primary"
                                                    link
                                                    size="small"
                                                    class="process-count"
                                                    @click.stop="openGPUProcesses(item)"
                                                >
                                                    {{ $t('aiTools.gpu.processCount') }}:
                                                    {{
                                                        item.processStatus === 'unavailable'
                                                            ? 'N/A'
                                                            : item.processes?.length || 0
                                                    }}
                                                </el-button>
                                                <el-button
                                                    type="primary"
                                                    link
                                                    @click="detailDeviceKey = deviceKey(item)"
                                                >
                                                    {{ $t('commons.button.view') }}
                                                </el-button>
                                                <el-button
                                                    type="primary"
                                                    link
                                                    :disabled="historyDisabled"
                                                    @click="openHistory(item)"
                                                >
                                                    {{ $t('aiTools.gpu.history') }}
                                                </el-button>
                                            </div>
                                        </div>
                                    </template>

                                    <div class="device-metrics">
                                        <div class="metric-item metric-primary">
                                            <span>
                                                {{ item.type === 'ascend' ? 'AICore(%)' : $t('aiTools.gpu.gpuUtil') }}
                                            </span>
                                            <strong>{{ deviceUtil(item) || 'N/A' }}</strong>
                                            <el-progress
                                                v-if="isAvailable(deviceUtil(item))"
                                                :percentage="percentage(deviceUtil(item))"
                                                :show-text="false"
                                                :stroke-width="2"
                                            />
                                        </div>
                                        <div class="metric-item metric-primary metric-memory">
                                            <span>
                                                {{
                                                    $t(
                                                        item.type !== 'ascend' && item.sharedMemory
                                                            ? 'aiTools.gpu.systemMemory'
                                                            : 'aiTools.gpu.memoryUsed',
                                                    )
                                                }}
                                                <el-tooltip
                                                    v-if="item.type !== 'ascend' && item.sharedMemory"
                                                    :content="$t('aiTools.gpu.systemMemoryHelper')"
                                                >
                                                    <el-icon><InfoFilled /></el-icon>
                                                </el-tooltip>
                                            </span>
                                            <strong>
                                                <span
                                                    v-for="(part, index) in formatMemory(
                                                        item.memUsed,
                                                        item.memTotal,
                                                    ).split(' / ')"
                                                    :key="index"
                                                    :class="{ 'memory-total': index > 0 }"
                                                >
                                                    {{ index ? ' / ' : '' }}{{ part }}
                                                </span>
                                            </strong>
                                            <el-progress
                                                v-if="memoryPercentage(item) !== null"
                                                :percentage="memoryPercentage(item) || 0"
                                                :show-text="false"
                                                :stroke-width="2"
                                            />
                                        </div>
                                        <div class="metric-item metric-temperature">
                                            <span>{{ $t('aiTools.gpu.temperature') }}</span>
                                            <strong>{{ formatTemperature(item.temperature) }}</strong>
                                        </div>
                                    </div>
                                    <div class="device-secondary-metrics">
                                        <div class="metric-item">
                                            <span>{{ $t('aiTools.gpu.powerUsage') }}</span>
                                            <strong :title="formatPower(item)">{{ item.powerDraw || 'N/A' }}</strong>
                                        </div>
                                        <div class="metric-item">
                                            <span>
                                                {{
                                                    item.type === 'ascend'
                                                        ? $t('aiTools.gpu.aiCPUUtil')
                                                        : $t('aiTools.gpu.frequency')
                                                }}
                                            </span>
                                            <strong>
                                                {{
                                                    (item.type === 'ascend' ? item.aiCPUUtil : item.frequency) || 'N/A'
                                                }}
                                            </strong>
                                        </div>
                                        <div class="metric-item">
                                            <span>
                                                {{
                                                    item.type === 'ascend'
                                                        ? $t('commons.table.status')
                                                        : $t('aiTools.gpu.fanSpeed')
                                                }}
                                            </span>
                                            <strong v-if="item.type === 'ascend'">{{ item.health || 'N/A' }}</strong>
                                            <strong v-else>
                                                {{
                                                    [item.fanSpeed, item.fanRPM].filter(isAvailable).join(' / ') ||
                                                    'N/A'
                                                }}
                                            </strong>
                                        </div>
                                    </div>
                                </el-card>
                            </div>
                        </section>
                        <section v-if="xpuInfo.xpu.length" class="gpu-group">
                            <div class="device-card-grid">
                                <el-card
                                    v-for="item in xpuInfo.xpu"
                                    :key="item.basic.deviceID"
                                    class="device-card"
                                    shadow="never"
                                >
                                    <template #header>
                                        <div class="device-card-header">
                                            <div class="device-title">
                                                <strong>XPU {{ item.basic.deviceID }}</strong>
                                                <span
                                                    v-if="!commonDeviceModels.intel && !commonDeviceModels.xpu"
                                                    class="card-product-name"
                                                    :title="item.basic.deviceName"
                                                >
                                                    · {{ item.basic.deviceName }}
                                                </span>
                                            </div>
                                            <div class="device-actions">
                                                <el-button
                                                    type="primary"
                                                    size="small"
                                                    link
                                                    class="process-count"
                                                    @click.stop="openXPUProcesses(item.basic.deviceID)"
                                                >
                                                    {{ $t('aiTools.gpu.processCount') }}:
                                                    {{
                                                        item.processStatus === 'unavailable'
                                                            ? 'N/A'
                                                            : item.processes?.length || 0
                                                    }}
                                                </el-button>
                                                <el-button
                                                    type="primary"
                                                    link
                                                    @click="detailDeviceKey = `xpu-${item.basic.deviceID}`"
                                                >
                                                    {{ $t('commons.button.view') }}
                                                </el-button>
                                                <el-button
                                                    type="primary"
                                                    link
                                                    :disabled="historyDisabled"
                                                    @click="openHistory(item)"
                                                >
                                                    {{ $t('aiTools.gpu.history') }}
                                                </el-button>
                                            </div>
                                        </div>
                                    </template>

                                    <div class="device-metrics">
                                        <div class="metric-item metric-primary">
                                            <span>{{ $t('aiTools.gpu.gpuUtil') }}</span>
                                            <strong>{{ item.stats.gpuUtil || 'N/A' }}</strong>
                                            <el-progress
                                                v-if="isAvailable(item.stats.gpuUtil)"
                                                :percentage="percentage(item.stats.gpuUtil)"
                                                :show-text="false"
                                                :stroke-width="2"
                                            />
                                        </div>
                                        <div class="metric-item metric-primary metric-memory">
                                            <span>{{ $t('aiTools.gpu.memoryUsed') }}</span>
                                            <strong>
                                                <span
                                                    v-for="(part, index) in formatMemory(
                                                        item.stats.memoryUsed,
                                                        item.basic.memory,
                                                    ).split(' / ')"
                                                    :key="index"
                                                    :class="{ 'memory-total': index > 0 }"
                                                >
                                                    {{ index ? ' / ' : '' }}{{ part }}
                                                </span>
                                            </strong>
                                            <el-progress
                                                v-if="
                                                    memoryPercentageValues(item.stats.memoryUsed, item.basic.memory) !==
                                                    null
                                                "
                                                :percentage="
                                                    memoryPercentageValues(item.stats.memoryUsed, item.basic.memory) ||
                                                    0
                                                "
                                                :show-text="false"
                                                :stroke-width="2"
                                            />
                                        </div>
                                        <div class="metric-item metric-temperature">
                                            <span>{{ $t('aiTools.gpu.temperature') }}</span>
                                            <strong>{{ formatTemperature(item.stats.temperature) }}</strong>
                                        </div>
                                    </div>
                                    <div class="device-secondary-metrics">
                                        <div class="metric-item">
                                            <span>{{ $t('aiTools.gpu.powerUsage') }}</span>
                                            <strong>{{ item.stats.power || 'N/A' }}</strong>
                                        </div>
                                        <div class="metric-item">
                                            <span>{{ $t('aiTools.gpu.frequency') }}</span>
                                            <strong>{{ item.stats.frequency || 'N/A' }}</strong>
                                        </div>
                                        <div class="metric-item">
                                            <span>{{ $t('aiTools.gpu.memoryTemperature') }}</span>
                                            <strong>{{ formatTemperature(item.stats.memoryTemperature || '') }}</strong>
                                        </div>
                                    </div>
                                </el-card>
                            </div>
                        </section>
                    </div>
                </template>
            </LayoutContent>
        </div>
        <DrawerPro
            :model-value="Boolean(detailDeviceKey)"
            :header="$t('commons.button.view')"
            :resource="deviceDetail?.title || ''"
            size="large"
            @close="detailDeviceKey = ''"
        >
            <DeviceDetails
                v-if="deviceDetail?.sections.length"
                :key="detailDeviceKey"
                :sections="deviceDetail.sections"
            />
            <el-empty v-else :description="$t('commons.msg.noneData')" />
        </DrawerPro>
        <DialogPro v-model="environmentDialogVisible" :title="$t('aiTools.gpu.environment.title')" size="large">
            <div v-loading="loading">
                <el-alert
                    v-if="loadFailed"
                    class="common-prompt"
                    :title="$t('aiTools.gpu.environment.loadFailed')"
                    type="error"
                    :closable="false"
                    show-icon
                >
                    <el-button type="primary" link :disabled="loading" @click="refresh()">
                        {{ $t('aiTools.gpu.environment.recheck') }}
                    </el-button>
                </el-alert>
                <section v-else-if="showEnvironment" class="environment-status">
                    <div class="environment-row">
                        <div class="environment-identity">
                            <el-icon :size="22"><Monitor /></el-icon>
                            <div>
                                <strong>{{ $t('aiTools.gpu.environment.hostGPU') }}</strong>
                                <div class="environment-description">
                                    <template v-if="driverChecks.length">
                                        <span v-for="check in driverChecks" :key="check.vendor">
                                            {{ checkTitle(check) }} · {{ checkStatus(check) }}
                                            {{ check.version }}
                                        </span>
                                    </template>
                                    <span v-else>
                                        {{ hasAccelerators ? deviceCountText : emptyDescription }}
                                    </span>
                                </div>
                            </div>
                        </div>
                        <el-tag :type="hostEnvironmentStatus.type">
                            {{ $t('aiTools.gpu.environment.' + hostEnvironmentStatus.label) }}
                        </el-tag>
                    </div>
                    <div v-if="containerChecks.length" class="environment-row">
                        <div class="environment-identity">
                            <el-icon :size="22"><Box /></el-icon>
                            <div>
                                <strong>NVIDIA · {{ $t('aiTools.gpu.environment.containerGPU') }}</strong>
                                <div class="environment-description">
                                    <span v-for="check in containerChecks" :key="check.item">
                                        {{ checkTitle(check) }} · {{ checkStatus(check) }}
                                    </span>
                                </div>
                            </div>
                        </div>
                        <el-tag :type="containerEnvironmentStatus.type">
                            {{ $t('aiTools.gpu.environment.' + containerEnvironmentStatus.label) }}
                        </el-tag>
                    </div>
                    <div v-if="environmentGuidance" class="environment-guidance">
                        <div>
                            <strong>
                                {{ $t('aiTools.gpu.environment.' + environmentGuidance.title) }}
                            </strong>
                            <p>
                                <template v-if="environmentGuidance.item !== 'driver' && hasAccelerators">
                                    {{ $t('aiTools.gpu.environment.monitorUnaffected') }}
                                </template>
                                {{ $t('aiTools.gpu.environment.' + environmentGuidance.description) }}
                            </p>
                        </div>
                        <el-button
                            v-if="environmentGuidance.action === 'docker'"
                            type="primary"
                            @click="
                                environmentDialogVisible = false;
                                routerToName('ContainerSetting');
                            "
                        >
                            {{ $t('aiTools.gpu.environment.goContainerSettings') }}
                        </el-button>
                        <el-button
                            v-else-if="environmentGuidance.action === 'details'"
                            type="primary"
                            @click="environmentDetails = ['environment']"
                        >
                            {{ $t('aiTools.gpu.environment.details') }}
                        </el-button>
                        <el-button
                            v-else
                            tag="a"
                            type="primary"
                            href="https://docs.fit2cloud.com/1panel/faq/gpu-driver"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            {{ $t('aiTools.gpu.environment.' + environmentGuidance.action) }}
                        </el-button>
                    </div>
                </section>
                <el-empty
                    v-if="!loadFailed && !showEnvironment && !environmentChecks.length && !gpuInfo.warnings?.length"
                    :description="$t(`aiTools.gpu.environment.${environmentStatus}`, { count: environmentIssueCount })"
                />
                <el-collapse
                    v-if="!loadFailed && (environmentChecks.length || gpuInfo.warnings?.length)"
                    v-model="environmentDetails"
                    class="environment-details"
                >
                    <el-collapse-item :title="$t('aiTools.gpu.environment.details')" name="environment">
                        <div v-for="check in environmentChecks" :key="`${check.vendor}-${check.item}`" class="mb-2">
                            <div class="flex flex-wrap items-center gap-2">
                                <span>{{ checkTitle(check) }}</span>
                                <el-tag :type="checkTagType(check)" size="small">
                                    {{ checkStatus(check) }}
                                </el-tag>
                                <span v-if="check.version">{{ check.version }}</span>
                            </div>
                            <p v-if="check.status !== 'passed'" class="environment-check-description">
                                {{ checkDescription(check) }}
                            </p>
                            <pre v-if="check.reason" class="whitespace-pre-wrap break-words mt-1">{{
                                check.reason
                            }}</pre>
                        </div>
                        <div v-for="warning in gpuInfo.warnings || []" :key="warning" class="break-words">
                            {{ warning }}
                        </div>
                    </el-collapse-item>
                </el-collapse>
                <div v-if="showEnvironment && !loadFailed && gpuInfo.environment?.checkedAt" class="input-help mb-4">
                    {{ $t('aiTools.gpu.environment.checkedAt') }}:
                    {{ new Date(gpuInfo.environment.checkedAt).toLocaleString() }}
                </div>
            </div>
            <template #footer>
                <el-button @click="environmentDialogVisible = false">{{ $t('commons.button.close') }}</el-button>
                <el-button type="primary" :loading="loading" @click="refresh()">
                    {{ $t('aiTools.gpu.environment.recheck') }}
                </el-button>
            </template>
        </DialogPro>
        <DialogPro v-model="processDrawerVisible" :title="processDrawerTitle" size="large">
            <template v-if="processGPU">
                <el-alert
                    v-if="processGPU.processStatus === 'unavailable'"
                    :title="$t('aiTools.gpu.processUnavailable')"
                    type="warning"
                    :closable="false"
                />
                <el-table v-if="processGPU.processes?.length" :data="processGPU.processes">
                    <el-table-column label="PID" prop="pid" />
                    <el-table-column :label="$t('aiTools.gpu.processName')" prop="processName" />
                    <el-table-column :label="$t('aiTools.gpu.processMemoryUsage')" prop="usedMemory" />
                </el-table>
                <el-empty
                    v-else-if="processGPU.processStatus !== 'unavailable'"
                    :description="$t('commons.msg.noneData')"
                />
            </template>
            <template v-else-if="processXPU">
                <el-alert
                    v-if="processXPU.processStatus === 'unavailable'"
                    :title="$t('aiTools.gpu.processUnavailable')"
                    type="warning"
                    :closable="false"
                />
                <el-table v-if="processXPU.processes?.length" :data="processXPU.processes">
                    <el-table-column label="PID" prop="pid" />
                    <el-table-column :label="$t('aiTools.gpu.processName')" prop="command" />
                    <el-table-column :label="$t('aiTools.gpu.shr')" prop="shr" />
                    <el-table-column :label="$t('aiTools.gpu.processMemoryUsage')" prop="memory" />
                </el-table>
                <el-empty
                    v-else-if="processXPU.processStatus !== 'unavailable'"
                    :description="$t('commons.msg.noneData')"
                />
            </template>
        </DialogPro>
    </div>
</template>

<script lang="ts" setup>
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Box, InfoFilled, Monitor } from '@element-plus/icons-vue';
import { loadGPUInfo, loadModelGPUDevices } from '@/api/modules/ai';
import RouterMenu from '@/views/ai/gpu/index.vue';
import GPUTopology from '@/views/ai/components/gpu-topology/index.vue';
import { AI } from '@/api/interface/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { routerToName } from '@/utils/router';
import i18n from '@/lang';
import { computeSizeForDocker } from '@/utils/size';
import DeviceDetails, { type DetailSection } from './components/DeviceDetails.vue';

const router = useRouter();
const route = useRoute();
const pagePath = route.path;
const { currentNode } = useGlobalStore();
const loading = ref(false);
const loadFailed = ref(false);
const topology = ref<{ devices: AI.ModelGPUDevice[]; numaNodes: AI.ModelNUMANode[]; warnings: string[] }>({
    devices: [],
    numaNodes: [],
    warnings: [],
});
const topologyLoading = ref(false);
const topologyFailed = ref(false);
let topologyRequest = 0;
const loadTopology = async () => {
    if (route.path !== pagePath) return;
    const version = ++topologyRequest;
    const node = currentNode.value;
    topologyLoading.value = true;
    topologyFailed.value = false;
    try {
        const { data } = await loadModelGPUDevices(node);
        if (version === topologyRequest && node === currentNode.value) topology.value = data;
    } catch {
        if (version === topologyRequest) topologyFailed.value = true;
    } finally {
        if (version === topologyRequest) topologyLoading.value = false;
    }
};
const environmentDetails = ref<string[]>(['environment']);
const environmentDialogVisible = ref(false);
const processDrawerVisible = ref(false);
const processDeviceKey = ref('');
const detailDeviceKey = ref('');
const processXPUId = ref<number | null>(null);
const emptyGPUInfo = (): AI.Info => ({
    cudaVersion: '',
    driverVersion: '',
    type: 'nvidia',
    gpu: [],
    npu: [],
    xpuDriverVersion: '',
    xpu: [],
});
const gpuInfo = ref<AI.Info>(emptyGPUInfo());
const xpuInfo = ref<AI.XpuInfo>({
    driverVersion: '',
    type: 'xpu',
    xpu: [],
});

type AcceleratorDevice = AI.GPU | AI.NPU;

interface GPUGroup {
    key: string;
    title: string;
    devices: AcceleratorDevice[];
}

const deviceKey = (item: AcceleratorDevice) => {
    return item.type === 'ascend'
        ? `ascend-${item.busID || item.npuIndex}-${item.chipIndex}`
        : `${item.type}-${item.uuid || item.busID || item.index}`;
};

const groupKey = (item: AcceleratorDevice) => {
    return item.type === 'ascend' ? `ascend-npu-${item.npuIndex}` : `${item.type}-devices`;
};

const gpuGroups = computed<GPUGroup[]>(() => {
    const groups = new Map<string, GPUGroup>();
    for (const item of [...gpuInfo.value.gpu, ...gpuInfo.value.npu]) {
        const key = groupKey(item);
        if (!groups.has(key)) {
            groups.set(key, {
                key,
                title: item.type === 'ascend' ? `NPU ${item.npuIndex}` : '',
                devices: [],
            });
        }
        const group = groups.get(key)!;
        group.devices.push(item);
    }
    return Array.from(groups.values());
});

const deviceDetail = computed(() => {
    const device = [...gpuInfo.value.gpu, ...gpuInfo.value.npu].find(
        (item) => deviceKey(item) === detailDeviceKey.value,
    );
    if (device) {
        const title =
            device.type === 'ascend' ? `NPU ${device.npuIndex} · ${deviceTitle(device)}` : deviceTitle(device);
        return { title: `${title} · ${device.productName}`, sections: deviceDetailSections(device) };
    }
    const xpu = xpuInfo.value.xpu.find((item) => `xpu-${item.basic.deviceID}` === detailDeviceKey.value);
    if (xpu) {
        return {
            title: `XPU ${xpu.basic.deviceID} · ${xpu.basic.deviceName}`,
            sections: xpuDetailSections(xpu),
        };
    }
    return null;
});

const processGPU = computed(() => {
    return [...gpuInfo.value.gpu, ...gpuInfo.value.npu].find((item) => deviceKey(item) === processDeviceKey.value);
});

const processXPU = computed(() => {
    return xpuInfo.value.xpu.find((item) => item.basic.deviceID === processXPUId.value);
});

const processDrawerTitle = computed(() => {
    if (processGPU.value) {
        const item = processGPU.value;
        return `${deviceTitle(item)} · ${item.productName} · ${i18n.global.t('aiTools.gpu.process')}`;
    }
    if (processXPU.value) {
        const item = processXPU.value;
        return `XPU ${item.basic.deviceID} · ${item.basic.deviceName} · ${i18n.global.t('aiTools.gpu.process')}`;
    }
    return i18n.global.t('aiTools.gpu.process');
});

const deviceCountText = computed(() => {
    const parts: string[] = [];
    if (gpuInfo.value.gpu.length) {
        parts.push(`${gpuInfo.value.gpu.length} GPU`);
    }
    if (gpuInfo.value.npu.length) {
        const npuCount = new Set(gpuInfo.value.npu.map((item) => item.npuIndex)).size;
        parts.push(`${npuCount} NPU`, `${gpuInfo.value.npu.length} Chip`);
    }
    if (xpuInfo.value.xpu.length) {
        parts.push(`${xpuInfo.value.xpu.length} XPU`);
    }
    return parts.join(' · ');
});

const hasAccelerators = computed(() => {
    return gpuGroups.value.length > 0 || xpuInfo.value.xpu.length > 0;
});

const vendorTitle = (vendor: string): string => {
    const titles: Record<string, string> = {
        nvidia: 'NVIDIA',
        intel: 'Intel',
        xpu: 'Intel',
        ascend: 'Huawei Ascend',
        amd: 'AMD',
    };
    if (vendor === 'mixed') {
        const vendors = new Set(gpuInfo.value.gpu.map((device) => device.type as string));
        if (gpuInfo.value.npu.length) vendors.add('ascend');
        if (xpuInfo.value.xpu.length) vendors.add('intel');
        return Array.from(vendors, vendorTitle).join(' / ') || '—';
    }
    return titles[vendor] || vendor || '—';
};

const displayValue = (value: string | undefined) => {
    return isAvailable(value) ? value : '—';
};

const formatSummaryMemory = (value: number | null) => {
    return value === null || !Number.isFinite(value) ? '—' : computeSizeForDocker(value);
};

const summaryMemoryUsage = (vendor?: string) => {
    const devices = [
        ...[...gpuInfo.value.gpu, ...gpuInfo.value.npu]
            .filter(
                (device) => (!vendor || device.type === vendor) && (device.type === 'ascend' || !device.sharedMemory),
            )
            .map((device) => ({ id: deviceKey(device), used: device.memUsed, total: device.memTotal })),
        ...(!vendor || vendor === 'intel' || vendor === 'xpu' ? xpuInfo.value.xpu : []).map((device) => ({
            id: `xpu-${device.basic.deviceID}`,
            used: device.stats.memoryUsed,
            total: device.basic.memory,
        })),
    ];
    if (!devices.length) return null;
    let used: number | null = 0;
    let total: number | null = 0;
    const seen = new Set<string>();
    for (const device of devices) {
        if (seen.has(device.id)) continue;
        seen.add(device.id);
        const deviceUsed = memoryMiB(device.used || '');
        const deviceTotal = memoryMiB(device.total || '');
        used = used !== null && deviceUsed !== null && deviceUsed >= 0 ? used + deviceUsed * 1024 ** 2 : null;
        total = total !== null && deviceTotal !== null && deviceTotal >= 0 ? total + deviceTotal * 1024 ** 2 : null;
    }
    return {
        used,
        total,
        percentage:
            used !== null && total !== null && total > 0
                ? Math.min(100, Number(((used / total) * 100).toFixed(1)))
                : null,
    };
};

const overviewRows = computed(() => {
    if (gpuInfo.value.summaries?.length) {
        return gpuInfo.value.summaries
            .filter(
                (summary) =>
                    summary.cardCount > 0 ||
                    isAvailable(summary.driverVersion) ||
                    summary.memoryTotal !== null ||
                    summary.sharedMemoryTotal !== null,
            )
            .map((summary) => {
                const devices = gpuInfo.value.gpu.filter((device) => device.type === summary.vendor);
                const sharedMemoryTotal =
                    devices.length && !devices.some((device) => device.sharedMemory) ? null : summary.sharedMemoryTotal;
                return {
                    ...summary,
                    memoryUsage: summaryMemoryUsage(summary.vendor),
                    sharedMemoryTotal,
                    cardCountText: summary.cardCount === null ? '—' : String(summary.cardCount),
                    sharedMemoryOnly:
                        summary.memoryTotal === null &&
                        sharedMemoryTotal !== null &&
                        devices.every((device) => device.sharedMemory),
                };
            });
    }
    if (!hasAccelerators.value) return [];
    const memoryDevices = [
        ...[...gpuInfo.value.gpu, ...gpuInfo.value.npu].map((device) => ({
            id: deviceKey(device),
            total: device.memTotal,
            shared: device.type !== 'ascend' && Boolean(device.sharedMemory),
        })),
        ...xpuInfo.value.xpu.map((device) => ({
            id: `xpu-${device.basic.deviceID}`,
            total: device.basic.memory,
            shared: false,
        })),
    ];
    let memoryTotal: number | null = null;
    let sharedMemoryTotal: number | null = null;
    let memoryPartial = false;
    const seen = new Set<string>();
    for (const device of memoryDevices) {
        if (seen.has(device.id)) continue;
        seen.add(device.id);
        const total = memoryMiB(device.total || '');
        if (total === null || !Number.isFinite(total) || total < 0) {
            memoryPartial = true;
            continue;
        }
        const bytes = total * 1024 * 1024;
        if (device.shared) {
            sharedMemoryTotal = Math.max(sharedMemoryTotal ?? 0, bytes);
        } else {
            memoryTotal = (memoryTotal ?? 0) + bytes;
        }
    }
    return [
        {
            vendor: gpuInfo.value.type,
            driverVersion: gpuInfo.value.driverVersion,
            cudaVersion: gpuInfo.value.cudaVersion,
            cardCountText: deviceCountText.value,
            memoryTotal,
            memoryUsage: summaryMemoryUsage(),
            sharedMemoryTotal,
            sharedMemoryOnly: sharedMemoryTotal !== null && memoryDevices.every((device) => device.shared),
            memoryPartial,
        },
    ];
});

const commonDeviceModels = computed<Record<string, string>>(() => {
    const models: Record<string, string> = {};
    for (const summary of overviewRows.value) {
        const names =
            summary.vendor === 'intel' || summary.vendor === 'xpu'
                ? xpuInfo.value.xpu.map((device) => device.basic.deviceName)
                : [...gpuInfo.value.gpu, ...gpuInfo.value.npu]
                      .filter((device) => device.type === summary.vendor)
                      .map((device) => device.productName);
        if (names.length && names[0] && names.every((name) => name === names[0])) {
            models[summary.vendor] = names[0];
        }
    }
    return models;
});

const environmentChecks = computed(() =>
    (gpuInfo.value.environment?.checks || []).filter((check) => check.item === 'driver' || check.vendor === 'nvidia'),
);
const environmentIssueCount = computed(
    () =>
        environmentChecks.value.filter((check) => check.status !== 'passed').length +
        (gpuInfo.value.warnings?.length || 0),
);
const environmentPassed = computed(
    () => !loadFailed.value && environmentChecks.value.length > 0 && environmentIssueCount.value === 0,
);
const environmentStatus = computed(() => {
    if (loadFailed.value) return 'loadFailed';
    if (environmentPassed.value) return 'passed';
    if (environmentIssueCount.value) return 'issues';
    return gpuInfo.value.environment?.hardwareStatus === 'not_detected' ? 'notDetected' : 'unknown';
});
const historyDisabled = computed(
    () =>
        (loading.value && !gpuInfo.value.environment && !hasAccelerators.value) ||
        environmentChecks.value.some(
            (check) => check.item === 'driver' && check.status !== 'passed' && (check.vendor || !hasAccelerators.value),
        ),
);
const driverChecks = computed(() => environmentChecks.value.filter((check) => check.item === 'driver'));
const containerChecks = computed(() => environmentChecks.value.filter((check) => check.item !== 'driver'));
const showEnvironment = computed(
    () =>
        gpuInfo.value.environment &&
        (gpuInfo.value.environment.hardwareStatus !== 'not_detected' || environmentChecks.value.length > 0),
);
const hostEnvironmentStatus = computed<{ type: 'success' | 'warning' | 'info'; label: string }>(() => {
    const hasDriverIssue = driverChecks.value.some((check) => check.status !== 'passed');
    if (hasAccelerators.value && (!hasDriverIssue || driverChecks.value.some((check) => check.status === 'passed'))) {
        return hasDriverIssue
            ? { type: 'warning', label: 'monitorPartial' }
            : { type: 'success', label: 'monitorAvailable' };
    }
    if (driverChecks.value.some((check) => ['missing', 'unconfigured', 'failed'].includes(check.status))) {
        return { type: 'warning', label: 'notReady' };
    }
    return { type: 'info', label: 'statusUnknown' };
});
const containerEnvironmentStatus = computed<{ type: 'success' | 'warning' | 'info'; label: string }>(() => {
    const nvidiaDriver = driverChecks.value.find((check) => check.vendor === 'nvidia');
    if (nvidiaDriver && nvidiaDriver.status !== 'passed') {
        return nvidiaDriver.status === 'unknown'
            ? { type: 'info', label: 'statusUnknown' }
            : { type: 'warning', label: 'notReady' };
    }
    if (containerChecks.value.every((check) => check.status === 'passed')) {
        return { type: 'success', label: 'ready' };
    }
    if (containerChecks.value.some((check) => ['missing', 'unconfigured', 'failed'].includes(check.status))) {
        return { type: 'warning', label: 'notReady' };
    }
    return { type: 'info', label: 'statusUnknown' };
});
const environmentGuidance = computed(() => {
    const issue =
        driverChecks.value.find((check) => check.status !== 'passed') ||
        containerChecks.value.find((check) => check.item === 'toolkit' && check.status !== 'passed') ||
        containerChecks.value.find((check) => check.status !== 'passed');
    if (!issue) return null;
    if (issue.status === 'unknown' || (issue.item === 'driver' && !issue.vendor)) {
        return { item: issue.item, title: 'unknown', description: 'checkDetailsHelp', action: 'details' };
    }
    if (issue.item === 'driver') {
        return { item: issue.item, title: 'driverActionTitle', description: 'driverActionHelp', action: 'driverGuide' };
    }
    if (issue.item === 'toolkit') {
        return {
            item: issue.item,
            title: 'toolkitActionTitle',
            description: 'toolkitActionHelp',
            action: 'toolkitGuide',
        };
    }
    return { item: issue.item, title: 'dockerActionTitle', description: 'dockerActionHelp', action: 'docker' };
});
const emptyDescription = computed(() => {
    if (!gpuInfo.value.environment) return i18n.global.t('aiTools.gpu.gpuHelper');
    return gpuInfo.value.environment.hardwareStatus === 'not_detected'
        ? i18n.global.t('aiTools.gpu.environment.notDetected')
        : i18n.global.t('aiTools.gpu.environment.unknown');
});
const checkStatus = (check: AI.GPUEnvironmentCheck) => {
    const status = check.status[0].toUpperCase() + check.status.slice(1);
    return i18n.global.t(`aiTools.gpu.environment.status${status}`);
};
const checkTagType = (check: AI.GPUEnvironmentCheck) => {
    if (check.status === 'passed') return 'success';
    return check.status === 'failed' ? 'danger' : 'warning';
};
const checkTitle = (check: AI.GPUEnvironmentCheck) => {
    if (check.vendor === 'nvidia' && check.item === 'driver') {
        return i18n.global.t('aiTools.gpu.environment.nvidiaDriver');
    }
    const title = i18n.global.t(`aiTools.gpu.environment.${check.item}`);
    return check.vendor && check.item === 'driver' ? `${vendorTitle(check.vendor)} · ${title}` : title;
};
const checkDescription = (check: AI.GPUEnvironmentCheck) => {
    const managementTools: Record<string, string> = {
        nvidia: 'nvidia-smi',
        intel: 'xpu-smi',
        ascend: 'npu-smi',
        amd: 'amd-smi',
    };
    const tool = managementTools[check.vendor];
    let key: string;
    if (check.item === 'driver') {
        if (!tool) return i18n.global.t('aiTools.gpu.environment.unknown');
        key =
            check.status === 'missing' || check.status === 'unconfigured'
                ? 'driverMissing'
                : check.status === 'failed'
                  ? 'driverFailed'
                  : 'driverUnknown';
    } else if (check.item === 'toolkit') {
        key = check.status === 'missing' || check.status === 'unconfigured' ? 'toolkitMissing' : 'toolkitFailed';
    } else {
        const configKeys = [
            'dockerConfigMissing',
            'dockerConfigInvalid',
            'dockerConfigPending',
            'dockerConfigUnreadable',
        ];
        const code = check.code || '';
        key = configKeys.includes(code)
            ? code
            : check.status === 'missing' || check.status === 'unconfigured'
              ? 'dockerUnconfigured'
              : 'dockerFailed';
    }
    const description = i18n.global.t(`aiTools.gpu.environment.${key}`, {
        tool: tool || '',
        vendor: vendorTitle(check.vendor),
    });
    return check.item === 'driver' && key !== 'driverMissing' ? `${checkTitle(check)}: ${description}` : description;
};

const summaryMemoryTooltip = (summary: { sharedMemoryOnly: boolean; memoryPartial: boolean }) =>
    [
        summary.sharedMemoryOnly && i18n.global.t('aiTools.gpu.systemMemoryHelper'),
        summary.memoryPartial && i18n.global.t('aiTools.gpu.environment.partialMemory'),
    ]
        .filter(Boolean)
        .join(' ');

const normalizeAcceleratorInfo = (data: AI.Info): AI.Info => {
    const devices = (data.gpu || []) as AcceleratorDevice[];
    const legacyNPUs = devices
        .filter((item): item is AI.NPU => item.type === 'ascend')
        .map((item) => {
            const legacyItem = item as AI.NPU & { gpuUtil?: string; performanceState?: string };
            return {
                ...item,
                aiCore: item.aiCore || legacyItem.gpuUtil || '',
                health: item.health || legacyItem.performanceState || '',
            };
        });
    return {
        ...data,
        gpu: devices.filter((item): item is AI.GPU => item.type !== 'ascend'),
        npu: data.npu || legacyNPUs,
        xpuDriverVersion: data.xpuDriverVersion || (data.type === 'xpu' ? data.driverVersion : ''),
        xpu: data.xpu || [],
    };
};

const applyAcceleratorInfo = (data: AI.Info) => {
    const normalized = normalizeAcceleratorInfo(data);
    gpuInfo.value = normalized;
    xpuInfo.value = {
        type: 'xpu',
        driverVersion: normalized.xpuDriverVersion,
        xpu: normalized.xpu,
    };
};

let requestID = 0;
const search = async (refreshEnvironment = false) => {
    if (route.path !== pagePath) return;
    if (loading.value) return;
    const id = ++requestID;
    const node = currentNode.value;
    loading.value = true;
    loadFailed.value = false;
    try {
        const res = await loadGPUInfo(refreshEnvironment, node);
        if (id !== requestID || node !== currentNode.value) return;
        applyAcceleratorInfo(res.data);
    } catch {
        if (id === requestID && node === currentNode.value) {
            loadFailed.value = true;
        }
    } finally {
        if (id === requestID) loading.value = false;
    }
};

const refresh = () => {
    search(true);
    loadTopology();
};

watch(currentNode, () => {
    topologyRequest++;
    topology.value = { devices: [], numaNodes: [], warnings: [] };
    topologyLoading.value = false;
    topologyFailed.value = false;
    loadTopology();
    environmentDialogVisible.value = false;
    environmentDetails.value = ['environment'];
    requestID++;
    loading.value = false;
    loadFailed.value = false;
    applyAcceleratorInfo(emptyGPUInfo());
    detailDeviceKey.value = '';
    processDrawerVisible.value = false;
    search(true);
});

const visibleSections = (sections: DetailSection[]) => {
    return sections
        .map((section) => ({
            ...section,
            items: section.items.filter((item) => isAvailable(item.value)),
        }))
        .filter((section) => section.items.length);
};

const deviceDetailSections = (item: AcceleratorDevice): DetailSection[] => {
    const t = i18n.global.t;
    if (item.type === 'ascend') {
        return visibleSections([
            {
                id: 'runtime',
                title: t('aiTools.gpu.runtimeInfo'),
                items: [
                    { label: t('aiTools.gpu.powerUsage'), value: item.powerDraw },
                    { label: t('commons.table.status'), value: item.health },
                ],
            },
            {
                id: 'memory',
                title: t('aiTools.gpu.memory'),
                items: [
                    {
                        label: t('aiTools.gpu.ddrUsage'),
                        value: hasUsage(item.memoryUsed, item.memoryTotal)
                            ? formatMemory(item.memoryUsed || 'N/A', item.memoryTotal || 'N/A')
                            : '',
                    },
                    {
                        label: t('aiTools.gpu.hbmUsage'),
                        value: hasUsage(item.hbmUsed, item.hbmTotal)
                            ? formatMemory(item.hbmUsed || 'N/A', item.hbmTotal || 'N/A')
                            : '',
                    },
                    {
                        label: t('aiTools.gpu.hugepagesUsage'),
                        value: hasUsage(item.hugepagesUsed, item.hugepagesTotal)
                            ? formatUsage(item.hugepagesUsed, item.hugepagesTotal)
                            : '',
                    },
                    { label: t('aiTools.gpu.ddrBandwidth'), value: item.ddrBandwidth },
                    { label: t('aiTools.gpu.hbmBandwidth'), value: item.hbmBandwidth },
                ],
            },
            {
                id: 'engines',
                title: t('aiTools.gpu.cpuUtil'),
                items: [
                    { label: t('aiTools.gpu.aiCPUUtil'), value: item.aiCPUUtil },
                    { label: t('aiTools.gpu.ctrlCPUUtil'), value: item.ctrlCPUUtil },
                ],
            },
            {
                id: 'device',
                title: t('aiTools.gpu.deviceInfo'),
                items: [{ label: t('aiTools.gpu.busID'), value: item.busID }],
            },
        ]);
    }
    return visibleSections([
        {
            id: 'runtime',
            title: t('aiTools.gpu.runtimeInfo'),
            items: [
                { label: t('aiTools.gpu.powerUsage'), value: formatPower(item) },
                { label: t('aiTools.gpu.frequency'), value: item.frequency },
                { label: t('aiTools.gpu.memoryFrequency'), value: item.memoryFrequency },
                { label: t('aiTools.gpu.mediaFrequency'), value: item.mediaFrequency },
                { label: t('aiTools.gpu.fanSpeed') + ' (%)', value: item.fanSpeed },
                { label: t('aiTools.gpu.fanRPM'), value: item.fanRPM },
                { label: t('aiTools.gpu.hotspotTemperature'), value: item.hotspotTemperature },
                { label: t('aiTools.gpu.memoryTemperature'), value: item.memoryTemperature },
                {
                    label: t('aiTools.gpu.performanceState'),
                    value: item.performanceState,
                    help: t('aiTools.gpu.performanceStateHelper'),
                },
                { label: t('aiTools.gpu.clockEvents'), value: item.clockEvents?.join(', ') },
            ],
        },
        {
            id: 'memory',
            title: t(item.sharedMemory ? 'aiTools.gpu.systemMemory' : 'aiTools.gpu.memory'),
            items: [
                { label: t('aiTools.gpu.memoryActivity'), value: item.memoryActivity },
                {
                    label: t(item.sharedMemory ? 'aiTools.gpu.systemMemoryAvailable' : 'aiTools.gpu.freeMemory'),
                    value: item.memoryFree,
                    help: item.sharedMemory ? t('aiTools.gpu.systemMemoryHelper') : undefined,
                },
                { label: t('aiTools.gpu.memoryReserved'), value: item.memoryReserved },
            ],
        },
        {
            id: 'engines',
            title: t('aiTools.gpu.engineUtil'),
            items: [
                { label: t('aiTools.gpu.encoderUtil'), value: item.encoderUtil },
                { label: t('aiTools.gpu.decoderUtil'), value: item.decoderUtil },
                { label: t('aiTools.gpu.jpegUtil'), value: item.jpegUtil },
                { label: t('aiTools.gpu.ofaUtil'), value: item.ofaUtil },
                { label: t('aiTools.gpu.mediaUtil'), value: item.mediaUtil },
            ],
        },
        {
            id: 'ecc',
            title: 'ECC',
            items: [
                { label: 'ECC', value: isAvailable(item.ecc) ? loadEcc(item.ecc) : '', help: t('aiTools.gpu.ecc') },
                {
                    label: t('aiTools.gpu.eccPending'),
                    value: isAvailable(item.eccPending) ? loadEcc(item.eccPending!) : '',
                },
                ...(item.eccErrors || []).flatMap((error) => [
                    { label: `${error.scope} · ${t('aiTools.gpu.eccCorrectable')}`, value: error.correctable },
                    { label: `${error.scope} · ${t('aiTools.gpu.eccUncorrectable')}`, value: error.uncorrectable },
                ]),
            ],
        },
        {
            id: 'device',
            title: t('aiTools.gpu.deviceInfo'),
            items: [
                { label: 'UUID', value: item.uuid, ellipsis: true },
                { label: t('aiTools.gpu.busID'), value: item.busID },
                { label: t('aiTools.gpu.driverVersion'), value: item.driverVersion },
                { label: t('aiTools.gpu.architecture'), value: item.architecture },
                { label: t('aiTools.gpu.powerDefaultLimit'), value: item.defaultPowerLimit },
                { label: t('aiTools.gpu.powerMaxLimit'), value: item.maxPowerLimit },
                {
                    label: t('aiTools.gpu.pcieGeneration'),
                    value: isAvailable(item.pcieGeneration)
                        ? `${item.pcieGeneration} / ${item.pcieMaxGeneration || 'N/A'}`
                        : '',
                },
                {
                    label: t('aiTools.gpu.pcieWidth'),
                    value: isAvailable(item.pcieWidth) ? `${item.pcieWidth} / ${item.pcieMaxWidth || 'N/A'}` : '',
                },
                {
                    label: t('aiTools.gpu.persistenceMode'),
                    value: isAvailable(item.persistenceMode) ? loadEcc(item.persistenceMode) : '',
                    help: t('aiTools.gpu.persistenceModeHelper'),
                },
                {
                    label: t('aiTools.gpu.displayActive'),
                    value: isAvailable(item.displayActive)
                        ? t(
                              'aiTools.gpu.' +
                                  (item.displayActive.toLowerCase() === 'disabled'
                                      ? 'displayActiveF'
                                      : 'displayActiveT'),
                          )
                        : '',
                },
                {
                    label: t('aiTools.gpu.computeMode'),
                    value: isAvailable(item.computeMode) ? loadComputeMode(item.computeMode) : '',
                    help: ['defaultHelper', 'exclusiveProcessHelper', 'exclusiveThreadHelper', 'prohibitedHelper']
                        .map((key) => t('aiTools.gpu.' + key))
                        .join('\n'),
                },
                {
                    label: 'MIG',
                    value: isAvailable(item.migMode) ? loadEcc(item.migMode) : '',
                    help: t('aiTools.gpu.migModeHelper'),
                },
            ],
        },
    ]);
};

const xpuDetailSections = (item: AI.XpuInfo['xpu'][number]): DetailSection[] => {
    const t = i18n.global.t;
    return visibleSections([
        {
            id: 'runtime',
            title: t('aiTools.gpu.runtimeInfo'),
            items: [
                { label: t('aiTools.gpu.powerUsage'), value: item.stats.power },
                { label: t('aiTools.gpu.frequency'), value: item.stats.frequency },
                { label: t('aiTools.gpu.mediaFrequency'), value: item.stats.mediaFrequency },
                { label: t('aiTools.gpu.memoryTemperature'), value: item.stats.memoryTemperature },
            ],
        },
        {
            id: 'memory',
            title: t('aiTools.gpu.memory'),
            items: [
                { label: t('aiTools.gpu.freeMemory'), value: item.basic.freeMemory },
                { label: t('aiTools.gpu.memoryUsage'), value: item.stats.memoryUtil },
                { label: t('aiTools.gpu.memoryBandwidth'), value: item.stats.memoryBandwidthUtil },
            ],
        },
        {
            id: 'engines',
            title: t('aiTools.gpu.engineUtil'),
            items: [
                { label: t('aiTools.gpu.computeUtil'), value: item.stats.computeUtil },
                { label: t('aiTools.gpu.mediaUtil'), value: item.stats.mediaUtil },
                { label: t('aiTools.gpu.copyUtil'), value: item.stats.copyUtil },
            ],
        },
        ...(item.tiles || []).map((tile) => ({
            id: `tile-${tile.tileID}`,
            title: `Tile ${tile.tileID}`,
            items: [
                { label: t('aiTools.gpu.gpuUtil'), value: tile.stats.gpuUtil },
                { label: t('aiTools.gpu.powerCurrent'), value: tile.stats.power },
                { label: t('aiTools.gpu.frequency'), value: tile.stats.frequency },
                { label: t('aiTools.gpu.mediaFrequency'), value: tile.stats.mediaFrequency },
                { label: t('aiTools.gpu.temperature'), value: tile.stats.temperature },
                { label: t('aiTools.gpu.memoryTemperature'), value: tile.stats.memoryTemperature },
                { label: t('aiTools.gpu.memoryUsed'), value: tile.stats.memoryUsed },
                { label: t('aiTools.gpu.memoryUsage'), value: tile.stats.memoryUtil },
                { label: t('aiTools.gpu.memoryBandwidth'), value: tile.stats.memoryBandwidthUtil },
                { label: t('aiTools.gpu.computeUtil'), value: tile.stats.computeUtil },
                { label: t('aiTools.gpu.mediaUtil'), value: tile.stats.mediaUtil },
                { label: t('aiTools.gpu.copyUtil'), value: tile.stats.copyUtil },
            ],
        })),
        {
            id: 'device',
            title: t('aiTools.gpu.deviceInfo'),
            items: [
                { label: 'UUID', value: item.basic.uuid, ellipsis: true },
                { label: t('aiTools.gpu.busID'), value: item.basic.pciBdfAddress },
                { label: t('aiTools.gpu.driverVersion'), value: item.basic.driverVersion },
            ],
        },
    ]);
};

const openGPUProcesses = (item: AcceleratorDevice) => {
    processDeviceKey.value = deviceKey(item);
    processXPUId.value = null;
    processDrawerVisible.value = true;
};

const openXPUProcesses = (deviceID: number) => {
    processDeviceKey.value = '';
    processXPUId.value = deviceID;
    processDrawerVisible.value = true;
};

const stableDeviceID = (vendor: string, identifier: string | undefined, fallback: string) => {
    return `${vendor}:${identifier && identifier.toUpperCase() !== 'N/A' ? identifier : fallback}`;
};

const openHistory = (item: AcceleratorDevice | AI.XpuInfo['xpu'][number]) => {
    if (historyDisabled.value) return;
    let deviceID: string;
    if ('basic' in item) {
        deviceID = stableDeviceID(
            'xpu',
            item.basic.uuid,
            stableDeviceID('pci', item.basic.pciBdfAddress, String(item.basic.deviceID)),
        );
    } else if (item.type === 'ascend') {
        deviceID = `${stableDeviceID('ascend', item.busID, String(item.npuIndex))}:${item.chipIndex}`;
    } else {
        deviceID = stableDeviceID(item.type, item.uuid, stableDeviceID('pci', item.busID, String(item.index)));
    }
    void router.push({ path: '/ai/gpu/history', query: { deviceID } });
};

const deviceTitle = (item: AcceleratorDevice) => {
    return item.type === 'ascend' ? `Chip ${item.chipIndex}` : `GPU ${item.index}`;
};

const isAvailable = (value?: string) => {
    return Boolean(
        value &&
        value.trim() !== '' &&
        !['N/A', 'NA', 'NOT SUPPORTED', '[NOT SUPPORTED]', 'NAN', 'INF'].includes(value.trim().toUpperCase()),
    );
};

interface Quantity {
    value: number;
    unit: string;
}

const parseQuantity = (value: string): Quantity | null => {
    const matched = value.trim().match(/^([+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+))\s*(.*?)$/);
    if (!matched) {
        return null;
    }
    const parsed = Number.parseFloat(matched[1]);
    return Number.isFinite(parsed) ? { value: parsed, unit: matched[2].replace(/\s+/g, '').toLowerCase() } : null;
};

const percentage = (value: string) => {
    const parsed = parseQuantity(value);
    if (!parsed || !['', '%', 'percent', 'pct'].includes(parsed.unit)) {
        return 0;
    }
    return Math.min(100, Math.max(0, parsed.value));
};

const formatTemperature = (value: string) => {
    if (!isAvailable(value)) {
        return value || 'N/A';
    }
    return value.replace(/\s*°?C\b/, ' °C');
};

const memoryPercentage = (item: AcceleratorDevice): number | null => {
    return memoryPercentageValues(item.memUsed, item.memTotal);
};

const memoryPercentageValues = (usedValue: string, totalValue: string): number | null => {
    const used = memoryMiB(usedValue);
    const total = memoryMiB(totalValue);
    if (used === null || total === null || total <= 0) {
        return null;
    }
    return Math.min(100, Math.max(0, Number(((used / total) * 100).toFixed(1))));
};

const formatMemory = (usedValue: string, totalValue: string) => {
    const used = memoryMiB(usedValue);
    const total = memoryMiB(totalValue);
    if (used === null || total === null) {
        return `${usedValue} / ${totalValue}`;
    }
    if (total >= 1024) {
        return `${formatQuantity(used / 1024)} / ${formatQuantity(total / 1024)} GiB`;
    }
    return `${formatQuantity(used)} / ${formatQuantity(total)} MiB`;
};

const memoryMiB = (value: string): number | null => {
    const parsed = parseQuantity(value);
    if (!parsed) {
        return null;
    }
    const factors: Record<string, number> = {
        '': 1,
        b: 1 / (1024 * 1024),
        byte: 1 / (1024 * 1024),
        bytes: 1 / (1024 * 1024),
        kb: 1000 / (1024 * 1024),
        kib: 1 / 1024,
        mb: 1_000_000 / (1024 * 1024),
        mib: 1,
        gb: 1_000_000_000 / (1024 * 1024),
        gib: 1024,
        tb: 1_000_000_000_000 / (1024 * 1024),
        tib: 1024 * 1024,
    };
    const factor = factors[parsed.unit];
    return factor === undefined ? null : parsed.value * factor;
};

const formatQuantity = (value: number) => Number(value.toFixed(1)).toString();

const hasUsage = (usedValue?: string, totalValue?: string) => {
    return isAvailable(usedValue) || isAvailable(totalValue);
};

const formatUsage = (usedValue?: string, totalValue?: string) => {
    return `${usedValue || 'N/A'} / ${totalValue || 'N/A'}`;
};

const deviceUtil = (item: AcceleratorDevice) => {
    return item.type === 'ascend' ? item.aiCore : item.gpuUtil;
};

const formatPower = (item: AcceleratorDevice) => {
    if (item.type === 'ascend') {
        return item.powerDraw;
    }
    return isAvailable(item.powerLimit) ? `${item.powerDraw || 'N/A'} / ${item.powerLimit}` : item.powerDraw;
};

const loadComputeMode = (val: string) => {
    switch (val) {
        case 'Default':
            return i18n.global.t('aiTools.gpu.default');
        case 'Exclusive Process':
            return i18n.global.t('aiTools.gpu.exclusiveProcess');
        case 'Exclusive Thread':
            return i18n.global.t('aiTools.gpu.exclusiveThread');
        case 'Prohibited':
            return i18n.global.t('aiTools.gpu.prohibited');
    }
};

const loadEcc = (val: string) => {
    if (val === 'N/A') {
        return i18n.global.t('aiTools.gpu.migModeNA');
    }
    if (val === 'Disabled') {
        return i18n.global.t('aiTools.gpu.disabled');
    }
    if (val === 'Enabled') {
        return i18n.global.t('aiTools.gpu.enabled');
    }
    return val || 'N/A';
};

onMounted(() => {
    loadTopology();
    search();
});
onActivated(() => {
    if (!topologyLoading.value) loadTopology();
    search();
});
const invalidateRequest = () => {
    topologyRequest++;
    topologyLoading.value = false;
    requestID++;
    loading.value = false;
};
onDeactivated(invalidateRequest);
onBeforeUnmount(invalidateRequest);
</script>

<style lang="scss" scoped>
.monitor-heading {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px 16px;
}
.environment-detail-link {
    font-size: 12px;
    color: var(--el-text-color-regular);
}
.collection-time {
    align-self: center;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}
.environment-heading {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
}
.environment-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 16px;
    padding: 16px 0;
    border-top: 1px solid var(--el-border-color-lighter);

    > .el-tag {
        flex-shrink: 0;
    }
}
.environment-identity {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;

    > .el-icon {
        flex-shrink: 0;
        color: var(--el-text-color-secondary);
    }

    strong {
        color: var(--el-text-color-primary);
        font-weight: 600;
    }
}
.environment-description {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 16px;
    margin-top: 5px;
    color: var(--el-text-color-regular);
    font-size: 12px;
    overflow-wrap: anywhere;
}
.environment-guidance {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 16px;
    margin: 12px 0 16px;
    border-radius: 6px;
    background: var(--el-color-warning-light-9);

    strong {
        color: var(--el-color-warning-dark-2);
    }

    p {
        margin: 6px 0 0;
        color: var(--el-text-color-regular);
        font-size: 12px;
        line-height: 1.7;
    }

    > .el-button {
        flex-shrink: 0;
        max-width: 100%;
        height: auto;
        min-height: 32px;
        white-space: normal;
        text-align: center;
    }
}
.environment-details {
    margin-bottom: 12px;
}
.environment-check-description {
    margin: 6px 0;
    color: var(--el-text-color-regular);
    line-height: 1.7;
}
@media (max-width: 767px) {
    .environment-guidance {
        align-items: flex-start;
        flex-direction: column;
    }
}
.device-overview {
    display: flex;
    align-items: center;
    gap: 0;
    flex-wrap: wrap;
    min-height: 44px;
    padding: 12px 16px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
}
.overview-item {
    display: flex;
    align-items: center;
    gap: 8px;
    white-space: nowrap;
    & + .overview-item {
        margin-left: 16px;
        padding-left: 16px;
        border-left: 1px solid var(--el-border-color-lighter);
    }
    span {
        color: var(--el-text-color-secondary);
        font-size: 12px;
        line-height: 18px;
    }
    strong {
        color: var(--el-text-color-primary);
        font-size: 12px;
        font-weight: 400;
        line-height: 18px;
    }
}
.overview-model {
    max-width: 100%;
    white-space: normal;
    strong {
        overflow-wrap: anywhere;
        font-size: 14px;
        font-weight: 500;
    }
}
.gpu-group-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-top: 16px;
}
.gpu-group {
    padding-top: 2px;
}
.gpu-group + .gpu-group {
    padding-top: 18px;
    border-top: 1px solid var(--el-border-color-lighter);
}
.group-header,
.device-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
}
.group-header {
    justify-content: flex-start;
    gap: 8px;
    margin-bottom: 14px;
}
.device-card-header {
    width: 100%;
    min-width: 0;
}
.group-identity,
.device-title {
    display: flex;
    align-items: center;
    gap: 8px;
}
.group-identity strong {
    color: var(--el-text-color-regular);
    font-size: 14px;
    font-weight: 600;
    line-height: 20px;
}
.card-product-name {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}
.status-dot {
    width: 8px;
    height: 8px;
    flex: 0 0 8px;
    border-radius: 50%;
}
.status-ok {
    background: var(--el-color-success);
    box-shadow: 0 0 0 3px var(--el-color-success-light-9);
}
.status-error {
    background: var(--el-color-danger);
    box-shadow: 0 0 0 3px var(--el-color-danger-light-9);
}
.device-card-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
    align-items: stretch;
}
.device-card {
    box-sizing: border-box;
    width: 100%;
    min-width: 0;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 7px;
    outline: none;
    background: var(--el-bg-color);
    transition: border-color 0.2s;
    &:hover {
        border-color: var(--el-border-color);
    }
    :deep(.el-card__header) {
        display: flex;
        align-items: center;
        box-sizing: border-box;
        padding: 16px 20px 0;
        border-bottom: 0;
    }
    :deep(.el-card__body) {
        padding: 16px 20px;
    }
}
.device-card:only-child {
    grid-column: 1 / -1;
}
.device-title {
    flex: 1;
    align-items: center;
    min-width: 0;
    strong {
        color: var(--el-text-color-primary);
        font-size: 14px;
        font-weight: 500;
        line-height: 20px;
        white-space: nowrap;
    }
}
.card-product-name {
    overflow: hidden;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.device-actions {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    gap: 12px;
    margin-left: 12px;
    .el-button {
        font-size: 12px;
        font-weight: 400;
        &:not(.is-disabled) {
            color: var(--el-text-color-regular);
            &:hover,
            &:focus-visible {
                color: var(--el-color-primary);
            }
        }
    }
    .el-button + .el-button {
        margin-left: 0;
    }
}
.device-actions .process-count {
    padding-right: 12px;
    border-right: 1px solid var(--el-border-color-lighter);
    border-radius: 0;
    white-space: nowrap;
}
.device-metrics {
    display: grid;
    grid-template-columns: 1fr 1.05fr 0.6fr;
    gap: 16px 24px;
}
.metric-item {
    min-width: 0;
    > span {
        display: block;
        margin-bottom: 6px;
        color: var(--el-text-color-secondary);
        font-size: 12px;
        line-height: 18px;
    }
    strong {
        display: block;
        overflow: hidden;
        color: var(--el-text-color-primary);
        font-size: 14px;
        font-weight: 500;
        font-variant-numeric: tabular-nums;
        line-height: 26px;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    :deep(.el-progress) {
        margin-top: 8px;
    }
}
.device-secondary-metrics {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 24px;
    margin-top: 14px;
    .metric-item {
        display: flex;
        align-items: baseline;
        gap: 6px;
        > span {
            margin-bottom: 0;
        }
        strong {
            color: var(--el-text-color-regular);
            font-size: 12px;
            font-weight: 400;
            line-height: 18px;
        }
    }
}
.metric-primary {
    min-height: 62px;
    strong {
        font-size: 20px;
        font-weight: 500;
        line-height: 26px;
    }
    :deep(.el-progress) {
        margin-top: 10px;
    }
}
.metric-memory {
    strong {
        font-size: 17px;
        font-weight: 400;
    }
    .memory-total {
        color: var(--el-text-color-secondary);
        font-size: 13px;
    }
}
.metric-temperature strong {
    font-size: 17px;
    font-weight: 400;
    line-height: 26px;
}
@media (max-width: 1200px) {
    .device-card-grid {
        grid-template-columns: 1fr;
    }
}

@media (max-width: 768px) {
    .collection-time {
        width: 100%;
    }
    .device-overview {
        gap: 8px 16px;
        padding: 10px 12px;
    }
    .overview-item + .overview-item {
        margin-left: 0;
        padding-left: 0;
        border-left: 0;
    }
    .overview-model {
        flex-basis: 100%;
    }
    .device-card-header {
        flex-wrap: wrap;
        gap: 12px;
    }
    .device-title {
        flex-basis: 100%;
    }
    .device-actions {
        flex-wrap: wrap;
        gap: 12px;
        margin-left: 0;
    }
    .device-metrics {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    .metric-temperature {
        display: flex;
        align-items: baseline;
        gap: 8px;
        grid-column: 1 / -1;
        > span {
            margin-bottom: 0;
        }
    }
}
</style>
