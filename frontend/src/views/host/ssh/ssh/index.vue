<template>
    <div v-loading="loading">
        <FireRouter />

        <div class="app-status card-interval">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4 ml-3">
                        <el-tag class="float-left" effect="dark" type="success">SSH</el-tag>
                        <Status class="mt-0.5" :status="form.isActive ? 'enable' : 'disable'" :msg="form.message" />
                    </div>
                    <div class="mt-0.5">
                        <el-button
                            v-if="form.isActive"
                            type="primary"
                            v-permission
                            v-node-admin
                            @click="onOperate('stop')"
                            link
                        >
                            {{ $t('commons.button.stop') }}
                        </el-button>
                        <el-button
                            v-if="!form.isActive"
                            type="primary"
                            v-permission
                            v-node-admin
                            @click="onOperate('start')"
                            link
                        >
                            {{ $t('commons.button.start') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button v-permission v-node-admin type="primary" @click="onOperate('restart')" link>
                            {{ $t('commons.button.restart') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" link>
                            {{ $t('ssh.autoStart') }}
                        </el-button>
                        <el-switch
                            v-permission
                            v-node-admin
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

        <LayoutContent>
            <template #leftToolBar>
                <el-radio-group v-model="confShowType" @change="changeMode">
                    <el-radio-button value="base">{{ $t('database.baseConf') }}</el-radio-button>
                    <el-radio-button value="all">{{ $t('database.allConf') }}</el-radio-button>
                </el-radio-group>

                <el-button v-permission v-node-admin @click="onOpenDrawer">{{ $t('ssh.pubkey') }}</el-button>
                <el-button v-permission v-node-admin @click="onOpenAuthKeys">{{ $t('ssh.authKeys') }}</el-button>
            </template>
            <template #main>
                <el-row class="mt-10" v-if="confShowType === 'base'">
                    <el-col :xs="24" :sm="20" :md="20" :lg="10" :xl="10">
                        <el-form :model="form" label-position="right" ref="formRef" label-width="120px">
                            <el-form-item :label="$t('ssh.port')" prop="port">
                                <el-input disabled v-model="form.port">
                                    <template #append>
                                        <el-button v-permission v-node-admin @click="onChangePort" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('ssh.portHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('ssh.listenAddress')" prop="listenAddress">
                                <el-input disabled v-model="form.listenAddressItem">
                                    <template #append>
                                        <el-button v-permission v-node-admin @click="onChangeAddress" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('ssh.addressHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('ssh.permitRootLogin')" prop="permitRootLoginItem">
                                <el-input disabled v-model="form.permitRootLoginItem">
                                    <template #append>
                                        <el-button v-permission v-node-admin @click="onChangeRoot" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('ssh.rootSettingHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('ssh.passwordAuthentication')" prop="passwordAuthentication">
                                <el-switch
                                    v-permission
                                    v-node-admin
                                    active-value="yes"
                                    inactive-value="no"
                                    @change="onSave(formRef, 'PasswordAuthentication', form.passwordAuthentication)"
                                    v-model="form.passwordAuthentication"
                                ></el-switch>
                                <span class="input-help">{{ $t('ssh.pwdAuthHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('ssh.pubkeyAuthentication')" prop="pubkeyAuthentication">
                                <el-switch
                                    v-permission
                                    v-node-admin
                                    active-value="yes"
                                    inactive-value="no"
                                    @change="onSave(formRef, 'PubkeyAuthentication', form.pubkeyAuthentication)"
                                    v-model="form.pubkeyAuthentication"
                                ></el-switch>
                                <span class="input-help">{{ $t('ssh.keyAuthHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('ssh.useDNS')" prop="useDNS">
                                <el-switch
                                    v-permission
                                    v-node-admin
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
                    <el-alert type="info" :closable="false" :title="$t('ssh.confFileOrderHelper')" />
                    <el-select v-model="allConfMode" class="mt-2 mini-form-item" @change="changeAllConfMode">
                        <el-option
                            v-for="item in sshConfOptions"
                            :key="item.value"
                            :value="item.value"
                            :label="item.label"
                        ></el-option>
                    </el-select>
                    <CodemirrorPro
                        :heightDiff="459"
                        :minHeight="350"
                        class="mt-5"
                        v-model="allConfContent"
                        mode="nginx"
                        placeholder="# The SSH configuration file does not exist or is empty (/etc/ssh/sshd_config)"
                    ></CodemirrorPro>
                    <el-button
                        v-permission
                        v-node-admin
                        :disabled="loading"
                        type="primary"
                        @click="onSaveFile"
                        class="mt-2.5"
                    >
                        {{ $t('commons.button.save') }}
                    </el-button>
                </div>
            </template>
        </LayoutContent>

        <AuthKeys ref="authKeyRef" />
        <Cert ref="pubKeyRef" @search="search" />
        <Port ref="portRef" @search="search" />
        <Address ref="addressRef" @search="search" />
        <Root ref="rootsRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';
import FireRouter from '@/views/host/ssh/index.vue';
import AuthKeys from '@/views/host/ssh/ssh/auth-keys/index.vue';
import Cert from '@/views/host/ssh/ssh/certification/index.vue';
import Root from '@/views/host/ssh/ssh/root/index.vue';
import Port from '@/views/host/ssh/ssh/port/index.vue';
import Address from '@/views/host/ssh/ssh/address/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { loadSSHFile, getSSHInfo, operateSSH, updateSSH, updateSSHByFile } from '@/api/modules/host';
import { ElMessageBox, FormInstance } from 'element-plus';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';

const loading = ref(false);
const formRef = ref();
const confShowType = ref('base');
const pubKeyRef = ref();
const portRef = ref();
const addressRef = ref();
const rootsRef = ref();
const authKeyRef = ref();

const autoStart = ref('enable');
const allConfMode = ref('');

const sshConfPath = ref('');
const sshConfOptions = ref<Array<{ value: string; label: string }>>([]);
const allConfContent = computed({
    get: () => sshConfPath.value,
    set: (value: string) => {
        sshConfPath.value = value;
    },
});
const form = reactive({
    isActive: false,
    message: '',
    port: '22',
    listenAddress: '',
    listenAddressItem: '',
    passwordAuthentication: 'yes',
    pubkeyAuthentication: 'yes',
    encryptionMode: '',
    primaryKey: '',
    permitRootLogin: 'yes',
    permitRootLoginItem: 'yes',
    useDNS: 'no',
    currentUser: 'root',
});

const onSaveFile = async () => {
    ElMessageBox.confirm(i18n.global.t('ssh.sshFileChangeHelper'), i18n.global.t('ssh.sshChange'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        if (!allConfMode.value) {
            loading.value = false;
            return;
        }
        await updateSSHByFile('sshdConfPath', sshConfPath.value, allConfMode.value)
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
    pubKeyRef.value.acceptParams(form.currentUser);
};
const onOpenAuthKeys = () => {
    authKeyRef.value.acceptParams();
};

const onChangePort = () => {
    portRef.value.acceptParams({ port: form.port });
};
const onChangeRoot = () => {
    rootsRef.value.acceptParams({ permitRootLogin: form.permitRootLogin });
};
const onChangeAddress = () => {
    addressRef.value.acceptParams({ address: form.listenAddress, port: loadPrimarySSHPort(form.port) });
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
    const optionsRes = await loadSSHFile('sshdConfOptions');
    let options: Array<{ path: string; priority: number }> = [];
    try {
        options = JSON.parse(optionsRes.data || '[]');
    } catch {
        options = [];
    }
    const fileOptions = options.map((item) => ({
        value: item.path,
        label: i18n.global.t('ssh.confFileOrderLabel', [item.path, item.priority]),
    }));
    sshConfOptions.value = fileOptions;
    if (!fileOptions.find((item) => item.value === allConfMode.value)) {
        allConfMode.value = fileOptions.length > 0 ? fileOptions[0].value : '';
    }
    if (allConfMode.value) {
        const fileRes = await loadSSHFile(`sshdConfPath:${allConfMode.value}`);
        sshConfPath.value = fileRes.data || '';
    } else {
        sshConfPath.value = '';
    }
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
    form.isActive = res.data.isActive;
    form.port = res.data.port || '22';
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
    form.currentUser = res.data.currentUser;
};

const changeAllConfMode = async () => {
    if (allConfMode.value) {
        const res = await loadSSHFile(`sshdConfPath:${allConfMode.value}`);
        sshConfPath.value = res.data || '';
    } else {
        sshConfPath.value = '';
    }
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

const loadPrimarySSHPort = (port: string) => {
    const firstPort = port.split(',')[0] || '22';
    const portNumber = Number(firstPort);
    return Number.isNaN(portNumber) ? 22 : portNumber;
};

onMounted(() => {
    search();
});
</script>
