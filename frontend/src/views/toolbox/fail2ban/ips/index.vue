<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="30%" :before-close="handleClose">
        <template #header>
            <Header :header="$t('toolbox.fail2ban.bannedIP')" :back="handleClose"></Header>
        </template>

        <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search()">
            <template #toolbar>
                <el-button type="primary" @click="onAdd">{{ $t('commons.button.add') }}</el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" prop="name">
                <template #default="{ row }">
                    <div v-if="!row.edit">
                        <span>{{ row.name }}</span>
                    </div>
                    <el-form @submit.prevent ref="formRef" v-if="row.edit" :model="row">
                        <el-form-item prop="name" v-if="row.edit" :rules="Rules.name">
                            <div style="margin-top: 20px; width: 100%"><el-input v-model="row.name" /></div>
                        </el-form-item>
                    </el-form>
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.operate')">
                <template #default="{ row, $index }">
                    <div>
                        <el-button link v-if="!row.edit" type="primary" @click="onEdit($index)">
                            {{ $t('commons.button.edit') }}
                        </el-button>
                        <el-button
                            link
                            v-if="!row.edit"
                            :disabled="row.isDefault"
                            type="primary"
                            @click="onDelete($index)"
                        >
                            {{ $t('commons.button.delete') }}
                        </el-button>
                        <el-button link v-if="row.edit" type="primary" @click="search">
                            {{ $t('commons.button.cancel') }}
                        </el-button>
                    </div>
                </template>
            </el-table-column>
        </ComplexTable>
    </el-drawer>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { searchFail2ban } from '@/api/modules/toolbox';
import Header from '@/components/drawer-header/index.vue';
import { Rules } from '@/global/form-rules';
import { FormInstance } from 'element-plus';

const open = ref(false);
const ipType = ref();
const data = ref([]);
const handleClose = () => {
    open.value = false;
    data.value = [];
};

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

interface DialogProps {
    ipType: string;
}
const formRef = ref<FormInstance>();
const acceptParams = (params: DialogProps): void => {
    ipType.value = params.ipType;
    open.value = true;
    search();
};

const search = () => {
    let params = {
        status: ipType.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    searchFail2ban(params).then((res) => {
        let dataItem = res.data.items || [];
        paginationConfig.total = res.data.total;
        data.value = [];
        for (const item of dataItem) {
            data.value.push({ name: item });
        }
    });
};

const onAdd = () => {
    for (const d of data.value) {
        if (d.name == '') {
            return;
        }
        if (d.edit) {
            d.edit = false;
        }
    }
    const g = {
        id: 0,
        name: '',
        isDefault: false,
        edit: true,
    };
    data.value.unshift(g);
};

const onDelete = (index: number) => {
    data.value.splice(index, 1);
};

const onEdit = (index: number) => {
    for (const i in data.value) {
        const d = data.value[i];
        if (d.name == '') {
            data.value.splice(Number(i), 1);
        }
        if (d.edit) {
            d.edit = false;
        }
    }
    data.value[index].edit = true;
};

defineExpose({ acceptParams });
</script>
