<template>
    <div v-loading="loading">
        <RouterButton
            :buttons="[
                {
                    label: i18n.global.t('menu.ssh'),
                    path: '/hosts/ssh',
                },
            ]"
        />
        <LayoutContent style="margin-top: 20px" :title="$t('menu.ssh')" :divider="true">
            <template #main>
                <el-radio-group v-model="confShowType" @change="changeMode">
                    <el-radio-button label="base">{{ $t('database.baseConf') }}</el-radio-button>
                    <el-radio-button label="all">{{ $t('database.allConf') }}</el-radio-button>
                </el-radio-group>
                <el-row style="margin-top: 20px" v-if="confShowType === 'base'">
                    <el-col :span="1"><br /></el-col>
                    <el-col :span="10">
                        <el-form :model="form" label-position="left" ref="formRef" label-width="120px">
                            <el-form-item :label="$t('ssh.port')" prop="port" :rules="Rules.port">
                                <el-input v-model.number="form.port">
                                    <template #append>
                                        <el-button icon="Collection" @click="onSave(formRef, 'Port', form.port + '')">
                                            {{ $t('commons.button.save') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('ssh.portHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('ssh.listenAddress')" prop="listenAddress">
                                <el-input v-model="form.listenAddress">
                                    <template #append>
                                        <el-button
                                            icon="Collection"
                                            @click="onSave(formRef, 'ListenAddress', form.listenAddress)"
                                        >
                                            {{ $t('commons.button.save') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('ssh.addressHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('ssh.permitRootLogin')" prop="permitRootLogin">
                                <el-select
                                    v-model="form.permitRootLogin"
                                    @change="onSave(formRef, 'PermitRootLogin', form.permitRootLogin)"
                                    style="width: 100%"
                                >
                                    <el-option :label="$t('ssh.rootHelper1')" value="yes" />
                                    <el-option :label="$t('ssh.rootHelper2')" value="no" />
                                    <el-option :label="$t('ssh.rootHelper3')" value="without-password" />
                                    <el-option :label="$t('ssh.rootHelper4')" value="forced-commands-only" />
                                </el-select>
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

        <PubKey ref="pubKeyRef" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import LayoutContent from '@/layout/layout-content.vue';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import PubKey from '@/views/host/ssh/pubkey/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { getSSHInfo, updateSSH } from '@/api/modules/host';
import { LoadFile, SaveFileContent } from '@/api/modules/files';
import { Rules } from '@/global/form-rules';
import { ElMessageBox, FormInstance } from 'element-plus';

const loading = ref(false);
const formRef = ref();
const extensions = [javascript(), oneDark];
const confShowType = ref('base');
const pubKeyRef = ref();

const sshConf = ref();
const form = reactive({
    port: 22,
    listenAddress: '',
    passwordAuthentication: 'yes',
    pubkeyAuthentication: 'yes',
    encryptionMode: '',
    primaryKey: '',
    permitRootLogin: 'yes',
    useDNS: 'no',
});

const onSaveFile = async () => {
    loading.value = true;
    await SaveFileContent({ path: '/etc/ssh/sshd_config', content: sshConf.value })
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const onOpenDrawer = () => {
    pubKeyRef.value.acceptParams();
};

const onSave = async (formEl: FormInstance | undefined, key: string, value: string) => {
    if (!formEl) return;
    let itemKey = key.replace(key[0], key[0].toLowerCase());
    const result = await formEl.validateField(itemKey, callback);
    if (!result) {
        return;
    }

    ElMessageBox.confirm(
        i18n.global.t('ssh.sshChangeHelper', [i18n.global.t('ssh.' + itemKey), value]),
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
    form.port = Number(res.data.port);
    form.listenAddress = res.data.listenAddress;
    form.passwordAuthentication = res.data.passwordAuthentication;
    form.pubkeyAuthentication = res.data.pubkeyAuthentication;
    form.permitRootLogin = res.data.permitRootLogin;
    form.useDNS = res.data.useDNS;
};

onMounted(() => {
    search();
});
</script>
