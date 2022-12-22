<template>
    <div>
        <Submenu activeName="panel" />
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
                            prop="userName"
                        >
                            <el-input clearable v-model="form.userName">
                                <template #append>
                                    <el-button
                                        @click="onSave(panelFormRef, 'UserName', form.userName)"
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
                            prop="password"
                        >
                            <el-input type="password" clearable disabled v-model="form.password">
                                <template #append>
                                    <el-button icon="Setting" @click="onChangePassword">
                                        {{ $t('commons.button.set') }}
                                    </el-button>
                                </template>
                            </el-input>
                        </el-form-item>

                        <el-form-item :label="$t('setting.title')" :rules="Rules.requiredInput" prop="panelName">
                            <el-input clearable v-model="form.panelName">
                                <template #append>
                                    <el-button
                                        @click="onSave(panelFormRef, 'PanelName', form.panelName)"
                                        icon="Collection"
                                    >
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                </template>
                            </el-input>
                        </el-form-item>

                        <el-form-item :label="$t('setting.theme')" :rules="Rules.requiredSelect" prop="theme">
                            <el-radio-group @change="onSave(panelFormRef, 'Theme', form.theme)" v-model="form.theme">
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

                        <el-form-item :label="$t('setting.language')" :rules="Rules.requiredSelect" prop="language">
                            <el-radio-group
                                style="width: 100%"
                                @change="onSave(panelFormRef, 'Language', form.language)"
                                v-model="form.language"
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

                        <el-form-item :label="$t('setting.sessionTimeout')" :rules="Rules.number" prop="sessionTimeout">
                            <el-input v-model.number="form.sessionTimeout">
                                <template #append>
                                    <el-button
                                        @click="onSave(panelFormRef, 'SessionTimeout', form.sessionTimeout)"
                                        icon="Collection"
                                    >
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                </template>
                            </el-input>
                            <div>
                                <span class="input-help">
                                    {{ $t('setting.sessionTimeoutHelper', [form.sessionTimeout]) }}
                                </span>
                            </div>
                        </el-form-item>

                        <el-form-item :label="$t('setting.syncTime')">
                            <el-input disabled v-model="form.localTime">
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
                    v-if="form.complexityVerification === 'disable'"
                    :label="$t('setting.newPassword')"
                    prop="newPassword"
                >
                    <el-input type="password" show-password clearable v-model="passForm.newPassword" />
                </el-form-item>
                <el-form-item
                    v-if="form.complexityVerification === 'enable'"
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
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { ElMessage, ElForm } from 'element-plus';
import { updatePassword, syncTime, getSettingInfo, updateSetting } from '@/api/modules/setting';
import Submenu from '@/views/setting/index.vue';
import router from '@/routers/router';
import { Rules } from '@/global/form-rules';
import { GlobalStore } from '@/store';
import { useI18n } from 'vue-i18n';
import { useTheme } from '@/hooks/use-theme';

const i18n = useI18n();
const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);
const { switchDark } = useTheme();

type FormInstance = InstanceType<typeof ElForm>;
const passFormRef = ref<FormInstance>();
const passRules = reactive({
    oldPassword: [Rules.requiredInput],
    newPassword: [Rules.requiredInput, { min: 6, message: i18n.t('commons.rule.commonPassword'), trigger: 'blur' }],
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
const form = reactive({
    userName: '',
    password: '',
    email: '',
    sessionTimeout: 0,
    localTime: '',
    panelName: '',
    theme: '',
    language: '',
    complexityVerification: '',
});

const search = async () => {
    const res = await getSettingInfo();
    form.userName = res.data.userName;
    form.password = '******';
    form.sessionTimeout = res.data.sessionTimeout;
    form.localTime = res.data.localTime;
    form.panelName = res.data.panelName;
    form.theme = res.data.theme;
    form.language = res.data.language;
    form.complexityVerification = res.data.complexityVerification;
};
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
    switch (key) {
        case 'Language':
            i18n.locale.value = val;
            globalStore.updateLanguage(val);
            break;
        case 'Theme':
            globalStore.setThemeConfig({ ...themeConfig.value, theme: val });
            switchDark();
            break;
        case 'SessionTimeout':
            if (Number(val) < 300) {
                ElMessage.error(i18n.t('setting.sessionTimeoutError'));
                search();
                return;
            }
        case 'PanelName':
            globalStore.setThemeConfig({ ...themeConfig.value, panelName: val });
            break;
        case 'MonitorStoreDays':
        case 'ServerPort':
            val = val + '';
            break;
    }
    let param = {
        key: key,
        value: val + '',
    };
    await updateSetting(param);
    ElMessage.success(i18n.t('commons.msg.operationSuccess'));
    search();
};
function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

function checkPassword(rule: any, value: any, callback: any) {
    let password = form.complexityVerification === 'disable' ? passForm.newPassword : passForm.newPasswordComplexity;
    if (password !== passForm.retryPassword) {
        return callback(new Error(i18n.t('commons.rule.rePassword')));
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
            form.complexityVerification === 'disable' ? passForm.newPassword : passForm.newPasswordComplexity;
        if (password === passForm.oldPassword) {
            ElMessage.error(i18n.t('setting.duplicatePassword'));
            return;
        }
        await updatePassword({ oldPassword: passForm.oldPassword, newPassword: password });
        passwordVisiable.value = false;
        ElMessage.success(i18n.t('commons.msg.operationSuccess'));
        router.push({ name: 'login', params: { code: '' } });
        globalStore.setLogStatus(false);
    });
};

const onSyncTime = async () => {
    const res = await syncTime();
    form.localTime = res.data;
    ElMessage.success(i18n.t('commons.msg.operationSuccess'));
};

onMounted(() => {
    search();
});
</script>
