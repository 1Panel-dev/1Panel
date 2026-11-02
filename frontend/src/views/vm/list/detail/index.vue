<template>
    <DrawerPro v-model="open" :header="vmItem?.name || $t('xpack.vm.title')" size="large">
        <div v-if="vmItem" v-loading="loading" class="vm-detail">
            <el-descriptions :column="2" border>
                <el-descriptions-item :label="$t('commons.table.name')">
                    {{ vmItem.name }}
                </el-descriptions-item>
                <el-descriptions-item label="UUID">{{ vmItem.uuid || '-' }}</el-descriptions-item>
                <el-descriptions-item :label="$t('commons.table.status')">
                    <Status :status="vmItem.status === 'shut off' ? 'stopped' : vmItem.status"></Status>
                </el-descriptions-item>
                <el-descriptions-item :label="$t('xpack.vm.cpu')">{{ vmItem.cpu }}</el-descriptions-item>
                <el-descriptions-item :label="$t('xpack.vm.memory')">
                    {{ computeSizeFromMB(vmItem.memory) }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('xpack.vm.diskBus')">
                    {{ vmItem.diskBus?.toUpperCase() || '-' }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('xpack.vm.network')">
                    {{ vmItem.network || '-' }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('xpack.vm.bootOrder')">
                    {{ vmItem.bootOrder || '-' }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('xpack.vm.isoPath')">
                    <div v-for="path in vmItem.isoPaths || []" :key="path">
                        {{ path }}
                    </div>
                    <span v-if="!vmItem.isoPaths?.length">-</span>
                </el-descriptions-item>
                <el-descriptions-item :label="$t('xpack.vm.console')">
                    <span v-if="vmItem.graphics">
                        {{ vmItem.graphics.type?.toUpperCase() }} {{ vmItem.graphics.listen || '-' }}:{{
                            vmItem.graphics.port || '-'
                        }}
                    </span>
                    <span v-else>-</span>
                </el-descriptions-item>
                <el-descriptions-item :label="$t('commons.table.description')">
                    {{ vmItem.description || '-' }}
                </el-descriptions-item>
            </el-descriptions>

            <div class="section-title">{{ $t('xpack.vm.disk') }}</div>
            <ComplexTable :data="vmItem.disks || []">
                <el-table-column prop="device" :label="$t('xpack.vm.device')" width="110" />
                <el-table-column prop="target" label="Target" width="110" />
                <el-table-column prop="bus" label="Bus" width="100" />
                <el-table-column prop="source" :label="$t('xpack.vm.path')" min-width="260" show-overflow-tooltip />
            </ComplexTable>

            <div class="section-title">{{ $t('xpack.vm.interface') }}</div>
            <ComplexTable :data="vmItem.interfaces || []">
                <el-table-column prop="type" :label="$t('commons.table.type')" width="110" />
                <el-table-column prop="source" :label="$t('xpack.vm.network')" min-width="160" />
                <el-table-column prop="model" label="Model" width="120" />
                <el-table-column prop="mac" label="MAC" min-width="160" />
                <el-table-column prop="ip" label="IP" min-width="140">
                    <template #default="{ row }">{{ row.ip || '-' }}</template>
                </el-table-column>
            </ComplexTable>
        </div>
    </DrawerPro>
</template>

<script setup lang="ts">
import { getVM } from '@/api/modules/vm';
import { computeSizeFromMB } from '@/utils/size';
import { ref } from 'vue';

const open = ref(false);
const loading = ref(false);
const vmItem = ref<any>(null);

const acceptParams = (name: string): void => {
    open.value = true;
    loadDetail(name);
};

const loadDetail = async (name: string) => {
    loading.value = true;
    await getVM(name)
        .then((res) => {
            vmItem.value = res.data;
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.vm-detail {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.section-title {
    margin-top: 8px;
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
}
</style>
