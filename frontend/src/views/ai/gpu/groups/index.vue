<template>
    <div>
        <RouterMenu />
        <LayoutContent v-loading="loading" :title="$t('website.groupSetting')">
            <template v-if="!loading && !loadFailed && devices.length > 1" #leftToolBar>
                <el-button v-permission="'ai_gpu_manage'" type="primary" :disabled="visible || loading" @click="open()">
                    {{ $t('commons.button.create') }}
                </el-button>
            </template>
            <template v-if="!noDevices && !singleGPU" #rightToolBar>
                <TableRefresh :disabled="loading || visible" @search="load" />
            </template>
            <template v-if="!loading && !loadFailed && devices.length > 1" #prompt>
                <p class="group-help">{{ $t('aiTools.gpuGroup.help') }}</p>
            </template>
            <template #main>
                <el-alert
                    v-if="!loading && loadFailed"
                    class="mb-3"
                    :closable="false"
                    type="warning"
                    :title="$t('aiTools.model.loadFailed')"
                />
                <el-empty v-if="noDevices" :description="$t('aiTools.gpu.environment.notDetected')" />
                <el-empty v-else-if="singleGPU" :description="$t('aiTools.gpuGroup.singleGPU')" />
                <GroupTopology
                    v-else-if="!loading && !loadFailed"
                    :key="currentNode"
                    :devices="numaDevices"
                    :nodes="numaNodes"
                    :groups="groups"
                    @edit="open"
                    @remove="remove"
                />
            </template>
        </LayoutContent>
        <GroupOperate
            v-if="visible"
            v-model:visible="visible"
            :group="editingGroup"
            :groups="groups"
            :devices="devices"
            @saved="load"
        />
    </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onActivated, ref, watch } from 'vue';
import { ElMessageBox } from 'element-plus';
import { AI } from '@/api/interface/ai';
import { deleteGPUGroup, listGPUGroups, loadModelGPUDevices } from '@/api/modules/ai';
import GroupOperate from './operate/index.vue';
import GroupTopology from './topology.vue';
import RouterMenu from '../index.vue';
import i18n from '@/lang';
import { useGlobalStore } from '@/composables/useGlobalStore';

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
const noDevices = computed(() => !loading.value && !loadFailed.value && devices.value.length === 0);
const singleGPU = computed(() => !loading.value && !loadFailed.value && devices.value.length === 1);
const visible = ref(false);
const editingGroup = ref<AI.GPUGroup>();
let requestVersion = 0;
const load = async () => {
    const version = ++requestVersion;
    const node = currentNode.value;
    loading.value = true;
    loadFailed.value = false;
    try {
        const deviceResult = await loadModelGPUDevices(node);
        if (version !== requestVersion || node !== currentNode.value) return;
        const groupResult = deviceResult.data.devices?.length > 1 ? await listGPUGroups(node) : undefined;
        if (version !== requestVersion || node !== currentNode.value) return;
        groups.value = groupResult?.data || [];
        devices.value = deviceResult.data.devices || [];
        numaNodes.value = deviceResult.data.numaNodes || [];
    } catch (error) {
        if (version === requestVersion) loadFailed.value = true;
        throw error;
    } finally {
        if (version === requestVersion) loading.value = false;
    }
};
const open = (group?: AI.GPUGroup) => {
    editingGroup.value = group;
    visible.value = true;
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
    editingGroup.value = undefined;
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
</style>
