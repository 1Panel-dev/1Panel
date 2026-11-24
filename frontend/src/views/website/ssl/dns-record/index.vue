<template>
    <div>
        <RouterButton :buttons="routerButton" />
        <LayoutContent backName="SSL" :title="$t('website.dnsRecordManage')">
            <template #leftToolBar>
                <el-select
                    v-model="accountId"
                    class="w-56 mr-2"
                    :placeholder="$t('website.dnsRecordAccount')"
                    @change="loadZones"
                >
                    <el-option v-for="a in accounts" :key="a.id" :label="`${a.name} (${a.type})`" :value="a.id" />
                </el-select>
                <el-select
                    v-model="zoneId"
                    class="w-56 mr-2"
                    filterable
                    :placeholder="$t('website.dnsRecordZone')"
                    :disabled="!accountId"
                    @change="loadRecords"
                >
                    <el-option v-for="z in zones" :key="z.id" :label="z.name" :value="z.id" />
                </el-select>
                <el-button v-permission type="primary" :disabled="!zoneId" @click="openCreate">
                    {{ $t('website.dnsRecordCreate') }}
                </el-button>
            </template>
            <template #main>
                <el-alert
                    v-if="loaded && accounts.length === 0"
                    :title="$t('website.dnsRecordNoAccount')"
                    type="warning"
                    :closable="false"
                    class="mb-2"
                />
                <ComplexTable
                    v-loading="loading"
                    :data="pagedRecords"
                    :cache-pagination="false"
                    :pagination-config="paginationConfig"
                >
                    <el-table-column :label="$t('commons.table.type')" prop="type" width="90" />
                    <el-table-column
                        :label="$t('website.dnsRecordHost')"
                        prop="name"
                        min-width="140"
                        show-overflow-tooltip
                    />
                    <el-table-column
                        :label="$t('website.dnsRecordValue')"
                        prop="value"
                        min-width="240"
                        show-overflow-tooltip
                    />
                    <el-table-column :label="$t('website.dnsRecordTTL')" prop="ttl" width="120" />
                    <el-table-column :label="$t('website.dnsRecordPriority')" prop="priority" width="100">
                        <template #default="{ row }">
                            <span>{{ row.type === 'MX' ? row.priority : '-' }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        v-if="accountType === 'CloudFlare'"
                        :label="$t('website.dnsRecordProxied')"
                        prop="proxied"
                        width="100"
                    >
                        <template #default="{ row }">
                            <span>{{ row.proxied ? $t('commons.status.enabled') : '-' }}</span>
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        :ellipsis="1"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <Create ref="createRef" @saved="loadRecords" />
        <OpDialog drawer ref="opRef" @search="loadRecords" />
    </div>
</template>

<script lang="ts" setup>
import Create from './create/index.vue';
import { Website } from '@/api/interface/website';
import { deleteDnsRecord, searchDnsAccount, searchDnsRecords, searchDnsZones } from '@/api/modules/website';
import { computed, onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';

// Providers that support DNS record management (keep in sync with agent/utils/dns).
const SUPPORTED_TYPES = ['CloudFlare', 'AliYun', 'AWSRoute53', 'TencentCloud'];

const accounts = ref<Website.DnsAccount[]>([]);
const zones = ref<Website.DnsZone[]>([]);
const records = ref<Website.DnsRecord[]>([]);
const accountId = ref<number | undefined>();
const zoneId = ref<string | undefined>();
const loading = ref(false);
const loaded = ref(false);
const createRef = ref();
const opRef = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});

const accountType = computed(() => accounts.value.find((a) => a.id === accountId.value)?.type || '');
const currentZone = computed(() => zones.value.find((z) => z.id === zoneId.value));

const pagedRecords = computed(() => {
    const start = (paginationConfig.currentPage - 1) * paginationConfig.pageSize;
    return records.value.slice(start, start + paginationConfig.pageSize);
});

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        click: function (row: Website.DnsRecord) {
            openEdit(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        click: function (row: Website.DnsRecord) {
            deleteRecord(row);
        },
    },
];

const loadAccounts = async () => {
    const res = await searchDnsAccount({ page: 1, pageSize: 500 });
    accounts.value = (res.data.items || []).filter((a: Website.DnsAccount) => SUPPORTED_TYPES.includes(a.type));
    loaded.value = true;
    if (accounts.value.length > 0) {
        accountId.value = accounts.value[0].id;
        await loadZones();
    }
};

const loadZones = async () => {
    zones.value = [];
    records.value = [];
    zoneId.value = undefined;
    paginationConfig.total = 0;
    if (!accountId.value) return;
    loading.value = true;
    try {
        const res = await searchDnsZones({ accountId: accountId.value });
        zones.value = res.data || [];
        if (zones.value.length > 0) {
            zoneId.value = zones.value[0].id;
            await loadRecords();
        }
    } finally {
        loading.value = false;
    }
};

const loadRecords = async () => {
    if (!accountId.value || !currentZone.value) return;
    loading.value = true;
    try {
        const res = await searchDnsRecords({
            accountId: accountId.value,
            zoneId: currentZone.value.id,
            zoneName: currentZone.value.name,
        });
        records.value = res.data || [];
        paginationConfig.total = records.value.length;
        const lastPage = Math.max(1, Math.ceil(paginationConfig.total / paginationConfig.pageSize));
        if (paginationConfig.currentPage > lastPage) {
            paginationConfig.currentPage = lastPage;
        }
    } finally {
        loading.value = false;
    }
};

const openCreate = () => {
    if (!accountId.value || !currentZone.value) return;
    createRef.value.acceptParams({
        mode: 'create',
        accountId: accountId.value,
        accountType: accountType.value,
        zone: currentZone.value,
    });
};

const openEdit = (row: Website.DnsRecord) => {
    if (!accountId.value || !currentZone.value) return;
    createRef.value.acceptParams({
        mode: 'edit',
        accountId: accountId.value,
        accountType: accountType.value,
        zone: currentZone.value,
        form: row,
    });
};

const deleteRecord = (row: Website.DnsRecord) => {
    if (!accountId.value || !currentZone.value) return;
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [`${row.type} ${row.name}`],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.dnsRecordManage'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteDnsRecord,
        params: {
            accountId: accountId.value,
            zoneId: currentZone.value.id,
            zoneName: currentZone.value.name,
            id: row.id,
            name: row.name,
            type: row.type,
        },
    });
};

onMounted(() => {
    loadAccounts();
});

const routerButton = [{ label: i18n.global.t('website.dnsRecordManage'), path: '/websites/ssl/dns-record' }];
</script>
