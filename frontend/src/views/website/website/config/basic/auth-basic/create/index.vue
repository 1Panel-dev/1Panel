<template>
    <DrawerPro v-model="open" :header="$t('commons.button.' + authBasic.operate)" :back="handleClose">
        <el-form-item>
            <el-alert
                v-if="authBasic.operate === 'edit'"
                :title="$t('website.editBasicAuthHelper')"
                type="info"
                :closable="false"
            />
        </el-form-item>
        <el-form ref="proxyForm" label-position="top" :model="authBasic" :rules="rules">
            <el-form-item :label="$t('commons.table.name')" prop="name" v-if="authBasic.scope != 'root'">
                <el-input v-model.trim="authBasic.name" :disabled="authBasic.operate === 'edit'"></el-input>
            </el-form-item>
            <el-form-item :label="$t('website.path')" prop="path" v-if="authBasic.scope != 'root'">
                <el-input v-model.trim="authBasic.path" :disabled="authBasic.operate === 'edit'"></el-input>
            </el-form-item>
            <el-form-item :label="$t('commons.login.username')" prop="username">
                <el-input
                    v-model.trim="authBasic.username"
                    :disabled="authBasic.scope == 'root' && authBasic.operate === 'edit'"
                ></el-input>
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
        <template #footer>
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="submit(proxyForm)" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { operateAuthConfig, operatePathAuthConfig } from '@/api/modules/website';
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
    name: [Rules.requiredInput],
    path: [Rules.requiredInput],
});
const open = ref(false);
const loading = ref(false);

const initData = (): Website.NginxAuthConfig => ({
    websiteID: 0,
    operate: 'create',
    username: '',
    password: '',
    remark: '',
    scope: 'root',
    path: '',
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
    await formEl.validate(async (valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        try {
            if (authBasic.value.scope == 'root') {
                await operateAuthConfig(authBasic.value);
            } else {
                const req = {
                    websiteID: authBasic.value.websiteID,
                    path: authBasic.value.path,
                    name: authBasic.value.name,
                    username: authBasic.value.username,
                    password: authBasic.value.password,
                    operate: authBasic.value.operate,
                    remark: authBasic.value.remark,
                };
                await operatePathAuthConfig(req);
            }
            if (authBasic.value.operate == 'create') {
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
            } else {
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            }
            handleClose();
        } catch (error) {
        } finally {
            loading.value = false;
        }
    });
};

defineExpose({
    acceptParams,
});
</script>
