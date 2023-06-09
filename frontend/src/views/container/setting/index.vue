<template>
    <div v-loading="loading">
        <div class="a-card" style="margin-top: 20px">
            <el-card>
                <div>
                    <el-tag style="float: left" effect="dark" type="success">Docker</el-tag>
                    <el-tag round class="status-content" v-if="form.status === 'Running'" type="success">
                        {{ $t('commons.status.running') }}
                    </el-tag>
                    <el-tag round class="status-content" v-if="form.status === 'Stopped'" type="info">
                        {{ $t('commons.status.stopped') }}
                    </el-tag>
                    <el-tag class="status-content">{{ $t('app.version') }}: {{ form.version }}</el-tag>

                    <span v-if="form.status === 'Running'" class="buttons">
                        <el-button type="primary" @click="onOperator('stop')" link>
                            {{ $t('container.stop') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" @click="onOperator('restart')" link>
                            {{ $t('container.restart') }}
                        </el-button>
                    </span>

                    <span v-if="form.status === 'Stopped'" class="buttons">
                        <el-button type="primary" @click="onOperator('start')" link>
                            {{ $t('container.start') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" @click="onOperator('restart')" link>
                            {{ $t('container.restart') }}
                        </el-button>
                    </span>
                </div>
            </el-card>
        </div>

        <LayoutContent style="margin-top: 20px" :title="$t('container.setting')" :divider="true">
            <template #main>
                <el-radio-group v-model="confShowType" @change="changeMode">
                    <el-radio-button label="base">{{ $t('database.baseConf') }}</el-radio-button>
                    <el-radio-button label="all">{{ $t('database.allConf') }}</el-radio-button>
                </el-radio-group>
                <el-row style="margin-top: 20px" v-if="confShowType === 'base'">
                    <el-col :span="1"><br /></el-col>
                    <el-col :xs="24" :sm="24" :md="15" :lg="12" :xl="10">
                        <el-form :model="form" label-position="left" :rules="rules" ref="formRef" label-width="120px">
                            <el-form-item :label="$t('container.mirrors')" prop="mirrors">
                                <div style="width: 100%" v-if="form.mirrors">
                                    <el-input
                                        type="textarea"
                                        :autosize="{ minRows: 3, maxRows: 5 }"
                                        disabled
                                        v-model="form.mirrors"
                                        style="width: calc(100% - 80px)"
                                    />
                                    <el-button class="append-button" @click="onChangeMirrors" icon="Setting">
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
                                <span class="input-help">
                                    {{ $t('container.mirrorsHelper2') }}
                                    <el-link
                                        style="font-size: 12px; margin-left: 5px"
                                        icon="Position"
                                        @click="toDoc()"
                                        type="primary"
                                    >
                                        {{ $t('firewall.quickJump') }}
                                    </el-link>
                                </span>
                            </el-form-item>
                            <el-form-item :label="$t('container.registries')" prop="registries">
                                <div style="width: 100%" v-if="form.registries">
                                    <el-input
                                        type="textarea"
                                        :autosize="{ minRows: 3, maxRows: 5 }"
                                        disabled
                                        v-model="form.registries"
                                        style="width: calc(100% - 80px)"
                                    />
                                    <el-button class="append-button" @click="onChangeRegistries" icon="Setting">
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
                                    <el-radio label="cgroupfs">cgroupfs</el-radio>
                                    <el-radio label="systemd">systemd</el-radio>
                                </el-radio-group>
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
                        style="margin-top: 10px; height: calc(100vh - 430px)"
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
            v-model="iptablesVisiable"
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
                            iptablesVisiable = false;
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
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import i18n from '@/lang';
import {
    dockerOperate,
    loadDaemonJson,
    loadDaemonJsonFile,
    updateDaemonJson,
    updateDaemonJsonByfile,
} from '@/api/modules/container';
import { MsgSuccess } from '@/utils/message';
import { checkNumberRange } from '@/global/form-rules';

const unset = ref(i18n.global.t('setting.unSetting'));
const submitInput = ref();

const loading = ref(false);
const showDaemonJsonAlert = ref(false);
const extensions = [javascript(), oneDark];
const confShowType = ref('base');

const logOptionRef = ref();
const confirmDialogRefLog = ref();
const mirrorRef = ref();
const registriesRef = ref();
const confirmDialogRefLive = ref();
const confirmDialogRefCgroup = ref();
const confirmDialogRefIptable = ref();
const logOptionShow = ref();

const form = reactive({
    isSwarm: false,
    status: '',
    version: '',
    mirrors: '',
    registries: '',
    liveRestore: false,
    iptables: true,
    cgroupDriver: '',
    logOptionShow: false,
    logMaxSize: '',
    logMaxFile: 3,
});
const rules = reactive({
    logMaxSize: [checkNumberRange(1, 1024000)],
    logMaxFile: [checkNumberRange(1, 100)],
});

const formRef = ref<FormInstance>();
const dockerConf = ref();
const confirmDialogRefFile = ref();

const iptablesVisiable = ref();

const onSaveFile = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRefFile.value!.acceptParams(params);
};

const onChangeMirrors = () => {
    mirrorRef.value.acceptParams({ mirrors: form.mirrors });
};
const onChangeRegistries = () => {
    registriesRef.value.acceptParams({ registries: form.registries });
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
        iptablesVisiable.value = true;
    }
};
const onSubmitCloseIPtable = () => {
    save('IPtables', 'disable');
    iptablesVisiable.value = false;
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
    save('Dirver', form.cgroupDriver);
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
    window.open('https://1panel.cn/docs/user_manual/containers/setting/', '_blank');
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
body {
    margin: 0;
}

.append-button {
    width: 80px;
    background-color: var(--el-fill-color-light);
    color: var(--el-color-info);
}
</style>
