<template>
    <div v-loading="loading">
        <div class="app-status" style="margin-top: 20px">
            <el-card>
                <div>
                    <el-tag style="float: left" effect="dark" type="success">Fail2ban</el-tag>
                    <el-tag round class="status-content" v-if="form.isActive" type="success">
                        {{ $t('commons.status.running') }}
                    </el-tag>
                    <el-popover
                        v-if="!form.isActive"
                        placement="top-start"
                        trigger="hover"
                        width="450"
                        :content="form.version"
                    >
                        <template #reference>
                            <el-tag round class="status-content" v-if="!form.isActive" type="info">
                                {{ $t('commons.status.stopped') }}
                            </el-tag>
                        </template>
                    </el-popover>
                    <el-tag class="status-content">{{ form.version }}</el-tag>
                    <span class="buttons">
                        <el-button v-if="form.isActive" type="primary" @click="onOperate('stop')" link>
                            {{ $t('commons.button.stop') }}
                        </el-button>
                        <el-button v-if="!form.isActive" type="primary" @click="onOperate('start')" link>
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
                            style="margin-left: 10px"
                            inactive-value="disable"
                            active-value="enable"
                            @change="onOperate(autoStart)"
                            v-model="autoStart"
                        />
                    </span>
                </div>
            </el-card>
        </div>

        <LayoutContent style="margin-top: 20px" :title="$t('menu.config')" :divider="true">
            <template #toolbar>
                <el-row>
                    <el-col :span="16">
                        <el-button type="primary" plain @click="onLoadList('ignore')">
                            {{ $t('toolbox.fail2ban.ignoreIP') }}
                        </el-button>
                        <el-button type="primary" plain @click="onLoadList('banned')">
                            {{ $t('toolbox.fail2ban.bannedIP') }}
                        </el-button>
                    </el-col>
                </el-row>
            </template>
            <template #main>
                <el-radio-group v-model="confShowType" @change="changeMode">
                    <el-radio-button label="base">{{ $t('database.baseConf') }}</el-radio-button>
                    <el-radio-button label="all">{{ $t('database.allConf') }}</el-radio-button>
                </el-radio-group>
                <el-row style="margin-top: 20px" v-if="confShowType === 'base'">
                    <el-col :span="1"><br /></el-col>
                    <el-col :xs="24" :sm="20" :md="20" :lg="10" :xl="10">
                        <el-form :model="form" label-position="left" ref="formRef" label-width="120px">
                            <el-form-item :label="$t('toolbox.fail2ban.maxRetry')" prop="maxRetry">
                                <el-input disabled v-model="form.maxRetry">
                                    <template #append>
                                        <el-button @click="onChangeMaxRetry" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('toolbox.fail2ban.banTime')" prop="banTime">
                                <el-input disabled v-model="form.banTime">
                                    <template #append>
                                        <el-button @click="onChangeBanTime" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('toolbox.fail2ban.banTimeHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('toolbox.fail2ban.findTime')" prop="findTime">
                                <el-input disabled v-model="form.findTime">
                                    <template #append>
                                        <el-button @click="onChangeFindTime" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('toolbox.fail2ban.banAction')" prop="banAction">
                                <el-input disabled v-model="form.banAction">
                                    <template #append>
                                        <el-button @click="onChangeBanAction" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                        </el-form>
                    </el-col>
                </el-row>

                <div v-if="confShowType === 'all'">
                    <codemirror
                        :autofocus="true"
                        placeholder="# The Fail2ban configuration file does not exist or is empty (/etc/ssh/sshd_config)"
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

        <MaxRetry ref="maxRetryRef" @search="search" />
        <BanTime ref="banTimeRef" @search="search" />
        <FindTime ref="findTimeRef" @search="search" />
        <BanAction ref="banActionRef" @search="search" />

        <IPs ref="listRef" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import MaxRetry from '@/views/toolbox/fail2ban/max-retry/index.vue';
import BanTime from '@/views/toolbox/fail2ban/ban-time/index.vue';
import FindTime from '@/views/toolbox/fail2ban/find-time/index.vue';
import BanAction from '@/views/toolbox/fail2ban/ban-action/index.vue';
import IPs from '@/views/toolbox/fail2ban/ips/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { getFail2banConf, operateFail2ban, updateFail2banByFile } from '@/api/modules/toolbox';
import { ElMessageBox } from 'element-plus';
import { getFail2banBase } from '@/api/modules/toolbox';

const loading = ref(false);
const formRef = ref();
const extensions = [javascript(), oneDark];
const confShowType = ref('base');

const maxRetryRef = ref();
const banTimeRef = ref();
const findTimeRef = ref();
const banActionRef = ref();
const listRef = ref();

const autoStart = ref('enable');

const sshConf = ref();
const form = reactive({
    isEnable: false,
    isActive: false,
    version: '-',

    port: 22,
    maxRetry: 5,
    banTime: '',
    findTime: '',
    banAction: '',
    logPath: '',
});

const onLoadList = async (type: string) => {
    listRef.value.acceptParams({ ipType: type });
};

const onSaveFile = async () => {
    ElMessageBox.confirm(i18n.global.t('ssh.sshFileChangeHelper'), i18n.global.t('ssh.sshChange'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await updateFail2banByFile(sshConf.value)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const onChangeMaxRetry = () => {
    maxRetryRef.value.acceptParams({ maxRetry: form.maxRetry });
};
const onChangeBanTime = () => {
    banTimeRef.value.acceptParams({ banTime: form.banTime });
};
const onChangeFindTime = () => {
    findTimeRef.value.acceptParams({ findTime: form.findTime });
};
const onChangeBanAction = () => {
    banActionRef.value.acceptParams({ banAction: form.banAction });
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
            await operateFail2ban(operation)
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
            autoStart.value = operation === 'enable' ? 'disable' : 'enable';
        });
};

const loadSSHConf = async () => {
    const res = await getFail2banConf();
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
    const res = await getFail2banBase();
    form.isEnable = res.data.isEnable;
    form.isActive = res.data.isActive;
    form.version = res.data.version;

    form.port = res.data.port;
    form.maxRetry = res.data.maxRetry;
    form.banTime = res.data.banTime;
    form.findTime = res.data.findTime;
    form.banAction = res.data.banAction;
    form.logPath = res.data.logPath;
};

onMounted(() => {
    search();
});
</script>
