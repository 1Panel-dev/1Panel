<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.panel')" :divider="true">
            <template #main>
                <el-form :model="form" ref="panelFormRef" label-position="left" v-loading="loading" label-width="160px">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :span="10">
                            <el-form-item :label="$t('setting.user')" :rules="Rules.requiredInput" prop="userName">
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

                            <el-form-item :label="$t('setting.passwd')" :rules="Rules.requiredInput" prop="password">
                                <el-input type="password" clearable disabled v-model="form.password">
                                    <template #append>
                                        <el-button icon="Setting" @click="onChangePassword">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>

                            <el-form-item :label="$t('setting.theme')" :rules="Rules.requiredSelect" prop="theme">
                                <el-radio-group
                                    @change="onSave(panelFormRef, 'Theme', form.theme)"
                                    v-model="form.theme"
                                >
                                    <el-radio-button label="light">
                                        <el-icon><Sunny /></el-icon>
                                        {{ $t('setting.light') }}
                                    </el-radio-button>
                                    <el-radio-button label="dark">
                                        <el-icon><Moon /></el-icon>
                                        {{ $t('setting.dark') }}
                                    </el-radio-button>
                                </el-radio-group>
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

                            <el-form-item :label="$t('setting.language')" :rules="Rules.requiredSelect" prop="language">
                                <el-radio-group
                                    style="width: 100%"
                                    @change="onSave(panelFormRef, 'Language', form.language)"
                                    v-model="form.language"
                                >
                                    <el-radio label="zh">中文</el-radio>
                                    <el-radio label="en">English</el-radio>
                                </el-radio-group>
                            </el-form-item>

                            <el-form-item
                                :label="$t('setting.sessionTimeout')"
                                :rules="Rules.number"
                                prop="sessionTimeout"
                            >
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
                                        <el-button v-show="!show" @click="onSyncTime" icon="Refresh">
                                            {{ $t('commons.button.sync') }}
                                        </el-button>
                                        <span v-show="show">{{ count }} S</span>
                                    </template>
                                </el-input>
                            </el-form-item>
                        </el-col>
                    </el-row>
                </el-form>
            </template>
        </LayoutContent>
        <Password ref="passwordRef" />
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { ElForm } from 'element-plus';
import LayoutContent from '@/layout/layout-content.vue';
import { syncTime, getSettingInfo, updateSetting } from '@/api/modules/setting';
import { Rules } from '@/global/form-rules';
import { GlobalStore } from '@/store';
import { useI18n } from 'vue-i18n';
import { useTheme } from '@/hooks/use-theme';
import { MsgError, MsgSuccess } from '@/utils/message';
import Password from '@/views/setting/panel/password/index.vue';

const loading = ref(false);
const i18n = useI18n();
const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);
const { switchDark } = useTheme();

type FormInstance = InstanceType<typeof ElForm>;

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

const timer = ref();
const TIME_COUNT = ref(10);
const count = ref();
const show = ref();

const passwordRef = ref();

const search = async () => {
    const res = await getSettingInfo();
    form.userName = res.data.userName;
    form.password = '******';
    form.sessionTimeout = Number(res.data.sessionTimeout);
    form.localTime = res.data.localTime;
    form.panelName = res.data.panelName;
    form.theme = res.data.theme;
    form.language = res.data.language;
    form.complexityVerification = res.data.complexityVerification;
};
const panelFormRef = ref<FormInstance>();

const onChangePassword = () => {
    passwordRef.value.acceptParams({ complexityVerification: form.complexityVerification });
};

const onSave = async (formEl: FormInstance | undefined, key: string, val: any) => {
    if (!formEl) return;
    const result = await formEl.validateField(key.replace(key[0], key[0].toLowerCase()), callback);
    if (!result) {
        return;
    }
    if (val === '') {
        return;
    }
    loading.value = true;
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
                MsgError(i18n.t('setting.sessionTimeoutError'));
                search();
                return;
            }
        case 'PanelName':
            globalStore.setThemeConfig({ ...themeConfig.value, panelName: val });
            document.title = val;
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
    await updateSetting(param)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

function countdown() {
    count.value = TIME_COUNT.value;
    show.value = true;
    timer.value = setInterval(() => {
        if (count.value > 0 && count.value <= TIME_COUNT.value) {
            count.value--;
        } else {
            show.value = false;
            clearInterval(timer.value);
            timer.value = null;
        }
    }, 1000);
}

function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

const onSyncTime = async () => {
    loading.value = true;
    await syncTime()
        .then((res) => {
            loading.value = false;
            form.localTime = res.data;
            countdown();
            MsgSuccess(i18n.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search();
});
</script>
