<template>
    <ComplexTable :data="data" @search="search" v-loading="loading">
        <template #toolbar>
            <el-button type="primary" plain @click="openCreate">{{ $t('website.addDomain') }}</el-button>
        </template>
        <el-table-column width="30px">
            <template #default="{ row }">
                <el-button link :icon="Promotion" @click="openUrl(row)"></el-button>
            </template>
        </el-table-column>
        <el-table-column :label="$t('website.domain')" prop="domain"></el-table-column>
        <el-table-column :label="$t('commons.table.port')" prop="port"></el-table-column>
        <fu-table-operations
            :ellipsis="1"
            :buttons="buttons"
            :label="$t('commons.table.operate')"
            :fixed="mobile ? false : 'right'"
            fix
        />
    </ComplexTable>
    <Domain ref="domainRef" @close="search(id)"></Domain>
    <OpDialog ref="opRef" @search="search(id)" />
</template>

<script lang="ts" setup>
import Domain from './create/index.vue';
import { Website } from '@/api/interface/website';
import { DeleteDomain, GetWebsite, ListDomains } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import i18n from '@/lang';
import { Promotion } from '@element-plus/icons-vue';
import { GlobalStore } from '@/store';
import { CheckAppInstalled } from '@/api/modules/app';
const globalStore = GlobalStore();

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const id = computed(() => {
    return props.id;
});
const mobile = computed(() => {
    return globalStore.isMobile();
});
let loading = ref(false);
const data = ref<Website.Domain[]>([]);
const domainRef = ref();
const website = ref<Website.WebsiteDTO>();
const opRef = ref();
const httpPort = ref(80);
const httpsPort = ref(443);

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.Domain) {
            deleteDomain(row);
        },
        disabled: () => {
            return data.value.length == 1;
        },
    },
];

const openCreate = () => {
    domainRef.value.acceptParams(id.value);
};

const openUrl = (domain: Website.Domain) => {
    const protocol = website.value.protocol.toLowerCase();
    let url = protocol + '://' + domain.domain;
    if (protocol == 'http' && domain.port != 80) {
        url = url + ':' + domain.port;
    }
    if (protocol == 'https') {
        if (httpsPort.value != 443) {
            url = url + ':' + httpsPort.value;
        }
        if (domain.port != 80) {
            url = 'http://' + domain.domain + ':' + domain.port;
        }
    }
    window.open(url);
};

const deleteDomain = async (row: Website.Domain) => {
    opRef.value.acceptParams({
        title: i18n.global.t('commons.msg.deleteTitle'),
        names: [row.domain],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.domain'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: DeleteDomain,
        params: { id: row.id },
    });
};

const search = (id: number) => {
    loading.value = true;
    ListDomains(id)
        .then((res) => {
            data.value = res.data;
        })
        .finally(() => {
            loading.value = false;
        });
    onCheck();
};

const getWebsite = (id: number) => {
    GetWebsite(id).then((res) => {
        website.value = res.data;
    });
};

const onCheck = async () => {
    await CheckAppInstalled('openresty', '')
        .then((res) => {
            httpPort.value = res.data.httpPort;
            httpsPort.value = res.data.httpsPort;
        })
        .catch(() => {});
};

onMounted(() => {
    search(id.value);
    getWebsite(id.value);
});
</script>
