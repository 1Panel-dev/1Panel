<template>
    <el-card class="disk-card shadow-sm">
        <div class="border-b pb-4">
            <div class="flex items-center space-x-4">
                <div>
                    <h3 class="disk-title text-lg">
                        <span class="disk-title__name">
                            {{ $t('home.disk') }}{{ $t('commons.table.name') }}: {{ diskInfo.device }}
                        </span>
                        <el-tag size="small" type="warning" v-if="scope === 'system'">
                            {{ $t('disk.systemDisk') }}
                        </el-tag>
                        <el-tag
                            size="small"
                            type="warning"
                            v-if="
                                scope == 'unpartitioned' &&
                                diskInfo.partitions == undefined &&
                                diskInfo.mountPoint == ''
                            "
                        >
                            {{ $t('disk.unpartitionedDisk') }}
                        </el-tag>
                    </h3>
                    <div class="disk-summary text-sm">
                        <el-text type="info" class="disk-summary__item">
                            {{ $t('container.size') }}: {{ diskInfo.size }}
                        </el-text>
                        <el-text type="info" class="disk-summary__item">
                            {{ $t('disk.partition') }}:
                            <span v-if="diskInfo.partitions">
                                {{ diskInfo.partitions?.length }}
                            </span>
                            <span v-else>0</span>
                        </el-text>
                        <el-text type="info" v-if="diskInfo.diskType" class="disk-summary__item">
                            {{ $t('disk.diskType') }}:
                            <el-tag class="ml-2" size="small" type="info">{{ diskInfo.diskType }}</el-tag>
                        </el-text>
                        <el-text type="info" v-if="diskInfo.model" class="disk-summary__item">
                            {{ $t('disk.model') }}:
                            <span class="disk-summary__value ml-2">{{ diskInfo.model }}</span>
                        </el-text>
                        <el-text type="info" v-if="diskInfo.serial" class="disk-summary__item">
                            {{ $t('disk.serial') }}:
                            <span class="disk-summary__value ml-2">{{ diskInfo.serial }}</span>
                        </el-text>
                        <div
                            v-if="
                                scope == 'unpartitioned' &&
                                diskInfo.partitions == undefined &&
                                diskInfo.mountPoint == ''
                            "
                        >
                            <el-button
                                v-permission
                                v-node-admin
                                type="primary"
                                size="small"
                                @click="handlePartition(diskInfo)"
                            >
                                {{ $t('disk.handlePartition') }}
                            </el-button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <div v-if="diskInfo.partitions && diskInfo.partitions.length > 0">
            <el-table :data="diskInfo.partitions" class="w-full" :scrollbar-always-on="isMobile">
                <el-table-column
                    prop="device"
                    :label="$t('disk.partition') + $t('commons.table.name')"
                    :min-width="columnMinWidth(100, 120)"
                >
                    <template #default="{ row }">
                        <span class="font-medium">{{ row.device.split('/').pop() }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="size" :label="$t('container.size')" :min-width="columnMinWidth(40, 80)" />
                <el-table-column prop="used" :label="$t('home.used')" :min-width="columnMinWidth(40, 80)" />
                <el-table-column prop="avail" :label="$t('home.available')" :min-width="columnMinWidth(40, 80)" />
                <el-table-column prop="usePercent" :label="$t('home.percent')" :min-width="columnMinWidth(60, 120)">
                    <template #default="{ row }">
                        <el-progress
                            :percentage="row.usePercent"
                            :status="row.usePercent >= 90 ? 'exception' : 'success'"
                            :text-inside="true"
                            :stroke-width="14"
                        />
                    </template>
                </el-table-column>
                <el-table-column prop="mountPoint" :label="$t('disk.mountPoint')" :min-width="columnMinWidth(120, 160)">
                    <template #default="{ row }">
                        <span v-if="row.mountPoint != ''">
                            {{ row.mountPoint }}
                        </span>
                        <el-tag v-else size="small" type="warning">{{ $t('disk.unmounted') }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="filesystem" :label="$t('disk.filesystem')" :min-width="columnMinWidth(80, 100)">
                    <template #default="{ row }">
                        <el-tag size="small" type="info" v-if="row.filesystem != ''">{{ row.filesystem }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')" width="150">
                    <template #default="{ row }">
                        <el-text type="info" v-if="scope === 'system'">{{ $t('disk.cannotOperate') }}</el-text>
                        <el-button
                            v-permission
                            v-node-admin
                            type="primary"
                            link
                            v-else-if="row.mountPoint != ''"
                            @click="unmount(row)"
                        >
                            {{ $t('disk.unmount') }}
                        </el-button>
                        <el-button v-permission v-node-admin type="primary" link v-else @click="mount(row)">
                            {{ $t('disk.mount') }}
                        </el-button>
                    </template>
                </el-table-column>
            </el-table>
            <el-text v-if="scope === 'system'">{{ $t('disk.systemDiskHelper') }}</el-text>
        </div>
        <div v-if="diskInfo.partitions == undefined && diskInfo.mountPoint != ''">
            <el-table :data="[diskInfo]" class="w-full" :scrollbar-always-on="isMobile">
                <el-table-column
                    prop="device"
                    :label="$t('disk.partition') + $t('commons.table.name')"
                    :min-width="columnMinWidth(100, 120)"
                >
                    <template #default="{ row }">
                        <span class="font-medium">{{ row.device.split('/').pop() }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="size" :label="$t('container.size')" :min-width="columnMinWidth(40, 80)" />
                <el-table-column prop="used" :label="$t('home.used')" :min-width="columnMinWidth(40, 80)" />
                <el-table-column prop="avail" :label="$t('home.available')" :min-width="columnMinWidth(40, 80)" />
                <el-table-column prop="usePercent" :label="$t('home.percent')" :min-width="columnMinWidth(60, 120)">
                    <template #default="{ row }">
                        <el-progress
                            :percentage="row.usePercent"
                            :status="row.usePercent >= 90 ? 'exception' : 'success'"
                            :text-inside="true"
                            :stroke-width="14"
                        />
                    </template>
                </el-table-column>
                <el-table-column prop="mountPoint" :label="$t('disk.mountPoint')" :min-width="columnMinWidth(120, 160)">
                    <template #default="{ row }">
                        <span v-if="row.mountPoint != ''">
                            {{ row.mountPoint }}
                        </span>
                        <el-tag v-else size="small" type="warning">{{ $t('disk.unmounted') }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="filesystem" :label="$t('disk.filesystem')" :min-width="columnMinWidth(80, 100)">
                    <template #default="{ row }">
                        <el-tag size="small" type="info" v-if="row.filesystem != ''">{{ row.filesystem }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')" width="150">
                    <template #default="{ row }">
                        <el-text type="info" v-if="scope === 'system'">{{ $t('disk.cannotOperate') }}</el-text>
                        <el-button
                            v-permission
                            v-node-admin
                            type="primary"
                            link
                            v-else-if="row.mountPoint != ''"
                            @click="unmount(row)"
                        >
                            {{ $t('disk.unmount') }}
                        </el-button>
                        <el-button v-permission v-node-admin type="primary" link v-else @click="mount(row)">
                            {{ $t('disk.mount') }}
                        </el-button>
                    </template>
                </el-table-column>
            </el-table>
        </div>
    </el-card>
</template>

<script lang="ts" setup>
import { Host } from '@/api/interface/host';
import i18n from '@/lang';
import { unmountDisk } from '@/api/modules/host';
import { MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';

const emit = defineEmits(['partition', 'search', 'mount']);
const { isMobile } = useGlobalStore();

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

const columnMinWidth = (desktop: number, mobile: number) => (isMobile.value ? mobile : desktop);

const mount = (diskInfo: Host.DiskInfo) => {
    emit('mount', diskInfo);
};

const unmount = (diskInfo: Host.DiskInfo) => {
    ElMessageBox.confirm(
        i18n.global.t('disk.unmountHelper', [diskInfo.device.split('/').pop()]),
        i18n.global.t('disk.unmount'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        },
    ).then(async () => {
        unmountDisk({
            mountPoint: diskInfo.mountPoint,
        }).then(() => {
            MsgSuccess(i18n.global.t('disk.unmount') + i18n.global.t('commons.status.success'));
            emit('search');
        });
    });
};
</script>

<style scoped lang="scss">
.disk-title {
    display: flex;
    min-width: 0;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
}

.disk-title__name,
.disk-summary__value {
    min-width: 0;
    overflow-wrap: anywhere;
}

.disk-summary {
    display: flex;
    min-width: 0;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px 24px;
}

.disk-summary__item {
    display: inline-flex;
    min-width: 0;
    max-width: 100%;
    align-items: center;
}

@media only screen and (max-width: 767px) {
    .disk-card {
        --el-card-padding: 12px;
    }

    .disk-summary {
        align-items: flex-start;
        gap: 8px 12px;
    }

    .disk-summary__item {
        flex: 1 1 140px;
    }
}
</style>
