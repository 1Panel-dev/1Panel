<template>
    <div>
        <el-drawer
            v-model="drawerVisiable"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            @close="handleClose"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('setting.mfa')" :back="handleClose" />
            </template>
            <el-alert :closable="false" type="warning">
                <template #default>
                    <span>
                        <span>{{ $t('setting.mfaAlert') }}</span>
                    </span>
                </template>
            </el-alert>
            <el-form
                :model="form"
                style="margin-top: 20px"
                ref="formRef"
                @submit.prevent
                v-loading="loading"
                label-position="top"
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
                            <el-image style="width: 120px; height: 120px" :src="qrImage" />
                            <span class="input-help">
                                <span style="float: left">{{ $t('setting.secret') }}: {{ form.secret }}</span>
                                <div style="float: left; margin-top: 2px">
                                    <el-icon
                                        color="#409EFC"
                                        style="cursor: pointer; margin-left: 10px"
                                        :size="18"
                                        @click="onCopy()"
                                    >
                                        <DocumentCopy />
                                    </el-icon>
                                </div>
                            </span>
                        </el-form-item>
                        <el-form-item :label="$t('setting.mfaCode')" prop="code" :rules="Rules.requiredInput">
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
import { bindMFA, getMFA } from '@/api/modules/setting';
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import useClipboard from 'vue-clipboard3';
const { toClipboard } = useClipboard();

const loading = ref();
const qrImage = ref();
const drawerVisiable = ref();
const formRef = ref();

const form = reactive({
    code: '',
    secret: '',
});

const emit = defineEmits<{ (e: 'search'): void }>();
const acceptParams = (): void => {
    loadMfaCode();
    drawerVisiable.value = true;
};

const onCopy = async () => {
    try {
        await toClipboard(form.secret);
        MsgSuccess(i18n.global.t('commons.msg.copySuccess'));
    } catch (e) {
        MsgError(i18n.global.t('commons.msg.copyfailed'));
    }
};

const loadMfaCode = async () => {
    const res = await getMFA();
    form.secret = res.data.secret;
    qrImage.value = res.data.qrImage;
};

const onBind = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await bindMFA(form)
            .then(() => {
                loading.value = false;
                drawerVisiable.value = false;
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
    drawerVisiable.value = false;
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
