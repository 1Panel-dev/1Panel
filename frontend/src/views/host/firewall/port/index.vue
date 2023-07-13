<template>
    <div v-loading="loading" style="position: relative">
        <FireRouter />

        <FireStatus
            v-show="fireName !== '-'"
            ref="fireStatuRef"
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

            <LayoutContent :title="$t('firewall.portRule')" :class="{ mask: fireStatus != 'running' }">
                <template #prompt>
                    <el-alert type="info" :closable="false">
                        <template #default>
                            <span>
                                <span>{{ $t('firewall.dockerHelper', [fireName]) }}</span>
                                <el-link
                                    style="font-size: 12px; margin-left: 5px"
                                    icon="Position"
                                    @click="quickJump()"
                                    type="primary"
                                >
                                    {{ $t('firewall.quickJump') }}
                                </el-link>
                            </span>
                        </template>
                    </el-alert>
                </template>
                <template #toolbar>
                    <el-row>
                        <el-col :span="16">
                            <el-button type="primary" @click="onOpenDialog('create')">
                                {{ $t('commons.button.create') }}{{ $t('firewall.portRule') }}
                            </el-button>
                            <el-button @click="onDelete(null)" plain :disabled="selects.length === 0">
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </el-col>
                        <el-col :span="8">
                            <TableSetting @search="search()" />
                            <div class="search-button">
                                <el-input
                                    v-model="searchName"
                                    clearable
                                    @clear="search()"
                                    suffix-icon="Search"
                                    @keyup.enter="search()"
                                    @change="search()"
                                    :placeholder="$t('commons.button.search')"
                                ></el-input>
                            </div>
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
                        <el-table-column type="selection" :selectable="selectable" fix />
                        <el-table-column :label="$t('commons.table.protocol')" :min-width="90" prop="protocol" />
                        <el-table-column :label="$t('commons.table.port')" :min-width="120" prop="port" />
                        <el-table-column :label="$t('commons.table.status')" :min-width="120">
                            <template #default="{ row }">
                                <el-tag type="info" v-if="row.isUsed">
                                    {{
                                        row.appName
                                            ? $t('firewall.used') + ' ( ' + row.appName + ' )'
                                            : $t('firewall.used')
                                    }}
                                </el-tag>
                                <el-tag type="success" v-else>{{ $t('firewall.unUsed') }}</el-tag>
                            </template>
                        </el-table-column>
                        <el-table-column :min-width="80" :label="$t('firewall.strategy')" prop="strategy">
                            <template #default="{ row }">
                                <el-button
                                    v-if="row.strategy === 'accept'"
                                    :disabled="row.appName === '1panel'"
                                    @click="onChangeStatus(row, 'drop')"
                                    link
                                    type="success"
                                >
                                    {{ $t('firewall.accept') }}
                                </el-button>
                                <el-button v-else link type="danger" @click="onChangeStatus(row, 'accept')">
                                    {{ $t('firewall.drop') }}
                                </el-button>
                            </template>
                        </el-table-column>
                        <el-table-column :min-width="80" :label="$t('firewall.address')" prop="address">
                            <template #default="{ row }">
                                <span v-if="row.address && row.address !== 'Anywhere'">{{ row.address }}</span>
                                <span v-else>{{ $t('firewall.allIP') }}</span>
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
        </div>
        <div v-else>
            <LayoutContent :title="$t('firewall.firewall')" :divider="true">
                <template #main>
                    <div class="app-warn">
                        <div>
                            <span>{{ $t('firewall.notSupport') }}</span>
                            <el-link
                                style="font-size: 12px; margin-left: 5px"
                                @click="toDoc"
                                icon="Position"
                                type="primary"
                            >
                                {{ $t('firewall.quickJump') }}
                            </el-link>
                            <div>
                                <img src="@/assets/images/no_app.svg" />
                            </div>
                        </div>
                    </div>
                </template>
            </LayoutContent>
        </div>

        <OperatrDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import FireRouter from '@/views/host/firewall/index.vue';
import TableSetting from '@/components/table-setting/index.vue';
import OperatrDialog from '@/views/host/firewall/port/operate/index.vue';
import FireStatus from '@/views/host/firewall/status/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { batchOperateRule, searchFireRule, updatePortRule } from '@/api/modules/host';
import { Host } from '@/api/interface/host';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import router from '@/routers';

const loading = ref();
const activeTag = ref('port');
const selects = ref<any>([]);
const searchName = ref();

const maskShow = ref(true);
const fireStatus = ref('running');
const fireName = ref();
const fireStatuRef = ref();

const data = ref();
const paginationConfig = reactive({
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
    rowData: Partial<Host.RulePort> = {
        protocol: 'tcp',
        source: 'anyWhere',
        strategy: 'accept',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const quickJump = () => {
    router.push({ name: 'AppInstalled' });
};
const toDoc = () => {
    window.open('https://1panel.cn/docs/user_manual/hosts/firewall/', '_blank');
};

const onChangeStatus = async (row: Host.RuleInfo, status: string) => {
    let operation =
        status === 'accept'
            ? i18n.global.t('firewall.changeStrategyPortHelper2')
            : i18n.global.t('firewall.changeStrategyPortHelper1');
    ElMessageBox.confirm(operation, i18n.global.t('firewall.changeStrategy', [i18n.global.t('commons.table.port')]), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        let params = {
            oldRule: {
                operation: 'remove',
                address: row.address,
                port: row.port,
                source: '',
                protocol: row.protocol,
                strategy: row.strategy,
            },
            newRule: {
                operation: 'add',
                address: row.address,
                port: row.port,
                source: '',
                protocol: row.protocol,
                strategy: status,
            },
        };
        loading.value = true;
        await updatePortRule(params)
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

const onDelete = async (row: Host.RuleInfo | null) => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.delete'), i18n.global.t('commons.msg.deleteTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'warning',
    }).then(async () => {
        let rules = [];
        if (row) {
            rules.push({
                operation: 'remove',
                address: row.address,
                port: row.port,
                source: '',
                protocol: row.protocol,
                strategy: row.strategy,
            });
        } else {
            for (const item of selects.value) {
                rules.push({
                    operation: 'remove',
                    address: item.address,
                    port: item.port,
                    source: '',
                    protocol: item.protocol,
                    strategy: item.strategy,
                });
            }
        }
        loading.value = true;
        await batchOperateRule({ type: 'port', rules: rules })
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

function selectable(row) {
    return row.appName !== '1panel';
}

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Host.RulePort) => {
            onOpenDialog('edit', row);
        },
        disabled: (row: any) => {
            return row.appName === '1panel';
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Host.RuleInfo) => {
            onDelete(row);
        },
        disabled: (row: any) => {
            return row.appName === '1panel';
        },
    },
];

onMounted(() => {
    if (fireName.value !== '-') {
        loading.value = true;
        fireStatuRef.value.acceptParams();
    }
});
</script>
