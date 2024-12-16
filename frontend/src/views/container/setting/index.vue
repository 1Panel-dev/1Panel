<template>
    <div v-loading="loading">
        <div class="app-status" style="margin-top: 20px">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4">
                        <el-tag style="float: left" effect="dark" type="success">Docker</el-tag>
                        <el-tag round v-if="form.status === 'Running'" type="success">
                            {{ $t('commons.status.running') }}
                        </el-tag>
                        <el-tag round v-if="form.status === 'Stopped'" type="info">
                            {{ $t('commons.status.stopped') }}
                        </el-tag>
                        <el-tag>{{ $t('app.version') }}: {{ form.version }}</el-tag>
                    </div>
                    <div class="mt-0.5">
                        <el-button type="primary" v-if="form.status === 'Running'" @click="onOperator('stop')" link>
                            {{ $t('container.stop') }}
                        </el-button>
                        <el-button type="primary" v-if="form.status === 'Stopped'" @click="onOperator('start')" link>
                            {{ $t('container.start') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" @click="onOperator('restart')" link>
                            {{ $t('container.restart') }}
                        </el-button>
                    </div>
                </div>
            </el-card>
        </div>

        <LayoutContent style="margin-top: 20px" :title="$t('container.setting', 2)" :divider="true">
            <template #main>
                <el-radio-group v-model="confShowType" @change="changeMode">
                    <el-radio-button value="base">{{ $t('database.baseConf') }}</el-radio-button>
                    <el-radio-button value="all">{{ $t('database.allConf') }}</el-radio-button>
                </el-radio-group>
                <el-row style="margin-top: 20px" v-if="confShowType === 'base'">
                    <el-col :span="1"><br /></el-col>
                    <el-col :xs="24" :sm="24" :md="15" :lg="12" :xl="10">
                        <el-form :model="form" label-position="left" :rules="rules" ref="formRef" label-width="auto">
                            <el-form-item :label="$t('container.mirrors')" prop="mirrors">
                                <div
                                    class="flex w-full justify-start flex-col sm:flex-row sm:items-end"
                                    v-if="form.mirrors"
                                >
                                    <el-input
                                        type="textarea"
                                        :rows="5"
                                        disabled
                                        v-model="form.mirrors"
                                        class="sm:calc(100% - 80px)"
                                    />
                                    <el-button @click="onChangeMirrors" icon="Setting" class="custom-input-textarea">
                                        {{ $t('commons.button.set') }}
                                    </el-button>
                                </div>
                                <el-input disabled v-if="!form.mirrors" v-model="unset">
                                    <template #append>
                                        <el-button @click="onChangeMirrors" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <div class="flex">
                                    <span>
                                        {{ $t('container.mirrorsHelper') }} {{ $t('container.mirrorsHelper2') }}
                                        <el-link
                                            style="font-size: 12px; margin-left: 5px"
                                            icon="Position"
                                            @click="toDoc()"
                                            type="primary"
                                        >
                                            {{ $t('firewall.quickJump') }}
                                        </el-link>
                                    </span>
                                </div>
                            </el-form-item>
                            <el-form-item :label="$t('container.registries')" prop="registries">
                                <div style="width: 100%" v-if="form.registries">
                                    <el-input
                                        type="textarea"
                                        :rows="5"
                                        disabled
                                        v-model="form.registries"
                                        style="width: calc(100% - 80px)"
                                    />
                                    <el-button @click="onChangeRegistries" icon="Setting">
                                        {{ $t('commons.button.set') }}
                                    </el-button>
                                </div>
                                <el-input disabled v-if="!form.registries" v-model="unset">
                                    <template #append>
                                        <el-button @click="onChangeRegistries" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>

                            <el-form-item label="IPv6" prop="ipv6">
                                <el-switch v-model="form.ipv6" @change="handleIPv6"></el-switch>
                                <span class="input-help"></span>
                                <div v-if="ipv6OptionShow">
                                    <el-tag>{{ $t('container.subnet') }}: {{ form.fixedCidrV6 }}</el-tag>
                                    <div>
                                        <el-button @click="handleIPv6" type="primary" link>
                                            {{ $t('commons.button.view') }}
                                        </el-button>
                                    </div>
                                </div>
                            </el-form-item>

                            <el-form-item :label="$t('container.cutLog')" prop="hasLogOption">
                                <el-switch v-model="form.logOptionShow" @change="handleLogOption"></el-switch>
                                <span class="input-help"></span>
                                <div v-if="logOptionShow">
                                    <el-tag>{{ $t('container.maxSize') }}: {{ form.logMaxSize }}</el-tag>
                                    <el-tag style="margin-left: 5px">
                                        {{ $t('container.maxFile') }}: {{ form.logMaxFile }}
                                    </el-tag>
                                    <div>
                                        <el-button @click="handleLogOption" type="primary" link>
                                            {{ $t('commons.button.view') }}
                                        </el-button>
                                    </div>
                                </div>
                            </el-form-item>

                            <el-form-item label="iptables" prop="iptables">
                                <el-switch v-model="form.iptables" @change="handleIptables"></el-switch>
                                <span class="input-help">{{ $t('container.iptablesHelper1') }}</span>
                            </el-form-item>
                            <el-form-item label="Live restore" prop="liveRestore">
                                <el-switch
                                    :disabled="form.isSwarm"
                                    v-model="form.liveRestore"
                                    @change="handleLive"
                                ></el-switch>
                                <span class="input-help">{{ $t('container.liveHelper') }}</span>
                                <span v-if="form.isSwarm" class="input-help">
                                    {{ $t('container.liveWithSwarmHelper') }}
                                </span>
                            </el-form-item>
                            <el-form-item label="cgroup driver" prop="cgroupDriver">
                                <el-radio-group v-model="form.cgroupDriver" @change="handleCgroup">
                                    <el-radio value="cgroupfs">cgroupfs</el-radio>
                                    <el-radio value="systemd">systemd</el-radio>
                                </el-radio-group>
                            </el-form-item>
                            <el-form-item :label="$t('container.sockPath')" prop="dockerSockPath">
                                <el-input disabled v-model="form.dockerSockPath">
                                    <template #append>
                                        <el-button @click="onChangeSockPath" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('container.sockPathHelper') }}</span>
                            </el-form-item>
                        </el-form>
                    </el-col>
                </el-row>

                <div v-if="confShowType === 'all'">
                    <codemirror
                        :autofocus="true"
                        placeholder="# The Docker configuration file does not exist or is empty (/etc/docker/daemon.json)"
                        :indent-with-tab="true"
                        :tabSize="4"
                        :style="{ height: `calc(100vh - ${loadHeight()})`, 'margin-top': '10px' }"
                        :lineWrapping="true"
                        :matchBrackets="true"
                        theme="cobalt"
                        :styleActiveLine="true"
                        :extensions="extensions"
                        v-model="dockerConf"
                    />
                    <el-button :disabled="loading" type="primary" @click="onSaveFile" style="margin-top: 5px">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </div>
            </template>
        </LayoutContent>

        <el-dialog
            v-model="iptablesVisible"
            :title="$t('container.iptablesDisable')"
            width="30%"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            :show-close="false"
        >
            <div style="margin-top: 10px">
                <span style="color: red">{{ $t('container.iptablesHelper2') }}</span>
                <div style="margin-top: 10px">
                    <span style="font-size: 12px">{{ $t('database.restartNowHelper') }}</span>
                </div>
                <div style="margin-top: 10px">
                    <span style="font-size: 12px">{{ $t('commons.msg.operateConfirm') }}</span>
                    <span style="font-size: 12px; color: red; font-weight: 500">'{{ $t('database.restartNow') }}'</span>
                </div>
                <el-input style="margin-top: 10px" v-model="submitInput"></el-input>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button
                        @click="
                            iptablesVisible = false;
                            search();
                        "
                    >
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button
                        :disabled="submitInput !== $t('database.restartNow')"
                        type="primary"
                        @click="onSubmitCloseIPtable"
                    >
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>

        <Mirror ref="mirrorRef" @search="search" />
        <Registry ref="registriesRef" @search="search" />
        <LogOption ref="logOptionRef" @search="search" />
        <Ipv6Option ref="ipv6OptionRef" @search="search" />
        <SockPath ref="sockPathRef" @search="search" />
        <ConfirmDialog ref="confirmDialogRefIpv6" @confirm="onSaveIPv6" @cancel="search" />
        <ConfirmDialog ref="confirmDialogRefIptable" @confirm="onSubmitOpenIPtable" @cancel="search" />
        <ConfirmDialog ref="confirmDialogRefLog" @confirm="onSubmitSaveLog" @cancel="search" />
        <ConfirmDialog ref="confirmDialogRefLive" @confirm="onSubmitSaveLive" @cancel="search" />
        <ConfirmDialog ref="confirmDialogRefCgroup" @confirm="onSubmitSaveCgroup" @cancel="search" />

        <ConfirmDialog ref="confirmDialogRefFile" @confirm="onSubmitSaveFile" @cancel="search" />
    </div>
</template>

<script lang="ts" setup>
import { ElMessageBox, FormInstance } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import Mirror from '@/views/container/setting/mirror/index.vue';
import Registry from '@/views/container/setting/registry/index.vue';
import LogOption from '@/views/container/setting/log/index.vue';
import Ipv6Option from '@/views/container/setting/ipv6/index.vue';
import SockPath from '@/views/container/setting/sock-path/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import i18n from '@/lang';
import {
    dockerOperate,
    loadDaemonJson,
    loadDaemonJsonFile,
    updateDaemonJson,
    updateDaemonJsonByfile,
} from '@/api/modules/container';
import { getSettingInfo } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';
import { checkNumberRange } from '@/global/form-rules';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const unset = ref(i18n.global.t('setting.unSetting'));
const submitInput = ref();

const loading = ref(false);
const showDaemonJsonAlert = ref(false);
const extensions = [javascript(), oneDark];
const confShowType = ref('base');

const logOptionRef = ref();
const ipv6OptionRef = ref();
const confirmDialogRefLog = ref();
const mirrorRef = ref();
const registriesRef = ref();
const confirmDialogRefLive = ref();
const confirmDialogRefCgroup = ref();
const confirmDialogRefIptable = ref();
const confirmDialogRefIpv6 = ref();
const logOptionShow = ref();
const ipv6OptionShow = ref();
const sockPathRef = ref();

const form = reactive({
    isSwarm: false,
    status: '',
    version: '',
    mirrors: '',
    registries: '',
    liveRestore: false,
    iptables: true,
    cgroupDriver: '',

    ipv6: false,
    fixedCidrV6: '',
    ip6Tables: false,
    experimental: false,

    logOptionShow: false,
    logMaxSize: '',
    logMaxFile: 3,

    dockerSockPath: '',
});
const rules = reactive({
    logMaxSize: [checkNumberRange(1, 1024000)],
    logMaxFile: [checkNumberRange(1, 100)],
});
const formRef = ref<FormInstance>();
const dockerConf = ref();
const confirmDialogRefFile = ref();

const iptablesVisible = ref();

const onSaveFile = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRefFile.value!.acceptParams(params);
};

const loadHeight = () => {
    return globalStore.openMenuTabs ? '450px' : '430px';
};

const onChangeMirrors = () => {
    mirrorRef.value.acceptParams({ mirrors: form.mirrors });
};
const onChangeRegistries = () => {
    registriesRef.value.acceptParams({ registries: form.registries });
};

const onChangeSockPath = () => {
    sockPathRef.value.acceptParams({ dockerSockPath: form.dockerSockPath });
};

const handleIPv6 = async () => {
    if (form.ipv6) {
        ipv6OptionRef.value.acceptParams({
            fixedCidrV6: form.fixedCidrV6,
            ip6Tables: form.ip6Tables,
            experimental: form.experimental,
        });
        return;
    }
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRefIpv6.value!.acceptParams(params);
};
const onSaveIPv6 = () => {
    save('Ipv6', 'disable');
};

const handleLogOption = async () => {
    if (form.logOptionShow) {
        logOptionRef.value.acceptParams({ logMaxSize: form.logMaxSize, logMaxFile: form.logMaxFile });
        return;
    }
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRefLog.value!.acceptParams(params);
};
const onSubmitSaveLog = async () => {
    save('LogOption', 'disable');
};

const handleIptables = () => {
    if (form.iptables) {
        let params = {
            header: i18n.global.t('database.confChange'),
            operationInfo: i18n.global.t('database.restartNowHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialogRefIptable.value!.acceptParams(params);
        return;
    } else {
        iptablesVisible.value = true;
    }
};
const onSubmitCloseIPtable = () => {
    save('IPtables', 'disable');
    iptablesVisible.value = false;
};
const onSubmitOpenIPtable = () => {
    save('IPtables', 'enable');
};

const handleLive = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRefLive.value!.acceptParams(params);
};
const onSubmitSaveLive = () => {
    save('LiveRestore', form.liveRestore ? 'enable' : 'disable');
};
const handleCgroup = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRefCgroup.value!.acceptParams(params);
};
const onSubmitSaveCgroup = () => {
    save('Driver', form.cgroupDriver);
};

const save = async (key: string, value: string) => {
    loading.value = true;
    await updateDaemonJson(key, value)
        .then(() => {
            loading.value = false;
            search();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            search();
            loading.value = false;
        });
};

const toDoc = () => {
    window.open(globalStore.docsUrl + '/user_manual/containers/setting/', '_blank', 'noopener,noreferrer');
};

const onOperator = async (operation: string) => {
    ElMessageBox.confirm(
        i18n.global.t('container.operatorStatusHelper', [i18n.global.t('commons.button.' + operation)]),
        i18n.global.t('commons.table.operate'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(async () => {
        loading.value = true;
        await dockerOperate(operation)
            .then(() => {
                loading.value = false;
                search();
                changeMode();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onSubmitSaveFile = async () => {
    let param = { file: dockerConf.value };
    loading.value = true;
    await updateDaemonJsonByfile(param)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
    return;
};

const loadDockerConf = async () => {
    const res = await loadDaemonJsonFile();
    if (res.data === 'daemon.json is not find in path') {
        showDaemonJsonAlert.value = true;
    } else {
        dockerConf.value = res.data;
    }
};

const changeMode = async () => {
    if (confShowType.value === 'all') {
        loadDockerConf();
    } else {
        showDaemonJsonAlert.value = false;
        search();
    }
};

const search = async () => {
    const res = await loadDaemonJson();
    form.isSwarm = res.data.isSwarm;
    form.status = res.data.status;
    form.version = res.data.version;
    form.cgroupDriver = res.data.cgroupDriver || 'cgroupfs';
    form.liveRestore = res.data.liveRestore;
    form.iptables = res.data.iptables;
    form.mirrors = res.data.registryMirrors ? res.data.registryMirrors.join('\n') : '';
    form.registries = res.data.insecureRegistries ? res.data.insecureRegistries.join('\n') : '';
    if (res.data.logMaxFile || res.data.logMaxSize) {
        form.logOptionShow = true;
        logOptionShow.value = true;
        form.logMaxFile = Number(res.data.logMaxFile);
        form.logMaxSize = res.data.logMaxSize;
    } else {
        form.logOptionShow = false;
        logOptionShow.value = false;
    }
    form.ipv6 = res.data.ipv6;
    ipv6OptionShow.value = form.ipv6;
    form.fixedCidrV6 = res.data.fixedCidrV6;
    form.ip6Tables = res.data.ip6Tables;
    form.experimental = res.data.experimental;

    const settingRes = await getSettingInfo();
    form.dockerSockPath = settingRes.data.dockerSockPath || 'unix:///var/run/docker-x.sock';
};

onMounted(() => {
    search();
});
</script>
