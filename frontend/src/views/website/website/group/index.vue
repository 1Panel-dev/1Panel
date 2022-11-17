<template>
    <el-dialog v-model="open" :title="$t('website.groupSetting')" width="40%" :before-close="handleClose">
        <ComplexTable :data="data" @search="search()">
            <template #toolbar>
                <el-button type="primary" plain @click="openCreate">{{ $t('commons.button.create') }}</el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" prop="name">
                <template #default="{ row }">
                    <span v-if="!row.edit" @click="row.edit = true">{{ row.name }}</span>
                    <el-input v-if="row.edit" v-model="row.name" @blur="row.edit = false"></el-input>
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.operate')">
                <template #default="{ row, $index }">
                    <el-button link :disabled="row.default" type="primary" @click="saveGroup(row)">
                        {{ $t('commons.button.save') }}
                    </el-button>
                    <el-button link :disabled="row.default" type="primary" @click="deleteGroup($index)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
            </el-table-column>
        </ComplexTable>
    </el-dialog>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import i18n from '@/lang';
import ComplexTable from '@/components/complex-table/index.vue';
import { ListGroups, CreateGroup, DeleteGroup, UpdateGroup } from '@/api/modules/website';
import { ElMessage } from 'element-plus';

interface groupData {
    id: number;
    name: string;
    edit: boolean;
    default: boolean;
}

let open = ref(false);
let data = ref<groupData[]>([]);

const handleClose = () => {
    open.value = false;
    data.value = [];
};

const search = () => {
    ListGroups().then((res) => {
        for (const d of res.data) {
            const g = {
                id: d.id,
                name: d.name,
                default: d.default,
                edit: false,
            };
            data.value.push(g);
        }
    });
};

const saveGroup = (create: groupData) => {
    const group = {
        name: create.name,
        id: create.id,
    };
    if (create.id == 0) {
        CreateGroup(group).then(() => {
            ElMessage.success(i18n.global.t('commons.msg.createSuccess'));
        });
    } else {
        UpdateGroup(group).then(() => {
            ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
        });
    }
};

const acceptParams = async () => {
    open.value = true;
    search();
};

const openCreate = () => {
    const g = {
        id: 0,
        name: '',
        default: false,
        edit: true,
    };
    data.value.push(g);
};

const deleteGroup = (index: number) => {
    const group = data.value[index];

    if (group.id > 0) {
        DeleteGroup(group.id).then(() => {
            data.value.splice(index, 1);
            ElMessage.success(i18n.global.t('commons.msg.deleteSuccess'));
        });
    } else {
        data.value.splice(index, 1);
    }
};

defineExpose({ acceptParams });
</script>
