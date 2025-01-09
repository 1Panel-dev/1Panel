<template>
    <div v-loading="loading">
        <FireRouter />

        <div class="app-status" style="margin-top: 20px">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4">
                        <el-tag style="float: left" effect="dark" type="success">SSH</el-tag>
                        <el-tag round v-if="form.status === 'Enable'" type="success">
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
                                <el-tag round v-if="form.status === 'Disable'" type="info">
                                    {{ $t('commons.status.stopped') }}
                                </el-tag>
                            </template>
                        </el-popover>
                    </div>
                    <div class="mt-0.5">
                        <el-button v-if="form.status === 'Enable'" type="primary" @click="onOperate('stop')" link>
                            {{ $t('commons.button.stop') }}
                        </el-button>
                        <el-button v-if="form.status === 'Disable'" type="primary" @click="onOperate('start')" link>
                            {{ $t('commons.button.start') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" @click="onOperate('restart')" link>
                            {{ $t('container.restart') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" link>
                            {{ $t('ssh.autoStart') }}
                        </el-button>
                        <el-switch
                            size="small"
                            class="ml-2"
                            inactive-value="disable"
                            active-value="enable"
                            @change="onOperate(autoStart)"
                            v-model="autoStart"
                        />
                    </div>
                </div>
            </el-card>
        </div>

        <LayoutContent style="margin-top: 20px" :title="$t('ssh.setting')" :divider="true">
            <template #main>
                <el-radio-group v-model="confShowType" @change="changeMode">
                    <el-radio-button value="base">{{ $t('database.baseConf') }}</el-radio-button>
                    <el-radio-button value="all">{{ $t('database.allConf') }}</el-radio-button>
                </el-radio-group>
                <el-row style="margin-top: 20px" v-if="confShowType === 'base'">
                    <el-col :span="1"><br /></el-col>
                    <el-col :xs="24" :sm="20" :md="20" :lg="10" :xl="10">
                        <el-form :model="form" label-position="left" ref="formRef" label-width="120px">
                            <el-form-item :label="$t('ssh.port')" prop="port">
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
                                <el-input disabled v-model="form.listenAddressItem">
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
                        style="margin-top: 10px; height: calc(100vh - 405px)"
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
import { getSSHConf, getSSHInfo, operateSSH, updateSSH, updateSSHByfile } from '@/api/modules/host';
import { ElMessageBox, FormInstance } from 'element-plus';

const loading = ref(false);
const formRef = ref();
const extensions = [javascript(), oneDark];
const confShowType = ref('base');
const pubKeyRef = ref();
const portRef = ref();
const addressRef = ref();
const rootsRef = ref();

const autoStart = ref('enable');

const sshConf = ref();
const form = reactive({
    status: 'enable',
    message: '',
    port: 22,
    listenAddress: '',
    listenAddressItem: '',
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
    addressRef.value.acceptParams({ address: form.listenAddress, port: form.port });
};

const onOperate = async (operation: string) => {
    let msg = operation === 'enable' || operation === 'disable' ? 'ssh.' : 'commons.button.';
    ElMessageBox.confirm(i18n.global.t('ssh.sshOperate', [i18n.global.t(msg + operation)]), 'SSH', {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    })
        .then(async () => {
            loading.value = true;
            await operateSSH(operation)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    search();
                })
                .catch(() => {
                    autoStart.value = operation === 'enable' ? 'disable' : 'enable';
                    loading.value = false;
                });
        })
        .catch(() => {
            search();
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
        i18n.global.t('ssh.sshChangeHelper', [i18n.global.t('ssh.' + itemKey), changeI18n(value)]),
        i18n.global.t('ssh.sshChange'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    )
        .then(async () => {
            let params = {
                key: key,
                oldValue: '',
                newValue: value,
            };
            loading.value = true;
            await updateSSH(params)
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

const changeI18n = (value: string) => {
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
    const res = await getSSHConf();
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
    autoStart.value = res.data.autoStart ? 'enable' : 'disable';
    form.listenAddress = res.data.listenAddress;
    form.listenAddressItem =
        form.listenAddress === '' || form.listenAddress === '0.0.0.0,::'
            ? i18n.global.t('ssh.allV4V6', [form.port])
            : form.listenAddress;
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
