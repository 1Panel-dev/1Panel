<template>
    <div>
        <FireRouter />

        <div v-loading="loading">
            <FireStatus
                ref="fireStatusRef"
                @search="search"
                v-model:loading="loading"
                v-model:mask-show="maskShow"
                v-model:is-active="isActive"
                v-model:name="fireName"
            />
            <div v-if="fireName !== '-'">
                <el-card v-if="!isActive && maskShow" class="mask-prompt">
                    <span>{{ $t('firewall.firewallNotStart') }}</span>
                </el-card>

                <LayoutContent :title="$t('firewall.forwardRule', 2)" :class="{ mask: !isActive }">
                    <template #leftToolBar>
                        <el-button type="primary" @click="onOpenDialog('create')">
                            {{ $t('firewall.createForwardRule') }}
                        </el-button>
                        <el-button @click="onDelete(null)" plain :disabled="selects.length === 0">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>
                    <template #rightToolBar>
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                        <TableRefresh @search="search()" />
                        <TableSetting title="firewall-forward-refresh" @search="search()" />
                    </template>
                    <template #main>
                        <ComplexTable
                            :pagination-config="paginationConfig"
                            v-model:selects="selects"
                            @search="search"
                            :data="data"
                            :heightDiff="370"
                        >
                            <el-table-column type="selection" fix />
                            <el-table-column :label="$t('commons.table.protocol')" :min-width="70" prop="protocol" />
                            <el-table-column :label="$t('firewall.sourcePort')" :min-width="70" prop="port" />
                            <el-table-column :min-width="80" :label="$t('firewall.targetIP')" prop="targetIP" />
                            <el-table-column :label="$t('firewall.targetPort')" :min-width="70" prop="targetPort" />
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
            </div>
        </div>

        <OpDialog ref="opRef" @search="search" @submit="onSubmitDelete()">
            <template #content>
                <el-form class="mt-4 mb-1" ref="deleteForm" label-position="left">
                    <el-form-item>
                        <el-checkbox v-model="forceDelete" :label="$t('website.forceDelete')" />
                        <span class="input-help">
                            {{ $t('website.forceDeleteHelper') }}
                        </span>
                    </el-form-item>
                </el-form>
            </template>
        </OpDialog>
        <OperateDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import FireRouter from '@/views/host/firewall/index.vue';
import OperateDialog from './operate/index.vue';
import FireStatus from '@/views/host/firewall/status/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { operateForwardRule, searchFireRule } from '@/api/modules/host';
import { Host } from '@/api/interface/host';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const activeTag = ref('forward');
const selects = ref<any>([]);
const searchName = ref();
const searchStatus = ref('');
const searchStrategy = ref('');

const maskShow = ref(true);
const isActive = ref(false);
const fireName = ref();
const fireStatusRef = ref();

const opRef = ref();
const forceDelete = ref(false);
const operateRules = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'firewall-forward-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const search = async () => {
    if (!isActive.value) {
        loading.value = false;
        data.value = [];
        paginationConfig.total = 0;
        return;
    }
    let params = {
        type: activeTag.value,
        status: searchStatus.value,
        strategy: searchStrategy.value,
        info: searchName.value,
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
    rowData: Partial<Host.RuleForward> = {
        protocol: 'tcp',
        port: '8080',
        targetIP: '',
        targetPort: '',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};
const onDelete = async (row: Host.RuleForward | null) => {
    let names = [];
    let rules = [];
    if (row) {
        rules.push({
            ...row,
            operation: 'remove',
        });
        names = [row.port + ' (' + row.protocol + ')'];
    } else {
        for (const item of selects.value) {
            names.push(item.port + ' (' + item.protocol + ')');
            rules.push({
                ...item,
                operation: 'remove',
            });
        }
    }
    operateRules.value = rules;
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('firewall.forwardRule'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: null,
        params: null,
    });
};
const onSubmitDelete = async () => {
    loading.value = true;
    await operateForwardRule({ rules: operateRules.value, forceDelete: forceDelete.value })
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Host.RuleForward) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Host.RuleForward) => {
            onDelete(row);
        },
    },
];

onMounted(() => {
    forceDelete.value = false;
    if (fireName.value !== '-') {
        loading.value = true;
        fireStatusRef.value.acceptParams();
    }
});
</script>

<style lang="scss" scoped>
.svg-icon {
    font-size: 8px;
    margin-bottom: -4px;
    cursor: pointer;
}
</style>
