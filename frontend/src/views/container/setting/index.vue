<template>
    <div>
        <div class="app-content" style="margin-top: 20px">
            <el-card class="app-card">
                <el-row :gutter="20">
                    <el-col :lg="3" :xl="2">
                        <div>
                            <el-tag effect="dark" type="success">Docker</el-tag>
                        </div>
                    </el-col>
                    <el-col :lg="3" :xl="2">
                        <div>
                            {{ $t('app.version') }}:
                            <el-tag type="info">{{ form.version }}</el-tag>
                        </div>
                    </el-col>
                    <el-col :lg="3" :xl="2">
                        <div>
                            {{ $t('commons.table.status') }}:
                            <el-tag v-if="form.status === 'Running'" type="success">
                                {{ $t('commons.status.running') }}
                            </el-tag>
                            <el-tag v-if="form.status === 'Stopped'" type="info">
                                {{ $t('commons.status.stopped') }}
                            </el-tag>
                        </div>
                    </el-col>
                    <el-col :lg="4" :xl="6">
                        <div v-if="form.status === 'Running'">
                            <el-button type="primary" @click="onOperator('stop')" link style="margin-left: 20px">
                                {{ $t('container.stop') }}
                            </el-button>
                            <el-divider direction="vertical" />
                            <el-button type="primary" @click="onOperator('restart')" link>
                                {{ $t('container.restart') }}
                            </el-button>
                        </div>
                        <div v-if="form.status === 'Stopped'">
                            <el-button type="primary" @click="onOperator('start')" link style="margin-left: 20px">
                                {{ $t('container.start') }}
                            </el-button>
                            <el-divider direction="vertical" />
                            <el-button type="primary" @click="onOperator('restart')" link>
                                {{ $t('container.restart') }}
                            </el-button>
                        </div>
                    </el-col>
                </el-row>
            </el-card>
        </div>

        <LayoutContent v-loading="loading" :title="$t('container.setting')" :divider="true">
            <template #main>
                <el-radio-group v-model="confShowType" @change="changeMode">
                    <el-radio-button label="base">{{ $t('database.baseConf') }}</el-radio-button>
                    <el-radio-button label="all">{{ $t('database.allConf') }}</el-radio-button>
                </el-radio-group>
                <el-row style="margin-top: 20px" v-if="confShowType === 'base'">
                    <el-col :span="1"><br /></el-col>
                    <el-col :span="10">
                        <el-form :model="form" ref="formRef" label-width="120px">
                            <el-form-item :label="$t('container.mirrors')" prop="mirrors">
                                <el-input
                                    type="textarea"
                                    :placeholder="$t('container.mirrorHelper')"
                                    :autosize="{ minRows: 3, maxRows: 10 }"
                                    v-model="form.mirrors"
                                />
                                <span class="input-help">{{ $t('container.mirrorsHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('container.registries')" prop="registries">
                                <el-input
                                    type="textarea"
                                    :placeholder="$t('container.registrieHelper')"
                                    :autosize="{ minRows: 3, maxRows: 10 }"
                                    v-model="form.registries"
                                />
                            </el-form-item>
                            <el-form-item label="live-restore" prop="liveRestore">
                                <el-switch v-model="form.liveRestore"></el-switch>
                                <span class="input-help">{{ $t('container.liveHelper') }}</span>
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
                        placeholder="None data"
                        :indent-with-tab="true"
                        :tabSize="4"
                        style="margin-top: 10px; height: calc(100vh - 380px)"
                        :lineWrapping="true"
                        :matchBrackets="true"
                        theme="cobalt"
                        :styleActiveLine="true"
                        :extensions="extensions"
                        v-model="dockerConf"
                        :readOnly="true"
                    />
                    <el-button :disabled="loading" type="primary" @click="onSaveFile" style="margin-top: 5px">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </div>
            </template>
        </LayoutContent>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmitSave"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import { ElMessage, FormInstance } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import LayoutContent from '@/layout/layout-content.vue';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { LoadFile } from '@/api/modules/files';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import i18n from '@/lang';
import { dockerOperate, loadDaemonJson, updateDaemonJson, updateDaemonJsonByfile } from '@/api/modules/container';

const loading = ref(false);

const extensions = [javascript(), oneDark];
const confShowType = ref('base');

const form = reactive({
    status: '',
    version: '',
    mirrors: '',
    registries: '',
    liveRestore: false,
    cgroupDriver: '',
});

const formRef = ref<FormInstance>();
const dockerConf = ref();
const confirmDialogRef = ref();

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
        operationInfo: i18n.global.t('database.restartNowHelper1'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRef.value!.acceptParams(params);
};

const onOperator = async (operation: string) => {
    let param = {
        operation: operation,
    };
    await dockerOperate(param);
    search();
    changeMode();
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

const onSubmitSave = async () => {
    if (confShowType.value === 'all') {
        let param = { file: dockerConf.value };
        loading.value = true;
        await updateDaemonJsonByfile(param)
            .then(() => {
                loading.value = false;
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
        return;
    }
    let itemMirrors = form.mirrors.split('\n');
    let itemRegistries = form.registries.split('\n');
    let param = {
        status: form.status,
        version: '',
        registryMirrors: itemMirrors.filter(function (el) {
            return el !== null && el !== '' && el !== undefined;
        }),
        insecureRegistries: itemRegistries.filter(function (el) {
            return el !== null && el !== '' && el !== undefined;
        }),
        liveRestore: form.liveRestore,
        cgroupDriver: form.cgroupDriver,
    };
    loading.value = true;
    await updateDaemonJson(param)
        .then(() => {
            loading.value = false;
            search();
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadDockerConf = async () => {
    const res = await LoadFile({ path: '/etc/docker/daemon.json' });
    dockerConf.value = res.data;
};

const changeMode = async () => {
    if (confShowType.value === 'all') {
        loadDockerConf();
    } else {
        search();
    }
};

const search = async () => {
    const res = await loadDaemonJson();
    form.status = res.data.status;
    form.version = res.data.version;
    form.cgroupDriver = res.data.cgroupDriver;
    form.liveRestore = res.data.liveRestore;
    form.mirrors = res.data.registryMirrors ? res.data.registryMirrors.join('\n') : '';
    form.registries = res.data.insecureRegistries ? res.data.insecureRegistries.join('\n') : '';
};

onMounted(() => {
    search();
});
</script>

<style lang="scss">
.app-card {
    font-size: 14px;
    height: 60px;
}

.app-content {
    height: 50px;
}

body {
    margin: 0;
}
</style>
