<template>
    <div>
        <FireRouter />
        <LayoutContent v-loading="loading" :title="$t('firewall.firewall')">
            <template #toolbar>
                <el-row>
                    <el-col :span="20">
                        <el-button type="primary" @click="onOpenDialog('create')">
                            {{ $t('commons.button.create') }} {{ $t('firewall.ipRule') }}
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
                    <el-table-column :min-width="80" :label="$t('firewall.address')" prop="address">
                        <template #default="{ row }">
                            <span v-if="row.address && row.address !== 'Anywhere'">{{ row.address }}</span>
                            <span v-else>{{ $t('firewall.allIP') }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :min-width="80" :label="$t('firewall.strategy')" prop="strategy">
                        <template #default="{ row }">
                            <el-tag v-if="row.strategy === 'accept'" type="success">{{ $t('firewall.accept') }}</el-tag>
                            <el-tag v-if="row.strategy === 'drop'" type="danger">{{ $t('firewall.drop') }}</el-tag>
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        width="200px"
                        :buttons="buttons"
                        :ellipsis="10"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OperatrDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import OperatrDialog from '@/views/host/firewall/ip/operate/index.vue';
import FireRouter from '@/views/host/firewall/index.vue';
import LayoutContent from '@/layout/layout-content.vue';
import { onMounted, reactive, ref } from 'vue';
import { operateIPRule, searchFireRule } from '@/api/modules/host';
import { Host } from '@/api/interface/host';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const activeTag = ref('address');
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

const dialogRef = ref();
const onOpenDialog = async (
    title: string,
    rowData: Partial<Host.RuleIP> = {
        strategy: 'accept',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const onDelete = async (row: Host.RuleInfo | null) => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.delete'), i18n.global.t('commons.msg.deleteTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'warning',
    }).then(async () => {
        let params = {
            operation: 'remove',
            address: row.address,
            strategy: row.strategy,
        };
        loading.value = true;
        await operateIPRule(params)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Host.RuleInfo) => {
            onDelete(row);
        },
    },
];

onMounted(() => {
    search();
});
</script>
