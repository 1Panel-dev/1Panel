<template>
    <el-row style="margin: 20px; margin-left: 20px" class="row-box" :gutter="20">
        <el-col :span="8">
            <el-card class="el-card">
                <el-button icon="Plus" @click="readyForCreate" size="small" />
                <el-button icon="FolderAdd" @click="(folderCreate = true), (newGroupName = '')" size="small" />
                <el-button icon="Expand" @click="setTreeStatus(true)" size="small" />
                <el-button icon="Fold" @click="setTreeStatus(false)" size="small" />
                <el-input size="small" @input="loadHost" clearable style="margin-top: 5px" v-model="searcConfig.info">
                    <template #append><el-button icon="search" @click="loadHost" /></template>
                </el-input>
                <el-input size="small" v-if="folderCreate" clearable style="margin-top: 5px" v-model="newGroupName">
                    <template #append>
                        <el-button-group>
                            <el-button icon="Check" @click="loadHost" />
                            <el-button icon="Close" @click="folderCreate = false" />
                        </el-button-group>
                    </template>
                </el-input>
                <el-tree ref="tree" :default-expand-all="true" :data="hostTree" :props="defaultProps" />
            </el-card>
        </el-col>
        <el-col :span="16">
            <el-card class="el-card">
                <el-form ref="hostInfoRef" label-width="100px" label-position="left" :model="hostInfo" :rules="rules">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input clearable v-model="hostInfo.name" />
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
                    <el-form-item>
                        <el-button type="primary" @click="submitAddHost(hostInfoRef)">
                            {{ $t('commons.button.create') }}
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
import { getHostList, addHost } from '@/api/modules/host';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';

type FormInstance = InstanceType<typeof ElForm>;
const hostInfoRef = ref<FormInstance>();
const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
    addr: [Rules.requiredInput, Rules.ip],
    port: [Rules.requiredInput, Rules.port],
    user: [Rules.requiredInput],
    authMode: [Rules.requiredSelect],
    password: [Rules.requiredInput],
    privateKey: [Rules.requiredInput],
});

let hostInfo = reactive<Host.HostOperate>({
    id: 0,
    name: '',
    addr: '',
    port: 22,
    user: '',
    authMode: 'password',
    password: '',
    privateKey: '',
    description: '',
});

let searcConfig = reactive<Host.ReqSearch>({
    info: '',
});
const tree = ref<any>(null);
const hostTree = ref<Array<Host.HostTree>>();
const defaultProps = {
    children: 'children',
    label: 'label',
};
const newGroupName = ref();
const folderCreate = ref<boolean>(false);

const loadHost = async () => {
    const res = await getHostList(searcConfig);
    hostTree.value = res.data;
};

function setTreeStatus(expend: boolean) {
    for (let i = 0; i < tree.value.store._getAllNodes().length; i++) {
        tree.value.store._getAllNodes()[i].expanded = expend;
    }
}

function readyForCreate() {
    if (hostInfoRef.value) {
        hostInfoRef.value.resetFields();
    }
}

const submitAddHost = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        try {
            await addHost(hostInfo);
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            // loadHost();
        } catch (error) {
            ElMessage.success(i18n.global.t('commons.msg.loginSuccess') + ':' + error);
        }
    });
};

onMounted(() => {
    loadHost();
});
</script>
