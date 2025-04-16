<template>
    <div v-loading="loading">
        <docker-status
            v-model:isActive="isActive"
            v-model:isExist="isExist"
            :is-hide="true"
            v-model:loading="loading"
            @search="search"
        />

        <div v-if="isExist" class="app-status card-interval">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4 ml-3">
                        <el-tag class="float-left" effect="dark" type="success">Docker</el-tag>
                        <Status class="mt-0.5" :status="isActive ? 'enable' : 'disable'" />
                        <el-tag>{{ $t('app.version') }}: {{ form.version }}</el-tag>
                    </div>
                    <div class="mt-0.5">
                        <el-button v-if="isActive" type="primary" @click="onOperator('stop')" link>
                            {{ $t('commons.operate.stop') }}
                        </el-button>
                        <el-button v-if="!isActive" type="primary" @click="onOperator('start')" link>
                            {{ $t('commons.operate.start') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" @click="onOperator('restart')" link>
                            {{ $t('commons.button.restart') }}
                        </el-button>
                    </div>
                </div>
            </el-card>
        </div>

        <LayoutContent v-if="isExist" class="card-interval" :title="$t('container.setting')">
            <template #main>
                <el-radio-group v-model="confShowType" @change="changeMode">
                    <el-radio-button value="base">{{ $t('database.baseConf') }}</el-radio-button>
                    <el-radio-button value="all">{{ $t('database.allConf') }}</el-radio-button>
                </el-radio-group>
                <el-row class="p-mt-20" v-if="confShowType === 'base'">
                    <el-col :span="1"><br /></el-col>
                    <el-col :xs="24" :sm="24" :md="15" :lg="12" :xl="10">
                        <el-form
                            :model="form"
                            :label-position="mobile ? 'top' : 'left'"
                            :rules="rules"
                            ref="formRef"
                            label-width="auto"
                        >
                            <el-form-item :label="$t('container.mirrors')" prop="mirrors">
                                <div class="w-full" v-if="form.mirrors">
                                    <el-input
                                        type="textarea"
                                        :rows="5"
                                        disabled
                                        v-model="form.mirrors"
                                        style="width: calc(100% - 80px)"
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
                                <span class="input-help">{{ $t('container.mirrorsHelper') }}</span>
                                <span class="input-help flex flx-align-center">
                                    {{ $t('container.mirrorsHelper2') }}
                                    <el-link class="p-ml-5 text-xs" icon="Position" @click="toDoc()" type="primary">
                                        {{ $t('firewall.quickJump') }}
                                    </el-link>
                                </span>
                            </el-form-item>
                            <el-form-item :label="$t('container.registries')" prop="registries">
                                <div class="w-full" v-if="form.registries">
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

                            <el-form-item label="ipv6" prop="ipv6">
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
                                    <el-tag class="p-ml-5">{{ $t('container.maxFile') }}: {{ form.logMaxFile }}</el-tag>
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
                            <el-form-item label="live-restore" prop="liveRestore">
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
                            <el-form-item label="cgroup-driver" prop="cgroupDriver">
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
                    <CodemirrorPro
                        class="mt-5"
                        :heightDiff="loadHeight()"
                        v-model="dockerConf"
                        mode="json"
                        placeholder="# The Docker configuration file does not exist or is empty (/etc/docker/daemon.json)"
                    ></CodemirrorPro>
                    <el-button :disabled="loading" type="primary" @click="onSaveFile" class="mt-2.5">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </div>
            </template>
        </LayoutContent>

        <DialogPro v-model="open" :title="$t('container.iptablesDisable')" size="small">
            <div class="mt-2.5">
                <span class="text-rose-500">{{ $t('container.iptablesHelper2') }}</span>
                <div class="mt-2.5">
                    <span class="text-xs">{{ $t('database.restartNowHelper') }}</span>
                </div>
                <div class="mt-2.5">
                    <span class="text-xs">{{ $t('commons.msg.operateConfirm') }}</span>
                    <span class="text-xs text-rose-500 font-medium">'{{ $t('database.restartNow') }}'</span>
                </div>
                <el-input class="mt-2.5" v-model="submitInput"></el-input>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button
                        @click="
                            open = false;
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
        </DialogPro>

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
import { onMounted, reactive, ref, computed } from 'vue';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import Mirror from '@/views/container/setting/mirror/index.vue';
import Registry from '@/views/container/setting/registry/index.vue';
import LogOption from '@/views/container/setting/log/index.vue';
import Ipv6Option from '@/views/container/setting/ipv6/index.vue';
import SockPath from '@/views/container/setting/sock-path/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';
import i18n from '@/lang';
import {
    dockerOperate,
    loadDaemonJson,
    loadDaemonJsonFile,
    updateDaemonJson,
    updateDaemonJsonByfile,
} from '@/api/modules/container';
import { getAgentSettingInfo } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';
import { checkNumberRange } from '@/global/form-rules';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});
const unset = ref(i18n.global.t('setting.unSetting'));
const submitInput = ref();

const isActive = ref(false);
const isExist = ref(false);

const loading = ref(false);
const showDaemonJsonAlert = ref(false);
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
    isActive: false,
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

const open = ref();

const onSaveFile = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRefFile.value!.acceptParams(params);
};

const loadHeight = () => {
    return globalStore.openMenuTabs ? 450 : 430;
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
        open.value = true;
    }
};
const onSubmitCloseIPtable = () => {
    save('IPtables', 'disable');
    open.value = false;
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

    const settingRes = await getAgentSettingInfo();
    form.dockerSockPath = settingRes.data.dockerSockPath || 'unix:///var/run/docker.sock';
};

onMounted(() => {
    search();
});
</script>
