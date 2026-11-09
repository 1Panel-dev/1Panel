<template>
    <div>
        <RouterMenu />
        <LayoutContent v-loading="loading" :title="$t('aiTools.gpuGroup.groups')">
            <template #leftToolBar>
                <el-button v-permission="'ai_gpu_manage'" type="primary" @click="open()">
                    {{ $t('commons.button.create') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <el-select v-model="viewMode" class="p-w-200">
                    <template #prefix>{{ $t('aiTools.gpuGroup.dimension') }}</template>
                    <el-option value="group" :label="$t('aiTools.gpuGroup.byGroup')" />
                    <el-option value="numa" :label="$t('aiTools.gpuGroup.byNUMA')" />
                </el-select>
                <TableRefresh :disabled="loading" @search="load" />
            </template>
            <template #prompt>
                <el-alert :closable="false" :title="$t('aiTools.gpuGroup.help')" type="info" />
            </template>
            <template #main>
                <el-empty v-if="viewMode === 'group' && !groups.length && !loading" />
                <div v-if="viewMode === 'group'" class="gpu-group-cards">
                    <el-card v-for="group in groups" :key="group.id" shadow="never" class="gpu-group-card">
                        <template #header>
                            <div class="group-card-header">
                                <strong>{{ group.name }}</strong>
                                <div class="group-card-actions">
                                    <el-button v-permission="'ai_gpu_manage'" link type="primary" @click="open(group)">
                                        {{ $t('commons.button.edit') }}
                                    </el-button>
                                    <el-button
                                        v-permission="'ai_gpu_manage'"
                                        link
                                        type="primary"
                                        @click="remove(group)"
                                    >
                                        {{ $t('commons.button.delete') }}
                                    </el-button>
                                </div>
                            </div>
                        </template>
                        <p v-if="group.description" class="group-description">{{ group.description }}</p>
                        <GPUTopology
                            v-if="devices.some((device) => group.gpu.deviceIDs.includes(device.id))"
                            :devices="devices.filter((device) => group.gpu.deviceIDs.includes(device.id))"
                            :nodes="
                                numaNodes.filter((node) =>
                                    devices.some(
                                        (device) =>
                                            group.gpu.deviceIDs.includes(device.id) && device.numaNode === node.id,
                                    ),
                                )
                            "
                        />
                        <div class="missing-group-devices">
                            <el-tag
                                v-for="id in group.gpu.deviceIDs.filter(
                                    (id) => !devices.some((device) => device.id === id),
                                )"
                                :key="id"
                                type="warning"
                            >
                                {{ id }} · {{ $t('aiTools.model.gpuMissing') }}
                            </el-tag>
                        </div>
                        <div class="group-card-meta">
                            <div class="group-binding">
                                <span>{{ $t('aiTools.gpuGroup.affinity') }}</span>
                                <strong>
                                    {{
                                        $t(
                                            group.gpu.numa?.mode === 'auto'
                                                ? 'aiTools.gpuGroup.followGPU'
                                                : group.gpu.numa?.mode === 'manual'
                                                  ? 'aiTools.model.numaManual'
                                                  : 'aiTools.model.numaOff',
                                        )
                                    }}
                                </strong>
                                <div v-if="group.gpu.numa && group.gpu.numa.mode !== 'off'">
                                    <div v-for="id in groupNodes(group)" :key="id">
                                        NUMA {{ id }} · CPU {{ numaNodes.find((node) => node.id === id)?.cpus || '—' }}
                                    </div>
                                </div>
                            </div>
                            <div class="group-deployments">
                                <span>{{ $t('aiTools.gpuGroup.deployments') }}</span>
                                <div>{{ group.deployments.join(', ') || '—' }}</div>
                            </div>
                        </div>
                    </el-card>
                </div>
                <GPUTopology v-else :devices="numaDevices" :nodes="numaNodes" cards>
                    <template #device-meta="{ device }">
                        <div class="device-memberships">
                            <template v-if="deviceGroups.get(device.id)?.length">
                                <span>{{ $t('aiTools.gpuGroup.groups') }}</span>
                                <el-button
                                    v-for="group in deviceGroups.get(device.id)"
                                    :key="group.id"
                                    v-permission="'ai_gpu_manage'"
                                    link
                                    type="primary"
                                    @click="open(group)"
                                >
                                    {{ group.name }}
                                </el-button>
                            </template>
                            <span v-else>{{ $t('aiTools.gpuGroup.ungrouped') }}</span>
                        </div>
                    </template>
                </GPUTopology>
            </template>
        </LayoutContent>
        <DrawerPro
            v-model="visible"
            :header="$t(form.id ? 'commons.button.edit' : 'commons.button.create')"
            :resource="$t('aiTools.gpuGroup.groups')"
            size="large"
            :auto-close="false"
            confirm-before-close
            @before-close="close"
        >
            <el-form
                ref="formRef"
                :model="form"
                :rules="rules"
                label-position="top"
                v-loading="saving"
                :disabled="saving"
                @submit.prevent
            >
                <el-form-item :label="$t('commons.table.name')" prop="name">
                    <el-input v-model.trim="form.name" maxlength="128" clearable />
                </el-form-item>
                <el-form-item :label="$t('aiTools.gpuGroup.vendor')" prop="imageType">
                    <el-select v-model="form.imageType" @change="changeVendor">
                        <el-option
                            v-for="vendor in ['nvidia', 'intel', 'ascend']"
                            :key="vendor"
                            :value="vendor"
                            :label="vendor === 'nvidia' ? 'NVIDIA' : vendor === 'intel' ? 'Intel' : 'Ascend'"
                        />
                    </el-select>
                </el-form-item>
                <GPUConfig ref="gpuRef" v-model="form.gpu" :image-type="form.imageType" group-editor>
                    <template #device-meta="{ device }">
                        <span
                            v-if="deviceGroups.get(device.id)?.some((group) => group.id !== form.id)"
                            class="device-existing-groups"
                        >
                            {{ $t('aiTools.gpuGroup.groups') }} ·
                            {{
                                deviceGroups
                                    .get(device.id)
                                    ?.filter((group) => group.id !== form.id)
                                    .map((group) => group.name)
                                    .join(', ')
                            }}
                        </span>
                    </template>
                </GPUConfig>
                <el-form-item :label="$t('commons.table.description')" prop="description">
                    <el-input v-model="form.description" type="textarea" :rows="2" maxlength="500" show-word-limit />
                </el-form-item>
                <div v-if="form.id" class="input-help">{{ $t('aiTools.gpuGroup.snapshotHelp') }}</div>
            </el-form>
            <template #footer>
                <el-button :disabled="saving" @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" :loading="saving" :disabled="gpuRef?.loading" @click="save">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </template>
        </DrawerPro>
    </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onActivated, ref, watch } from 'vue';
import { ElMessageBox, FormInstance } from 'element-plus';
import { AI } from '@/api/interface/ai';
import { deleteGPUGroup, listGPUGroups, loadModelGPUDevices, saveGPUGroup } from '@/api/modules/ai';
import GPUConfig from '@/views/ai/components/gpu-config/index.vue';
import GPUTopology from '@/views/ai/components/gpu-topology/index.vue';
import RouterMenu from '../index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { Rules } from '@/global/form-rules';
import DrawerPro from '@/components/drawer-pro/index.vue';

const { currentNode } = useGlobalStore();

const viewMode = ref<'group' | 'numa'>('group');
const groups = ref<AI.GPUGroup[]>([]);
const devices = ref<AI.ModelGPUDevice[]>([]);
const numaNodes = ref<AI.ModelNUMANode[]>([]);
const deviceGroups = computed(() => {
    const memberships = new Map<string, AI.GPUGroup[]>();
    for (const group of groups.value) {
        for (const id of new Set(group.gpu.deviceIDs)) {
            const entries = memberships.get(id) || [];
            entries.push(group);
            memberships.set(id, entries);
        }
    }
    return memberships;
});
const numaDevices = computed(() => {
    const result = new Map(devices.value.map((device) => [device.id, device]));
    for (const [id, memberships] of deviceGroups.value) {
        if (!result.has(id)) result.set(id, { id, name: id, type: memberships[0].imageType, available: false });
    }
    return [...result.values()];
});
const loading = ref(false);
const saving = ref(false);
const visible = ref(false);
const formRef = ref<FormInstance>();
const gpuRef = ref<InstanceType<typeof GPUConfig>>();
const emptyGPU = (): AI.ModelGPUConfig => ({
    enabled: true,
    mode: 'devices',
    deviceIDs: [],
    capabilities: ['gpu', 'utility', 'compute'],
    numa: { mode: 'off', nodes: [] },
});
const form = ref({ id: 0, name: '', description: '', imageType: 'nvidia', gpu: emptyGPU() });
const rules = {
    name: [
        Rules.requiredInput,
        {
            validator: (_rule: unknown, value: string, callback: (error?: Error) => void) => {
                callback(
                    groups.value.some((group) => group.id !== form.value.id && group.name === value.trim())
                        ? new Error(i18n.global.t('commons.rule.duplicate'))
                        : undefined,
                );
            },
            trigger: 'blur',
        },
    ],
    imageType: [Rules.requiredSelect],
};
let requestVersion = 0;
const load = async () => {
    const version = ++requestVersion;
    const node = currentNode.value;
    loading.value = true;
    try {
        const [groupResult, deviceResult] = await Promise.all([listGPUGroups(node), loadModelGPUDevices(node)]);
        if (version !== requestVersion || node !== currentNode.value) return;
        groups.value = groupResult.data || [];
        devices.value = deviceResult.data.devices || [];
        numaNodes.value = deviceResult.data.numaNodes || [];
    } finally {
        if (version === requestVersion) loading.value = false;
    }
};
const groupNodes = (group: AI.GPUGroup) =>
    group.gpu.numa?.mode === 'auto'
        ? [
              ...new Set(
                  devices.value
                      .filter((device) => group.gpu.deviceIDs.includes(device.id) && device.numaNode !== undefined)
                      .map((device) => device.numaNode),
              ),
          ]
        : group.gpu.numa?.nodes || [];
const open = async (group?: AI.GPUGroup) => {
    form.value = group
        ? {
              id: group.id,
              name: group.name,
              description: group.description || '',
              imageType: group.imageType,
              gpu: JSON.parse(JSON.stringify(group.gpu)),
          }
        : { id: 0, name: '', description: '', imageType: 'nvidia', gpu: emptyGPU() };
    visible.value = true;
    await nextTick();
    formRef.value?.clearValidate();
    await gpuRef.value?.init(false);
};
const changeVendor = async () => {
    form.value.gpu = emptyGPU();
    await nextTick();
    formRef.value?.clearValidate();
    await gpuRef.value?.init(false);
};
const close = (done: () => void) => {
    if (!saving.value) done();
};
const save = async () => {
    if (saving.value || gpuRef.value?.loading) return;
    const node = currentNode.value;
    if (!(await formRef.value?.validate().catch(() => false))) return;
    if (node !== currentNode.value) return;
    saving.value = true;
    try {
        await saveGPUGroup(form.value);
        visible.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await load();
    } finally {
        saving.value = false;
    }
};
const remove = async (group: AI.GPUGroup) => {
    const node = currentNode.value;
    await ElMessageBox.confirm(
        i18n.global.t('aiTools.gpuGroup.deleteHelp', { name: group.name }),
        i18n.global.t('commons.button.delete'),
        { type: 'warning' },
    );
    if (node !== currentNode.value) return;
    await deleteGPUGroup(group.id);
    await load();
};
onMounted(load);
onActivated(() => {
    if (!loading.value) load();
});
watch(currentNode, () => {
    visible.value = false;
    groups.value = [];
    devices.value = [];
    numaNodes.value = [];
    load();
});
</script>

<style lang="scss" scoped>
.group-description {
    margin: 0 0 16px;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    line-height: 20px;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
}
.device-existing-groups {
    width: 100%;
    font-size: 12px;
    line-height: 20px;
    color: var(--el-text-color-secondary);
}

.device-memberships {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 4px 10px;
    width: 100%;
    font-size: 12px;
    color: var(--el-text-color-secondary);
}
.device-memberships :deep(.el-button) {
    margin-left: 0;
    height: auto;
    white-space: normal;
    text-align: left;
    font-size: 12px;
    min-width: 0;
}
.device-memberships :deep(.el-button > span) {
    overflow-wrap: anywhere;
}

.gpu-group-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(440px, 100%), 1fr));
    gap: 20px;
    align-items: stretch;
}
.gpu-group-card {
    display: flex;
    flex-direction: column;
    min-width: 0;
    border-color: var(--el-border-color-lighter);
    border-radius: 10px;
}
.gpu-group-card :deep(.el-card__header) {
    border-bottom: 0;
    padding: 20px 20px 16px;
}
.gpu-group-card :deep(.el-card__body) {
    display: flex;
    flex: 1;
    flex-direction: column;
    padding: 0 20px 20px;
}
.group-card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
    overflow-wrap: anywhere;
}
.group-card-header > strong {
    min-width: 0;
    flex: 1;
    font-size: 16px;
    font-weight: 600;
}
.group-card-actions {
    flex-shrink: 0;
}
.group-card-meta {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    gap: 16px;
    margin-top: auto;
    padding-top: 24px;
}
.group-binding,
.group-deployments {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 12px;
    line-height: 20px;
    overflow-wrap: anywhere;
}
.group-binding > strong {
    font-weight: 500;
    font-size: 13px;
}
.group-binding > span,
.group-deployments > span {
    color: var(--el-text-color-secondary);
}
.missing-group-devices :deep(.el-tag) {
    max-width: 100%;
    height: auto;
}
.missing-group-devices :deep(.el-tag__content) {
    white-space: normal;
    overflow-wrap: anywhere;
}
.missing-group-devices:not(:empty) {
    margin-top: 8px;
}
.missing-group-devices {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
}
@media (max-width: 480px) {
    .group-card-meta {
        grid-template-columns: minmax(0, 1fr);
    }
}
</style>
