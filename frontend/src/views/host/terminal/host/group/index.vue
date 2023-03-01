<template>
    <el-drawer :close-on-click-modal="false" v-model="drawerVisiable" size="50%" :before-close="handleClose">
        <template #header>
            <Header :header="$t('website.group')" :back="handleClose"></Header>
        </template>

        <ComplexTable v-loading="loading" :data="data" @search="search()">
            <template #toolbar>
                <el-button type="primary" @click="openCreate">{{ $t('website.createGroup') }}</el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" prop="name">
                <template #default="{ row }">
                    <span v-if="!row.edit">
                        {{ row.name }}
                    </span>
                    <el-input v-if="row.edit" v-model="row.name"></el-input>
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.operate')">
                <template #default="{ row, $index }">
                    <div>
                        <el-button link v-if="row.edit" type="primary" @click="onSaveGroup(row)">
                            {{ $t('commons.button.save') }}
                        </el-button>
                        <el-button
                            link
                            v-if="!row.edit"
                            :disabled="row.name === 'default'"
                            type="primary"
                            @click="onEditGroup($index)"
                        >
                            {{ $t('commons.button.edit') }}
                        </el-button>
                        <el-button
                            link
                            v-if="!row.edit"
                            :disabled="row.name === 'default'"
                            type="primary"
                            @click="onDeleteGroup($index)"
                        >
                            {{ $t('commons.button.delete') }}
                        </el-button>
                        <el-button link v-if="row.edit" type="primary" @click="cancelEdit($index)">
                            {{ $t('commons.button.cancel') }}
                        </el-button>
                    </div>
                </template>
            </el-table-column>
        </ComplexTable>
    </el-drawer>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import i18n from '@/lang';
import ComplexTable from '@/components/complex-table/index.vue';
import Header from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { addGroup, deleteGroup, editGroup, getGroupList } from '@/api/modules/host';
import { useDeleteData } from '@/hooks/use-delete-data';

const loading = ref();

interface groupData {
    id: number;
    name: string;
}

let drawerVisiable = ref(false);
let data = ref();
const handleClose = () => {
    drawerVisiable.value = false;
    data.value = [];
    emit('search');
};

const search = () => {
    data.value = [];
    getGroupList({ type: 'host' }).then((res) => {
        for (const d of res.data) {
            const g = {
                id: d.id,
                name: d.name,
                edit: false,
            };
            data.value.push(g);
        }
    });
};

const onSaveGroup = (create: groupData) => {
    const group = {
        id: create.id,
        type: 'host',
        name: create.name,
    };
    loading.value = true;
    if (group.id == 0) {
        addGroup(group)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    } else {
        editGroup(group)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    }
};

const acceptParams = async () => {
    drawerVisiable.value = true;
    search();
};
const emit = defineEmits<{ (e: 'search'): void }>();

const openCreate = () => {
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
        type: 'host',
        name: '',
        edit: true,
    };
    data.value.push(g);
};

const onDeleteGroup = async (index: number) => {
    const group = data.value[index];
    if (group.id > 0) {
        await useDeleteData(deleteGroup, group.id, 'terminal.groupDeleteHelper')
            .then(() => {
                loading.value = false;
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    } else {
        data.value.splice(index, 1);
    }
};

const onEditGroup = (index: number) => {
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

const cancelEdit = (index: number) => {
    if (data.value[index].id == 0) {
        data.value.splice(index, 1);
    } else {
        data.value[index].edit = false;
    }
    search();
};

defineExpose({ acceptParams });
</script>
