<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.safe')" :divider="true">
            <template #main>
                <el-form :model="form" ref="panelFormRef" v-loading="loading" label-position="left" label-width="180px">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :span="10">
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

                            <el-form-item
                                :label="$t('setting.expirationTime')"
                                prop="expirationTime"
                                :rules="Rules.requiredInput"
                            >
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
                            <el-form-item
                                :label="$t('setting.complexity')"
                                prop="complexityVerification"
                                :rules="Rules.requiredSelect"
                            >
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
                            <el-form-item
                                :label="$t('setting.mfa')"
                                prop="securityEntrance"
                                :rules="Rules.requiredSelect"
                            >
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
                        </el-col>
                    </el-row>
                </el-form>
            </template>
        </LayoutContent>
        <el-drawer v-model="timeoutVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('setting.expirationTime')" :back="handleClose" />
            </template>
            <el-form ref="timeoutFormRef" label-position="top" :model="timeoutForm">
                <el-form-item :label="$t('setting.days')" prop="days" :rules="Rules.number">
                    <el-input clearable v-model.number="timeoutForm.days" />
                    <span class="input-help">{{ $t('setting.expirationHelper') }}</span>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="timeoutVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="submitTimeout(timeoutFormRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue';
import { ElForm, ElMessageBox } from 'element-plus';
import { Setting } from '@/api/interface/setting';
import LayoutContent from '@/layout/layout-content.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { updateSetting, getMFA, bindMFA, getSettingInfo, updatePort } from '@/api/modules/setting';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { dateFormatSimple } from '@/utils/util';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
const form = reactive({
    serverPort: 9999,
    securityEntrance: '',
    expirationDays: 0,
    expirationTime: '',
    complexityVerification: '',
    mfaStatus: 'disable',
    mfaSecret: 'disable',
});
type FormInstance = InstanceType<typeof ElForm>;
const timeoutFormRef = ref<FormInstance>();
const timeoutVisiable = ref<boolean>(false);
const timeoutForm = reactive({
    days: 0,
});

const search = async () => {
    const res = await getSettingInfo();
    form.serverPort = Number(res.data.serverPort);
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
                window.open(`${href.split('//')[0]}//${ip}:${val}/1panel/login`, '_self');
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const handleMFA = async () => {
    console.log('dawdwda');
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

const handleClose = () => {
    timeoutVisiable.value = false;
};

const onBind = async () => {
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
    timeoutForm.days = form.expirationDays;
    timeoutVisiable.value = true;
};

const submitTimeout = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let time = new Date(new Date().getTime() + 3600 * 1000 * 24 * timeoutForm.days);
        loading.value = true;
        await updateSetting({ key: 'ExpirationDays', value: timeoutForm.days + '' })
            .then(() => {
                loading.value = false;
                search();
                loadTimeOut();
                form.expirationTime = dateFormatSimple(time);
                timeoutVisiable.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    });
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
});
</script>
