<template>
    <div>
        <el-card width="30%" v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span style="font-size: 14px">{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link style="font-size: 14px; margin-bottom: 5px" @click="goSetting">
                【 {{ $t('container.setting') }} 】
            </el-button>
            <span style="font-size: 14px">{{ $t('container.startIn') }}</span>
        </el-card>

        <LayoutContent
            v-loading="loading"
            :title="$t('container.network')"
            :class="{ mask: dockerStatus != 'Running' }"
        >
            <template #toolbar>
                <el-row>
                    <el-col :span="20">
                        <el-button type="primary" @click="onCreate()">
                            {{ $t('container.createNetwork') }}
                        </el-button>
                        <el-button type="primary" plain :disabled="selects.length === 0" @click="batchDelete(null)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </el-col>
                    <el-col :span="4">
                        <div class="search-button">
                            <el-input
                                v-model="searchName"
                                clearable
                                @clear="search()"
                                suffix-icon="Search"
                                @keyup.enter="search()"
                                @blur="search()"
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
                    :data="data"
                    @search="search"
                >
                    <el-table-column type="selection" :selectable="selectable" fix />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        show-overflow-tooltip
                        min-width="80"
                        prop="name"
                        fix
                    >
                        <template #default="{ row }">
                            <el-link @click="onInspect(row.id)" type="primary">{{ row.name }}</el-link>
                            <el-tag effect="dark" round v-if="row.isSystem" style="margin-left: 5px">system</el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.driver')"
                        show-overflow-tooltip
                        min-width="40"
                        prop="driver"
                    />
                    <el-table-column :label="$t('container.subnet')" min-width="80" prop="subnet" fix />
                    <el-table-column :label="$t('container.gateway')" min-width="80" prop="gateway" fix />
                    <el-table-column :label="$t('container.tag')" min-width="140" fix>
                        <template #default="{ row }">
                            <div v-for="(item, index) in row.labels" :key="index">
                                <div v-if="row.expand || (!row.expand && index < 3)">
                                    <el-tag>{{ item }}</el-tag>
                                </div>
                            </div>
                            <div v-if="!row.expand && row.labels.length > 3">
                                <el-button type="primary" link @click="row.expand = true">
                                    {{ $t('commons.button.expand') }}...
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        min-width="90"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                    />
                    <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>

        <CodemirrorDialog ref="codemirror" />
        <CreateDialog @search="search" ref="dialogCreateRef" />
    </div>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import CreateDialog from '@/views/container/network/create/index.vue';
import CodemirrorDialog from '@/components/codemirror-dialog/codemirror.vue';
import { reactive, onMounted, ref } from 'vue';
import { dateFormat } from '@/utils/util';
import { deleteNetwork, searchNetwork, inspect, loadDockerStatus } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';
import router from '@/routers';

const loading = ref();

const detailInfo = ref();
const codemirror = ref();

const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();

const dockerStatus = ref();
const loadStatus = async () => {
    const res = await loadDockerStatus();
    dockerStatus.value = res.data;
    if (dockerStatus.value === 'Running') {
        search();
    }
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
    if (row === null) {
        selects.value.forEach((item: Container.NetworkInfo) => {
            names.push(item.name);
        });
    } else {
        names.push(row.name);
    }
    await useDeleteData(deleteNetwork, { names: names }, 'commons.msg.delete');
    search();
};

const onInspect = async (id: string) => {
    const res = await inspect({ id: id, type: 'network' });
    detailInfo.value = JSON.stringify(JSON.parse(res.data), null, 2);
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo.value,
    };
    codemirror.value!.acceptParams(param);
};

function isSystem(val: string) {
    return val === 'bridge' || val === '1panel' || val === 'none' || val === 'host';
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
