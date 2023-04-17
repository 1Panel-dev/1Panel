<template>
    <el-drawer
        :close-on-click-modal="false"
        v-model="open"
        :title="$t('app.install')"
        size="50%"
        :before-close="handleClose"
    >
        <template #header>
            <Header :header="$t('app.install')" :back="handleClose" />
        </template>

        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form
                    @submit.prevent
                    ref="paramForm"
                    label-position="top"
                    :model="form"
                    label-width="150px"
                    :rules="rules"
                    :validate-on-rule-change="false"
                >
                    <el-form-item :label="$t('app.name')" prop="NAME">
                        <el-input v-model.trim="form['NAME']"></el-input>
                    </el-form-item>
                    <Params
                        v-if="open"
                        v-model:form="form"
                        v-model:params="installData.params"
                        v-model:rules="rules"
                    ></Params>
                </el-form>
            </el-col>
        </el-row>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(paramForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup name="appInstall">
import { App } from '@/api/interface/app';
import { InstallApp } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import { FormInstance, FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import Params from '../params/index.vue';
import Header from '@/components/drawer-header/index.vue';

const router = useRouter();

interface InstallRrops {
    appDetailId: number;
    params?: App.AppParams;
}

const installData = ref<InstallRrops>({
    appDetailId: 0,
});
let open = ref(false);
let form = ref<{ [key: string]: any }>({});
let rules = ref<FormRules>({
    NAME: [Rules.appName],
});
let loading = ref(false);
const paramForm = ref<FormInstance>();
const req = reactive({
    appDetailId: 0,
    params: {},
    name: '',
});

const handleClose = () => {
    open.value = false;
    resetForm();
};

const resetForm = () => {
    if (paramForm.value) {
        paramForm.value.clearValidate();
        paramForm.value.resetFields();
    }
};

const acceptParams = (props: InstallRrops): void => {
    installData.value = props;
    resetForm();
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        req.appDetailId = installData.value.appDetailId;
        req.params = form.value;
        req.name = form.value['NAME'];
        loading.value = true;
        InstallApp(req)
            .then(() => {
                handleClose();
                router.push({ path: '/apps/installed' });
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
