<template>
    <el-drawer :close-on-click-modal="false" :close-on-press-escape="false" v-model="open" size="50%">
        <template #header>
            <Header
                :header="$t('commons.button.' + operateReq.operate)"
                :resource="resourceName"
                :back="handleClose"
            ></Header>
        </template>
        <el-row :gutter="10">
            <el-col :span="22" :offset="1">
                <div>
                    <el-descriptions direction="vertical">
                        <el-descriptions-item>
                            <el-link @click="toLink(app.website)">
                                <el-icon><OfficeBuilding /></el-icon>
                                <span>{{ $t('app.appOfficeWebsite') }}</span>
                            </el-link>
                        </el-descriptions-item>
                        <el-descriptions-item>
                            <el-link @click="toLink(app.document)">
                                <el-icon><Document /></el-icon>
                                <span>{{ $t('app.document') }}</span>
                            </el-link>
                        </el-descriptions-item>
                        <el-descriptions-item>
                            <el-link @click="toLink(app.github)">
                                <el-icon><Link /></el-icon>
                                <span>{{ $t('app.github') }}</span>
                            </el-link>
                        </el-descriptions-item>
                    </el-descriptions>
                </div>
            </el-col>
            <el-col :span="22" :offset="1">
                <el-form
                    @submit.prevent
                    ref="updateRef"
                    :rules="rules"
                    label-position="top"
                    :model="operateReq"
                    v-loading="loading"
                >
                    <el-form-item :label="$t('app.versionSelect')" prop="detailId">
                        <el-select v-model="operateReq.version" @change="getVersions(operateReq.version)">
                            <el-option
                                v-for="(version, index) in versions"
                                :key="index"
                                :value="version.version"
                                :label="version.version"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item prop="backup" v-if="operateReq.operate === 'upgrade'">
                        <el-checkbox v-model="operateReq.backup" :label="$t('app.backupApp')" />
                        <span class="input-help">
                            <el-text type="warning">{{ $t('app.backupAppHelper') }}</el-text>
                        </span>
                    </el-form-item>
                    <el-form-item pro="pullImage" v-if="operateReq.operate === 'upgrade'">
                        <el-checkbox v-model="operateReq.pullImage" :label="$t('app.pullImage')" size="large" />
                        <span class="input-help">{{ $t('app.pullImageHelper') }}</span>
                    </el-form-item>
                </el-form>
            </el-col>

            <el-col :span="22" :offset="1" v-if="operateReq.operate === 'upgrade'">
                <el-text type="warning">{{ $t('app.upgradeWarn') }}</el-text>
                <el-button class="ml-1.5" type="text" @click="openDiff()">{{ $t('app.showDiff') }}</el-button>
                <div>
                    <el-checkbox v-model="useNewCompose" :label="$t('app.useCustom')" size="large" />
                </div>
                <div v-if="useNewCompose">
                    <el-text type="danger">{{ $t('app.useCustomHelper') }}</el-text>
                </div>
                <codemirror
                    v-if="useNewCompose"
                    :autofocus="true"
                    placeholder=""
                    :indent-with-tab="true"
                    :tabSize="4"
                    style="width: 100%; height: calc(100vh - 500px); margin-top: 10px"
                    :lineWrapping="true"
                    :matchBrackets="true"
                    theme="cobalt"
                    :styleActiveLine="true"
                    :extensions="extensions"
                    v-model="newCompose"
                />
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onOperate" :disabled="versions == null || loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
        <Diff ref="composeDiffRef" @confirm="getNewCompose" />
    </el-drawer>
</template>
<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { GetAppUpdateVersions, IgnoreUpgrade, InstalledOp } from '@/api/modules/app';
import i18n from '@/lang';
import { ElMessageBox, FormInstance } from 'element-plus';
import { reactive, ref, onBeforeUnmount } from 'vue';
import Header from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import Diff from './diff/index.vue';
import bus from '../../bus';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
const extensions = [javascript(), oneDark];

const composeDiffRef = ref();
const updateRef = ref<FormInstance>();
const open = ref(false);
const loading = ref(false);
const versions = ref<App.VersionDetail[]>();
const operateReq = reactive({
    detailId: 0,
    operate: 'upgrade',
    installId: 0,
    backup: true,
    pullImage: true,
    version: '',
    dockerCompose: '',
});
const resourceName = ref('');
const rules = ref<any>({
    detailId: [Rules.requiredSelect],
});
const app = ref();
const oldContent = ref('');
const newContent = ref('');
const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    em('close', open);
};

const newCompose = ref('');
const useNewCompose = ref(false);
const appInstallID = ref(0);

const toLink = (link: string) => {
    window.open(link, '_blank');
};

const openDiff = async () => {
    if (newContent.value === '') {
        await getVersions(operateReq.version);
    }
    composeDiffRef.value.acceptParams(oldContent.value, newContent.value);
};

const getNewCompose = (compose: string) => {
    if (compose !== '') {
        newCompose.value = compose;
        useNewCompose.value = true;
    } else {
        newCompose.value = newContent.value;
        useNewCompose.value = false;
    }
};

const initData = () => {
    newCompose.value = '';
    useNewCompose.value = false;
    operateReq.backup = true;
    operateReq.pullImage = true;
    operateReq.dockerCompose = '';
};

const acceptParams = (id: number, name: string, dockerCompose: string, op: string, appDetail: App.AppDetail) => {
    initData();
    operateReq.installId = id;
    operateReq.operate = op;
    resourceName.value = name;
    app.value = appDetail;
    oldContent.value = dockerCompose;
    appInstallID.value = id;
    getVersions('');
    open.value = true;
};

const getVersions = async (version: string) => {
    const req = {
        appInstallID: appInstallID.value,
    };
    if (version !== '') {
        req['updateVersion'] = version;
    }
    try {
        const res = await GetAppUpdateVersions(req);
        versions.value = res.data || [];
        if (res.data != null && res.data.length > 0) {
            let item = res.data[0];
            if (version != '') {
                item = res.data.find((v) => v.version === version);
            }
            operateReq.detailId = item.detailId;
            operateReq.version = item.version;
            newContent.value = item.dockerCompose;
            newCompose.value = item.dockerCompose;
            useNewCompose.value = false;
        }
    } catch (error) {}
};

const operate = async () => {
    loading.value = true;
    if (operateReq.operate === 'upgrade') {
        if (useNewCompose.value) {
            operateReq.dockerCompose = newCompose.value;
        }
        await InstalledOp(operateReq)
            .then(() => {
                MsgSuccess(i18n.global.t('app.upgradeStart'));
                bus.emit('upgrade', true);
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    } else {
        await IgnoreUpgrade(operateReq)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                bus.emit('upgrade', true);
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    }
};

const onOperate = async () => {
    ElMessageBox.confirm(
        i18n.global.t('app.operatorHelper', [i18n.global.t('commons.button.' + operateReq.operate)]),
        i18n.global.t('commons.button.upgrade'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(() => {
        operate();
    });
};

onBeforeUnmount(() => {
    bus.off('upgrade');
});

defineExpose({
    acceptParams,
});
</script>
