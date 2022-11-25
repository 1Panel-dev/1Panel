<template>
    <el-dialog v-model="open" :title="$t('website.groupSetting')" width="40%" :before-close="handleClose">
        <ComplexTable :data="data" @search="search()">
            <template #toolbar>
                <el-button type="primary" plain @click="openCreate">{{ $t('commons.button.create') }}</el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" prop="name">
                <template #default="{ row }">
                    <span v-if="!row.edit">
                        {{ row.name }}
                        <span v-if="row.default">({{ $t('website.default') }})</span>
                    </span>
                    <el-input v-if="row.edit" v-model="row.name"></el-input>
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.operate')">
                <template #default="{ row, $index }">
                    <div>
                        <el-button link v-if="row.edit" type="primary" @click="saveGroup(row, false)">
                            {{ $t('commons.button.save') }}
                        </el-button>
                        <el-button link v-if="!row.edit" type="primary" @click="editGroup($index)">
                            {{ $t('commons.button.edit') }}
                        </el-button>
                        <el-button
                            link
                            v-if="!row.edit"
                            :disabled="row.default"
                            type="primary"
                            @click="deleteGroup($index)"
                        >
                            {{ $t('commons.button.delete') }}
                        </el-button>
                        <el-button link v-if="row.edit" type="primary" @click="cancelEdit($index)">
                            {{ $t('commons.button.cancel') }}
                        </el-button>
                        <el-button link v-if="!row.edit && !row.default" type="primary" @click="saveGroup(row, true)">
                            {{ $t('website.setDefault') }}
                        </el-button>
                    </div>
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
    data.value = [];
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

const saveGroup = (create: groupData, isDefault: boolean) => {
    const group = {
        name: create.name,
        id: create.id,
        default: create.default,
    };
    if (isDefault) {
        group.default = isDefault;
    }
    if (create.id == 0) {
        CreateGroup(group).then(() => {
            ElMessage.success(i18n.global.t('commons.msg.createSuccess'));
            search();
        });
    } else {
        UpdateGroup(group).then(() => {
            ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
            search();
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

const editGroup = (index: number) => {
    data.value[index].edit = true;
};

const cancelEdit = (index: number) => {
    if (data.value[index].id == 0) {
        data.value.splice(index, 1);
    } else {
        data.value[index].edit = false;
    }
};

defineExpose({ acceptParams });
</script>
