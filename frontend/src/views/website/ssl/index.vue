<template>
    <div>
        <RouterButton :buttons="routerButton" />
        <LayoutContent :title="$t('website.ssl', 2)">
            <template #leftToolBar>
                <el-button type="primary" @click="openSSL()">
                    {{ $t('ssl.create') }}
                </el-button>
                <el-button type="primary" @click="openUpload()">
                    {{ $t('ssl.upload') }}
                </el-button>
                <el-button type="primary" plain @click="openCA()">
                    {{ $t('ssl.selfSigned') }}
                </el-button>
                <el-button type="primary" plain @click="openAcmeAccount()">
                    {{ $t('website.acmeAccountManage') }}
                </el-button>
                <el-button type="primary" plain @click="openDnsAccount()">
                    {{ $t('website.dnsAccountManage') }}
                </el-button>
                <el-button plain @click="deletessl(null)" :disabled="selects.length === 0">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="req.domain" />
                <TableRefresh @search="search()" />
                <fu-table-column-select
                    :columns="columns"
                    trigger="hover"
                    :title="$t('commons.table.selectColumn')"
                    popper-class="popper-class"
                    :only-icon="true"
                />
            </template>
            <template #main>
                <ComplexTable
                    :data="data"
                    :pagination-config="paginationConfig"
                    @search="search()"
                    v-model:selects="selects"
                    v-loading="loading"
                    :columns="columns"
                    localKey="sslColumn"
                    :height-diff="260"
                >
                    <el-table-column type="selection" width="30" />
                    <el-table-column
                        :label="$t('website.domain')"
                        show-overflow-tooltip
                        prop="primaryDomain"
                        min-width="150px"
                    ></el-table-column>
                    <el-table-column
                        :label="$t('website.otherDomains')"
                        show-overflow-tooltip
                        prop="domains"
                        min-width="90px"
                    ></el-table-column>
                    <el-table-column :label="$t('ssl.applyType')" show-overflow-tooltip prop="provider" width="120px">
                        <template #default="{ row }">{{ getProvider(row.provider) }}</template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('ssl.acmeAccount')"
                        show-overflow-tooltip
                        prop="acmeAccount.email"
                        width="150px"
                    ></el-table-column>
                    <el-table-column
                        :label="$t('commons.table.status')"
                        show-overflow-tooltip
                        prop="status"
                        width="100px"
                    >
                        <template #default="{ row }">
                            <el-popover
                                v-if="
                                    row.status === 'error' ||
                                    row.status === 'applyError' ||
                                    row.status === 'systemRestart'
                                "
                                placement="bottom"
                                :width="400"
                                trigger="hover"
                            >
                                <template #reference>
                                    <Status :key="row.status" :status="row.status"></Status>
                                </template>
                                <div class="max-h-96 overflow-auto">
                                    <span>{{ row.message }}</span>
                                </div>
                            </el-popover>
                            <div v-else>
                                <Status :key="row.status" :status="row.status"></Status>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.button.log')" width="80px">
                        <template #default="{ row }">
                            <el-button
                                @click="openSSLLog(row)"
                                link
                                type="primary"
                                v-if="row.provider != 'manual' && row.provider !== 'fromMaster'"
                            >
                                {{ $t('website.check') }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('website.brand')"
                        show-overflow-tooltip
                        prop="organization"
                        width="150px"
                    ></el-table-column>
                    <el-table-column :label="$t('website.remark')" prop="description" width="100px">
                        <template #default="{ row }">
                            <fu-read-write-switch>
                                <template #read>
                                    <MsgInfo :info="row.description" width="200" />
                                </template>
                                <template #default="{ read }">
                                    <el-input v-model="row.description" @blur="updateDesc(row, read)" />
                                </template>
                            </fu-read-write-switch>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('ssl.autoRenew')" width="100px">
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
                        width="180px"
                    />
                    <fu-table-operations
                        :ellipsis="3"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        :fixed="mobile ? false : 'right'"
                        width="300px"
                        fix
                    />
                </ComplexTable>
            </template>
            <DnsAccount ref="dnsAccountRef" />
            <AcmeAccount ref="acmeAccountRef" />
            <Create ref="sslCreateRef" @close="search()" @submit="openLog" />
            <Detail ref="detailRef" />
            <SSLUpload ref="sslUploadRef" @close="search()" />
            <Apply ref="applyRef" @search="search" @submit="openLog" />
            <OpDialog ref="opRef" @search="search" @cancel="search" />
            <Log ref="logRef" @close="search()" :heightDiff="220" />
            <CA ref="caRef" @close="search()" />
            <Obtain ref="obtainRef" @close="search()" @submit="openLog" />
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref, computed } from 'vue';
import { deleteSSL, downloadFile, searchSSL, updateSSL } from '@/api/modules/website';
import DnsAccount from './dns-account/index.vue';
import AcmeAccount from './acme-account/index.vue';
import CA from './ca/index.vue';
import Create from './create/index.vue';
import Detail from './detail/index.vue';
import { dateFormat, getProvider } from '@/utils/util';
import i18n from '@/lang';
import { Website } from '@/api/interface/website';
import { MsgError, MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
import SSLUpload from './upload/index.vue';
import Apply from './apply/index.vue';
import Log from '@/components/log/file-drawer/index.vue';
import Obtain from './obtain/index.vue';
import MsgInfo from '@/components/msg-info/index.vue';

const globalStore = GlobalStore();
const paginationConfig = reactive({
    cacheSizeKey: 'ssl-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('ssl-page-size')) || 10,
    total: 0,
});
const acmeAccountRef = ref();
const dnsAccountRef = ref();
const sslCreateRef = ref();
const detailRef = ref();
const data = ref();
const loading = ref(false);
const opRef = ref();
const sslUploadRef = ref();
const applyRef = ref();
const logRef = ref();
const caRef = ref();
const obtainRef = ref();
let selects = ref<any>([]);
const columns = ref([]);
const req = reactive({
    domain: '',
});

const routerButton = [
    {
        label: i18n.global.t('website.ssl', 2),
        path: '/websites/ssl',
    },
];

const buttons = [
    {
        label: i18n.global.t('ssl.detail'),
        disabled: function (row: Website.SSLDTO) {
            return row.status === 'init' || row.status === 'error';
        },
        click: function (row: Website.SSLDTO) {
            openDetail(row.id);
        },
    },
    {
        label: i18n.global.t('ssl.apply'),
        disabled: function (row: Website.SSLDTO) {
            return row.status === 'applying' || row.provider === 'manual' || row.provider === 'fromMaster';
        },
        click: function (row: Website.SSLDTO) {
            if (row.provider === 'dnsManual') {
                applyRef.value.acceptParams({ ssl: row });
            } else {
                applySSL(row);
            }
        },
        show: function (row: Website.SSLDTO) {
            return row.provider != 'manual';
        },
    },
    {
        label: i18n.global.t('commons.button.update'),
        click: function (row: Website.SSLDTO) {
            sslUploadRef.value.acceptParams(row);
        },
        show: function (row: Website.SSLDTO) {
            return row.provider == 'manual';
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        disabled: function (row: Website.SSLDTO) {
            return row.provider === 'fromMaster';
        },
        click: function (row: Website.SSLDTO) {
            onEdit(row);
        },
        show: function (row: Website.SSLDTO) {
            return row.provider != 'manual';
        },
    },
    {
        label: i18n.global.t('commons.button.download'),
        click: function (row: Website.SSLDTO) {
            onDownload(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.SSLDTO) {
            deletessl(row);
        },
    },
];

const onDownload = (ssl: Website.SSLDTO) => {
    loading.value = true;
    downloadFile({ id: ssl.id })
        .then((res) => {
            const downloadUrl = window.URL.createObjectURL(new Blob([res]));
            const a = document.createElement('a');
            a.style.display = 'none';
            a.href = downloadUrl;
            a.download = ssl.primaryDomain + '.zip';
            const event = new MouseEvent('click');
            a.dispatchEvent(event);
        })
        .finally(() => {
            loading.value = false;
        });
};

const mobile = computed(() => {
    return globalStore.isMobile();
});

const search = () => {
    const request = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        domain: req.domain,
    };
    loading.value = true;
    searchSSL(request)
        .then((res) => {
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .finally(() => {
            loading.value = false;
        });
};

const updateDesc = (row: Website.SSLDTO, bulr: Function) => {
    bulr();
    if (row.description && row.description.length > 128) {
        MsgError(i18n.global.t('commons.rule.length128Err'));
        return;
    }
    updateConfig(row);
};

const updateConfig = (row: Website.SSLDTO) => {
    loading.value = true;
    updateSSL(row)
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
    sslCreateRef.value.acceptParams('create');
};
const onEdit = (row: Website.SSL) => {
    sslCreateRef.value.acceptParams('edit', row);
};

const openUpload = () => {
    sslUploadRef.value.acceptParams();
};
const openDetail = (id: number) => {
    detailRef.value.acceptParams(id);
};
const openLog = (id: number) => {
    logRef.value.acceptParams({ id: id, type: 'ssl', tail: true });
};
const openSSLLog = (row: Website.SSL) => {
    logRef.value.acceptParams({ id: row.id, type: 'ssl', tail: row.status === 'applying' });
};

const openCA = () => {
    caRef.value.acceptParams();
};

const applySSL = (row: Website.SSLDTO) => {
    obtainRef.value.acceptParams({ ssl: row });
};

const deletessl = async (row: any) => {
    let names = [];
    let params = {};
    if (row == null) {
        names = selects.value.map((item: Website.SSLDTO) => item.primaryDomain);
        params = { ids: selects.value.map((item: Website.SSLDTO) => item.id) };
    } else {
        names = [row.primaryDomain];
        params = { ids: [row.id] };
    }

    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.ssl'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteSSL,
        params: params,
    });
    search();
};

onMounted(() => {
    search();
});
</script>
