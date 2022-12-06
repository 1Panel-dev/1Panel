<template>
    <div>
        <Submenu activeName="setting" />
        <el-card style="margin-top: 20px">
            <el-radio-group v-model="confShowType" @change="changeMode">
                <el-radio-button label="base">{{ $t('database.baseConf') }}</el-radio-button>
                <el-radio-button label="all">{{ $t('database.allConf') }}</el-radio-button>
            </el-radio-group>
            <el-form v-if="confShowType === 'base'" :model="form" ref="formRef" label-width="120px">
                <el-row style="margin-top: 20px">
                    <el-col :span="1"><br /></el-col>
                    <el-col :span="10">
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
                        <el-form-item label="bip">
                            <el-input clearable v-model="form.bip" />
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
                            <el-button type="primary" @click="onSave(formRef)" style="width: 90px; margin-top: 5px">
                                {{ $t('commons.button.save') }}
                            </el-button>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <div v-if="confShowType === 'all'">
                <codemirror
                    :autofocus="true"
                    placeholder="None data"
                    :indent-with-tab="true"
                    :tabSize="4"
                    style="margin-top: 10px; height: calc(100vh - 280px)"
                    :lineWrapping="true"
                    :matchBrackets="true"
                    theme="cobalt"
                    :styleActiveLine="true"
                    :extensions="extensions"
                    v-model="dockerConf"
                    :readOnly="true"
                />
                <el-button type="primary" @click="onSaveFile" style="width: 90px; margin-top: 5px">
                    {{ $t('commons.button.save') }}
                </el-button>
            </div>
        </el-card>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmitSave"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import { ElMessage, FormInstance } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import Submenu from '@/views/container/index.vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { LoadFile } from '@/api/modules/files';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import i18n from '@/lang';
import { loadDaemonJson, updateDaemonJson, updateDaemonJsonByfile } from '@/api/modules/container';

const extensions = [javascript(), oneDark];
const confShowType = ref('base');

const form = reactive({
    status: '',
    bip: '',
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

const onSubmitSave = async () => {
    if (confShowType.value === 'all') {
        let param = {
            file: dockerConf.value,
            path: '/opt/1Panel/docker/conf/daemon.json',
        };
        await updateDaemonJsonByfile(param);
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        return;
    }
    let itemMirrors = form.mirrors.split('\n');
    let itemRegistries = form.registries.split('\n');
    let param = {
        status: form.status,
        bip: form.bip,
        registryMirrors: itemMirrors.filter(function (el) {
            return el !== null && el !== '' && el !== undefined;
        }),
        insecureRegistries: itemRegistries.filter(function (el) {
            return el !== null && el !== '' && el !== undefined;
        }),
        liveRestore: form.liveRestore,
        cgroupDriver: form.cgroupDriver,
    };
    await updateDaemonJson(param);
    search();
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

const loadMysqlConf = async () => {
    const res = await LoadFile({ path: '/opt/1Panel/docker/conf/daemon.json' });
    dockerConf.value = res.data;
};

const changeMode = async () => {
    if (confShowType.value === 'all') {
        loadMysqlConf();
    } else {
        search();
    }
};

const search = async () => {
    const res = await loadDaemonJson();
    form.bip = res.data.bip;
    form.status = res.data.status;
    form.cgroupDriver = res.data.cgroupDriver;
    form.liveRestore = res.data.liveRestore;
    form.mirrors = res.data.registryMirrors.join('\n');
    form.registries = res.data.insecureRegistries.join('\n');
};

onMounted(() => {
    search();
});
</script>
