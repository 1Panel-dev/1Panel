<template>
    <div v-loading="loading" style="position: relative">
        <FireRouter />
        <FireStatus
            v-show="fireName !== '-'"
            ref="fireStatuRef"
            @search="search"
            v-model:loading="loading"
            v-model:name="fireName"
            v-model:mask-show="maskShow"
            v-model:status="fireStatus"
        />

        <div v-if="fireName !== '-'">
            <el-card v-if="fireStatus != 'running' && maskShow" class="mask-prompt">
                <span>{{ $t('firewall.firewallNotStart') }}</span>
            </el-card>

            <LayoutContent :title="$t('firewall.ipRule')" :class="{ mask: fireStatus != 'running' }">
                <template #toolbar>
                    <el-row>
                        <el-col :span="16">
                            <el-button type="primary" @click="onOpenDialog('create')">
                                {{ $t('commons.button.create') }} {{ $t('firewall.ipRule') }}
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
                        <el-table-column type="selection" fix />
                        <el-table-column :min-width="80" :label="$t('firewall.address')" prop="address">
                            <template #default="{ row }">
                                <span v-if="row.address && row.address !== 'Anywhere'">{{ row.address }}</span>
                                <span v-else>{{ $t('firewall.allIP') }}</span>
                            </template>
                        </el-table-column>
                        <el-table-column :min-width="80" :label="$t('firewall.strategy')" prop="strategy">
                            <template #default="{ row }">
                                <el-button
                                    v-if="row.strategy === 'accept'"
                                    @click="onChangeStatus(row, 'drop')"
                                    link
                                    type="success"
                                >
                                    {{ $t('firewall.allow') }}
                                </el-button>
                                <el-button v-else link type="danger" @click="onChangeStatus(row, 'accept')">
                                    {{ $t('firewall.deny') }}
                                </el-button>
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
import OperatrDialog from '@/views/host/firewall/ip/operate/index.vue';
import FireRouter from '@/views/host/firewall/index.vue';
import TableSetting from '@/components/table-setting/index.vue';
import FireStatus from '@/views/host/firewall/status/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { batchOperateRule, searchFireRule, updateAddrRule } from '@/api/modules/host';
import { Host } from '@/api/interface/host';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const activeTag = ref('address');
const selects = ref<any>([]);
const searchName = ref();
const fireName = ref();

const maskShow = ref(true);
const fireStatus = ref('running');
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

const toDoc = () => {
    window.open('https://1panel.cn/docs/user_manual/hosts/firewall/', '_blank');
};

const onChangeStatus = async (row: Host.RuleInfo, status: string) => {
    let operation =
        status === 'accept'
            ? i18n.global.t('firewall.changeStrategyIPHelper2')
            : i18n.global.t('firewall.changeStrategyIPHelper1');
    ElMessageBox.confirm(operation, i18n.global.t('firewall.changeStrategy', [' IP ']), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        let params = {
            oldRule: {
                operation: 'remove',
                address: row.address,
                strategy: row.strategy,
            },
            newRule: {
                operation: 'add',
                address: row.address,
                strategy: status,
            },
        };
        loading.value = true;
        await updateAddrRule(params)
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

const onDelete = async (row: Host.RuleIP | null) => {
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
                port: '',
                source: '',
                protocol: '',
                strategy: row.strategy,
            });
        } else {
            for (const item of selects.value) {
                rules.push({
                    operation: 'remove',
                    address: item.address,
                    port: '',
                    source: '',
                    protocol: '',
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

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Host.RuleIP) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Host.RuleIP) => {
            onDelete(row);
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
