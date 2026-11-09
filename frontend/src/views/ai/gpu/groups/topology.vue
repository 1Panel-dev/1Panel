<template>
    <el-card shadow="never" class="relation-card">
        <div ref="board" class="relation-board">
            <svg class="relation-lines" aria-hidden="true">
                <path v-for="line in lines" :key="line.key" :d="line.path" :class="{ active: line.active }" />
            </svg>
            <div ref="hardware" class="hardware-column">
                <h3>{{ $t('aiTools.model.gpuDevices') }}</h3>
                <section
                    v-for="section in sections"
                    :key="section.key"
                    class="numa-section"
                    :class="{
                        unassigned: section.key === 'unassigned',
                        separated: section.key === 'unassigned' && nodes.length > 0,
                    }"
                >
                    <header v-if="section.label">
                        <div>
                            <strong>{{ section.label }}</strong>
                            <div v-if="section.node" class="input-help">CPU {{ section.node.cpus || '—' }}</div>
                        </div>
                        <el-button
                            v-if="section.node"
                            v-permission="'ai_gpu_manage'"
                            link
                            type="primary"
                            @click="createFromNode(section.node.id, section.devices)"
                        >
                            {{ $t('commons.button.create') }}
                        </el-button>
                    </header>
                    <div
                        v-for="device in section.devices"
                        :key="device.id"
                        :data-device="device.id"
                        class="device-node"
                        :class="{ active: activeDevices.has(device.id), unavailable: !device.available }"
                    >
                        <button
                            type="button"
                            class="node-button"
                            :aria-pressed="focus?.type === 'device' && focus.id === device.id"
                            @click="
                                focus =
                                    focus?.type === 'device' && focus.id === device.id
                                        ? undefined
                                        : { type: 'device', id: device.id }
                            "
                        >
                            {{ device.name }}
                        </button>
                        <el-tag v-if="!device.available" size="small" type="warning">
                            {{ $t('aiTools.model.gpuMissing') }}
                        </el-tag>
                        <el-button
                            v-permission="'ai_gpu_manage'"
                            link
                            type="primary"
                            class="device-create"
                            :disabled="!device.available"
                            :aria-label="`${$t('commons.button.create')} · ${device.name}`"
                            @click="$emit('create', [device.id])"
                        >
                            {{ $t('commons.button.create') }}
                        </el-button>
                        <span class="relation-port" />
                    </div>
                    <span v-if="!section.devices.length" class="input-help">
                        {{ $t('aiTools.gpuGroup.noDevices') }}
                    </span>
                </section>
            </div>
            <div ref="groupColumn" class="groups-column">
                <h3>{{ $t('commons.table.group') }}</h3>
                <div class="group-nodes">
                    <el-empty v-if="!groups.length" :image-size="60" />
                    <div
                        v-for="group in groups"
                        :key="group.id"
                        :data-group="group.id"
                        class="relation-group"
                        :class="{ active: activeGroups.has(group.id) }"
                    >
                        <button
                            type="button"
                            class="node-button"
                            :aria-pressed="focus?.type === 'group' && focus.id === group.id"
                            @click="
                                focus =
                                    focus?.type === 'group' && focus.id === group.id
                                        ? undefined
                                        : { type: 'group', id: group.id }
                            "
                        >
                            <strong>{{ group.name || $t('commons.button.create') }}</strong>
                            <span class="input-help">{{ group.gpu.deviceIDs.length }} GPU · {{ group.imageType }}</span>
                        </button>
                        <span class="relation-port" />
                    </div>
                </div>
            </div>
        </div>
        <div v-if="focusedGroup" class="relation-detail">
            <div class="detail-heading">
                <div class="detail-title">
                    <strong>{{ focusedGroup.name }}</strong>
                    <el-tag size="small" type="info" effect="plain">{{ focusedGroup.gpu.deviceIDs.length }} GPU</el-tag>
                </div>
                <div class="detail-actions">
                    <el-button v-permission="'ai_gpu_manage'" link type="primary" @click="$emit('edit', focusedGroup)">
                        {{ $t('commons.button.edit') }}
                    </el-button>
                    <el-button
                        v-permission="'ai_gpu_manage'"
                        link
                        type="primary"
                        @click="$emit('remove', focusedGroup)"
                    >
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </div>
            </div>
            <p v-if="focusedGroup.description" class="detail-description">{{ focusedGroup.description }}</p>
            <dl class="detail-fields">
                <div>
                    <dt>{{ $t('aiTools.gpuGroup.affinity') }}</dt>
                    <dd>
                        <span>
                            {{
                                $t(
                                    focusedGroup.gpu.numa?.mode === 'auto'
                                        ? 'aiTools.gpuGroup.followGPU'
                                        : focusedGroup.gpu.numa?.mode === 'manual'
                                          ? 'aiTools.model.numaManual'
                                          : 'aiTools.model.numaOff',
                                )
                            }}
                        </span>
                        <div v-if="focusedGroup.gpu.numa?.mode === 'manual'" class="detail-tags">
                            <el-tag
                                v-for="node in focusedGroup.gpu.numa.nodes"
                                :key="node"
                                size="small"
                                type="info"
                                effect="plain"
                            >
                                NUMA {{ node }}
                            </el-tag>
                        </div>
                    </dd>
                </div>
                <div>
                    <dt>{{ $t('aiTools.gpuGroup.deployments') }}</dt>
                    <dd class="detail-tags">
                        <el-tag
                            v-for="deployment in focusedGroup.deployments"
                            :key="deployment"
                            size="small"
                            type="info"
                            effect="plain"
                        >
                            {{ deployment }}
                        </el-tag>
                        <span v-if="!focusedGroup.deployments.length" class="detail-empty">—</span>
                    </dd>
                </div>
            </dl>
        </div>
        <div v-else-if="focusedDevice" class="relation-detail">
            <div class="detail-heading">
                <div class="detail-title">
                    <strong>{{ focusedDevice.name }}</strong>
                    <el-tag v-if="!focusedDevice.available" size="small" type="warning">
                        {{ $t('aiTools.model.gpuMissing') }}
                    </el-tag>
                </div>
                <el-button
                    v-permission="'ai_gpu_manage'"
                    link
                    type="primary"
                    :disabled="!focusedDevice.available"
                    @click="$emit('create', [focusedDevice.id])"
                >
                    {{ $t('commons.button.create') }}
                </el-button>
            </div>
            <dl class="detail-fields device-fields">
                <div>
                    <dt>{{ $t('commons.table.group') }}</dt>
                    <dd class="detail-tags">
                        <el-tag
                            v-for="group in groups.filter((group) => group.gpu.deviceIDs.includes(focusedDevice.id))"
                            :key="group.id"
                            size="small"
                            type="info"
                            effect="plain"
                        >
                            {{ group.name }}
                        </el-tag>
                        <span
                            v-if="!groups.some((group) => group.gpu.deviceIDs.includes(focusedDevice.id))"
                            class="detail-empty"
                        >
                            {{ $t('aiTools.gpuGroup.ungrouped') }}
                        </span>
                    </dd>
                </div>
            </dl>
        </div>
    </el-card>
</template>

<script setup lang="ts">
import { computed, h, nextTick, ref, watch } from 'vue';
import { ElMessageBox } from 'element-plus';
import { useResizeObserver } from '@vueuse/core';
import { AI } from '@/api/interface/ai';
import i18n from '@/lang';

const props = defineProps<{
    devices: AI.ModelGPUDevice[];
    nodes: AI.ModelNUMANode[];
    groups: AI.GPUGroup[];
}>();
const emit = defineEmits<{ create: [ids: string[]]; edit: [group: AI.GPUGroup]; remove: [group: AI.GPUGroup] }>();
const focus = ref<{ type: 'device'; id: string } | { type: 'group'; id: number }>();
const board = ref<HTMLElement>();
const hardware = ref<HTMLElement>();
const groupColumn = ref<HTMLElement>();
const lines = ref<{ key: string; path: string; active: boolean }[]>([]);
const sections = computed(() => {
    const available = props.devices.filter((device) => device.available);
    const result: { key: string; label: string; node?: AI.ModelNUMANode; devices: AI.ModelGPUDevice[] }[] =
        props.nodes.map((node) => ({
            key: `numa-${node.id}`,
            label: `NUMA ${node.id}`,
            node,
            devices: available.filter((device) => device.numaNode === node.id),
        }));
    const unassigned = available.filter((device) => !props.nodes.some((node) => node.id === device.numaNode));
    const missing = props.devices.filter((device) => !device.available);
    if (missing.length)
        result.push({
            key: 'missing',
            label: i18n.global.t('aiTools.model.gpuMissing'),
            node: undefined,
            devices: missing,
        });
    if (unassigned.length) result.push({ key: 'unassigned', label: '', devices: unassigned });
    return result;
});
const activeGroups = computed(
    () =>
        new Set(
            props.groups
                .filter((group) =>
                    focus.value?.type === 'group'
                        ? group.id === focus.value.id
                        : focus.value?.type === 'device' && group.gpu.deviceIDs.includes(focus.value.id),
                )
                .map((group) => group.id),
        ),
);
const activeDevices = computed(
    () =>
        new Set(
            focus.value?.type === 'device'
                ? [focus.value.id]
                : props.groups
                      .filter((group) => activeGroups.value.has(group.id))
                      .flatMap((group) => group.gpu.deviceIDs),
        ),
);
const focusedGroup = computed(() =>
    props.groups.find((group) => focus.value?.type === 'group' && group.id === focus.value.id),
);
const focusedDevice = computed(() =>
    props.devices.find((device) => focus.value?.type === 'device' && device.id === focus.value.id),
);
const createFromNode = async (node: number, devices: AI.ModelGPUDevice[]) => {
    const available = devices.filter((device) => device.available);
    if (!available.length) {
        await ElMessageBox.alert(
            h('div', [
                h('p', i18n.global.t('aiTools.gpuGroup.emptyNodeCreate', { node })),
                h('p', { class: 'mt-3' }, i18n.global.t('aiTools.gpuGroup.inspectNUMA')),
                h(
                    'pre',
                    { class: 'my-3', style: { whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' } },
                    'lspci -Dnn\ngrep -H . /sys/bus/pci/devices/*/numa_node',
                ),
                h('p', i18n.global.t('aiTools.gpuGroup.numaQueryHint')),
            ]),
            i18n.global.t('commons.msg.infoTitle'),
            {
                type: 'info',
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                customStyle: { width: '640px', maxWidth: 'calc(100vw - 32px)' },
            },
        ).catch(() => {});
        return;
    }
    emit(
        'create',
        available.filter((device) => device.type === available[0].type).map((device) => device.id),
    );
};
const drawLines = () => {
    if (!board.value) return;
    const rect = board.value.getBoundingClientRect();
    const devices = new Map(
        [...board.value.querySelectorAll<HTMLElement>('[data-device]')].map((element) => [
            element.dataset.device,
            element.getBoundingClientRect(),
        ]),
    );
    const groups = new Map(
        [...board.value.querySelectorAll<HTMLElement>('[data-group]')].map((element) => [
            Number(element.dataset.group),
            element.getBoundingClientRect(),
        ]),
    );
    const next: typeof lines.value = [];
    for (const group of props.groups) {
        const target = groups.get(group.id);
        if (!target) continue;
        for (const id of group.gpu.deviceIDs) {
            const source = devices.get(id);
            if (!source) continue;
            const x1 = source.right - rect.left;
            const y1 = source.top + source.height / 2 - rect.top;
            const x2 = target.left - rect.left;
            const y2 = target.top + target.height / 2 - rect.top;
            const mid = (x1 + x2) / 2;
            const active = activeGroups.value.has(group.id) && activeDevices.value.has(id);
            next.push({
                key: `${group.id}:${id}`,
                path: `M ${x1} ${y1} C ${mid} ${y1}, ${mid} ${y2}, ${x2} ${y2}`,
                active,
            });
        }
    }
    lines.value = next.sort((a, b) => Number(a.active) - Number(b.active));
};
useResizeObserver([board, hardware, groupColumn], drawLines);
watch(
    [() => props.groups, sections, activeDevices, activeGroups],
    async () => {
        await nextTick();
        drawLines();
    },
    { immediate: true },
);
</script>

<style scoped lang="scss">
.detail-heading,
.numa-section header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
}
.relation-card {
    border-color: var(--el-border-color-lighter);
    border-radius: 8px;
}
.relation-board {
    position: relative;
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(48px, 12%) minmax(0, 1fr);
    align-items: stretch;
}
.hardware-column {
    grid-column: 1;
}
.groups-column {
    grid-column: 3;
    display: flex;
    flex-direction: column;
    gap: 20px;
}
.group-nodes {
    display: flex;
    flex: 1;
    flex-direction: column;
    justify-content: space-around;
    gap: 20px;
    padding-bottom: 16px;
}
.hardware-column,
.groups-column {
    position: relative;
    z-index: 2;
    min-width: 0;
}
h3 {
    margin: 0 0 16px;
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-secondary);
}
.groups-column h3 {
    margin-bottom: 0;
}
.numa-section {
    padding: 16px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    margin-bottom: 16px;
}
.numa-section header {
    margin-bottom: 12px;
}
.numa-section.unassigned {
    border: 0;
    padding: 0;
}
.numa-section.unassigned.separated {
    border-top: 1px solid var(--el-border-color-lighter);
    border-radius: 0;
    padding-top: 16px;
}
.device-node,
.relation-group {
    position: relative;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    background: var(--el-bg-color);
}
.device-node {
    margin-top: 8px;
    padding: 10px 12px;
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    align-items: center;
}
.device-node > .node-button {
    flex: 1;
    min-width: 0;
    width: auto;
}
.device-create {
    flex-shrink: 0;
    margin-left: auto;
}
.device-node.active,
.relation-group.active {
    border-color: var(--el-color-primary-light-5);
    background: var(--el-color-primary-light-9);
}
.device-node.unavailable {
    color: var(--el-text-color-secondary);
}
.relation-group {
    padding: 16px;
}
.node-button {
    position: relative;
    border: 0;
    background: transparent;
    color: inherit;
    font: inherit;
    font-size: 13px;
    line-height: 20px;
    text-align: left;
    cursor: pointer;
    padding: 0;
    width: 100%;
    overflow-wrap: anywhere;
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-height: 24px;
}
.relation-port {
    position: absolute;
    right: -4px;
    top: calc(50% - 4px);
    width: 8px;
    height: 8px;
    border: 1px solid var(--el-border-color);
    border-radius: 50%;
    background: var(--el-bg-color);
}
.relation-group .relation-port {
    left: -4px;
    right: auto;
}
.active > .relation-port {
    background: var(--el-color-primary);
    border-color: var(--el-color-primary);
}
.relation-lines {
    position: absolute;
    z-index: 1;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    overflow: visible;
}
.relation-lines path {
    fill: none;
    stroke: var(--el-border-color);
    stroke-width: 1.5;
}
.relation-lines path.active {
    stroke: var(--el-color-primary);
    stroke-width: 2;
}
.relation-detail {
    border-top: 1px solid var(--el-border-color-lighter);
    padding-top: 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    overflow-wrap: anywhere;
    font-size: 13px;
    line-height: 20px;
}
.detail-title {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
    flex: 1;
    min-width: 0;
}
.detail-title > strong {
    font-size: 15px;
    font-weight: 600;
}
.detail-actions {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-shrink: 0;
}
.detail-actions :deep(.el-button + .el-button) {
    margin-left: 0;
}
.detail-description {
    margin: 0;
    color: var(--el-text-color-secondary);
    white-space: pre-wrap;
}
.detail-fields {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 20px 32px;
    margin: 0;
}
.detail-fields > div {
    min-width: 0;
}
.detail-fields dt {
    margin-bottom: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
}
.detail-fields dd {
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
}
.detail-fields.device-fields {
    grid-template-columns: minmax(0, 1fr);
}
.detail-tags,
.detail-fields dd.detail-tags {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    gap: 6px;
}
.detail-tags :deep(.el-tag) {
    max-width: 100%;
    height: auto;
    min-height: 22px;
}
.detail-tags :deep(.el-tag__content) {
    white-space: normal;
    overflow-wrap: anywhere;
}
.detail-empty {
    color: var(--el-text-color-secondary);
}
@media (max-width: 600px) {
    .device-node > .node-button {
        flex-basis: 100%;
    }
    .detail-fields {
        grid-template-columns: minmax(0, 1fr);
        gap: 16px;
    }
    .relation-card :deep(.el-card__body) {
        padding: 12px;
    }
    .numa-section {
        padding: 8px;
    }
    .device-node,
    .relation-group {
        padding: 10px 8px;
    }
    .relation-board {
        grid-template-columns: minmax(0, 1fr) 32px minmax(0, 1fr);
    }
}
</style>
