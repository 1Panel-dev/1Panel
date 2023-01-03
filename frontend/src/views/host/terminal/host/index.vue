<template>
    <div>
        <el-row class="row-box" style="margin-top: 20px" :gutter="20">
            <el-col :span="8">
                <el-card class="el-card">
                    <el-tooltip
                        class="box-item"
                        effect="dark"
                        :content="$t('terminal.createConn')"
                        placement="top-start"
                    >
                        <el-button icon="Plus" @click="restHostForm" />
                    </el-tooltip>
                    <el-tooltip
                        class="box-item"
                        effect="dark"
                        :content="$t('terminal.createGroup')"
                        placement="top-start"
                    >
                        <el-button icon="FolderAdd" @click="onGroupCreate" />
                    </el-tooltip>
                    <el-tooltip class="box-item" effect="dark" :content="$t('terminal.expand')" placement="top-start">
                        <el-button icon="Expand" @click="setTreeStatus(true)" />
                    </el-tooltip>
                    <el-tooltip class="box-item" effect="dark" :content="$t('terminal.fold')" placement="top-start">
                        <el-button icon="Fold" @click="setTreeStatus(false)" />
                    </el-tooltip>
                    <el-input @input="loadHostTree" clearable style="margin-top: 5px" v-model="searcConfig.info">
                        <template #append><el-button icon="search" @click="loadHostTree" /></template>
                    </el-input>
                    <el-input v-if="groupInputShow" clearable style="margin-top: 5px" v-model="groupInputValue">
                        <template #append>
                            <el-button-group>
                                <el-button icon="Check" @click="onCreateGroup(groupInputValue)" />
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
                    >
                        <template #default="{ node, data }">
                            <span class="custom-tree-node" @mouseover="hover = data.id" @mouseleave="hover = null">
                                <div v-if="node.label !== currentGroup || !data.onEdit">
                                    <span v-if="node.label.length <= 35">{{ node.label }}</span>
                                    <el-tooltip v-else :content="node.label" placement="top-start">
                                        <span>{{ node.label.substring(0, 32) }}...</span>
                                    </el-tooltip>
                                </div>
                                <el-input v-else v-model="currentGroupValue"></el-input>
                                <div
                                    style="margin-left: 10px"
                                    v-if="!(node.level === 1 && data.label === 'default') && data.id === hover"
                                >
                                    <el-button v-if="!data.onEdit" icon="Edit" link @click="onEdit(node, data)" />
                                    <el-button
                                        v-if="!data.onEdit && node.label.indexOf('@127.0.0.1:') === -1"
                                        icon="Delete"
                                        link
                                        @click="onDelete(node, data)"
                                    />
                                    <el-button v-if="data.onEdit" icon="Check" link @click="onUpdateGroup()" />
                                    <el-button v-if="data.onEdit" icon="Close" link @click="data.onEdit = false" />
                                </div>
                            </span>
                        </template>
                    </el-tree>
                </el-card>
            </el-col>
            <el-col :span="16">
                <el-card class="el-card">
                    <el-form ref="hostInfoRef" label-width="100px" :model="hostInfo" :rules="rules">
                        <el-form-item :label="$t('terminal.ip')" prop="addr">
                            <span v-if="hostInfo.addr === '127.0.0.1' && hostOperation === 'edit'">
                                {{ hostInfo.addr }}
                            </span>
                            <el-input v-else clearable v-model="hostInfo.addr" />
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
                        <el-form-item :label="$t('terminal.port')" prop="port">
                            <el-input clearable v-model.number="hostInfo.port" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.group')" prop="groupBelong">
                            <el-select filterable v-model="hostInfo.groupBelong" clearable style="width: 100%">
                                <el-option
                                    v-for="item in groupList"
                                    :key="item.id"
                                    :label="item.name"
                                    :value="item.name"
                                />
                            </el-select>
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.title')" prop="name">
                            <el-input clearable v-model="hostInfo.name" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.description')" prop="description">
                            <el-input clearable type="textarea" v-model="hostInfo.description" />
                        </el-form-item>
                        <el-form-item>
                            <el-button @click="restHostForm">
                                {{ $t('commons.button.reset') }}
                            </el-button>
                            <el-button @click="submitAddHost(hostInfoRef, 'testconn')">
                                {{ $t('terminal.testConn') }}
                            </el-button>
                            <el-button
                                v-if="hostOperation === 'create'"
                                type="primary"
                                @click="submitAddHost(hostInfoRef, 'create')"
                            >
                                {{ $t('commons.button.save') }}
                            </el-button>
                            <el-button
                                v-if="hostOperation === 'edit'"
                                type="primary"
                                @click="submitAddHost(hostInfoRef, 'edit')"
                            >
                                {{ $t('commons.button.confirm') }}
                            </el-button>
                        </el-form-item>
                    </el-form>
                </el-card>
            </el-col>
        </el-row>
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive } from 'vue';
import type { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { Host } from '@/api/interface/host';
import { Group } from '@/api/interface/group';
import { getHostTree, getHostInfo, addHost, editHost, deleteHost, testByInfo } from '@/api/modules/host';
import { getGroupList, addGroup, editGroup, deleteGroup } from '@/api/modules/group';
import { useDeleteData } from '@/hooks/use-delete-data';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';
import type Node from 'element-plus/es/components/tree/src/model/node';

type FormInstance = InstanceType<typeof ElForm>;
const hostInfoRef = ref<FormInstance>();
const rules = reactive({
    groupBelong: [Rules.requiredSelect],
    addr: [Rules.requiredInput],
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
    onEdit: boolean;
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

const groupInputValue = ref();
const currentGroup = ref();
const currentGroupID = ref();
const currentGroupValue = ref();

let groupOperation = ref<string>('create');
let groupInputShow = ref<boolean>(false);

const loadHostTree = async () => {
    const res = await getHostTree(searcConfig);
    hostTree.value = res.data;
};

const acceptParams = () => {
    loadHostTree();
    loadGroups();
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

const submitAddHost = (formEl: FormInstance | undefined, ops: string) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        switch (ops) {
            case 'create':
                await addHost(hostInfo);
                restHostForm();
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
                loadHostTree();
                break;
            case 'edit':
                await editHost(hostInfo);
                restHostForm();
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
                loadHostTree();
                break;
            case 'testconn':
                await testByInfo(hostInfo).then((res) => {
                    if (res.data) {
                        ElMessage.success(i18n.global.t('terminal.connTestOk'));
                    } else {
                        ElMessage.success(i18n.global.t('terminal.connTestFailed'));
                    }
                });
                break;
        }
    });
};

const onGroupCreate = () => {
    groupInputShow.value = true;
    groupInputValue.value = '';
    groupOperation.value = 'create';
};
const onCreateGroup = async (name: string) => {
    if (groupOperation.value === 'create') {
        let group = { id: 0, name: name, type: 'host' };
        await addGroup(group);
        groupOperation.value = '';
        groupInputShow.value = false;
    }
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    groupOperation.value = '';
    groupInputShow.value = false;
    loadHostTree();
    loadGroups();
};

const onUpdateGroup = async () => {
    if (currentGroup.value === currentGroupValue.value) {
        currentGroup.value = '';
        return;
    }
    let group = { id: currentGroupID.value, name: currentGroupValue.value, type: 'host' };
    await editGroup(group);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    loadHostTree();
    loadGroups();
};

const onDelete = async (node: Node, data: Tree) => {
    if (node.level === 1 && data.label === 'default') {
        return;
    }
    if (node.level === 1) {
        await useDeleteData(deleteGroup, data.id - 10000, i18n.global.t('terminal.groupDeleteHelper'));
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
    if (node.level === 1) {
        currentGroup.value = data.label;
        currentGroupValue.value = data.label;
        currentGroupID.value = data.id - 10000;
        groupOperation.value = 'edit';
        data.onEdit = true;
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

defineExpose({
    acceptParams,
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
