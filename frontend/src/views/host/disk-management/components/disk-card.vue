<template>
    <el-card class="shadow-sm">
        <div class="border-b pb-4">
            <div class="flex items-center space-x-4">
                <div>
                    <h3 class="text-lg">
                        {{ $t('home.disk') }}{{ $t('commons.table.name') }}: {{ diskInfo.device }}
                        <el-tag size="small" type="warning" v-if="scope === 'system'">
                            {{ $t('disk.systemDisk') }}
                        </el-tag>
                        <el-tag size="small" type="warning" v-if="scope === 'unpartitioned'">
                            {{ $t('disk.unpartitionedDisk') }}
                        </el-tag>
                    </h3>
                    <div class="flex items-center space-x-6 text-sm">
                        <el-text type="info">{{ $t('container.size') }}: {{ diskInfo.size }}</el-text>
                        <el-text type="info">
                            {{ $t('disk.partition') }}:
                            <span v-if="diskInfo.partitions">
                                {{ diskInfo.partitions?.length }}
                            </span>
                            <span v-else>0</span>
                        </el-text>
                        <el-text type="info" v-if="diskInfo.diskType" class="flex items-center">
                            {{ $t('disk.diskType') }}:
                            <el-tag class="ml-2" size="small" type="info">{{ diskInfo.diskType }}</el-tag>
                        </el-text>
                        <el-text type="info" v-if="diskInfo.model" class="flex items-center">
                            {{ $t('disk.model') }}:
                            <span class="ml-2">{{ diskInfo.model }}</span>
                        </el-text>
                        <el-text type="info" v-if="diskInfo.serial" class="flex items-center">
                            {{ $t('disk.model') }}:
                            <span class="ml-2">{{ diskInfo.serial }}</span>
                        </el-text>
                        <div v-if="scope == 'unpartitioned'">
                            <el-button type="primary" size="small" @click="handlePartition(diskInfo)">
                                {{ $t('disk.handlePartition') }}
                            </el-button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <div v-if="diskInfo.partitions && diskInfo.partitions.length > 0">
            <el-table :data="diskInfo.partitions" class="w-full">
                <el-table-column prop="device" :label="$t('disk.partition') + $t('commons.table.name')" min-width="100">
                    <template #default="{ row }">
                        <span class="font-medium">{{ row.device.split('/').pop() }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="size" :label="$t('container.size')" min-width="40" />
                <el-table-column prop="used" :label="$t('home.used')" min-width="40" />
                <el-table-column prop="avail" :label="$t('home.available')" min-width="40" />
                <el-table-column prop="usePercent" :label="$t('home.percent')" min-width="60">
                    <template #default="{ row }">
                        <el-progress
                            :percentage="row.usePercent"
                            :status="row.usePercent >= 90 ? 'exception' : 'success'"
                            :text-inside="true"
                            :stroke-width="14"
                        />
                    </template>
                </el-table-column>
                <el-table-column prop="mountPoint" :label="$t('disk.mountPoint')" min-width="120">
                    <template #default="{ row }">
                        <span v-if="row.mountPoint != ''">
                            {{ row.mountPoint }}
                        </span>
                        <el-tag v-else size="small" type="warning">{{ $t('disk.unmounted') }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="filesystem" :label="$t('disk.filesystem')" min-width="80">
                    <template #default="{ row }">
                        <el-tag size="small" type="info" v-if="row.filesystem != ''">{{ row.filesystem }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')" width="150">
                    <template #default="{ row }">
                        <el-text type="info" v-if="scope === 'system'">{{ $t('disk.cannotOperate') }}</el-text>
                        <el-button type="primary" link v-else-if="row.mountPoint != ''" @click="unmount(row)">
                            {{ $t('disk.unmount') }}
                        </el-button>
                        <el-button type="primary" link v-else @click="mount(row)">{{ $t('disk.mount') }}</el-button>
                    </template>
                </el-table-column>
            </el-table>
            <el-text v-if="scope === 'system'">{{ $t('disk.systemDiskHelper') }}</el-text>
        </div>
    </el-card>
</template>

<script lang="ts" setup>
import { Host } from '@/api/interface/host';
import i18n from '@/lang';
import { unmountDisk } from '@/api/modules/host';
import { MsgSuccess } from '@/utils/message';

const emit = defineEmits(['partition', 'search', 'mount']);

defineProps({
    diskInfo: {
        type: Object as () => Host.DiskInfo,
        required: true,
    },
    scope: {
        type: String,
        required: false,
    },
});

const handlePartition = (diskInfo: Host.DiskInfo) => {
    emit('partition', diskInfo);
};

const mount = (diskInfo: Host.DiskInfo) => {
    emit('mount', diskInfo);
};

const unmount = (diskInfo: Host.DiskInfo) => {
    ElMessageBox.confirm(i18n.global.t('disk.unmountHelper', [diskInfo.device]), i18n.global.t('disk.unmount'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        unmountDisk({
            mountPoint: diskInfo.mountPoint,
        }).then(() => {
            MsgSuccess(i18n.global.t('disk.unmount') + i18n.global.t('commons.status.success'));
            emit('search');
        });
    });
};
</script>
