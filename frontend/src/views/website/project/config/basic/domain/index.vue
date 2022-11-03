<template>
    <ComplexTable :data="data" @search="search" v-loading="loading">
        <template #toolbar>
            <el-button type="primary" plain>{{ $t('website.addDomain') }}</el-button>
        </template>
        <el-table-column :label="$t('website.domain')" prop="domain"></el-table-column>
        <el-table-column :label="$t('website.port')" prop="port"></el-table-column>
        <fu-table-operations :ellipsis="1" :buttons="buttons" :label="$t('commons.table.operate')" fixed="right" fix />
    </ComplexTable>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { WebSite } from '@/api/interface/website';
import { DeleteDomain, ListDomains } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});

let loading = ref(false);
const data = ref<WebSite.Domain[]>([]);

const buttons = [
    {
        label: i18n.global.t('app.delete'),
        click: function (row: WebSite.Domain) {
            deleteDoamin(row.id);
        },
    },
];

const deleteDoamin = async (domainId: number) => {
    await useDeleteData(DeleteDomain, domainId, 'commons.msg.delete', loading.value);
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

onMounted(() => search(id.value));
</script>
