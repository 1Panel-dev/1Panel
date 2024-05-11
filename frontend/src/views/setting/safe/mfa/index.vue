<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            @close="handleClose"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('setting.mfa')" :back="handleClose" />
            </template>
            <el-alert class="common-prompt" :closable="false" type="warning">
                <template #default>
                    <span>
                        <span>{{ $t('setting.mfaAlert') }}</span>
                    </span>
                </template>
            </el-alert>
            <el-form
                :model="form"
                ref="formRef"
                @submit.prevent
                v-loading="loading"
                label-position="top"
                :rules="rules"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.mfaHelper1')">
                            <ul class="help-ul">
                                <li>Google Authenticator</li>
                                <li>Microsoft Authenticator</li>
                                <li>1Password</li>
                                <li>LastPass</li>
                                <li>Authenticator</li>
                            </ul>
                        </el-form-item>
                        <el-form-item :label="$t('setting.mfaHelper2')">
                            <el-image class="w-32 h-32" :src="qrImage" />
                            <span class="input-help flex items-center">
                                <span>{{ $t('setting.secret') }}: {{ form.secret }}</span>
                                <CopyButton :content="form.secret" type="icon" />
                            </span>
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.title')" prop="title">
                            <el-input v-model="form.title">
                                <template #append>
                                    <el-button @click="loadMfaCodeBefore(formRef)">
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                </template>
                            </el-input>
                            <span class="input-help">{{ $t('setting.mfaTitleHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('setting.mfaInterval')" prop="interval">
                            <el-input v-model.number="form.interval">
                                <template #append>
                                    <el-button @click="loadMfaCodeBefore(formRef)">
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                </template>
                            </el-input>
                            <span class="input-help">{{ $t('setting.mfaIntervalHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('setting.mfaCode')" prop="code">
                            <el-input v-model="form.code"></el-input>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onBind(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { bindMFA, loadMFA } from '@/api/modules/setting';
import { reactive, ref } from 'vue';
import { Rules, checkNumberRange } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';

const loading = ref();
const qrImage = ref();
const drawerVisible = ref();
const formRef = ref();

const form = reactive({
    title: '1Panel',
    code: '',
    secret: '',
    interval: 30,
});

const rules = reactive({
    code: [Rules.requiredInput],
    title: [Rules.simpleName],
    interval: [Rules.number, checkNumberRange(15, 60)],
});

interface DialogProps {
    interval: number;
}
const emit = defineEmits<{ (e: 'search'): void }>();
const acceptParams = (params: DialogProps): void => {
    form.interval = params.interval;
    loadMfaCode();
    drawerVisible.value = true;
};

const loadMfaCodeBefore = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    const result = await formEl.validateField('interval', callback);
    if (!result) {
        return;
    }
    const result2 = await formEl.validateField('title', callback);
    if (!result2) {
        return;
    }
    loadMfaCode();
};
const loadMfaCode = async () => {
    let param = {
        title: form.title,
        interval: form.interval,
    };
    const res = await loadMFA(param);
    form.secret = res.data.secret;
    qrImage.value = res.data.qrImage;
};

function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

const onBind = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let param = {
            code: form.code,
            secret: form.secret,
            interval: form.interval + '',
        };
        loading.value = true;
        await bindMFA(param)
            .then(() => {
                loading.value = false;
                drawerVisible.value = false;
                emit('search');
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
<style scoped>
.help-ul {
    color: #8f959e;
}
</style>
