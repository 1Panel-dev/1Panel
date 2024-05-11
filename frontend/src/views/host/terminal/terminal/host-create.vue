<template>
    <div>
        <el-drawer
            v-model="dialogVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="50%"
        >
            <template #header>
                <DrawerHeader :header="$t('terminal.addHost')" :back="handleClose" />
            </template>
            <el-form ref="hostRef" label-width="100px" label-position="top" :model="hostInfo" :rules="rules">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-alert
                            v-if="isLocal"
                            class="common-prompt"
                            center
                            :title="$t('terminal.connLocalErr')"
                            :closable="false"
                            type="warning"
                        />
                        <el-form-item :label="$t('terminal.ip')" prop="addr">
                            <el-input @change="isOK = false" v-if="!isLocal" clearable v-model.trim="hostInfo.addr" />
                            <el-tag v-if="isLocal">{{ hostInfo.addr }}</el-tag>
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
                        <el-form-item
                            :label="$t('commons.login.password')"
                            v-if="hostInfo.authMode === 'password'"
                            prop="password"
                        >
                            <el-input
                                @change="isOK = false"
                                clearable
                                show-password
                                type="password"
                                v-model="hostInfo.password"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('terminal.key')" v-if="hostInfo.authMode === 'key'" prop="privateKey">
                            <el-input @change="isOK = false" clearable type="textarea" v-model="hostInfo.privateKey" />
                        </el-form-item>
                        <el-form-item
                            :label="$t('terminal.keyPassword')"
                            v-if="hostInfo.authMode === 'key'"
                            prop="passPhrase"
                        >
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
                        <el-form-item style="margin-top: 10px" :label="$t('commons.table.port')" prop="port">
                            <el-input @change="isOK = false" clearable v-model.number="hostInfo.port" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.title')" prop="name">
                            <el-input clearable v-model="hostInfo.name" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.description')" prop="description">
                            <el-input clearable v-model="hostInfo.description" />
                        </el-form-item>
                    </el-col>
                </el-row>
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
        </el-drawer>
    </div>
</template>

<script setup lang="ts">
import { ElForm } from 'element-plus';
import { Host } from '@/api/interface/host';
import { Rules } from '@/global/form-rules';
import { addHost, testByInfo } from '@/api/modules/host';
import DrawerHeader from '@/components/drawer-header/index.vue';
import i18n from '@/lang';
import { reactive, ref } from 'vue';
import { MsgError, MsgSuccess } from '@/utils/message';

const dialogVisible = ref();
const isOK = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const hostRef = ref<FormInstance>();

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
    hostInfo.addr = '';
    hostInfo.name = '';
    hostInfo.groupID = 0;
    hostInfo.addr = '';
    hostInfo.port = 22;
    hostInfo.user = '';
    hostInfo.authMode = 'password';
    hostInfo.password = '';
    hostInfo.privateKey = '';
    hostInfo.description = '';
    isLocal.value = props.isLocal;
    if (props.isLocal) {
        hostInfo.addr = '127.0.0.1';
        hostInfo.user = 'root';
    }
    dialogVisible.value = true;
};

const handleClose = () => {
    dialogVisible.value = false;
};

const emit = defineEmits(['on-conn-terminal', 'load-host-tree']);

const submitAddHost = (formEl: FormInstance | undefined, ops: string) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        hostInfo.groupID = 0;
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
                const res = await addHost(hostInfo);
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
