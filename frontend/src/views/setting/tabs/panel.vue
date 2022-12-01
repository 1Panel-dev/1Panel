<template>
    <el-form :model="form" ref="panelFormRef" label-position="left" label-width="160px">
        <el-card style="margin-top: 20px">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('setting.panel') }}</span>
                </div>
            </template>
            <el-row>
                <el-col :span="1"><br /></el-col>
                <el-col :span="10">
                    <el-form-item
                        :label="$t('commons.login.username')"
                        :rules="Rules.requiredInput"
                        prop="settingInfo.userName"
                    >
                        <el-input clearable v-model="form.settingInfo.userName">
                            <template #append>
                                <el-button
                                    @click="onSave(panelFormRef, 'UserName', form.settingInfo.userName)"
                                    icon="Collection"
                                >
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>

                    <el-form-item
                        :label="$t('commons.login.password')"
                        :rules="Rules.requiredInput"
                        prop="settingInfo.password"
                    >
                        <el-input type="password" clearable disabled v-model="form.settingInfo.password">
                            <template #append>
                                <el-button icon="Setting" @click="onChangePassword">
                                    {{ $t('commons.button.set') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>

                    <el-form-item
                        :label="$t('setting.title')"
                        :rules="Rules.requiredInput"
                        prop="settingInfo.panelName"
                    >
                        <el-input clearable v-model="form.settingInfo.panelName">
                            <template #append>
                                <el-button
                                    @click="onSave(panelFormRef, 'PanelName', form.settingInfo.panelName)"
                                    icon="Collection"
                                >
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>

                    <el-form-item :label="$t('setting.theme')" :rules="Rules.requiredSelect" prop="settingInfo.theme">
                        <el-radio-group
                            @change="onSave(panelFormRef, 'Theme', form.settingInfo.theme)"
                            v-model="form.settingInfo.theme"
                        >
                            <el-radio-button label="dark">
                                <el-icon><Moon /></el-icon>
                                {{ $t('setting.dark') }}
                            </el-radio-button>
                            <el-radio-button label="light">
                                <el-icon><Sunny /></el-icon>
                                {{ $t('setting.light') }}
                            </el-radio-button>
                        </el-radio-group>
                    </el-form-item>

                    <el-form-item
                        :label="$t('setting.language')"
                        :rules="Rules.requiredSelect"
                        prop="settingInfo.language"
                    >
                        <el-radio-group
                            style="width: 100%"
                            @change="onSave(panelFormRef, 'Language', form.settingInfo.language)"
                            v-model="form.settingInfo.language"
                        >
                            <el-radio-button label="zh">中文</el-radio-button>
                            <el-radio-button label="en">English</el-radio-button>
                        </el-radio-group>
                        <div>
                            <span class="input-help">
                                {{ $t('setting.languageHelper') }}
                            </span>
                        </div>
                    </el-form-item>

                    <el-form-item
                        :label="$t('setting.sessionTimeout')"
                        :rules="sessionTimeoutRules"
                        prop="settingInfo.sessionTimeout"
                    >
                        <el-input v-model.number="form.settingInfo.sessionTimeout">
                            <template #append>
                                <el-button
                                    @click="onSave(panelFormRef, 'SessionTimeout', form.settingInfo.sessionTimeout)"
                                    icon="Collection"
                                >
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </template>
                        </el-input>
                        <div>
                            <span class="input-help">
                                {{ $t('setting.sessionTimeoutHelper', [form.settingInfo.sessionTimeout]) }}
                            </span>
                        </div>
                    </el-form-item>

                    <el-form-item :label="$t('setting.syncTime')">
                        <el-input disabled v-model="form.settingInfo.localTime">
                            <template #append>
                                <el-button @click="onSyncTime" icon="Refresh">
                                    {{ $t('commons.button.sync') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-card>
    </el-form>
    <el-dialog
        v-model="passwordVisiable"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :title="$t('setting.changePassword')"
        width="30%"
    >
        <el-form ref="passFormRef" label-width="80px" label-position="left" :model="passForm" :rules="passRules">
            <el-form-item :label="$t('setting.oldPassword')" prop="oldPassword">
                <el-input type="password" show-password clearable v-model="passForm.oldPassword" />
            </el-form-item>
            <el-form-item
                v-if="form.settingInfo.complexityVerification === 'disable'"
                :label="$t('setting.newPassword')"
                prop="newPassword"
            >
                <el-input type="password" show-password clearable v-model="passForm.newPassword" />
            </el-form-item>
            <el-form-item
                v-if="form.settingInfo.complexityVerification === 'enable'"
                :label="$t('setting.newPassword')"
                prop="newPasswordComplexity"
            >
                <el-input type="password" show-password clearable v-model="passForm.newPasswordComplexity" />
            </el-form-item>
            <el-form-item :label="$t('setting.retryPassword')" prop="retryPassword">
                <el-input type="password" show-password clearable v-model="passForm.retryPassword" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="passwordVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button @click="submitChangePassword(passFormRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { ref, reactive } from 'vue';
import { ElMessage, ElForm } from 'element-plus';
import { updatePassword, syncTime } from '@/api/modules/setting';
import router from '@/routers/router';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const emit = defineEmits(['on-save', 'search']);

type FormInstance = InstanceType<typeof ElForm>;
const passFormRef = ref<FormInstance>();
const passRules = reactive({
    oldPassword: [Rules.requiredInput],
    newPassword: [
        Rules.requiredInput,
        { min: 6, message: i18n.global.t('commons.rule.commonPassword'), trigger: 'blur' },
    ],
    newPasswordComplexity: [Rules.requiredInput, Rules.password],
    retryPassword: [Rules.requiredInput, { validator: checkPassword, trigger: 'blur' }],
});
const passwordVisiable = ref<boolean>(false);
const passForm = reactive({
    oldPassword: '',
    newPassword: '',
    newPasswordComplexity: '',
    retryPassword: '',
});

const sessionTimeoutRules = [
    Rules.number,
    { min: 300, type: 'number', message: i18n.global.t('setting.sessionTimeoutError'), trigger: 'blur' },
];

interface Props {
    settingInfo: any;
}
const form = withDefaults(defineProps<Props>(), {
    settingInfo: {
        userName: '',
        password: '',
        email: '',
        sessionTimeout: 0,
        localTime: '',
        panelName: '',
        theme: '',
        language: '',
        complexityVerification: '',
    },
});

const panelFormRef = ref<FormInstance>();

function onSave(formEl: FormInstance | undefined, key: string, val: any) {
    emit('on-save', formEl, key, val);
}

function checkPassword(rule: any, value: any, callback: any) {
    let password =
        form.settingInfo.complexityVerification === 'disable' ? passForm.newPassword : passForm.newPasswordComplexity;
    if (password !== passForm.retryPassword) {
        return callback(new Error(i18n.global.t('commons.rule.rePassword')));
    }
    callback();
}

const onChangePassword = async () => {
    passForm.oldPassword = '';
    passForm.newPassword = '';
    passForm.newPasswordComplexity = '';
    passForm.retryPassword = '';
    passwordVisiable.value = true;
};

const submitChangePassword = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let password =
            form.settingInfo.complexityVerification === 'disable'
                ? passForm.newPassword
                : passForm.newPasswordComplexity;
        await updatePassword({ oldPassword: passForm.oldPassword, newPassword: password });
        passwordVisiable.value = false;
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        router.push({ name: 'login', params: { code: '' } });
        globalStore.setLogStatus(false);
    });
};

const onSyncTime = async () => {
    const res = await syncTime();
    emit('search');
    form.settingInfo.localTime = res.data;
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};
</script>
