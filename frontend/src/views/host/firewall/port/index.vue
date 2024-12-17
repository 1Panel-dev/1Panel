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

                <LayoutContent :title="$t('firewall.portRule', 2)" :class="{ mask: fireStatus != 'running' }">
                    <template #prompt>
                        <el-alert type="info" :closable="false">
                            <template #default>
                                <span class="flx-align-center">
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
                    <template #search>
                        <div class="flx-align-center">
                            <el-select v-model="searchStatus" @change="search()" clearable class="p-w-200">
                                <template #prefix>{{ $t('commons.table.status') }}</template>
                                <el-option :label="$t('commons.table.all')" value=""></el-option>
                                <el-option :label="$t('firewall.unUsed')" value="free"></el-option>
                                <el-option :label="$t('firewall.used')" value="used"></el-option>
                            </el-select>
                            <el-select v-model="searchStrategy" @change="search()" clearable class="p-w-200 ml-2.5">
                                <template #prefix>{{ $t('firewall.strategy') }}</template>
                                <el-option :label="$t('commons.table.all')" value=""></el-option>
                                <el-option :label="$t('firewall.accept')" value="accept"></el-option>
                                <el-option :label="$t('firewall.drop')" value="drop"></el-option>
                            </el-select>
                        </div>
                    </template>
                    <template #toolbar>
                        <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                            <div class="flex flex-wrap gap-3">
                                <el-button type="primary" @click="onOpenDialog('create')">
                                    {{ $t('firewall.createPortRule') }}
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
                            <el-table-column :label="$t('commons.table.port')" :min-width="70" prop="port" />
                            <el-table-column :label="$t('commons.table.status')" :min-width="120">
                                <template #default="{ row }">
                                    <div v-if="row.port.indexOf('-') !== -1 && row.usedStatus">
                                        <el-tag type="info" class="mt-1">
                                            {{ $t('firewall.used') + ' * ' + row.usedPorts.length }}
                                        </el-tag>
                                        <el-popover placement="right" popper-class="limit-height-popover" :width="250">
                                            <template #default>
                                                <ul v-for="(item, index) in row.usedPorts" :key="index">
                                                    <li>{{ item }}</li>
                                                </ul>
                                            </template>
                                            <template #reference>
                                                <svg-icon iconName="p-xiangqing" class="svg-icon"></svg-icon>
                                            </template>
                                        </el-popover>
                                    </div>
                                    <div v-else>
                                        <el-tag type="info" v-if="row.usedStatus">
                                            <span v-if="row.usedStatus === 'inUsed'">{{ $t('firewall.used') }}</span>
                                            <span v-else>{{ $t('firewall.used') + ' ' + row.usedStatus }}</span>
                                        </el-tag>
                                        <el-tag type="success" v-else>{{ $t('firewall.unUsed') }}</el-tag>
                                    </div>
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
                            <el-table-column
                                :min-width="150"
                                :label="$t('commons.table.description')"
                                prop="description"
                                show-overflow-tooltip
                            >
                                <template #default="{ row }">
                                    <fu-input-rw-switch v-model="row.description" @blur="onChange(row)" />
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
import OperateDialog from '@/views/host/firewall/port/operate/index.vue';
import FireStatus from '@/views/host/firewall/status/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { batchOperateRule, searchFireRule, updateFirewallDescription, updatePortRule } from '@/api/modules/host';
import { Host } from '@/api/interface/host';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import router from '@/routers';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();

const loading = ref();
const activeTag = ref('port');
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
    cacheSizeKey: 'firewall-port-page-size',
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
            for (const item of data.value) {
                item.usedPorts = item.usedStatus ? item.usedStatus.split(',') : [];
            }
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
    window.open(globalStore.docsUrl + '/user_manual/hosts/firewall/', '_blank', 'noopener,noreferrer');
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
                description: row.description,
            },
            newRule: {
                operation: 'add',
                address: row.address,
                port: row.port,
                source: '',
                protocol: row.protocol,
                strategy: status,
                description: row.description,
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

const onChange = async (info: any) => {
    info.type = 'port';
    await updateFirewallDescription(info);
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const onDelete = async (row: Host.RuleInfo | null) => {
    let names = [];
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
        names = [row.port + ' (' + row.protocol + ')'];
    } else {
        for (const item of selects.value) {
            names.push(item.port + ' (' + item.protocol + ')');
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
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('firewall.portRule'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: batchOperateRule,
        params: { type: 'port', rules: rules },
    });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Host.RulePort) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Host.RuleInfo) => {
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
