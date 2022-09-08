<template>
    <el-form size="small" :model="form" label-position="left" label-width="120px">
        <el-card style="margin-top: 20px">
            <template #header>
                <div class="card-header">
                    <span>面板</span>
                </div>
            </template>
            <el-row>
                <el-col :span="1"><br /></el-col>
                <el-col :span="8">
                    <el-form-item label="用户名">
                        <el-input clearable v-model="form.settingInfo.userName">
                            <template #append>
                                <el-button
                                    @click="SaveSetting('UserName', form.settingInfo.userName)"
                                    icon="Collection"
                                >
                                    保存
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item label="密码">
                        <el-input type="password" clearable disabled v-model="form.settingInfo.password">
                            <template #append>
                                <el-button icon="Setting" @click="passwordVisiable = true">设置</el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item label="邮箱">
                        <el-input clearable v-model="form.settingInfo.email">
                            <template #append>
                                <el-button @click="SaveSetting('Email', form.settingInfo.email)" icon="Collection">
                                    保存
                                </el-button>
                            </template>
                        </el-input>
                        <div><span class="input-help">用于密码找回</span></div>
                    </el-form-item>
                    <el-form-item label="面板别名">
                        <el-input clearable v-model="form.settingInfo.panelName">
                            <template #append>
                                <el-button
                                    @click="SaveSetting('PanelName', form.settingInfo.panelName)"
                                    icon="Collection"
                                >
                                    保存
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item label="主题色">
                        <el-radio-group
                            @change="SaveSetting('Theme', form.settingInfo.theme)"
                            v-model="form.settingInfo.theme"
                        >
                            <el-radio-button label="black">
                                <el-icon><Moon /></el-icon>黑金
                            </el-radio-button>
                            <el-radio-button label="auto" icon="Sunny">
                                <el-icon><MagicStick /></el-icon>自动
                            </el-radio-button>
                            <el-radio-button label="write">
                                <el-icon><Sunny /></el-icon>白金
                            </el-radio-button>
                        </el-radio-group>
                        <div>
                            <span class="input-help">
                                选择自动设置，将会在晚 6 点到次日早 6 点间自动切换到黑金主题。
                            </span>
                        </div>
                    </el-form-item>
                    <el-form-item label="系统语言">
                        <el-radio-group
                            @change="SaveSetting('Language', form.settingInfo.language)"
                            v-model="form.settingInfo.language"
                        >
                            <el-radio-button label="ch">中文 </el-radio-button>
                            <el-radio-button label="en">English </el-radio-button>
                        </el-radio-group>
                        <div>
                            <span class="input-help">
                                默认跟随浏览器语言，设置后只对当前浏览器生效，更换浏览器后需要重新设置。
                            </span>
                        </div>
                    </el-form-item>
                    <el-form-item label="超时时间">
                        <el-input v-model="form.settingInfo.sessionTimeout">
                            <template #append>
                                <el-button
                                    @click="SaveSetting('SessionTimeout', form.settingInfo.sessionTimeout)"
                                    icon="Collection"
                                >
                                    保存
                                </el-button>
                            </template>
                        </el-input>
                        <div>
                            <span class="input-help">如果用户超过 86400秒未操作面板，面板将自动退出登录 </span>
                        </div>
                    </el-form-item>
                    <el-form-item label="同步时间">
                        <el-input v-model="form.settingInfo.localTime">
                            <template #append>
                                <el-button icon="Refresh">同步</el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-card>
    </el-form>
    <el-dialog v-model="passwordVisiable" title="密码修改" width="30%">
        <el-form
            size="small"
            ref="passFormRef"
            label-width="80px"
            label-position="left"
            :model="passForm"
            :rules="passRules"
        >
            <el-form-item label="原密码" prop="oldPassword">
                <el-input type="password" show-password clearable v-model="passForm.oldPassword" />
            </el-form-item>
            <el-form-item label="新密码" prop="newPassword">
                <el-input type="password" show-password clearable v-model="passForm.newPassword" />
            </el-form-item>
            <el-form-item label="确认密码" prop="retryPassword">
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
import { ref, reactive } from 'vue';
import { ElMessage, ElForm } from 'element-plus';
import { updateSetting, updatePassword } from '@/api/modules/setting';
import { Setting } from '@/api/interface/setting';
import { useI18n } from 'vue-i18n';
import { GlobalStore } from '@/store';
import { useTheme } from '@/hooks/use-theme';
import { Rules } from '@/global/form-rues';
import router from '@/routers/router';

const i18n = useI18n();
const globalStore = GlobalStore();

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
        case 'theme':
            switchDark();
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
</script>
