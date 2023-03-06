<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="50%" :before-close="handleClose">
        <template #header>
            <Header :header="$t('website.group')" :back="handleClose"></Header>
        </template>

        <ComplexTable :data="data" @search="search()">
            <template #toolbar>
                <el-button type="primary" @click="openCreate">{{ $t('website.createGroup') }}</el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" prop="name">
                <template #default="{ row }">
                    <span v-if="!row.edit">
                        {{ row.name }}
                        <span v-if="row.isDefault">({{ $t('website.default') }})</span>
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
                            :disabled="row.isDefault"
                            type="primary"
                            @click="deleteGroup($index)"
                        >
                            {{ $t('commons.button.delete') }}
                        </el-button>
                        <el-button link v-if="row.edit" type="primary" @click="cancelEdit($index)">
                            {{ $t('commons.button.cancel') }}
                        </el-button>
                        <el-button link v-if="!row.edit && !row.isDefault" type="primary" @click="saveGroup(row, true)">
                            {{ $t('website.setDefault') }}
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
import { CreateGroup, DeleteGroup, GetGroupList, UpdateGroup } from '@/api/modules/host';
import Header from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { Group } from '@/api/interface/group';

const open = ref(false);
const type = ref();
const data = ref();
const handleClose = () => {
    open.value = false;
    data.value = [];
    emit('search');
};

interface DialogProps {
    type: string;
}
const acceptParams = (params: DialogProps): void => {
    type.value = params.type;
    open.value = true;
    search();
};
const emit = defineEmits<{ (e: 'search'): void }>();

const search = () => {
    data.value = [];
    GetGroupList({ type: type.value }).then((res) => {
        for (const d of res.data) {
            const g = {
                id: d.id,
                name: d.name,
                isDefault: d.isDefault,
                edit: false,
            };
            data.value.push(g);
        }
    });
};

const saveGroup = (group: Group.GroupInfo, isDefault: boolean) => {
    group.type = type.value;
    if (group.id == 0) {
        CreateGroup(group).then(() => {
            MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
            search();
        });
    } else {
        group.isDefault = isDefault;
        UpdateGroup(group).then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            search();
        });
    }
};

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
        name: '',
        isDefault: false,
        edit: true,
    };
    data.value.push(g);
};

const deleteGroup = (index: number) => {
    const group = data.value[index];

    if (group.id > 0) {
        DeleteGroup(group.id).then(() => {
            data.value.splice(index, 1);
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
        });
    } else {
        data.value.splice(index, 1);
    }
};

const editGroup = (index: number) => {
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
};

defineExpose({ acceptParams });
</script>
