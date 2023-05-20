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
                    <el-col :span="10">
                        <el-form :model="form" label-position="left" :rules="rules" ref="formRef" label-width="150px">
                            <el-form-item :label="$t('container.mirrors')" prop="mirrors">
                                <el-input
                                    type="textarea"
                                    :placeholder="$t('container.mirrorHelper')"
                                    :autosize="{ minRows: 3, maxRows: 10 }"
                                    v-model="form.mirrors"
                                />
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
                                <el-input
                                    type="textarea"
                                    :placeholder="$t('container.registrieHelper')"
                                    :autosize="{ minRows: 3, maxRows: 10 }"
                                    v-model="form.registries"
                                />
                            </el-form-item>

                            <el-form-item :label="$t('container.cutLog')" prop="hasLogOption">
                                <el-switch v-model="form.logOptionShow"></el-switch>
                            </el-form-item>
                            <div v-if="form.logOptionShow">
                                <el-form-item prop="logMaxSize">
                                    <el-input v-model.number="form.logMaxSize">
                                        <template #prepend>{{ $t('container.maxSize') }}</template>
                                        <template #append>
                                            <el-select v-model="form.sizeUnit" style="width: 70px">
                                                <el-option label="b" value="b"></el-option>
                                                <el-option label="k" value="k"></el-option>
                                                <el-option label="m" value="m"></el-option>
                                                <el-option label="g" value="g"></el-option>
                                                <el-option label="t" value="t"></el-option>
                                            </el-select>
                                        </template>
                                    </el-input>
                                </el-form-item>
                                <el-form-item prop="logMaxFile">
                                    <el-input v-model.number="form.logMaxFile">
                                        <template #prepend>{{ $t('container.maxFile') }}</template>
                                    </el-input>
                                </el-form-item>
                            </div>

                            <el-form-item label="iptables" prop="iptables">
                                <el-switch v-model="form.iptables" @change="onChangeIptables"></el-switch>
                            </el-form-item>
                            <el-form-item label="live-restore" prop="liveRestore">
                                <el-switch :disabled="form.isSwarm" v-model="form.liveRestore"></el-switch>
                                <span class="input-help">{{ $t('container.liveHelper') }}</span>
                                <span v-if="form.isSwarm" class="input-help">
                                    {{ $t('container.liveWithSwarmHelper') }}
                                </span>
                            </el-form-item>
                            <el-form-item label="cgroup-driver" prop="cgroupDriver">
                                <el-radio-group v-model="form.cgroupDriver">
                                    <el-radio label="cgroupfs">cgroupfs</el-radio>
                                    <el-radio label="systemd">systemd</el-radio>
                                </el-radio-group>
                            </el-form-item>
                            <el-form-item>
                                <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                                    {{ $t('commons.button.save') }}
                                </el-button>
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
            <span>{{ $t('container.iptablesHelper1') }}</span>
            <div style="margin-top: 10px">
                <span style="color: red; font-weight: 500">{{ $t('container.iptablesHelper2') }}</span>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="onSaveIptables(true)">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="onSaveIptables(false)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmitSave"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import { ElMessageBox, FormInstance } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import LayoutContent from '@/layout/layout-content.vue';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
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

const loading = ref(false);
const showDaemonJsonAlert = ref(false);
const extensions = [javascript(), oneDark];
const confShowType = ref('base');

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
    logMaxSize: 10,
    sizeUnit: 'm',
    logMaxFile: 3,
});
const rules = reactive({
    logMaxSize: [checkNumberRange(1, 1024000)],
    logMaxFile: [checkNumberRange(1, 100)],
});

const formRef = ref<FormInstance>();
const dockerConf = ref();
const confirmDialogRef = ref();

const iptablesVisiable = ref();

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!valid) return;
        let params = {
            header: i18n.global.t('database.confChange'),
            operationInfo: i18n.global.t('database.restartNowHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialogRef.value!.acceptParams(params);
    });
};
const onSaveFile = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRef.value!.acceptParams(params);
};

const toDoc = () => {
    window.open('https://1panel.cn/docs/user_manual/containers/setting/', '_blank');
};

const onChangeIptables = () => {
    if (!form.iptables) {
        iptablesVisiable.value = true;
    }
};

const onSaveIptables = (status: boolean) => {
    form.iptables = status;
    iptablesVisiable.value = false;
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

const onSubmitSave = async () => {
    if (confShowType.value === 'all') {
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
    }
    let itemMirrors = form.mirrors.split('\n');
    let itemRegistries = form.registries.split('\n');
    let param = {
        isSwarm: form.isSwarm,
        status: form.status,
        version: '',
        registryMirrors: itemMirrors.filter(function (el) {
            return el !== null && el !== '' && el !== undefined;
        }),
        insecureRegistries: itemRegistries.filter(function (el) {
            return el !== null && el !== '' && el !== undefined;
        }),
        liveRestore: form.liveRestore,
        iptables: form.iptables,
        cgroupDriver: form.cgroupDriver,
        logMaxSize: form.logMaxSize + form.sizeUnit,
        logMaxFile: form.logMaxFile + '',
    };
    if (!form.logOptionShow) {
        param.logMaxFile = '';
        param.logMaxSize = '';
    }

    loading.value = true;
    await updateDaemonJson(param)
        .then(() => {
            loading.value = false;
            search();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
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
    loading.value = true;
    await loadDaemonJson()
        .then((res) => {
            loading.value = false;
            form.isSwarm = res.data.isSwarm;
            form.status = res.data.status;
            form.version = res.data.version;
            form.cgroupDriver = res.data.cgroupDriver;
            form.liveRestore = res.data.liveRestore;
            form.iptables = res.data.iptables;
            form.mirrors = res.data.registryMirrors ? res.data.registryMirrors.join('\n') : '';
            form.registries = res.data.insecureRegistries ? res.data.insecureRegistries.join('\n') : '';
            if (res.data.logMaxFile || res.data.logMaxSize) {
                form.logOptionShow = true;
            }
            form.logMaxFile = Number(res.data.logMaxFile);
            form.logMaxSize = loadSize(res.data.logMaxSize);
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadSize = (value: string) => {
    if (value.indexOf('b') !== -1 || value.indexOf('B') !== -1) {
        form.sizeUnit = 'b';
        return Number(value.replaceAll('b', '').replaceAll('B', ''));
    }
    if (value.indexOf('k') !== -1 || value.indexOf('K') !== -1) {
        form.sizeUnit = 'k';
        return Number(value.replaceAll('k', '').replaceAll('K', ''));
    }
    if (value.indexOf('m') !== -1 || value.indexOf('M') !== -1) {
        form.sizeUnit = 'm';
        return Number(value.replaceAll('m', '').replaceAll('M', ''));
    }
    if (value.indexOf('g') !== -1 || value.indexOf('G') !== -1) {
        form.sizeUnit = 'g';
        return Number(value.replaceAll('g', '').replaceAll('G', ''));
    }
    if (value.indexOf('t') !== -1 || value.indexOf('T') !== -1) {
        form.sizeUnit = 't';
        return Number(value.replaceAll('t', '').replaceAll('T', ''));
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
</style>
