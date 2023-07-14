<template>
    <div v-loading="loading">
        <FireRouter />

        <div class="a-card" style="margin-top: 20px">
            <el-card>
                <div>
                    <el-tag style="float: left" effect="dark" type="success">SSH</el-tag>
                    <el-tag round class="status-content" v-if="form.status === 'Enable'" type="success">
                        {{ $t('commons.status.running') }}
                    </el-tag>
                    <el-popover
                        v-if="form.status === 'Disable'"
                        placement="top-start"
                        trigger="hover"
                        width="450"
                        :content="form.message"
                    >
                        <template #reference>
                            <el-tag round class="status-content" v-if="form.status === 'Disable'" type="info">
                                {{ $t('commons.status.stopped') }}
                            </el-tag>
                        </template>
                    </el-popover>
                    <span v-if="form.status === 'Enable'" class="buttons">
                        <el-button type="primary" @click="onOperate('stop')" link>
                            {{ $t('commons.button.stop') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" @click="onOperate('restart')" link>
                            {{ $t('container.restart') }}
                        </el-button>
                    </span>
                    <span v-if="form.status === 'Disable'" class="buttons">
                        <el-button type="primary" @click="onOperate('start')" link>
                            {{ $t('commons.button.start') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" @click="onOperate('restart')" link>
                            {{ $t('container.restart') }}
                        </el-button>
                    </span>
                </div>
            </el-card>
        </div>

        <LayoutContent style="margin-top: 20px" :title="$t('menu.config')" :divider="true">
            <template #main>
                <el-radio-group v-model="confShowType" @change="changeMode">
                    <el-radio-button label="base">{{ $t('database.baseConf') }}</el-radio-button>
                    <el-radio-button label="all">{{ $t('database.allConf') }}</el-radio-button>
                </el-radio-group>
                <el-row style="margin-top: 20px" v-if="confShowType === 'base'">
                    <el-col :span="1"><br /></el-col>
                    <el-col :xs="24" :sm="20" :md="20" :lg="10" :xl="10">
                        <el-form :model="form" label-position="left" ref="formRef" label-width="120px">
                            <el-form-item :label="$t('commons.table.port')" prop="port">
                                <el-input disabled v-model.number="form.port">
                                    <template #append>
                                        <el-button @click="onChangePort" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('ssh.portHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('ssh.listenAddress')" prop="listenAddress">
                                <el-input disabled v-model="form.listenAddress">
                                    <template #append>
                                        <el-button @click="onChangeAddress" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('ssh.addressHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('ssh.permitRootLogin')" prop="permitRootLoginItem">
                                <el-input disabled v-model="form.permitRootLoginItem">
                                    <template #append>
                                        <el-button @click="onChangeRoot" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('ssh.rootSettingHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('ssh.passwordAuthentication')" prop="passwordAuthentication">
                                <el-switch
                                    active-value="yes"
                                    inactive-value="no"
                                    @change="onSave(formRef, 'PasswordAuthentication', form.passwordAuthentication)"
                                    v-model="form.passwordAuthentication"
                                ></el-switch>
                                <span class="input-help">{{ $t('ssh.pwdAuthHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('ssh.pubkeyAuthentication')" prop="pubkeyAuthentication">
                                <el-switch
                                    active-value="yes"
                                    inactive-value="no"
                                    @change="onSave(formRef, 'PubkeyAuthentication', form.pubkeyAuthentication)"
                                    v-model="form.pubkeyAuthentication"
                                ></el-switch>
                                <span class="input-help">{{ $t('ssh.keyAuthHelper') }}</span>
                                <el-button link @click="onOpenDrawer" type="primary">
                                    {{ $t('ssh.pubkey') }}
                                </el-button>
                            </el-form-item>
                            <el-form-item :label="$t('ssh.useDNS')" prop="useDNS">
                                <el-switch
                                    active-value="yes"
                                    inactive-value="no"
                                    @change="onSave(formRef, 'UseDNS', form.useDNS)"
                                    v-model="form.useDNS"
                                ></el-switch>
                                <span class="input-help">{{ $t('ssh.dnsHelper') }}</span>
                            </el-form-item>
                        </el-form>
                    </el-col>
                </el-row>

                <div v-if="confShowType === 'all'">
                    <codemirror
                        :autofocus="true"
                        placeholder="# The SSH configuration file does not exist or is empty (/etc/ssh/sshd_config)"
                        :indent-with-tab="true"
                        :tabSize="4"
                        style="margin-top: 10px; height: calc(100vh - 330px)"
                        :lineWrapping="true"
                        :matchBrackets="true"
                        theme="cobalt"
                        :styleActiveLine="true"
                        :extensions="extensions"
                        v-model="sshConf"
                    />
                    <el-button :disabled="loading" type="primary" @click="onSaveFile" style="margin-top: 5px">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </div>
            </template>
        </LayoutContent>

        <PubKey ref="pubKeyRef" @search="search" />
        <Port ref="portRef" @search="search" />
        <Address ref="addressRef" @search="search" />
        <Root ref="rootsRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import FireRouter from '@/views/host/ssh/index.vue';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import PubKey from '@/views/host/ssh/ssh/pubkey/index.vue';
import Root from '@/views/host/ssh/ssh/root/index.vue';
import Port from '@/views/host/ssh/ssh/port/index.vue';
import Address from '@/views/host/ssh/ssh/address/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { getSSHInfo, operateSSH, updateSSH, updateSSHByfile } from '@/api/modules/host';
import { LoadFile } from '@/api/modules/files';
import { ElMessageBox, FormInstance } from 'element-plus';

const loading = ref(false);
const formRef = ref();
const extensions = [javascript(), oneDark];
const confShowType = ref('base');
const pubKeyRef = ref();
const portRef = ref();
const addressRef = ref();
const rootsRef = ref();

const sshConf = ref();
const form = reactive({
    status: 'enable',
    message: '',
    port: 22,
    listenAddress: '',
    passwordAuthentication: 'yes',
    pubkeyAuthentication: 'yes',
    encryptionMode: '',
    primaryKey: '',
    permitRootLogin: 'yes',
    permitRootLoginItem: 'yes',
    useDNS: 'no',
});

const onSaveFile = async () => {
    ElMessageBox.confirm(i18n.global.t('ssh.sshFileChangeHelper'), i18n.global.t('ssh.sshChange'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await updateSSHByfile(sshConf.value)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onOpenDrawer = () => {
    pubKeyRef.value.acceptParams();
};

const onChangePort = () => {
    portRef.value.acceptParams({ port: form.port });
};
const onChangeRoot = () => {
    rootsRef.value.acceptParams({ permitRootLogin: form.permitRootLogin });
};
const onChangeAddress = () => {
    addressRef.value.acceptParams({ address: form.listenAddress });
};

const onOperate = async (operation: string) => {
    ElMessageBox.confirm(i18n.global.t('ssh.sshOperate', [i18n.global.t('commons.button.' + operation)]), 'SSH', {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await operateSSH(operation)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onSave = async (formEl: FormInstance | undefined, key: string, value: string) => {
    if (!formEl) return;
    let itemKey = key.replace(key[0], key[0].toLowerCase());
    const result = await formEl.validateField(itemKey, callback);
    if (!result) {
        return;
    }

    ElMessageBox.confirm(
        i18n.global.t('ssh.sshChangeHelper', [i18n.global.t('ssh.' + itemKey), changei18n(value)]),
        i18n.global.t('ssh.sshChange'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    )
        .then(async () => {
            loading.value = true;
            await updateSSH(key, value)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        })
        .catch(() => {
            search();
        });
};
function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

const changei18n = (value: string) => {
    switch (value) {
        case 'yes':
            return i18n.global.t('commons.button.enable');
        case 'no':
            return i18n.global.t('commons.button.disable');
        default:
            return value;
    }
};

const loadSSHConf = async () => {
    const res = await LoadFile({ path: '/etc/ssh/sshd_config' });
    sshConf.value = res.data || '';
};

const changeMode = async () => {
    if (confShowType.value === 'all') {
        loadSSHConf();
    } else {
        search();
    }
};

const search = async () => {
    const res = await getSSHInfo();
    form.status = res.data.status;
    form.port = Number(res.data.port);
    form.listenAddress = res.data.listenAddress;
    form.passwordAuthentication = res.data.passwordAuthentication;
    form.pubkeyAuthentication = res.data.pubkeyAuthentication;
    form.permitRootLogin = res.data.permitRootLogin;
    form.permitRootLoginItem = loadPermitLabel(res.data.permitRootLogin);
    form.useDNS = res.data.useDNS;
};

const loadPermitLabel = (value: string) => {
    switch (value) {
        case 'yes':
            return i18n.global.t('ssh.rootHelper1');
        case 'no':
            return i18n.global.t('ssh.rootHelper2');
        case 'without-password':
            return i18n.global.t('ssh.rootHelper3');
        case 'forced-commands-only':
            return i18n.global.t('ssh.rootHelper4');
    }
};

onMounted(() => {
    search();
});
</script>

<style lang="scss" scoped>
.a-card {
    font-size: 17px;
    .el-card {
        --el-card-padding: 12px;
        .buttons {
            margin-left: 100px;
        }
    }
}
.status-content {
    float: left;
    margin-left: 50px;
}
</style>
