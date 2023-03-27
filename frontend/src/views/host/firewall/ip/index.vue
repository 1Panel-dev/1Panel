<template>
    <div>
        <FireRouter />
        <LayoutContent v-loading="loading" :title="$t('firewall.firewall')">
            <template #toolbar>
                <el-row>
                    <el-col :span="20">
                        <el-button
                            class="tag-button"
                            :class="activeTag === 'port' ? '' : 'no-active'"
                            @click="changeTag('port')"
                            :type="activeTag === 'port' ? 'primary' : ''"
                            :plain="activeTag !== 'port'"
                        >
                            {{ $t('firewall.portRule') }}
                        </el-button>
                        <el-button
                            class="tag-button"
                            :class="activeTag === 'ip' ? '' : 'no-active'"
                            @click="changeTag('ip')"
                            :type="activeTag === 'ip' ? 'primary' : ''"
                            :plain="activeTag !== 'ip'"
                        >
                            {{ $t('firewall.ipRule') }}
                        </el-button>
                    </el-col>
                </el-row>
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    @search="search"
                    :data="data"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column :label="$t('firewall.protocol')" :min-width="90" prop="protocol" />
                    <el-table-column :label="$t('firewall.port')" :min-width="120" prop="port" />
                    <el-table-column :min-width="80" :label="$t('firewall.strategy')" prop="strategy">
                        <template #default="{ row }">
                            <el-tag v-if="row.strategy === 'accept'" type="success">{{ $t('firewall.accept') }}</el-tag>
                            <el-tag v-if="row.strategy === 'drop'" type="danger">{{ $t('firewall.drop') }}</el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column :min-width="80" :label="$t('firewall.address')" prop="address">
                        <template #default="{ row }">
                            <span v-if="row.address && row.address !== 'Anywhere'">{{ row.address }}</span>
                            <span v-else>{{ $t('firewall.allIP') }}</span>
                        </template>
                    </el-table-column>
                </ComplexTable>
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import FireRouter from '@/views/host/firewall/index.vue';
import LayoutContent from '@/layout/layout-content.vue';
import { onMounted, reactive, ref } from 'vue';
import { searchFireRule } from '@/api/modules/host';

const loading = ref();
const activeTag = ref('port');
const selects = ref<any>([]);

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const search = async () => {
    let params = {
        type: activeTag.value,
        info: '',
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchFireRule(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const changeTag = async (type: string) => {
    activeTag.value = type;
};

onMounted(() => {
    search();
});
</script>
