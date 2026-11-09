<template>
    <div>
        <RouterMenu />
        <LayoutContent v-loading="loading" :title="$t('website.groupSetting')">
            <template v-if="!noDevices" #leftToolBar>
                <el-button v-permission="'ai_gpu_manage'" type="primary" :disabled="visible || loading" @click="open()">
                    {{ $t('commons.button.create') }}
                </el-button>
            </template>
            <template v-if="!noDevices" #rightToolBar>
                <TableRefresh :disabled="loading || visible" @search="load" />
            </template>
            <template v-if="!noDevices" #prompt>
                <p class="group-help">{{ $t('aiTools.gpuGroup.help') }}</p>
            </template>
            <template #main>
                <el-alert
                    v-if="!loading && ['issues', 'loadFailed'].includes(environment.status)"
                    class="mb-3"
                    :closable="false"
                    type="warning"
                    :title="$t(`aiTools.gpu.environment.${environment.status}`, { count: environment.issueCount })"
                />
                <el-empty v-if="noDevices" :description="$t('aiTools.gpu.environment.notDetected')" />
                <GroupTopology
                    v-else-if="!loading && !loadFailed"
                    :key="currentNode"
                    :devices="numaDevices"
                    :nodes="numaNodes"
                    :groups="groups"
                    @edit="open"
                    @remove="remove"
                    @create="open(undefined, $event)"
                />
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
                <h3 class="group-form-heading">{{ $t('home.baseInfo') }}</h3>
                <div class="group-form-basics">
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
                </div>
                <el-form-item :label="$t('commons.table.description')" prop="description">
                    <el-input v-model="form.description" type="textarea" :rows="2" maxlength="500" show-word-limit />
                </el-form-item>
                <GPUConfig ref="gpuRef" v-model="form.gpu" :image-type="form.imageType" group-editor>
                    <template #device-meta="{ device }">
                        <span
                            v-if="
                                groups.some((group) => group.id !== form.id && group.gpu.deviceIDs.includes(device.id))
                            "
                            class="device-existing-groups"
                        >
                            {{ $t('aiTools.gpuGroup.groups') }} ·
                            {{
                                groups
                                    .filter((group) => group.id !== form.id && group.gpu.deviceIDs.includes(device.id))
                                    .map((group) => group.name)
                                    .join(', ')
                            }}
                        </span>
                    </template>
                </GPUConfig>
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
import { deleteGPUGroup, listGPUGroups, loadGPUInfo, loadModelGPUDevices, saveGPUGroup } from '@/api/modules/ai';
import GPUConfig from '@/views/ai/components/gpu-config/index.vue';
import GroupTopology from './topology.vue';
import RouterMenu from '../index.vue';
import { getGPUEnvironmentState } from '../environment';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { Rules } from '@/global/form-rules';
import DrawerPro from '@/components/drawer-pro/index.vue';

const { currentNode } = useGlobalStore();

const groups = ref<AI.GPUGroup[]>([]);
const devices = ref<AI.ModelGPUDevice[]>([]);
const numaNodes = ref<AI.ModelNUMANode[]>([]);
const numaDevices = computed(() => {
    const result = new Map(devices.value.map((device) => [device.id, device]));
    for (const group of groups.value) {
        for (const id of group.gpu.deviceIDs) {
            if (!result.has(id)) result.set(id, { id, name: id, type: group.imageType, available: false });
        }
    }
    return [...result.values()];
});
const loading = ref(false);
const loadFailed = ref(false);
const gpuInfo = ref<AI.Info>();
const environment = computed(() => getGPUEnvironmentState(gpuInfo.value, loadFailed.value));
const noDevices = computed(() => !loading.value && environment.value.status === 'notDetected' && !groups.value.length);
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
    loadFailed.value = false;
    try {
        const [groupResult, deviceResult, infoResult] = await Promise.all([
            listGPUGroups(node),
            loadModelGPUDevices(node),
            loadGPUInfo(true, node),
        ]);
        if (version !== requestVersion || node !== currentNode.value) return;
        gpuInfo.value = infoResult.data;
        groups.value = groupResult.data || [];
        devices.value = deviceResult.data.devices || [];
        numaNodes.value = deviceResult.data.numaNodes || [];
    } catch (error) {
        if (version === requestVersion) loadFailed.value = true;
        throw error;
    } finally {
        if (version === requestVersion) loading.value = false;
    }
};
const open = async (group?: AI.GPUGroup, deviceIDs: string[] = []) => {
    form.value = group
        ? {
              id: group.id,
              name: group.name,
              description: group.description || '',
              imageType: group.imageType,
              gpu: JSON.parse(JSON.stringify(group.gpu)),
          }
        : {
              id: 0,
              name: '',
              description: '',
              imageType:
                  devices.value.find((device) => deviceIDs.includes(device.id))?.type ||
                  devices.value.find((device) => device.available)?.type ||
                  'nvidia',
              gpu: { ...emptyGPU(), deviceIDs },
          };
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
    if (!loading.value && !visible.value) load();
});
watch(currentNode, () => {
    visible.value = false;
    form.value = { id: 0, name: '', description: '', imageType: 'nvidia', gpu: emptyGPU() };
    gpuInfo.value = undefined;
    groups.value = [];
    devices.value = [];
    numaNodes.value = [];
    load();
});
</script>

<style lang="scss" scoped>
.group-help {
    margin: 0;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 20px;
}
.group-form-heading {
    font-size: 14px;
    font-weight: 600;
    margin: 0 0 16px;
}
.group-form-basics {
    display: grid;
    grid-template-columns: minmax(0, 2fr) minmax(0, 1fr);
    gap: 16px;
}
.device-existing-groups {
    width: 100%;
    font-size: 12px;
    line-height: 20px;
    color: var(--el-text-color-secondary);
}
@media (max-width: 480px) {
    .group-form-basics {
        grid-template-columns: minmax(0, 1fr);
    }
}
</style>
