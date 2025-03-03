<template>
    <DrawerPro v-model="open" :header="$t('commons.table.group')" @close="handleClose" size="large" :back="handleClose">
        <template #content>
            <ComplexTable :data="data" @search="search()">
                <template #toolbar>
                    <el-button type="primary" @click="openCreate">{{ $t('website.createGroup') }}</el-button>
                </template>
                <el-table-column :label="$t('commons.table.name')" prop="name">
                    <template #default="{ row }">
                        <div v-if="!row.edit">
                            <span v-if="row.name === 'default'">
                                {{ $t('commons.table.default') }}
                            </span>
                            <span v-if="row.name !== 'default'">{{ row.name }}</span>
                            <el-tag v-if="row.isDefault" type="success" class="ml-2" size="small">
                                ({{ $t('commons.table.default') }})
                            </el-tag>

                            <el-tag type="warning" size="small" class="ml-4" v-if="row.isDelete">
                                {{ $t('app.takeDown') }}
                            </el-tag>
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
                                @click="removeGroup($index)"
                            >
                                {{ $t('commons.button.delete') }}
                            </el-button>
                            <el-button link v-if="row.edit" type="primary" @click="search()">
                                {{ $t('commons.button.cancel') }}
                            </el-button>
                            <el-button
                                link
                                v-if="!row.edit && !row.isDefault && !row.isDelete"
                                type="primary"
                                @click="setDefault(row)"
                            >
                                {{ $t('website.setDefault') }}
                            </el-button>
                        </div>
                    </template>
                </el-table-column>
            </ComplexTable>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import i18n from '@/lang';
import { createAgentGroup, deleteAgentGroup, getAgentGroupList, updateAgentGroup } from '@/api/modules/group';
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
    getAgentGroupList(type.value).then((res) => {
        data.value = res.data || [];
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
            createAgentGroup(group).then(() => {
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                search();
            });
        } else {
            updateAgentGroup(group).then(() => {
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                search();
            });
        }
    });
};

const setDefault = (group: Group.GroupInfo) => {
    group.isDefault = true;
    group.type = type.value;
    updateAgentGroup(group).then(() => {
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
        status: 'Enable',
    };
    data.value.unshift(g);
};

const removeGroup = (index: number) => {
    const group = data.value[index];

    if (group.id > 0) {
        deleteAgentGroup(group.id).then(() => {
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
