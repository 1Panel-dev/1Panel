<template>
    <div>
        <RouterButton :buttons="routerButton" />
        <LayoutContent :title="$t('website.ssl', 2)">
            <template #toolbar>
                <div class="flex flex-wrap gap-3">
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
                    <el-button plain @click="deleteSSL(null)" :disabled="selects.length === 0">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </div>
            </template>
            <template #main>
                <br />
                <ComplexTable
                    :data="data"
                    :pagination-config="paginationConfig"
                    @search="search()"
                    v-model:selects="selects"
                    v-loading="loading"
                >
                    <el-table-column type="selection" width="30" />
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
                        min-width="150px"
                    ></el-table-column>
                    <el-table-column
                        :label="$t('ssl.applyType')"
                        fix
                        show-overflow-tooltip
                        prop="provider"
                        min-width="180px"
                    >
                        <template #default="{ row }">{{ getProvider(row.provider) }}</template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('ssl.acmeAccount')"
                        fix
                        show-overflow-tooltip
                        prop="acmeAccount.email"
                        min-width="180px"
                    ></el-table-column>
                    <el-table-column
                        :label="$t('commons.table.status')"
                        fix
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
                    <el-table-column :label="$t('website.log')" width="100px">
                        <template #default="{ row }">
                            <el-button @click="openSSLLog(row)" link type="primary" v-if="row.provider != 'manual'">
                                {{ $t('website.check') }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('website.brand')"
                        fix
                        show-overflow-tooltip
                        prop="organization"
                        min-width="110px"
                    ></el-table-column>
                    <el-table-column :label="$t('website.remark')" fix prop="description" min-width="100px">
                        <template #default="{ row }">
                            <fu-read-write-switch>
                                <template #read>
                                    <MsgInfo :info="row.description" />
                                </template>
                                <template #default="{ read }">
                                    <el-input v-model="row.description" @blur="updateDesc(row, read)" />
                                </template>
                            </fu-read-write-switch>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('ssl.autoRenew')" fix min-width="110px">
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
                        width="200px"
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
            <Log ref="logRef" @close="search()" />
            <CA ref="caRef" @close="search()" />
            <Obtain ref="obtainRef" @close="search()" @submit="openLog" />
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref, computed } from 'vue';
import { DeleteSSL, DownloadFile, SearchSSL, UpdateSSL } from '@/api/modules/website';
import DnsAccount from './dns-account/index.vue';
import AcmeAccount from './acme-account/index.vue';
import CA from './ca/index.vue';
import Create from './create/index.vue';
import Detail from './detail/index.vue';
import { dateFormat, getProvider } from '@/utils/util';
import i18n from '@/lang';
import { Website } from '@/api/interface/website';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
import SSLUpload from './upload/index.vue';
import Apply from './apply/index.vue';
import Log from '@/components/log-dialog/index.vue';
import Obtain from './obtain/index.vue';
import MsgInfo from '@/components/msg-info/index.vue';

const globalStore = GlobalStore();
const paginationConfig = reactive({
    cacheSizeKey: 'ssl-page-size',
    currentPage: 1,
    pageSize: 10,
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
            return row.status === 'applying' || row.provider === 'manual';
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
        label: i18n.global.t('commons.operate.update'),
        click: function (row: Website.SSLDTO) {
            sslUploadRef.value.acceptParams(row);
        },
        show: function (row: Website.SSLDTO) {
            return row.provider == 'manual';
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Website.SSLDTO) {
            onEdit(row);
        },
        show: function (row: Website.SSLDTO) {
            return row.provider != 'manual';
        },
    },
    {
        label: i18n.global.t('file.download'),
        click: function (row: Website.SSLDTO) {
            onDownload(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.SSLDTO) {
            deleteSSL(row);
        },
    },
];

const onDownload = (ssl: Website.SSLDTO) => {
    loading.value = true;
    DownloadFile({ id: ssl.id })
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

const updateDesc = (row: Website.SSLDTO, bulr: Function) => {
    bulr();
    updateConfig(row);
};

const updateConfig = (row: Website.SSLDTO) => {
    loading.value = true;
    UpdateSSL(row)
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

const deleteSSL = async (row: any) => {
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
        api: DeleteSSL,
        params: params,
    });
    search();
};

onMounted(() => {
    search();
});
</script>
