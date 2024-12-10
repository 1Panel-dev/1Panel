<template>
    <el-drawer
        :close-on-click-modal="false"
        v-model="open"
        size="50%"
        :close-on-press-escape="false"
        :before-close="handleClose"
    >
        <template #header>
            <Header :header="$t('website.manageGroup')" :back="handleClose"></Header>
        </template>

        <ComplexTable :data="data" @search="search()">
            <template #toolbar>
                <el-button type="primary" @click="openCreate">{{ $t('website.createGroup') }}</el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" prop="name">
                <template #default="{ row }">
                    <div v-if="!row.edit">
                        <span v-if="row.name === 'default'">
                            {{ $t('website.default') }}
                        </span>
                        <span v-if="row.name !== 'default'">{{ row.name }}</span>
                        <span v-if="row.isDefault">({{ $t('website.default') }})</span>
                    </div>

                    <el-form @submit.prevent ref="groupForm" v-if="row.edit" :model="row">
                        <el-form-item prop="name" v-if="row.edit" :rules="Rules.name">
                            <div style="margin-top: 20px; width: 100%"><el-input v-model="row.name" /></div>
                        </el-form-item>
                    </el-form>
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.operate')">
                <template #default="{ row, $index }">
                    <div>
                        <el-button link v-if="row.edit" type="primary" @click="saveGroup(groupForm, row)">
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
                        <el-button link v-if="row.edit" type="primary" @click="search()">
                            {{ $t('commons.button.cancel') }}
                        </el-button>
                        <el-button link v-if="!row.edit && !row.isDefault" type="primary" @click="setDefault(row)">
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
import { CreateGroup, DeleteGroup, GetGroupList, UpdateGroup } from '@/api/modules/group';
import Header from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { Group } from '@/api/interface/group';
import { Rules } from '@/global/form-rules';
import { FormInstance } from 'element-plus';

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

const groupForm = ref<FormInstance>();
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

const saveGroup = async (formEl: FormInstance, group: Group.GroupInfo) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        group.type = type.value;
        if (group.id == 0) {
            CreateGroup(group).then(() => {
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                search();
            });
        } else {
            UpdateGroup(group).then(() => {
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                search();
            });
        }
    });
};

const setDefault = (group: Group.GroupInfo) => {
    group.isDefault = true;
    group.type = type.value;
    UpdateGroup(group).then(() => {
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        search();
    });
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
    data.value.unshift(g);
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

defineExpose({ acceptParams });
</script>
