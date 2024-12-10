<template>
    <div v-loading="loading">
        <el-card v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span>{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link class="bt" @click="goSetting">【 {{ $t('container.setting') }} 】</el-button>
            <span>{{ $t('container.startIn') }}</span>
        </el-card>

        <LayoutContent :title="$t('container.network', 2)" :class="{ mask: dockerStatus != 'Running' }">
            <template #toolbar>
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div class="flex flex-wrap gap-3">
                        <el-button type="primary" @click="onCreate()">
                            {{ $t('container.createNetwork') }}
                        </el-button>
                        <el-button type="primary" plain @click="onClean()">
                            {{ $t('container.networkPrune') }}
                        </el-button>
                        <el-button :disabled="selects.length === 0" @click="batchDelete(null)">
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
                    :data="data"
                    @search="search"
                >
                    <el-table-column type="selection" :selectable="selectable" fix />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        width="130"
                        prop="name"
                        fix
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="onInspect(row.id)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column width="90">
                        <template #default="{ row }">
                            <el-tag round v-if="row.isSystem || row.name === '1panel-network'">system</el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.driver')"
                        show-overflow-tooltip
                        min-width="60"
                        prop="driver"
                    />
                    <el-table-column :label="$t('container.subnet')" min-width="80" prop="subnet" fix />
                    <el-table-column :label="$t('container.gateway')" min-width="80" prop="gateway" fix />
                    <el-table-column :label="$t('container.tag')" min-width="140" fix>
                        <template #default="{ row }">
                            <div v-for="(item, index) in row.labels" :key="index">
                                <div v-if="row.expand || (!row.expand && index < 3)">
                                    <el-tag type="info">{{ item }}</el-tag>
                                </div>
                            </div>
                            <div v-if="!row.expand && row.labels.length > 3">
                                <el-button link @click="row.expand = true">
                                    {{ $t('commons.button.expand') }}...
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        show-overflow-tooltip
                        min-width="90"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                    />
                    <fu-table-operations width="100" :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OpDialog ref="opRef" @search="search" />
        <CodemirrorDialog ref="codemirror" />
        <CreateDialog @search="search" ref="dialogCreateRef" />
    </div>
</template>

<script lang="ts" setup>
import CreateDialog from '@/views/container/network/create/index.vue';
import CodemirrorDialog from '@/components/codemirror-dialog/index.vue';
import { reactive, onMounted, ref } from 'vue';
import { dateFormat } from '@/utils/util';
import { deleteNetwork, searchNetwork, inspect, loadDockerStatus, containerPrune } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import i18n from '@/lang';
import router from '@/routers';
import { ElMessageBox } from 'element-plus';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const codemirror = ref();

const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'container-network-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();

const opRef = ref();

const dockerStatus = ref('Running');
const loadStatus = async () => {
    loading.value = true;
    await loadDockerStatus()
        .then((res) => {
            loading.value = false;
            dockerStatus.value = res.data;
            if (dockerStatus.value === 'Running') {
                search();
            }
        })
        .catch(() => {
            dockerStatus.value = 'Failed';
            loading.value = false;
        });
};
const goSetting = async () => {
    router.push({ name: 'ContainerSetting' });
};

const dialogCreateRef = ref<DialogExpose>();

interface DialogExpose {
    acceptParams: () => void;
}
const onCreate = async () => {
    dialogCreateRef.value!.acceptParams();
};

const onClean = () => {
    ElMessageBox.confirm(i18n.global.t('container.networkPruneHelper'), i18n.global.t('container.networkPrune'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        let params = {
            pruneType: 'network',
            withTagAll: false,
        };
        await containerPrune(params)
            .then((res) => {
                loading.value = false;
                MsgSuccess(i18n.global.t('container.cleanSuccess', [res.data.deletedNumber]));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

function selectable(row) {
    return !row.isSystem;
}

const search = async () => {
    const params = {
        info: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchNetwork(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            for (const item of data.value) {
                item.isSystem = isSystem(item.name);
            }
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const batchDelete = async (row: Container.NetworkInfo | null) => {
    let names: Array<string> = [];
    let hasPanelNetwork;
    if (row === null) {
        selects.value.forEach((item: Container.NetworkInfo) => {
            if (item.name === '1panel-network') {
                hasPanelNetwork = true;
            }
            names.push(item.name);
        });
    } else {
        if (row.name === '1panel-network') {
            hasPanelNetwork = true;
        }
        names.push(row.name);
    }
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: hasPanelNetwork
            ? i18n.global.t('container.networkHelper')
            : i18n.global.t('commons.msg.operatorHelper', [
                  i18n.global.t('container.network'),
                  i18n.global.t('commons.button.delete'),
              ]),
        api: deleteNetwork,
        params: { names: names },
    });
};

const onInspect = async (id: string) => {
    const res = await inspect({ id: id, type: 'network' });
    let detailInfo = JSON.stringify(JSON.parse(res.data), null, 2);
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo,
    };
    codemirror.value!.acceptParams(param);
};

function isSystem(val: string) {
    return val === 'bridge' || val === 'none' || val === 'host';
}

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Container.NetworkInfo) => {
            batchDelete(row);
        },
        disabled: (row: any) => {
            return row.isSystem;
        },
    },
];

onMounted(() => {
    loadStatus();
});
</script>
