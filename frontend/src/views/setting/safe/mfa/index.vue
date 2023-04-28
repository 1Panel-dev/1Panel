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
            <el-form :model="form" ref="formRef" v-loading="loading" label-position="top">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.mfaHelper1')">
                            <ul>
                                <li>Google Authenticator</li>
                                <li>Microsoft Authenticator</li>
                                <li>1Password</li>
                                <li>LastPass</li>
                                <li>Authenticator</li>
                            </ul>
                        </el-form-item>
                        <el-form-item :label="$t('setting.mfaTypeOption')">
                            <el-radio-group v-model="mode" @change="form.secret = ''">
                                <el-radio label="scan">{{ $t('setting.qrCode') }}</el-radio>
                                <el-radio label="input">{{ $t('setting.manualInput') }}</el-radio>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item :label="$t('setting.mfaHelper2')" v-if="mode === 'scan'">
                            <el-image style="width: 120px; height: 120px" :src="qrImage" />
                        </el-form-item>
                        <el-form-item
                            :label="$t('setting.mfaSecret')"
                            v-if="mode === 'input'"
                            prop="secret"
                            :rules="Rules.requiredInput"
                        >
                            <el-input v-model="form.secret"></el-input>
                        </el-form-item>

                        <el-form-item
                            :label="mode === 'scan' ? $t('setting.mfaHelper3') : $t('setting.mfaCode')"
                            prop="code"
                            :rules="Rules.requiredInput"
                        >
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
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';

const loading = ref();
const qrImage = ref();
const mode = ref('scan');
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
