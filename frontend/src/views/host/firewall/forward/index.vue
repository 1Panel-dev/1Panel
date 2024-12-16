<template>
    <div>
        <FireRouter />

        <div v-loading="loading">
            <FireStatus
                v-show="fireName !== '-'"
                ref="fireStatusRef"
                @search="search"
                v-model:loading="loading"
                v-model:mask-show="maskShow"
                v-model:status="fireStatus"
                v-model:name="fireName"
            />
            <div v-if="fireName !== '-'">
                <el-card v-if="fireStatus != 'running' && maskShow" class="mask-prompt">
                    <span>{{ $t('firewall.firewallNotStart') }}</span>
                </el-card>

                <LayoutContent :title="$t('firewall.forwardRule', 2)" :class="{ mask: fireStatus != 'running' }">
                    <template #toolbar>
                        <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                            <div class="flex flex-wrap gap-3">
                                <el-button type="primary" @click="onOpenDialog('create')">
                                    {{ $t('firewall.createForwardRule') }}
                                </el-button>
                                <el-button @click="onDelete(null)" plain :disabled="selects.length === 0">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                            <div class="flex flex-wrap gap-3">
                                <TableSetting @search="search()" />
                                <TableSearch @search="search()" v-model:searchName="searchName" />
                            </div>
                        </div>
                    </template>
                    <template #main>
                        <ComplexTable
                            :pagination-config="paginationConfig"
                            v-model:selects="selects"
                            @search="search"
                            :data="data"
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
            <div v-else>
                <LayoutContent :title="$t('firewall.firewall')" :divider="true">
                    <template #main>
                        <div class="app-warn">
                            <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                                <span>{{ $t('firewall.notSupport') }}</span>
                                <span @click="toDoc" class="flex items-center justify-center gap-0.5">
                                    <el-icon><Position /></el-icon>
                                    {{ $t('firewall.quickJump') }}
                                </span>
                            </div>
                            <div>
                                <img src="@/assets/images/no_app.svg" />
                            </div>
                        </div>
                    </template>
                </LayoutContent>
            </div>
        </div>

        <OpDialog ref="opRef" @search="search" />
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
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();

const loading = ref();
const activeTag = ref('forward');
const selects = ref<any>([]);
const searchName = ref();
const searchStatus = ref('');
const searchStrategy = ref('');

const maskShow = ref(true);
const fireStatus = ref('running');
const fireName = ref();
const fireStatusRef = ref();

const opRef = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'firewall-forward-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const search = async () => {
    if (fireStatus.value !== 'running') {
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
const toDoc = () => {
    window.open(globalStore.docsUrl + '/user_manual/hosts/firewall/', '_blank', 'noopener,noreferrer');
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
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('firewall.forwardRule'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: operateForwardRule,
        params: { rules: rules },
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
