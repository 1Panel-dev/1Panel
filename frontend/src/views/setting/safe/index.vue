<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.safe')" :divider="true">
            <template #main>
                <el-form :model="form" ref="panelFormRef" v-loading="loading" label-position="left" label-width="180px">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :span="16">
                            <el-form-item :label="$t('setting.panelPort')" :rules="Rules.port" prop="serverPort">
                                <el-input clearable v-model.number="form.serverPort">
                                    <template #append>
                                        <el-button
                                            style="width: 85px"
                                            @click="onSavePort(panelFormRef, 'ServerPort', form.serverPort)"
                                            icon="Collection"
                                        >
                                            {{ $t('commons.button.save') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>

                            <el-form-item :label="$t('setting.entrance')">
                                <el-input
                                    type="password"
                                    disabled
                                    v-if="form.securityEntrance"
                                    v-model="form.securityEntrance"
                                >
                                    <template #append>
                                        <el-button style="width: 85px" @click="onChangeEntrance" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <el-input disabled v-if="!form.securityEntrance" v-model="unset">
                                    <template #append>
                                        <el-button style="width: 85px" @click="onChangeEntrance" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('setting.entranceHelper') }}</span>
                            </el-form-item>

                            <el-form-item :label="$t('setting.expirationTime')" prop="expirationTime">
                                <el-input disabled v-model="form.expirationTime">
                                    <template #append>
                                        <el-button style="width: 85px" @click="onChangeExpirationTime" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <div>
                                    <span class="input-help" v-if="form.expirationTime !== $t('setting.unSetting')">
                                        {{ $t('setting.timeoutHelper', [loadTimeOut()]) }}
                                    </span>
                                    <span class="input-help" v-else>
                                        {{ $t('setting.noneSetting') }}
                                    </span>
                                </div>
                            </el-form-item>
                            <el-form-item :label="$t('setting.complexity')" prop="complexityVerification">
                                <el-switch
                                    @change="
                                        onSave(panelFormRef, 'ComplexityVerification', form.complexityVerification)
                                    "
                                    v-model="form.complexityVerification"
                                    active-value="enable"
                                    inactive-value="disable"
                                />
                                <span class="input-help">
                                    {{ $t('setting.complexityHelper') }}
                                </span>
                            </el-form-item>
                            <el-form-item :label="$t('setting.mfa')" prop="securityEntrance">
                                <el-switch
                                    @change="handleMFA"
                                    v-model="form.mfaStatus"
                                    active-value="enable"
                                    inactive-value="disable"
                                />
                                <span class="input-help">
                                    {{ $t('setting.mfaHelper') }}
                                </span>
                            </el-form-item>
                            <el-form-item v-if="isMFAShow">
                                <el-card style="width: 100%">
                                    <ul style="line-height: 24px">
                                        <li>
                                            {{ $t('setting.mfaHelper1') }}
                                            <ul>
                                                <li>Google Authenticator</li>
                                                <li>Microsoft Authenticator</li>
                                                <li>1Password</li>
                                                <li>LastPass</li>
                                                <li>Authenticator</li>
                                            </ul>
                                        </li>
                                        <li>{{ $t('setting.mfaHelper2') }}</li>
                                        <el-image
                                            style="margin-left: 15px; width: 100px; height: 100px"
                                            :src="otp.qrImage"
                                        />
                                        <li>{{ $t('setting.mfaHelper3') }}</li>
                                        <el-input v-model="mfaCode"></el-input>
                                        <div style="margin-top: 10px; margin-bottom: 10px; float: right">
                                            <el-button @click="onCancelMfaBind">
                                                {{ $t('commons.button.cancel') }}
                                            </el-button>
                                            <el-button type="primary" @click="onBind">
                                                {{ $t('commons.button.saveAndEnable') }}
                                            </el-button>
                                        </div>
                                    </ul>
                                </el-card>
                            </el-form-item>

                            <el-form-item label="https" prop="ssl">
                                <el-switch
                                    @change="handleSSL"
                                    v-model="form.ssl"
                                    active-value="enable"
                                    inactive-value="disable"
                                />
                                <span class="input-help">{{ $t('setting.https') }}</span>
                                <SSLSetting
                                    :type="form.sslType"
                                    :status="form.ssl"
                                    v-if="sslShow"
                                    style="width: 100%"
                                />
                            </el-form-item>
                        </el-col>
                    </el-row>
                </el-form>
            </template>
        </LayoutContent>

        <EntranceSetting ref="entranceRef" @search="search" />
        <TimeoutSetting ref="timeoutref" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue';
import { ElForm, ElMessageBox } from 'element-plus';
import { Setting } from '@/api/interface/setting';
import LayoutContent from '@/layout/layout-content.vue';
import SSLSetting from '@/views/setting/safe/ssl/index.vue';
import TimeoutSetting from '@/views/setting/safe/timeout/index.vue';
import EntranceSetting from '@/views/setting/safe/entrance/index.vue';
import {
    updateSetting,
    getMFA,
    bindMFA,
    getSettingInfo,
    updatePort,
    getSystemAvailable,
    updateSSL,
} from '@/api/modules/setting';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { MsgError, MsgSuccess } from '@/utils/message';

const loading = ref(false);
const entranceRef = ref();
const timeoutref = ref();

const form = reactive({
    serverPort: 9999,
    ssl: 'disable',
    sslType: 'self',
    securityEntrance: '',
    expirationDays: 0,
    expirationTime: '',
    complexityVerification: '',
    mfaStatus: 'disable',
    mfaSecret: 'disable',
});
type FormInstance = InstanceType<typeof ElForm>;

const sslShow = ref();
const oldSSLStatus = ref();

const unset = ref(i18n.global.t('setting.unSetting'));

const search = async () => {
    const res = await getSettingInfo();
    form.serverPort = Number(res.data.serverPort);
    form.ssl = res.data.ssl;
    oldSSLStatus.value = res.data.ssl;
    form.sslType = res.data.sslType;
    if (form.ssl === 'enable') {
        sslShow.value = true;
    }
    form.securityEntrance = res.data.securityEntrance;
    form.expirationDays = Number(res.data.expirationDays);
    form.expirationTime = res.data.expirationTime;
    form.complexityVerification = res.data.complexityVerification;
    form.mfaStatus = res.data.mfaStatus;
    form.mfaSecret = res.data.mfaSecret;
};

const isMFAShow = ref<boolean>(false);
const otp = reactive<Setting.MFAInfo>({
    secret: '',
    qrImage: '',
});
const mfaCode = ref();
const panelFormRef = ref<FormInstance>();

const onSave = async (formEl: FormInstance | undefined, key: string, val: any) => {
    if (!formEl) return;
    const result = await formEl.validateField(key.replace(key[0], key[0].toLowerCase()), callback);
    if (!result) {
        return;
    }
    if (val === '') {
        return;
    }
    let param = {
        key: key,
        value: val + '',
    };
    loading.value = true;
    await updateSetting(param)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};
function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}
const onSavePort = async (formEl: FormInstance | undefined, key: string, val: any) => {
    if (!formEl) return;
    const result = await formEl.validateField(key.replace(key[0], key[0].toLowerCase()), callback);
    if (!result) {
        return;
    }
    ElMessageBox.confirm(i18n.global.t('setting.portChangeHelper'), i18n.global.t('setting.portChange'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        let param = {
            serverPort: val,
        };
        await updatePort(param)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                let href = window.location.href;
                let ip = href.split('//')[1].split(':')[0];
                window.open(`${href.split('//')[0]}//${ip}:${val}/`, '_self');
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const handleMFA = async () => {
    if (form.mfaStatus === 'enable') {
        const res = await getMFA();
        otp.secret = res.data.secret;
        otp.qrImage = res.data.qrImage;
        isMFAShow.value = true;
    } else {
        isMFAShow.value = false;
        loading.value = true;
        await updateSetting({ key: 'MFAStatus', value: 'disable' })
            .then(() => {
                loading.value = false;
                search();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    }
};

const onChangeEntrance = async () => {
    entranceRef.value.acceptParams({ securityEntrance: form.securityEntrance });
};

const handleSSL = async () => {
    if (form.ssl === 'enable') {
        sslShow.value = true;
        return;
    }
    if (form.ssl === oldSSLStatus.value) {
        sslShow.value = false;
        return;
    }
    ElMessageBox.confirm(i18n.global.t('setting.sslDisableHelper'), i18n.global.t('setting.sslDisable'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    })
        .then(async () => {
            sslShow.value = false;
            await updateSSL({ ssl: 'disable', domain: '', sslType: '', key: '', cert: '', sslID: 0 });
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            let href = window.location.href;
            let address = href.split('://')[1];
            window.open(`http://${address}/`, '_self');
        })
        .catch(() => {
            form.ssl = 'enable';
        });
};

const onBind = async () => {
    if (!mfaCode.value) {
        MsgError(i18n.global.t('commons.msg.comfimNoNull', ['code']));
        return;
    }
    loading.value = true;
    await bindMFA({ code: mfaCode.value, secret: otp.secret })
        .then(() => {
            loading.value = false;
            search();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            isMFAShow.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onCancelMfaBind = async () => {
    form.mfaStatus = 'disable';
    isMFAShow.value = false;
};

const onChangeExpirationTime = async () => {
    timeoutref.value.acceptParams({ expirationDays: form.expirationDays });
};

function loadTimeOut() {
    if (form.expirationDays === 0) {
        form.expirationTime = i18n.global.t('setting.unSetting');
        return i18n.global.t('setting.unSetting');
    }
    let staytimeGap = new Date(form.expirationTime).getTime() - new Date().getTime();
    if (staytimeGap < 0) {
        form.expirationTime = i18n.global.t('setting.unSetting');
        return i18n.global.t('setting.unSetting');
    }
    return Math.floor(staytimeGap / (3600 * 1000 * 24));
}

onMounted(() => {
    search();
    getSystemAvailable();
});
</script>
