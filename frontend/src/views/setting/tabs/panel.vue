<template>
    <el-form size="small" :model="form" label-position="left" label-width="160px">
        <el-card style="margin-top: 20px">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('setting.panel') }}</span>
                </div>
            </template>
            <el-row>
                <el-col :span="1"><br /></el-col>
                <el-col :span="10">
                    <el-form-item :label="$t('auth.username')">
                        <el-input clearable v-model="form.settingInfo.userName">
                            <template #append>
                                <el-button
                                    @click="SaveSetting('UserName', form.settingInfo.userName)"
                                    icon="Collection"
                                >
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('auth.password')">
                        <el-input type="password" clearable disabled v-model="form.settingInfo.password">
                            <template #append>
                                <el-button icon="Setting" @click="passwordVisiable = true">
                                    {{ $t('commons.button.set') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('auth.email')">
                        <el-input clearable v-model="form.settingInfo.email">
                            <template #append>
                                <el-button @click="SaveSetting('Email', form.settingInfo.email)" icon="Collection">
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </template>
                        </el-input>
                        <div>
                            <span class="input-help">{{ $t('setting.emailHelper') }}</span>
                        </div>
                    </el-form-item>
                    <el-form-item :label="$t('setting.title')">
                        <el-input clearable v-model="form.settingInfo.panelName">
                            <template #append>
                                <el-button
                                    @click="SaveSetting('PanelName', form.settingInfo.panelName)"
                                    icon="Collection"
                                >
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('setting.theme')">
                        <el-radio-group
                            @change="SaveSetting('Theme', form.settingInfo.theme)"
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
                    <el-form-item :label="$t('setting.language')">
                        <el-radio-group
                            @change="SaveSetting('Language', form.settingInfo.language)"
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
                    <el-form-item :label="$t('setting.sessionTimeout')">
                        <el-input v-model="form.settingInfo.sessionTimeout">
                            <template #append>
                                <el-button
                                    @click="SaveSetting('SessionTimeout', form.settingInfo.sessionTimeout)"
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
    <el-dialog v-model="passwordVisiable" :title="$t('setting.changePassword')" width="30%">
        <el-form
            size="small"
            ref="passFormRef"
            label-width="80px"
            label-position="left"
            :model="passForm"
            :rules="passRules"
        >
            <el-form-item :label="$t('setting.oldPassword')" prop="oldPassword">
                <el-input type="password" show-password clearable v-model="passForm.oldPassword" />
            </el-form-item>
            <el-form-item :label="$t('setting.newPassword')" prop="newPassword">
                <el-input type="password" show-password clearable v-model="passForm.newPassword" />
            </el-form-item>
            <el-form-item :label="$t('setting.retryPassword')" prop="retryPassword">
                <el-input type="password" show-password clearable v-model="passForm.retryPassword" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="passwordVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button @click="submitChangePassword()">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { ref, reactive, computed } from 'vue';
import { ElMessage, ElForm } from 'element-plus';
import { updateSetting, updatePassword, syncTime } from '@/api/modules/setting';
import { Setting } from '@/api/interface/setting';
import { useI18n } from 'vue-i18n';
import { GlobalStore } from '@/store';
import { useTheme } from '@/hooks/use-theme';
import { Rules } from '@/global/form-rues';
import router from '@/routers/router';

const i18n = useI18n();
const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);

type FormInstance = InstanceType<typeof ElForm>;
const passFormRef = ref<FormInstance>();
const passRules = reactive({
    oldPassword: [Rules.requiredInput],
    newPassword: [Rules.requiredInput],
    retryPassword: [Rules.requiredInput, { validator: checkPassword, trigger: 'blur' }],
});
const passwordVisiable = ref<boolean>(false);
const passForm = reactive<Setting.PasswordUpdate>({
    oldPassword: '',
    newPassword: '',
    retryPassword: '',
});

interface Props {
    settingInfo: any;
}
const form = withDefaults(defineProps<Props>(), {
    settingInfo: {
        userName: '',
        password: '',
        email: '',
        sessionTimeout: '',
        localTime: '',
        panelName: '',
        theme: '',
        language: '',
    },
});

const { switchDark } = useTheme();

const SaveSetting = async (key: string, val: string) => {
    switch (key) {
        case 'Language':
            i18n.locale.value = val;
            globalStore.updateLanguage(val);
            break;
        case 'Theme':
            globalStore.setThemeConfig({ ...themeConfig.value, theme: val });
            switchDark();
            break;
        case 'PanelName':
            globalStore.setThemeConfig({ ...themeConfig.value, panelName: val });
            break;
    }
    let param = {
        key: key,
        value: val,
    };
    await updateSetting(param);
    ElMessage.success(i18n.t('commons.msg.operationSuccess'));
};

function checkPassword(rule: any, value: any, callback: any) {
    if (passForm.newPassword !== passForm.retryPassword) {
        return callback(new Error(i18n.t('commons.rule.rePassword')));
    }
    callback();
}
const submitChangePassword = async () => {
    await updatePassword(passForm);
    passwordVisiable.value = false;
    ElMessage.success(i18n.t('commons.msg.operationSuccess'));
    router.push({ name: 'login' });
    globalStore.setLogStatus(false);
};
const onSyncTime = async () => {
    const res = await syncTime();
    form.settingInfo.localTime = res.data;
    ElMessage.success(i18n.t('commons.msg.operationSuccess'));
};
</script>
