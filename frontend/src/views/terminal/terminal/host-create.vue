<template>
    <DrawerPro v-model="dialogVisible" :header="$t('terminal.addHost')" @close="handleClose" size="large">
        <el-form ref="hostRef" label-width="100px" label-position="top" :model="hostInfo" :rules="rules">
            <el-alert
                v-if="isLocal"
                class="common-prompt"
                center
                :title="$t('terminal.connLocalErr')"
                :closable="false"
                type="warning"
            />
            <el-form-item :label="$t('terminal.ip')" prop="addr">
                <el-input @change="isOK = false" clearable v-model.trim="hostInfo.addr" />
            </el-form-item>
            <el-form-item :label="$t('commons.login.username')" prop="user">
                <el-input @change="isOK = false" clearable v-model="hostInfo.user" />
            </el-form-item>
            <el-form-item :label="$t('terminal.authMode')" prop="authMode">
                <el-radio-group @change="isOK = false" v-model="hostInfo.authMode">
                    <el-radio value="password">{{ $t('terminal.passwordMode') }}</el-radio>
                    <el-radio value="key">{{ $t('terminal.keyMode') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item :label="$t('commons.login.password')" v-if="hostInfo.authMode === 'password'" prop="password">
                <el-input @change="isOK = false" clearable show-password type="password" v-model="hostInfo.password" />
            </el-form-item>
            <el-form-item :label="$t('terminal.key')" v-if="hostInfo.authMode === 'key'" prop="privateKey">
                <el-input @change="isOK = false" clearable type="textarea" v-model="hostInfo.privateKey" />
            </el-form-item>
            <el-form-item :label="$t('terminal.keyPassword')" v-if="hostInfo.authMode === 'key'" prop="passPhrase">
                <el-input
                    @change="isOK = false"
                    type="password"
                    show-password
                    clearable
                    v-model="hostInfo.passPhrase"
                />
            </el-form-item>
            <el-checkbox clearable v-model.number="hostInfo.rememberPassword">
                {{ $t('terminal.rememberPassword') }}
            </el-checkbox>
            <el-form-item class="mt-2.5" :label="$t('commons.table.port')" prop="port">
                <el-input @change="isOK = false" clearable v-model.number="hostInfo.port" />
            </el-form-item>
            <el-form-item :label="$t('commons.table.group')" prop="groupID">
                <el-select filterable v-model="hostInfo.groupID" clearable style="width: 100%">
                    <div v-for="item in groupList" :key="item.id">
                        <el-option
                            v-if="item.name === 'default'"
                            :label="$t('commons.table.default')"
                            :value="item.id"
                        />
                        <el-option v-else :label="item.name" :value="item.id" />
                    </div>
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input clearable v-model="hostInfo.description" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button @click="submitAddHost(hostRef, 'testConn')">
                    {{ $t('terminal.testConn') }}
                </el-button>
                <el-button type="primary" :disabled="!isOK" @click="submitAddHost(hostRef, 'saveAndConn')">
                    {{ $t('terminal.saveAndConn') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { ElForm } from 'element-plus';
import { Host } from '@/api/interface/host';
import { Rules } from '@/global/form-rules';
import { addHost, editHost, getHostByID, testByInfo } from '@/api/modules/terminal';
import i18n from '@/lang';
import { reactive, ref } from 'vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { getGroupList } from '@/api/modules/group';

const dialogVisible = ref();
const isOK = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const hostRef = ref<FormInstance>();

const groupList = ref();
const defaultGroup = ref();

let hostInfo = reactive<Host.HostOperate>({
    id: 0,
    name: '',
    groupID: 0,
    addr: '',
    port: 22,
    user: '',
    authMode: 'password',
    password: '',
    privateKey: '',
    passPhrase: '',
    rememberPassword: false,
    description: '',
});

const rules = reactive({
    addr: [Rules.ipV4V6OrDomain],
    port: [Rules.requiredInput, Rules.port],
    user: [Rules.requiredInput],
    authMode: [Rules.requiredSelect],
    password: [Rules.requiredInput],
    privateKey: [Rules.requiredInput],
});

const isLocal = ref(false);
interface DialogProps {
    isLocal: boolean;
}
const acceptParams = (props: DialogProps) => {
    isLocal.value = props.isLocal;
    loadGroups();
    dialogVisible.value = true;
};

const handleClose = () => {
    dialogVisible.value = false;
};

const emit = defineEmits(['on-conn-terminal', 'load-host-tree']);

const loadGroups = async () => {
    const res = await getGroupList('host');
    groupList.value = res.data;
    for (const item of groupList.value) {
        if (item.isDefault) {
            defaultGroup.value = item.id;
            break;
        }
    }
    if (isLocal.value) {
        loadLocal();
    } else {
        setDefault();
    }
};
const loadLocal = async () => {
    await getHostByID(0)
        .then((res) => {
            hostInfo.id = res.data.id || 0;
            hostInfo.addr = res.data.addr || '';
            hostInfo.name = 'local';
            hostInfo.groupID = res.data.groupID || defaultGroup.value;
            hostInfo.port = res.data.port || 22;
            hostInfo.user = res.data.user || 'root';
            hostInfo.authMode = res.data.authMode || 'password';
            hostInfo.password = res.data.password || '';
            hostInfo.privateKey = res.data.privateKey || '';
            hostInfo.description = res.data.description || '';
        })
        .catch(() => {
            setDefault();
        });
};

const setDefault = () => {
    hostInfo.addr = '';
    hostInfo.name = 'local';
    hostInfo.groupID = defaultGroup.value;
    hostInfo.port = 22;
    hostInfo.user = '';
    hostInfo.authMode = 'password';
    hostInfo.password = '';
    hostInfo.privateKey = '';
    hostInfo.description = '';
};

const submitAddHost = (formEl: FormInstance | undefined, ops: string) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        switch (ops) {
            case 'testConn':
                await testByInfo(hostInfo).then((res) => {
                    if (res.data) {
                        isOK.value = true;
                        MsgSuccess(i18n.global.t('terminal.connTestOk'));
                    } else {
                        isOK.value = false;
                        MsgError(i18n.global.t('terminal.connTestFailed'));
                    }
                });
                break;
            case 'saveAndConn':
                let res;
                if (hostInfo.id == 0) {
                    res = await addHost(hostInfo);
                } else {
                    res = await editHost(hostInfo);
                }
                dialogVisible.value = false;
                let title = res.data.user + '@' + res.data.addr + ':' + res.data.port;
                if (res.data.name.length !== 0) {
                    title = res.data.name + '-' + title;
                }
                let isLocal = hostInfo.addr === '127.0.0.1';
                emit('on-conn-terminal', title, res.data.id, isLocal);
                emit('load-host-tree');
                break;
        }
    });
};

defineExpose({
    acceptParams,
});
</script>
