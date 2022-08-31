<template>
    <el-row style="margin: 20px; margin-left: 20px" class="row-box" :gutter="20">
        <el-col :span="8">
            <el-card class="el-card">
                <el-tooltip class="box-item" effect="dark" content="创建连接" placement="top-start">
                    <el-button icon="Plus" @click="restHostForm" size="small" />
                </el-tooltip>
                <el-tooltip class="box-item" effect="dark" content="创建分组" placement="top-start">
                    <el-button icon="FolderAdd" @click="onGroupCreate" size="small" />
                </el-tooltip>
                <el-tooltip class="box-item" effect="dark" content="展开" placement="top-start">
                    <el-button icon="Expand" @click="setTreeStatus(true)" size="small" />
                </el-tooltip>
                <el-tooltip class="box-item" effect="dark" content="收缩" placement="top-start">
                    <el-button icon="Fold" @click="setTreeStatus(false)" size="small" />
                </el-tooltip>
                <el-input
                    size="small"
                    @input="loadHostTree"
                    clearable
                    style="margin-top: 5px"
                    v-model="searcConfig.info"
                >
                    <template #append><el-button icon="search" @click="loadHostTree" /></template>
                </el-input>
                <el-input
                    size="small"
                    v-if="groupInputShow"
                    clearable
                    style="margin-top: 5px"
                    v-model="groupInputValue"
                >
                    <template #append>
                        <el-button-group>
                            <el-button icon="Check" @click="onCreateGroup" />
                            <el-button icon="Close" @click="groupInputShow = false" />
                        </el-button-group>
                    </template>
                </el-input>
                <el-tree
                    ref="tree"
                    :expand-on-click-node="false"
                    node-key="id"
                    :default-expand-all="true"
                    :data="hostTree"
                    :props="defaultProps"
                    draggable
                >
                    <template #default="{ node, data }">
                        <span class="custom-tree-node" @mouseover="hover = data.id" @mouseleave="hover = null">
                            <span>
                                <a @click="onEdit(node, data)">{{ node.label }}</a>
                            </span>
                            <el-button-group
                                v-if="!(node.level === 1 && data.label === 'default') && data.id === hover"
                            >
                                <el-button icon="Edit" size="small" @click="onEdit(node, data)" />
                                <el-button icon="Delete" size="small" @click="onDelete(node, data)" />
                            </el-button-group>
                        </span>
                    </template>
                </el-tree>
            </el-card>
        </el-col>
        <el-col :span="16">
            <el-card class="el-card">
                <el-form ref="hostInfoRef" label-width="100px" label-position="left" :model="hostInfo" :rules="rules">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input clearable v-model="hostInfo.name" />
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.group')" prop="groupBelong">
                        <el-select v-model="hostInfo.groupBelong" clearable style="width: 100%">
                            <el-option v-for="item in groupList" :key="item.id" :label="item.name" :value="item.name" />
                        </el-select>
                    </el-form-item>
                    <el-form-item label="IP" prop="addr">
                        <el-input clearable v-model="hostInfo.addr" />
                    </el-form-item>
                    <el-form-item :label="$t('terminal.port')" prop="port">
                        <el-input clearable v-model.number="hostInfo.port" />
                    </el-form-item>
                    <el-form-item :label="$t('terminal.user')" prop="user">
                        <el-input clearable v-model="hostInfo.user" />
                    </el-form-item>
                    <el-form-item :label="$t('terminal.authMode')" prop="authMode">
                        <el-radio-group v-model="hostInfo.authMode">
                            <el-radio label="password">{{ $t('terminal.passwordMode') }}</el-radio>
                            <el-radio label="key">{{ $t('terminal.keyMode') }}</el-radio>
                        </el-radio-group>
                    </el-form-item>
                    <el-form-item
                        :label="$t('terminal.password')"
                        v-if="hostInfo.authMode === 'password'"
                        prop="password"
                    >
                        <el-input clearable show-password type="password" v-model="hostInfo.password" />
                    </el-form-item>
                    <el-form-item :label="$t('terminal.key')" v-if="hostInfo.authMode === 'key'" prop="privateKey">
                        <el-input clearable type="textarea" v-model="hostInfo.privateKey" />
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.description')" prop="description">
                        <el-input clearable type="textarea" v-model="hostInfo.description" />
                    </el-form-item>
                    <el-form-item>
                        <el-button @click="restHostForm">
                            {{ $t('commons.button.reset') }}
                        </el-button>
                        <el-button v-if="hostOperation === 'create'" type="primary" @click="submitAddHost(hostInfoRef)">
                            {{ $t('commons.button.create') }}
                        </el-button>
                        <el-button v-if="hostOperation === 'edit'" type="primary" @click="submitAddHost(hostInfoRef)">
                            {{ $t('commons.button.confirm') }}
                        </el-button>
                    </el-form-item>
                </el-form>
            </el-card>
        </el-col>
    </el-row>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue';
import type { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rues';
import { Host } from '@/api/interface/host';
import { Group } from '@/api/interface/group';
import { getHostList, getHostInfo, addHost, editHost, deleteHost } from '@/api/modules/host';
import { getGroupList, addGroup, editGroup, deleteGroup } from '@/api/modules/group';
import { useDeleteData } from '@/hooks/use-delete-data';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';
import type Node from 'element-plus/es/components/tree/src/model/node';

type FormInstance = InstanceType<typeof ElForm>;
const hostInfoRef = ref<FormInstance>();
const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
    group: [Rules.requiredSelect],
    addr: [Rules.requiredInput, Rules.ip],
    port: [Rules.requiredInput, Rules.port],
    user: [Rules.requiredInput],
    authMode: [Rules.requiredSelect],
    password: [Rules.requiredInput],
    privateKey: [Rules.requiredInput],
});
let hostOperation = ref<string>('create');
let hostInfo = reactive<Host.HostOperate>({
    id: 0,
    name: '',
    groupBelong: 'default',
    addr: '',
    port: 22,
    user: '',
    authMode: 'password',
    password: '',
    privateKey: '',
    description: '',
});

interface Tree {
    id: number;
    label: string;
    children?: Tree[];
}

let searcConfig = reactive<Host.ReqSearch>({
    info: '',
});
const tree = ref<any>(null);
const hover = ref();
const hostTree = ref<Array<Host.HostTree>>();
const defaultProps = {
    label: 'label',
    children: 'children',
};

const groupList = ref<Array<Group.GroupInfo>>();

let groupInputValue = ref();
let currentGroupID = ref();
let groupOperation = ref<string>('create');
let groupInputShow = ref<boolean>(false);

const loadHostTree = async () => {
    const res = await getHostList(searcConfig);
    hostTree.value = res.data;
};

const loadGroups = async () => {
    const res = await getGroupList({ type: 'host' });
    groupList.value = res.data;
};

function setTreeStatus(expend: boolean) {
    for (let i = 0; i < tree.value.store._getAllNodes().length; i++) {
        tree.value.store._getAllNodes()[i].expanded = expend;
    }
}

function restHostForm() {
    if (hostInfoRef.value) {
        hostInfoRef.value.resetFields();
    }
}

const submitAddHost = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (hostOperation.value === 'create') {
            await addHost(hostInfo);
        } else {
            await editHost(hostInfo);
        }
        restHostForm();
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        loadHostTree();
    });
};

const onGroupCreate = () => {
    groupInputShow.value = true;
    groupInputValue.value = '';
    groupOperation.value = 'create';
};
const onCreateGroup = async () => {
    console.log(groupOperation.value);
    if (groupOperation.value === 'create') {
        let group = { id: 0, name: groupInputValue.value, type: 'host' };
        await addGroup(group);
        groupOperation.value = '';
        groupInputShow.value = false;
    } else {
        let group = { id: currentGroupID.value, name: groupInputValue.value, type: 'host' };
        await editGroup(group);
    }
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    groupOperation.value = '';
    groupInputShow.value = false;
    loadHostTree();
    loadGroups();
};

const onDelete = async (node: Node, data: Tree) => {
    if (node.level === 1 && data.label === 'default') {
        return;
    }
    if (node.level === 1) {
        await useDeleteData(deleteGroup, data.id - 10000, '移除组后，组内所有连接将迁移到 default 组内，是否确认？');
        loadGroups();
    } else {
        await useDeleteData(deleteHost, data.id, 'commons.msg.delete');
    }
    loadHostTree();
    loadGroups();
};

const onEdit = async (node: Node, data: Tree) => {
    if (node.level === 1 && data.label === 'default') {
        return;
    }
    console.log(node.level === 1);
    if (node.level === 1) {
        groupInputShow.value = true;
        groupInputValue.value = data.label;
        currentGroupID.value = data.id - 10000;
        groupOperation.value = 'edit';
        console.log(groupOperation.value);
        return;
    } else {
        const res = await getHostInfo(data.id);
        hostInfo.id = res.data.id;
        hostInfo.name = res.data.name;
        hostInfo.groupBelong = res.data.groupBelong;
        hostInfo.addr = res.data.addr;
        hostInfo.port = res.data.port;
        hostInfo.user = res.data.user;
        hostInfo.description = res.data.description;
        hostOperation.value = 'edit';
    }
};

onMounted(() => {
    loadHostTree();
    loadGroups();
});
</script>

<style>
.custom-tree-node {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 14px;
    padding-right: 8px;
}
</style>
