<template>
    <ComplexTable :data="data" @search="search" v-loading="loading">
        <template #toolbar>
            <el-button type="primary" plain @click="openCreate">{{ $t('website.addDomain') }}</el-button>
        </template>
        <el-table-column width="30px">
            <template #default="{ row }">
                <el-button link :icon="Promotion" @click="openUrl(row.domain, row.port)"></el-button>
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
</template>

<script lang="ts" setup>
import Domain from './create/index.vue';
import { Website } from '@/api/interface/website';
import { DeleteDomain, GetWebsite, ListDomains } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';
import { Promotion } from '@element-plus/icons-vue';
import { GlobalStore } from '@/store';
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

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.Domain) {
            deleteDoamin(row.id);
        },
        disabled: () => {
            return data.value.length == 1;
        },
    },
];

const openCreate = () => {
    domainRef.value.acceptParams(id.value);
};

const openUrl = (domain: string, port: string) => {
    let url = website.value.protocol.toLowerCase() + '://' + domain;
    if (port != '80') {
        url = url + ':' + port;
    }
    window.open(url);
};

const deleteDoamin = async (domainId: number) => {
    await useDeleteData(DeleteDomain, { id: domainId }, 'commons.msg.delete');
    search(id.value);
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
};

const getWebsite = (id: number) => {
    GetWebsite(id).then((res) => {
        website.value = res.data;
    });
};

onMounted(() => {
    search(id.value);
    getWebsite(id.value);
});
</script>
