<template>
    <DrawerPro v-model="visible" :header="$t('commons.button.view')" size="large" @close="handleClose">
        <el-tabs v-model="activeTab" type="border-card">
            <el-tab-pane :label="$t('container.network')" name="overview">
                <el-descriptions :column="1" border :title="$t('home.baseInfo')">
                    <el-descriptions-item :label="$t('commons.table.name')">
                        <el-text>{{ networkData?.Name || '-' }}</el-text>
                        <CopyButton :content="networkData?.Name || ''" />
                    </el-descriptions-item>
                    <el-descriptions-item label="ID">
                        <el-text class="text-xs">{{ networkData?.Id?.substring(0, 12) }}</el-text>
                        <CopyButton :content="networkData?.Id || ''" />
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('container.driver')">
                        {{ networkData?.Driver || '-' }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('commons.table.createdAt')">
                        {{ formatDate(networkData?.Created) }}
                    </el-descriptions-item>
                    <el-descriptions-item label="IPv4">
                        <el-tag :type="networkData?.EnableIPv4 ? 'success' : 'info'">
                            {{ networkData?.EnableIPv4 ? $t('commons.status.enable') : $t('commons.status.disable') }}
                        </el-tag>
                    </el-descriptions-item>
                    <el-descriptions-item label="IPv6">
                        <el-tag :type="networkData?.EnableIPv6 ? 'success' : 'info'">
                            {{ networkData?.EnableIPv6 ? $t('commons.status.enable') : $t('commons.status.disable') }}
                        </el-tag>
                    </el-descriptions-item>
                </el-descriptions>

                <el-descriptions class="mt-4" :column="1" border title="IPAM">
                    <el-descriptions-item :label="$t('container.driver')">
                        {{ networkData?.IPAM?.Driver || '-' }}
                    </el-descriptions-item>
                    <el-descriptions-item
                        v-for="(config, index) in networkData?.IPAM?.Config"
                        :key="index"
                        :label="$t('container.subnet') + (index > 0 ? ' ' + (index + 1) : '')"
                    >
                        <div v-if="config">
                            <div v-if="config.Subnet">
                                <el-text class="mr-2">{{ $t('container.subnet') }}:</el-text>
                                <el-tag>{{ config.Subnet }}</el-tag>
                            </div>
                            <div v-if="config.Gateway" class="mt-1">
                                <el-text class="mr-2">{{ $t('container.gateway') }}:</el-text>
                                <el-tag>{{ config.Gateway }}</el-tag>
                            </div>
                            <div v-if="config.IPRange" class="mt-1">
                                <el-text class="mr-2">{{ $t('container.scope') }}:</el-text>
                                <el-tag>{{ config.IPRange }}</el-tag>
                            </div>
                        </div>
                        <span v-else>-</span>
                    </el-descriptions-item>
                </el-descriptions>

                <div class="mt-4">
                    <span class="block text-base font-medium text-el-color-primary">{{ $t('menu.container') }}</span>
                    <el-table :data="containerList" border class="mt-2">
                        <el-table-column :label="$t('commons.table.name')" min-width="120">
                            <template #default="{ row }">
                                <el-text>{{ row.name }}</el-text>
                                <CopyButton :content="row.name" />
                            </template>
                        </el-table-column>
                        <el-table-column label="IPv4" min-width="100">
                            <template #default="{ row }">
                                {{ row.ipv4 || '-' }}
                            </template>
                        </el-table-column>
                        <el-table-column label="IPv6" min-width="100">
                            <template #default="{ row }">
                                {{ row.ipv6 || '-' }}
                            </template>
                        </el-table-column>
                        <el-table-column label="MAC" min-width="120">
                            <template #default="{ row }">
                                {{ row.mac || '-' }}
                            </template>
                        </el-table-column>
                        <el-table-column label="Endpoint ID" width="120">
                            <template #default="{ row }">
                                <el-text class="text-xs">{{ row.endpointId?.substring(0, 12) || '-' }}</el-text>
                            </template>
                        </el-table-column>
                    </el-table>
                </div>
            </el-tab-pane>

            <el-tab-pane :label="$t('commons.button.view')" name="raw">
                <CodemirrorPro v-model="rawJson" :height-diff="240" :disabled="true" mode="json" />
            </el-tab-pane>
        </el-tabs>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';

const visible = ref(false);
const activeTab = ref('overview');
const networkData = ref<any>(null);
const rawJson = ref('');

interface DialogProps {
    data: any;
}

const acceptParams = (props: DialogProps): void => {
    visible.value = true;
    activeTab.value = 'overview';

    try {
        if (typeof props.data === 'string') {
            networkData.value = JSON.parse(props.data);
        } else {
            networkData.value = props.data;
        }
        rawJson.value = JSON.stringify(networkData.value, null, 2);
    } catch (e) {
        console.error('Failed to parse network data:', e);
    }
};

const handleClose = () => {
    visible.value = false;
    networkData.value = null;
    rawJson.value = '';
};

const formatDate = (dateStr: string): string => {
    if (!dateStr || dateStr === '0001-01-01T00:00:00Z') {
        return '-';
    }
    try {
        const date = new Date(dateStr);
        const y = date.getFullYear();
        const m = String(date.getMonth() + 1).padStart(2, '0');
        const d = String(date.getDate()).padStart(2, '0');
        const h = String(date.getHours()).padStart(2, '0');
        const minute = String(date.getMinutes()).padStart(2, '0');
        const second = String(date.getSeconds()).padStart(2, '0');
        return `${y}-${m}-${d} ${h}:${minute}:${second}`;
    } catch {
        return dateStr;
    }
};

const containerList = computed(() => {
    if (!networkData.value?.Containers) {
        return [];
    }

    return Object.entries(networkData.value.Containers).map(([id, container]: [string, any]) => ({
        id,
        name: container.Name || '-',
        ipv4: container.IPv4Address || '',
        ipv6: container.IPv6Address || '',
        mac: container.MacAddress || '',
        endpointId: container.EndpointID || '',
    }));
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
:deep(.el-descriptions__label) {
    width: 180px;
    background-color: transparent !important;
}
.text-el-color-primary {
    color: var(--el-text-color-primary);
}
</style>
