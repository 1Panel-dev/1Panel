<template>
    <div>
        <RouterButton :buttons="routerButton" />
        <LayoutContent :title="$t('website.ssl')">
            <template #prompt>
                <el-alert type="info" :closable="false">
                    <template #default>
                        <span><span v-html="$t('website.encryptHelper')"></span></span>
                    </template>
                </el-alert>
            </template>
            <template #toolbar>
                <el-button type="primary" @click="openSSL()">
                    {{ $t('ssl.create') }}
                </el-button>
                <el-button type="primary" plain @click="openAcmeAccount()">
                    {{ $t('website.acmeAccountManage') }}
                </el-button>
                <el-button type="primary" plain @click="openDnsAccount()">
                    {{ $t('website.dnsAccountManage') }}
                </el-button>
            </template>
            <template #main>
                <br />
                <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()">
                    <el-table-column
                        :label="$t('website.domain')"
                        fix
                        show-overflow-tooltip
                        prop="primaryDomain"
                    ></el-table-column>
                    <el-table-column
                        :label="$t('website.otherDomains')"
                        fix
                        show-overflow-tooltip
                        prop="domains"
                    ></el-table-column>
                    <el-table-column :label="$t('ssl.provider')" fix show-overflow-tooltip prop="provider">
                        <template #default="{ row }">{{ getProvider(row.provider) }}</template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('ssl.acmeAccount')"
                        fix
                        show-overflow-tooltip
                        prop="acmeAccount.email"
                    ></el-table-column>
                    <el-table-column
                        :label="$t('website.brand')"
                        fix
                        show-overflow-tooltip
                        prop="organization"
                    ></el-table-column>
                    <el-table-column :label="$t('ssl.autoRenew')" fix>
                        <template #default="{ row }">
                            <el-switch
                                :disabled="row.provider === 'dnsManual' || row.provider === 'manual'"
                                v-model="row.autoRenew"
                                @change="updateConfig(row)"
                            />
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="expireDate"
                        :label="$t('website.expireDate')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                    <fu-table-operations
                        :ellipsis="3"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        :fixed="mobile ? false : 'right'"
                        fix
                    />
                </ComplexTable>
            </template>
            <DnsAccount ref="dnsAccountRef"></DnsAccount>
            <AcmeAccount ref="acmeAccountRef"></AcmeAccount>
            <Create ref="sslCreateRef" @close="search()"></Create>
            <Renew ref="renewRef" @close="search()"></Renew>
            <Detail ref="detailRef"></Detail>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref, computed } from 'vue';
import { DeleteSSL, SearchSSL, UpdateSSL } from '@/api/modules/website';
import DnsAccount from './dns-account/index.vue';
import AcmeAccount from './acme-account/index.vue';
import Renew from './renew/index.vue';
import Create from './create/index.vue';
import Detail from './detail/index.vue';
import { dateFormat, getProvider } from '@/utils/util';
import i18n from '@/lang';
import { Website } from '@/api/interface/website';
import { useDeleteData } from '@/hooks/use-delete-data';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const acmeAccountRef = ref();
const dnsAccountRef = ref();
const sslCreateRef = ref();
const renewRef = ref();
const detailRef = ref();
let data = ref();
let loading = ref(false);

const routerButton = [
    {
        label: i18n.global.t('website.ssl'),
        path: '/websites/ssl',
    },
];

const buttons = [
    {
        label: i18n.global.t('ssl.detail'),
        click: function (row: Website.SSL) {
            openDetail(row.id);
        },
    },
    {
        label: i18n.global.t('website.renewSSL'),
        disabled: function (row: Website.SSL) {
            return row.provider === 'manual';
        },
        click: function (row: Website.SSL) {
            openRenewSSL(row.id, row.websites);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.SSL) {
            deleteSSL(row.id);
        },
    },
];

const mobile = computed(() => {
    return globalStore.isMobile();
});

const search = () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    SearchSSL(req)
        .then((res) => {
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .finally(() => {
            loading.value = false;
        });
};

const updateConfig = (row: Website.SSL) => {
    loading.value = true;
    UpdateSSL({ id: row.id, autoRenew: row.autoRenew })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const openAcmeAccount = () => {
    acmeAccountRef.value.acceptParams();
};
const openDnsAccount = () => {
    dnsAccountRef.value.acceptParams();
};
const openSSL = () => {
    sslCreateRef.value.acceptParams();
};
const openRenewSSL = (id: number, websites: Website.Website[]) => {
    renewRef.value.acceptParams({ id: id, websites: websites });
};
const openDetail = (id: number) => {
    detailRef.value.acceptParams(id);
};

const deleteSSL = async (id: number) => {
    loading.value = true;
    await useDeleteData(DeleteSSL, { id: id }, 'commons.msg.delete');
    loading.value = false;
    search();
};

onMounted(() => {
    search();
});
</script>
