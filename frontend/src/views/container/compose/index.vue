<template>
    <div>
        <Submenu activeName="compose" />
        <el-card style="margin-top: 20px">
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
                <template #toolbar>
                    <el-button icon="Plus" type="primary" @click="onOpenDialog()">
                        {{ $t('commons.button.create') }}
                    </el-button>
                    <el-button-group style="margin-left: 10px">
                        <el-button :disabled="selects.length === 0" @click="onOperate('up')">
                            {{ $t('container.start') }}
                        </el-button>
                        <el-button :disabled="selects.length === 0" @click="onOperate('stop')">
                            {{ $t('container.stop') }}
                        </el-button>
                        <el-button :disabled="selects.length === 0" @click="onOperate('pause')">
                            {{ $t('container.pause') }}
                        </el-button>
                        <el-button :disabled="selects.length === 0" @click="onOperate('unpause')">
                            {{ $t('container.unpause') }}
                        </el-button>
                        <el-button :disabled="selects.length === 0" @click="onOperate('restart')">
                            {{ $t('container.restart') }}
                        </el-button>
                        <el-button :disabled="selects.length === 0" @click="onOperate('down')">
                            {{ $t('container.down') }}
                        </el-button>
                    </el-button-group>
                </template>
                <el-table-column type="selection" fix></el-table-column>
                <el-table-column
                    :label="$t('commons.table.name')"
                    show-overflow-tooltip
                    min-width="100"
                    prop="name"
                    fix
                >
                    <template #default="{ row }">
                        <el-link @click="goContainer(row.name)" type="primary">{{ row.name }}</el-link>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('container.from')" prop="createdBy" min-width="80" fix />
                <el-table-column :label="$t('container.containerNumber')" prop="containerNumber" min-width="80" fix />
                <el-table-column :label="$t('container.container')" prop="contaienrs" min-width="80" fix>
                    <template #default="{ row }">
                        <div v-for="(item, index) in row.containers" :key="index">
                            <div v-if="row.expand || (!row.expand && index < 3)">
                                <el-tag>{{ item.name }} [{{ item.state }}]</el-tag>
                            </div>
                        </div>
                        <div v-if="!row.expand && row.containers.length > 3">
                            <el-button type="primary" link @click="row.expand = true">
                                {{ $t('commons.button.expand') }}...
                            </el-button>
                        </div>
                    </template>
                </el-table-column>

                <el-table-column :label="$t('commons.table.createdAt')" prop="createdAt" min-width="80" fix />
            </ComplexTable>
        </el-card>

        <OperatorDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { reactive, onMounted, ref } from 'vue';
import OperatorDialog from '@/views/container/compose/operator/index.vue';
import Submenu from '@/views/container/index.vue';
import { ComposeOperator, searchCompose } from '@/api/modules/container';
import i18n from '@/lang';
import { ElMessage, ElMessageBox } from 'element-plus';
import router from '@/routers';

const data = ref();
const selects = ref<any>([]);

const paginationConfig = reactive({
    page: 1,
    pageSize: 10,
    total: 0,
});

const search = async () => {
    let params = {
        page: paginationConfig.page,
        pageSize: paginationConfig.pageSize,
    };
    await searchCompose(params).then((res) => {
        if (res.data) {
            data.value = res.data.items;
            paginationConfig.total = res.data.total;
        }
    });
};

const goContainer = async (name: string) => {
    router.push({ name: 'Container', params: { filters: 'com.docker.compose.project=' + name } });
};

const dialogRef = ref();
const onOpenDialog = async () => {
    dialogRef.value!.acceptParams();
};

const onOperate = async (operation: string) => {
    ElMessageBox.confirm(
        i18n.global.t('container.operatorComposeHelper', [i18n.global.t('container.' + operation)]),
        i18n.global.t('container.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(() => {
        let ps = [];
        for (const item of selects.value) {
            const param = {
                path: item.path,
                operation: operation,
            };
            ps.push(ComposeOperator(param));
        }
        Promise.all(ps)
            .then(() => {
                search();
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                search();
            });
    });
};

onMounted(() => {
    search();
});
</script>
