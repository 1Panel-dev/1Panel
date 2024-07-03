<template>
    <el-drawer
        v-model="open"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="40%"
        :before-close="handleClose"
    >
        <template #header>
            <DrawerHeader :header="$t('commons.button.' + authBasic.operate)" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form-item>
                    <el-alert
                        v-if="authBasic.operate === 'edit'"
                        :title="$t('website.editBasicAuthHelper')"
                        type="info"
                        :closable="false"
                    />
                </el-form-item>
                <el-form ref="proxyForm" label-position="top" :model="authBasic" :rules="rules">
                    <el-form-item :label="$t('commons.login.username')" prop="username">
                        <el-input v-model.trim="authBasic.username" :disabled="authBasic.operate === 'edit'"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('commons.login.password')" prop="password">
                        <el-input type="password" clearable show-password v-model.trim="authBasic.password">
                            <template #append>
                                <el-button @click="random">
                                    {{ $t('commons.button.random') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.remark')" prop="remark">
                        <el-input v-model.trim="authBasic.remark"></el-input>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(proxyForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { OperateAuthConfig } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { Website } from '@/api/interface/website';
import { getRandomStr } from '@/utils/util';

const proxyForm = ref<FormInstance>();
const rules = ref({
    username: [Rules.requiredInput, Rules.name],
    password: [Rules.requiredInput],
});
const open = ref(false);
const loading = ref(false);

const initData = (): Website.NginxAuthConfig => ({
    websiteID: 0,
    operate: 'create',
    username: '',
    password: '',
    remark: '',
});

let authBasic = ref(initData());
const em = defineEmits(['close']);
const handleClose = () => {
    proxyForm.value?.resetFields();
    open.value = false;
    em('close', false);
};

const random = async () => {
    authBasic.value.password = getRandomStr(16);
};

const acceptParams = (proxyParam: Website.NginxAuthConfig) => {
    authBasic.value = proxyParam;
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        OperateAuthConfig(authBasic.value)
            .then(() => {
                if (authBasic.value.operate == 'create') {
                    MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                } else {
                    MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                }
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
