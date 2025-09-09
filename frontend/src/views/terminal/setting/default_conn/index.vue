<template>
    <DrawerPro v-model="drawerVisible" :header="$t('terminal.defaultConn')" @close="handleClose" size="large">
        <el-form ref="formRef" label-position="top" :model="form" :rules="rules" v-loading="loading">
            <el-form-item :label="$t('terminal.ip')" prop="addr">
                <el-input @change="isOK = false" clearable v-model.trim="form.addr" />
            </el-form-item>
            <el-form-item :label="$t('commons.login.username')" prop="user">
                <el-input @change="isOK = false" clearable v-model="form.user" />
            </el-form-item>
            <el-form-item :label="$t('terminal.authMode')" prop="authMode">
                <el-radio-group @change="isOK = false" v-model="form.authMode">
                    <el-radio value="password">{{ $t('terminal.passwordMode') }}</el-radio>
                    <el-radio value="key">{{ $t('terminal.keyMode') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item :label="$t('commons.login.password')" v-if="form.authMode === 'password'" prop="password">
                <el-input @change="isOK = false" clearable show-password type="password" v-model="form.password" />
            </el-form-item>
            <el-form-item :label="$t('terminal.key')" v-if="form.authMode === 'key'" prop="privateKey">
                <el-input @change="isOK = false" clearable type="textarea" v-model="form.privateKey" />
            </el-form-item>
            <el-form-item :label="$t('terminal.keyPassword')" v-if="form.authMode === 'key'" prop="passPhrase">
                <el-input @change="isOK = false" type="password" show-password clearable v-model="form.passPhrase" />
            </el-form-item>
            <el-form-item style="margin-top: 10px" :label="$t('commons.table.port')" prop="port">
                <el-input @change="isOK = false" clearable v-model.number="form.port" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button @click="onTest(formRef)">
                    {{ $t('terminal.testConn') }}
                </el-button>
                <el-button type="primary" :disabled="!isOK" @click="onSave(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref, reactive } from 'vue';
import type { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { addHost, loadLocalConn, testByInfo } from '@/api/modules/terminal';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Base64 } from 'js-base64';

const loading = ref();
const isOK = ref(false);
const drawerVisible = ref(false);

const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const form = reactive({
    user: 'root',
    addr: '127.0.0.1',
    port: 22,
    authMode: 'password',
    password: '',
    privateKey: '',
    passPhrase: '',
    isLocal: true,

    id: 0,
    name: '',
    groupID: 0,
    description: '',
    rememberPassword: false,
});
const rules = reactive({
    addr: [Rules.ipV4V6OrDomain],
    user: [Rules.requiredInput],
    port: [Rules.requiredInput, Rules.port],
    authMode: [Rules.requiredSelect],
    password: [Rules.requiredInput],
    privateKey: [Rules.requiredInput],
});

const acceptParams = (): void => {
    search();
    drawerVisible.value = true;
};

const search = async () => {
    await loadLocalConn().then((res) => {
        if (res.data) {
            form.addr = res.data.addr;
            form.port = res.data.port;
            form.authMode = res.data.authMode;
            form.password = Base64.decode(res.data.password);
            form.privateKey = Base64.decode(res.data.privateKey);
            form.passPhrase = Base64.decode(res.data.passPhrase);
        }
    });
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

const onSave = (formEl: FormInstance) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await addHost(form)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                drawerVisible.value = false;
                emit('search');
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onTest = (formEl: FormInstance) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await testByInfo(form).then((res) => {
            loading.value = false;
            if (res.data) {
                isOK.value = true;
                MsgSuccess(i18n.global.t('terminal.connTestOk'));
            } else {
                isOK.value = false;
                MsgError(i18n.global.t('terminal.connTestFailed'));
            }
        });
    });
};

defineExpose({
    acceptParams,
});
</script>
