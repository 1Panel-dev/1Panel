<template>
    <div v-loading="loading">
        <div v-if="!dnsStatus.enabled" class="mt-5">
            <LayoutContent :title="$t('menu.dns')">
                <template #main>
                    <div class="flex items-center justify-center" style="height: 300px">
                        <div class="text-center">
                            <el-icon :size="60" color="#409eff"><Connection /></el-icon>
                            <p class="mt-4 mb-2">{{ $t('dns.setupHelper') }}</p>
                            <el-button type="primary" @click="onSetup" :loading="setupLoading">
                                {{ $t('dns.setup') }}
                            </el-button>
                        </div>
                    </div>
                </template>
            </LayoutContent>
        </div>

        <div v-else>
            <LayoutContent :title="$t('menu.dns')">
                <template #leftToolBar>
                    <el-button type="primary" @click="onCreate">
                        {{ $t('dns.createZone') }}
                    </el-button>
                </template>
                <template #rightToolBar>
                    <el-tag v-if="dnsStatus.containerUp" type="success" class="mr-2">
                        {{ $t('dns.pdnsRunning') }}
                    </el-tag>
                    <el-tag v-else type="danger" class="mr-2">
                        {{ $t('dns.pdnsNotRunning') }}
                    </el-tag>
                    <TableSearch @search="search()" v-model:searchName="searchName" />
                    <TableRefresh @search="search()" />
                </template>
                <template #main>
                    <ComplexTable
                        :pagination-config="paginationConfig"
                        @sort-change="search"
                        @search="search"
                        :data="data"
                    >
                        <el-table-column :label="$t('dns.zoneName')" prop="name" sortable min-width="180">
                            <template #default="{ row }">
                                <el-link type="primary" @click="onDetail(row)">{{ row.name }}</el-link>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('dns.zoneType')" prop="type" min-width="100">
                            <template #default="{ row }">
                                <el-tag v-if="row.type === 'Native'" type="primary">{{ $t('dns.native') }}</el-tag>
                                <el-tag v-else-if="row.type === 'Slave'" type="warning">{{ $t('dns.slave') }}</el-tag>
                                <el-tag v-else>{{ row.type }}</el-tag>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('dns.recordCount')" prop="recordCount" min-width="80" />
                        <el-table-column :label="$t('dns.status')" prop="status" min-width="80">
                            <template #default="{ row }">
                                <el-tag v-if="row.status === 'active'" type="success">{{ $t('dns.enabled') }}</el-tag>
                                <el-tag v-else type="info">{{ $t('dns.disabled') }}</el-tag>
                            </template>
                        </el-table-column>
                        <el-table-column
                            :label="$t('commons.table.createdAt')"
                            prop="createdAt"
                            sortable
                            min-width="150"
                        >
                            <template #default="{ row }">
                                {{ dateFormat(0, 0, row.createdAt) }}
                            </template>
                        </el-table-column>
                        <fu-table-operations
                            :ellipsis="10"
                            :buttons="buttons"
                            :label="$t('commons.table.operate')"
                            fixed="right"
                            fix
                        />
                    </ComplexTable>
                </template>
            </LayoutContent>
        </div>

        <CreateZone ref="createRef" @search="search()" />
        <OpDialog ref="opRef" @search="search()" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { searchDnsZones, getDnsStatus, setupDns, deleteDnsZone } from '@/api/modules/dns';
import { Dns } from '@/api/interface/dns';
import { dateFormat } from '@/utils/util';
import CreateZone from './create.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const router = useRouter();
const loading = ref(false);
const setupLoading = ref(false);
const data = ref<Dns.ZoneRes[]>([]);
const searchName = ref('');
const createRef = ref();
const opRef = ref();

const dnsStatus = ref<Dns.Status>({
    enabled: false,
    containerUp: false,
    containerName: '',
    version: '',
});

const paginationConfig = reactive({
    cacheSizeKey: 'dns-zone-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('dns-zone-page-size')) || 20,
    total: 0,
    orderBy: 'createdAt',
    order: 'null',
});

const buttons = [
    {
        label: i18n.global.t('dns.viewRecords'),
        click: (row: Dns.ZoneRes) => {
            onDetail(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Dns.ZoneRes) => {
            onDelete(row);
        },
    },
];

const search = async () => {
    const req: Dns.ZoneSearch = {
        info: searchName.value || '',
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
    };
    loading.value = true;
    try {
        const res = await searchDnsZones(req);
        data.value = res.data.items || [];
        paginationConfig.total = res.data.total;
    } finally {
        loading.value = false;
    }
};

const loadStatus = async () => {
    try {
        const res = await getDnsStatus();
        dnsStatus.value = res.data;
    } catch (e) {
        dnsStatus.value.enabled = false;
    }
};

const onSetup = async () => {
    setupLoading.value = true;
    try {
        await setupDns({ enable: true });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await loadStatus();
        if (dnsStatus.value.enabled) {
            search();
        }
    } finally {
        setupLoading.value = false;
    }
};

const onCreate = () => {
    createRef.value.acceptParams();
};

const onDetail = (row: Dns.ZoneRes) => {
    router.push({ name: 'DnsZoneDetail', params: { id: row.id } });
};

const onDelete = (row: Dns.ZoneRes) => {
    opRef.value.acceptParams({
        title: i18n.global.t('dns.deleteZone'),
        names: [row.name],
        msg: i18n.global.t('dns.deleteZoneHelper'),
        api: deleteDnsZone,
        params: { id: row.id },
    });
};

onMounted(async () => {
    await loadStatus();
    if (dnsStatus.value.enabled) {
        search();
    }
});
</script>
