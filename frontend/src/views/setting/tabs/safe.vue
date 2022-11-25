<template>
    <div>
        <el-form :model="form" ref="panelFormRef" label-position="left" label-width="160px">
            <el-card style="margin-top: 20px">
                <template #header>
                    <div class="card-header">
                        <span>{{ $t('setting.safe') }}</span>
                    </div>
                </template>
                <el-row>
                    <el-col :span="1"><br /></el-col>
                    <el-col :span="10">
                        <el-form-item
                            :label="$t('setting.expirationTime')"
                            prop="settingInfo.expirationTime"
                            :rules="Rules.requiredInput"
                        >
                            <el-input disabled v-model="form.settingInfo.expirationTime">
                                <template #append>
                                    <el-button @click="timeoutVisiable = true" icon="Collection">
                                        {{ $t('commons.button.set') }}
                                    </el-button>
                                </template>
                            </el-input>
                            <div>
                                <span
                                    class="input-help"
                                    v-if="form.settingInfo.expirationTime !== $t('setting.unSetting')"
                                >
                                    {{ $t('setting.timeoutHelper', [loadTimeOut()]) }}
                                </span>
                                <span class="input-help" v-else>
                                    {{ $t('setting.noneSetting') }}
                                </span>
                            </div>
                        </el-form-item>
                        <el-form-item
                            :label="$t('setting.complexity')"
                            prop="settingInfo.complexityVerification"
                            :rules="Rules.requiredSelect"
                        >
                            <el-radio-group
                                style="width: 100%"
                                @change="
                                    onSave(
                                        panelFormRef,
                                        'ComplexityVerification',
                                        form.settingInfo.complexityVerification,
                                    )
                                "
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
                        <el-form-item
                            :label="$t('setting.mfa')"
                            prop="settingInfo.securityEntrance"
                            :rules="Rules.requiredSelect"
                        >
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
                                        <el-button @click="onCancelMfaBind">
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
        <el-dialog v-model="timeoutVisiable" :title="$t('setting.expirationTime')" width="30%">
            <el-form ref="timeoutFormRef" label-width="80px" label-position="left" :model="timeoutForm">
                <el-form-item :label="$t('setting.days')" prop="days" :rules="Rules.number">
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
import { Rules } from '@/global/form-rules';
import { dateFromat } from '@/utils/util';

const emit = defineEmits(['on-save', 'search']);

interface Props {
    settingInfo: any;
}
const form = withDefaults(defineProps<Props>(), {
    settingInfo: {
        serverPort: '',
        securityEntrance: '',
        expirationTime: '',
        complexityVerification: '',
        mfaStatus: '',
        mfaSecret: '',
    },
});
type FormInstance = InstanceType<typeof ElForm>;
const timeoutFormRef = ref<FormInstance>();
const timeoutVisiable = ref<boolean>(false);
const timeoutForm = reactive({
    days: 0,
});

const isMFAShow = ref<boolean>(false);
const otp = reactive<Setting.MFAInfo>({
    secret: '',
    qrImage: '',
});
const mfaCode = ref();
const panelFormRef = ref<FormInstance>();

function onSave(formEl: FormInstance | undefined, key: string, val: any) {
    emit('on-save', formEl, key, val);
}

const handleMFA = async () => {
    if (form.settingInfo.mfaStatus === 'enable') {
        const res = await getMFA();
        otp.secret = res.data.secret;
        otp.qrImage = res.data.qrImage;
        isMFAShow.value = true;
    } else {
        await updateSetting({ key: 'MFAStatus', value: 'disable' });
        emit('search');
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    }
};

const onBind = async () => {
    await bindMFA({ code: mfaCode.value, secret: otp.secret });
    emit('search');
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    isMFAShow.value = false;
};

const onCancelMfaBind = async () => {
    form.settingInfo.mfaStatus = 'disable';
    isMFAShow.value = false;
};

const submitTimeout = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let time = new Date(new Date().getTime() + 3600 * 1000 * 24 * timeoutForm.days);
        await updateSetting({ key: 'ExpirationDays', value: timeoutForm.days + '' });
        emit('search');
        loadTimeOut();
        form.settingInfo.expirationTime = dateFromat(0, 0, time);
        timeoutVisiable.value = false;
    });
};

function loadTimeOut() {
    if (form.settingInfo.expirationDays === 0) {
        form.settingInfo.expirationTime = i18n.global.t('setting.unSetting');
        return i18n.global.t('setting.unSetting');
    }
    let staytimeGap = new Date(form.settingInfo.expirationTime).getTime() - new Date().getTime();
    if (staytimeGap < 0) {
        form.settingInfo.expirationTime = i18n.global.t('setting.unSetting');
        return i18n.global.t('setting.unSetting');
    }
    return Math.floor(staytimeGap / (3600 * 1000 * 24));
}
</script>
