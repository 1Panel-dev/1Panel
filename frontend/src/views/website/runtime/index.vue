<template>
    <div>
        <RouterButton
            :buttons="[
                {
                    label: 'PHP',
                    path: '/runtimes',
                },
            ]"
        />
        <LayoutContent :title="$t('runtime.runtime')" v-loading="loading">
            <template #toolbar></template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="items" @search="search()">
                    <el-table-column :label="$t('commons.table.name')" fix prop="name" min-width="120px">
                        <template #default="{ row }">
                            <Tooltip :text="row.name" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('runtime.image')" prop="image"></el-table-column>
                    <el-table-column :label="$t('runtime.workDir')" prop="workDir"></el-table-column>
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { Runtime } from '@/api/interface/runtime';
import { SearchRuntimes } from '@/api/modules/runtime';
import RouterButton from '@/components/router-button/index.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import LayoutContent from '@/layout/layout-content.vue';
import { dateFormat } from '@/utils/util';

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 15,
    total: 0,
});
let req = reactive<Runtime.RuntimeReq>({
    name: '',
    page: 1,
    pageSize: 15,
});
const loading = ref(false);
const items = ref<Runtime.RuntimeDTO[]>([]);

const search = async () => {
    req.page = paginationConfig.currentPage;
    req.pageSize = paginationConfig.pageSize;
    loading.value = true;
    try {
        const res = await SearchRuntimes(req);
        items.value = res.data.items;
        paginationConfig.total = res.data.total;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

onMounted(() => {
    search();
});
</script>
