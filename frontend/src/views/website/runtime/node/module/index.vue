<template>
    <el-drawer :close-on-click-modal="false" :close-on-press-escape="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader :header="$t('runtime.moduleManager')" :back="handleClose" />
        </template>
        <el-row :gutter="20" v-loading="loading">
            <el-col :span="10">
                <el-input v-model="module">
                    <template #prepend>{{ packageManager }}</template>
                </el-input>
            </el-col>
            <el-col :span="14">
                <el-button @click="operateModule('install', module)" type="primary" :disabled="module === ''">
                    {{ $t('commons.operate.install') }}
                </el-button>
            </el-col>
            <el-col>
                <ComplexTable :data="data" @search="search()" class="mt-5" :height="800">
                    <el-table-column :label="$t('commons.table.name')" prop="name" min-width="100px"></el-table-column>
                    <el-table-column :label="$t('container.version')" prop="version" width="80px"></el-table-column>
                    <el-table-column
                        :label="$t('commons.table.protocol')"
                        prop="license"
                        width="120px"
                    ></el-table-column>
                    <el-table-column
                        :label="$t('container.description')"
                        fix
                        min-width="120px"
                        prop="description"
                    ></el-table-column>
                    <fu-table-operations
                        :ellipsis="10"
                        width="150px"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </el-col>
        </el-row>
    </el-drawer>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import { GetNodeModules, OperateNodeModule } from '@/api/modules/runtime';
import { MsgError, MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { Runtime } from '@/api/interface/runtime';

interface NodeProps {
    packageManager: string;
    id: number;
}

const open = ref(false);
const id = ref(0);
const data = ref([]);
const module = ref('');
const packageManager = ref('');
const loading = ref(false);

const buttons = [
    {
        label: i18n.global.t('commons.operate.update'),
        click: function (row: Runtime.Runtime) {
            operateModule('update', row.name);
        },
    },
    {
        label: i18n.global.t('commons.operate.uninstall'),
        click: function (row: Runtime.Runtime) {
            operateModule('uninstall', row.name);
        },
    },
];

const acceptParams = async (props: NodeProps) => {
    id.value = props.id;
    packageManager.value = props.packageManager;
    module.value = '';
    data.value = [];
    open.value = true;
    loading.value = true;
    await search();
    loading.value = false;
};

const search = async () => {
    try {
        const res = await GetNodeModules({ ID: id.value });
        data.value = res.data;
    } catch (error) {}
};

const operateModule = (operate: string, moduleName: string) => {
    ElMessageBox.confirm(
        i18n.global.t('runtime.nodeOperatorHelper', [i18n.global.t('commons.operate.' + operate), moduleName]),
        i18n.global.t('commons.operate.' + operate),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(async () => {
        loading.value = true;
        try {
            await OperateNodeModule({
                ID: id.value,
                Operate: operate,
                Module: moduleName,
                PkgManager: packageManager.value,
            });
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.operate.' + operate) + i18n.global.t('commons.status.success'));
            await search();
            module.value = '';
        } catch (error) {
            MsgError(
                i18n.global.t('commons.operate.' + operate) +
                    i18n.global.t('commons.status.failed') +
                    '  ' +
                    error.message,
            );
        } finally {
            loading.value = false;
        }
    });
};

const handleClose = () => {
    open.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.table-border {
    border: 1px solid var(--el-border-color);
}
</style>
