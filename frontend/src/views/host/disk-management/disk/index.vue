<template>
    <div>
        <DiskRouter />
        <MainDiv class="mt-2" :height-diff="140" v-loading="loading">
            <div v-if="diskInfo?.systemDisk">
                <DiskCard class="mt-2" :diskInfo="diskInfo.systemDisk" scope="system" />
            </div>
            <div v-if="diskInfo?.unpartitionedDisks">
                <DiskCard
                    class="mt-2"
                    v-for="(disk, index) in diskInfo.unpartitionedDisks"
                    :key="index"
                    :diskInfo="disk"
                    scope="unpartitioned"
                    @partition="(diskInfo) => partitionRef.acceptParams(diskInfo, 'partition')"
                    @search="() => getDisk()"
                />
            </div>
            <div v-if="diskInfo?.disks">
                <DiskCard
                    class="mt-2"
                    v-for="(disk, index) in diskInfo.disks"
                    :key="index"
                    :diskInfo="disk"
                    @mount="(diskInfo) => partitionRef.acceptParams(diskInfo, 'mount')"
                    scope="normal"
                    @search="() => getDisk()"
                />
            </div>
        </MainDiv>
        <Partition ref="partitionRef" @search="getDisk" />
    </div>
</template>
<script lang="ts" setup>
import { Host } from '@/api/interface/host';
import { listDisks } from '@/api/modules/host';
import DiskRouter from '@/views/host/disk-management/index.vue';
import DiskCard from '@/views/host/disk-management/components/disk-card.vue';
import Partition from '@/views/host/disk-management/partition/index.vue';

const loading = ref(false);
const partitionRef = ref();
const diskInfo = ref<Host.CompleteDiskInfo>();

const getDisk = async () => {
    try {
        loading.value = true;
        const res = await listDisks();
        diskInfo.value = res.data;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

onMounted(() => {
    getDisk();
});
</script>
