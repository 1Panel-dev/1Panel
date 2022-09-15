<template>
    <div>
        <el-form :model="form" label-position="left" label-width="160px">
            <el-card style="margin-top: 10px">
                <template #header>
                    <div class="card-header">
                        <span>{{ $t('setting.safe') }}</span>
                    </div>
                </template>
                <el-row>
                    <el-col :span="1"><br /></el-col>
                    <el-col :span="10">
                        <el-form-item :label="$t('setting.panelPort')">
                            <el-input clearable v-model="form.settingInfo.serverPort">
                                <template #append>
                                    <el-button
                                        @click="SaveSetting('ServerPort', form.settingInfo.serverPort)"
                                        icon="Collection"
                                    >
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                </template>
                                <el-tooltip
                                    class="box-item"
                                    effect="dark"
                                    content="Top Left prompts info"
                                    placement="top-start"
                                >
                                    <el-icon style="font-size: 14px; margin-top: 8px"><WarningFilled /></el-icon>
                                </el-tooltip>
                            </el-input>
                            <div>
                                <span class="input-help">
                                    {{ $t('setting.portHelper') }}
                                </span>
                            </div>
                        </el-form-item>
                        <el-form-item :label="$t('setting.safeEntrance')">
                            <el-input clearable v-model="form.settingInfo.securityEntrance">
                                <template #append>
                                    <el-button
                                        @click="SaveSetting('SecurityEntrance', form.settingInfo.securityEntrance)"
                                        icon="Collection"
                                    >
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                </template>
                            </el-input>
                            <div>
                                <span class="input-help">
                                    {{ $t('setting.safeEntranceHelper') }}
                                </span>
                            </div>
                        </el-form-item>
                        <el-form-item :label="$t('setting.passwordTimeout')">
                            <el-input clearable v-model="form.settingInfo.passwordTimeOut">
                                <template #append>
                                    <el-button @click="timeoutVisiable = true" icon="Collection">
                                        {{ $t('commons.button.set') }}
                                    </el-button>
                                </template>
                            </el-input>
                            <div>
                                <span class="input-help">
                                    {{ $t('setting.timeoutHelper', [loadTimeOut()]) }}
                                </span>
                            </div>
                        </el-form-item>
                        <el-form-item :label="$t('setting.complexity')">
                            <el-radio-group
                                @change="SaveSetting('ComplexityVerification', form.settingInfo.complexityVerification)"
                                v-model="form.settingInfo.complexityVerification"
                            >
                                <el-radio-button label="enable">{{ $t('commons.button.enable') }}</el-radio-button>
                                <el-radio-button label="disable">{{ $t('commons.button.disable') }}</el-radio-button>
                            </el-radio-group>
                            <div>
                                <span class="input-help">
                                    {{ $t('setting.complexityHelper') }}
                                </span>
                            </div>
                        </el-form-item>
                        <el-form-item :label="$t('setting.mfa')">
                            <el-radio-group @change="handleMFA()" v-model="form.settingInfo.mfaStatus">
                                <el-radio-button label="enable">{{ $t('commons.button.enable') }}</el-radio-button>
                                <el-radio-button label="disable">{{ $t('commons.button.disable') }}</el-radio-button>
                            </el-radio-group>
                        </el-form-item>
                        <div v-if="isMFAShow">
                            <el-card>
                                <ul style="margin-left: 120px; line-height: 24px">
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
                                        <el-button @click="form.settingInfo.mfaStatus = 'disable'">
                                            {{ $t('commons.button.cancel') }}
                                        </el-button>
                                        <el-button @click="onBind">{{ $t('commons.button.saveAndEnable') }}</el-button>
                                    </div>
                                </ul>
                            </el-card>
                        </div>
                    </el-col>
                </el-row>
            </el-card>
        </el-form>
        <el-dialog v-model="timeoutVisiable" :title="$t('setting.changePassword')" width="30%">
            <el-form ref="timeoutFormRef" label-width="80px" label-position="left" :model="timeoutForm">
                <el-form-item :label="$t('setting.oldPassword')" prop="days" :rules="Rules.number">
                    <el-input clearable v-model.number="timeoutForm.days" />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="timeoutVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button @click="submitTimeout(timeoutFormRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive } from 'vue';
import { ElMessage, ElForm } from 'element-plus';
import { Setting } from '@/api/interface/setting';
import { updateSetting, getMFA, bindMFA } from '@/api/modules/setting';
import i18n from '@/lang';
import { Rules } from '@/global/form-rues';
import { dateFromat } from '@/utils/util';

interface Props {
    settingInfo: any;
}
const form = withDefaults(defineProps<Props>(), {
    settingInfo: {
        serverPort: '',
        securityEntrance: '',
        passwordTimeOut: '',
        complexityVerification: '',
        mfaStatus: '',
        mfaSecret: '',
    },
});
type FormInstance = InstanceType<typeof ElForm>;
const timeoutFormRef = ref<FormInstance>();
const timeoutVisiable = ref<boolean>(false);
const timeoutForm = reactive({
    days: 10,
});

const isMFAShow = ref<boolean>(false);
const otp = reactive<Setting.MFAInfo>({
    secret: '',
    qrImage: '',
});
const mfaCode = ref();

const SaveSetting = async (key: string, val: string) => {
    let param = {
        key: key,
        value: val,
    };
    await updateSetting(param);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

const handleMFA = async () => {
    if (form.settingInfo.mfaStatus === 'enable') {
        const res = await getMFA();
        otp.secret = res.data.secret;
        otp.qrImage = res.data.qrImage;
        isMFAShow.value = true;
    } else {
        await updateSetting({ key: 'MFAStatus', value: 'disable' });
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    }
};

const onBind = async () => {
    await bindMFA({ code: mfaCode.value, secret: otp.secret });
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    isMFAShow.value = false;
};

const submitTimeout = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let time = new Date(new Date().getTime() + 3600 * 1000 * 24 * timeoutForm.days);
        SaveSetting('PasswordTimeOut', dateFromat(0, 0, time));
        form.settingInfo.passwordTimeOut = dateFromat(0, 0, time);
        timeoutVisiable.value = false;
    });
};

function loadTimeOut() {
    let staytimeGap = new Date(form.settingInfo.passwordTimeOut).getTime() - new Date().getTime();
    return Math.floor(staytimeGap / (3600 * 1000 * 24));
}
</script>
