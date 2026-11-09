<template>
    <div class="gpu-topology" :class="{ selectable, 'topology-cards': cards }">
        <section
            v-for="section in sections"
            :key="section.id ?? 'unknown'"
            class="numa-node"
            :class="{ unknown: section.id === undefined }"
        >
            <header class="numa-node-header">
                <div>
                    <strong>
                        {{ section.id === undefined ? $t('aiTools.gpuGroup.unknownNode') : `NUMA ${section.id}` }}
                    </strong>
                    <div class="numa-resources">
                        <span>CPU {{ section.node?.cpus || '—' }}</span>
                        <span>
                            {{ $t('aiTools.gpuGroup.memoryNode') }}
                            {{ section.node?.memory ? `NUMA ${section.id}` : '—' }}
                        </span>
                    </div>
                </div>
                <el-checkbox
                    v-if="selectable && section.id !== undefined && section.devices.some((device) => device.available)"
                    :model-value="
                        section.devices
                            .filter((device) => device.available)
                            .every((device) => selected.includes(device.id))
                    "
                    :indeterminate="
                        section.devices.some((device) => device.available && selected.includes(device.id)) &&
                        section.devices.some((device) => device.available && !selected.includes(device.id))
                    "
                    @change="toggleNode(section.devices, Boolean($event))"
                >
                    {{ $t('aiTools.gpuGroup.selectNode') }}
                </el-checkbox>
            </header>
            <div v-if="section.devices.length" class="numa-devices">
                <div
                    v-for="device in section.devices"
                    :key="device.id"
                    class="topology-device"
                    :class="{ selected: selectable && selected.includes(device.id), unavailable: !device.available }"
                >
                    <el-checkbox
                        v-if="selectable"
                        :model-value="selected.includes(device.id)"
                        :disabled="!device.available && !selected.includes(device.id)"
                        @change="toggleDevice(device.id, Boolean($event))"
                    >
                        {{ device.name }}
                    </el-checkbox>
                    <strong v-else>{{ device.name }}</strong>
                    <el-tag v-if="!device.available" size="small" type="warning">
                        {{ $t('aiTools.model.gpuMissing') }}
                    </el-tag>
                    <slot name="device-meta" :device="device" />
                </div>
            </div>
            <div v-else class="numa-empty">{{ $t('aiTools.gpuGroup.noDevices') }}</div>
        </section>
        <div v-if="selectable && missingDevices.length" class="missing-devices">
            <el-tag v-for="id in missingDevices" :key="id" type="warning" closable @close="toggleDevice(id, false)">
                {{ id }} · {{ $t('aiTools.model.gpuMissing') }}
            </el-tag>
        </div>
        <el-empty v-if="!sections.length" :description="$t('aiTools.model.gpuNotDetected')" :image-size="60" />
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { AI } from '@/api/interface/ai';

const props = withDefaults(
    defineProps<{ devices: AI.ModelGPUDevice[]; nodes: AI.ModelNUMANode[]; selectable?: boolean; cards?: boolean }>(),
    { selectable: false, cards: false },
);
const selected = defineModel<string[]>({ default: () => [] });
const sections = computed(() => {
    const result: { id?: number; node?: AI.ModelNUMANode; devices: AI.ModelGPUDevice[] }[] = props.nodes.map(
        (node) => ({
            id: node.id,
            node,
            devices: props.devices.filter((device) => device.numaNode === node.id),
        }),
    );
    const unknown = props.devices.filter((device) => !props.nodes.some((node) => node.id === device.numaNode));
    if (unknown.length) result.push({ devices: unknown });
    return result;
});
const missingDevices = computed(() => selected.value.filter((id) => !props.devices.some((device) => device.id === id)));
const toggleDevice = (id: string, checked: boolean) => {
    selected.value = checked ? [...new Set([...selected.value, id])] : selected.value.filter((value) => value !== id);
};
const toggleNode = (devices: AI.ModelGPUDevice[], checked: boolean) => {
    const ids = devices.filter((device) => device.available).map((device) => device.id);
    selected.value = checked
        ? [...new Set([...selected.value, ...ids])]
        : selected.value.filter((id) => !ids.includes(id));
};
</script>

<style lang="scss" scoped>
.gpu-topology {
    width: 100%;
    display: grid;
    gap: 24px;
}
.topology-cards {
    grid-template-columns: repeat(auto-fit, minmax(min(440px, 100%), 1fr));
    gap: 20px;
    align-items: stretch;
}
.topology-cards > .numa-node {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    padding: 20px;
}
.topology-cards .numa-node-header {
    margin-bottom: 16px;
}
.topology-cards .numa-node-header strong {
    font-size: 16px;
}
.topology-cards .topology-device {
    align-content: start;
}
.topology-cards.selectable {
    grid-template-columns: minmax(0, 1fr);
    gap: 12px;
}
.topology-cards.selectable > .numa-node {
    padding: 16px;
    border-radius: 8px;
}
.topology-cards.selectable .numa-node-header strong {
    font-size: 13px;
}
.numa-node {
    min-width: 0;
}
.numa-node-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px 16px;
    flex-wrap: wrap;
    margin-bottom: 10px;
}
.numa-node-header > div {
    min-width: 0;
}
.numa-node-header strong {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    font-weight: 600;
    line-height: 20px;
}
.numa-node-header strong::before {
    content: '';
    width: 6px;
    height: 6px;
    flex-shrink: 0;
    border-radius: 50%;
    background: var(--el-color-primary);
    opacity: 0.65;
}
.unknown .numa-node-header strong::before {
    background: var(--el-text-color-placeholder);
}
.numa-resources {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 16px;
    font-size: 12px;
    line-height: 20px;
    color: var(--el-text-color-secondary);
    margin-top: 3px;
    margin-left: 14px;
    overflow-wrap: anywhere;
}
.numa-devices {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(210px, 100%), 1fr));
    gap: 8px;
}
.topology-device {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px 10px;
    padding: 12px 14px;
    border-radius: 6px;
    min-width: 0;
    overflow-wrap: anywhere;
    background: var(--el-fill-color-light);
}
.topology-device > strong {
    font-size: 13px;
    font-weight: 500;
    line-height: 20px;
}
.topology-device.selected {
    background: var(--el-color-primary-light-9);
}
.topology-device.unavailable {
    color: var(--el-text-color-secondary);
}
.topology-device :deep(.el-checkbox) {
    height: auto;
    width: 100%;
    align-items: flex-start;
    white-space: normal;
}
.topology-device :deep(.el-checkbox__input) {
    margin-top: 3px;
}
.topology-device :deep(.el-checkbox__label) {
    white-space: normal;
    line-height: 20px;
    font-size: 13px;
    overflow-wrap: anywhere;
}
.numa-empty {
    color: var(--el-text-color-placeholder);
    font-size: 12px;
    padding-left: 14px;
    line-height: 20px;
}
.missing-devices :deep(.el-tag) {
    max-width: 100%;
    height: auto;
}
.missing-devices :deep(.el-tag__content) {
    white-space: normal;
    overflow-wrap: anywhere;
}
.missing-devices {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
}
</style>
