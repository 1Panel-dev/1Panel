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
                <el-alert type="info" :closable="false">
                    <p>{{ $t('app.appInstallWarn') }}</p>
                </el-alert>
                <el-form
                    @submit.prevent
                    ref="paramForm"
                    label-position="top"
                    :model="req"
                    label-width="150px"
                    :rules="rules"
                    :validate-on-rule-change="false"
                >
                    <el-form-item :label="$t('app.name')" prop="name">
                        <el-input v-model.trim="req.name"></el-input>
                    </el-form-item>
                    <Params
                        v-if="open"
                        v-model:form="req.params"
                        v-model:params="installData.params"
                        v-model:rules="rules.params"
                        :propStart="'params.'"
                    ></Params>
                    <el-form-item prop="advanced">
                        <el-checkbox v-model="req.advanced" :label="$t('app.advanced')" size="large" />
                    </el-form-item>
                    <div v-if="req.advanced">
                        <el-form-item :label="$t('app.containerName')" prop="containerName">
                            <el-input
                                v-model.trim="req.containerName"
                                :placeholder="$t('app.conatinerNameHelper')"
                            ></el-input>
                        </el-form-item>
                        <el-form-item :label="$t('container.cpuQuota')" prop="cpuQuota">
                            <el-input type="number" style="width: 40%" v-model.number="req.cpuQuota" maxlength="5">
                                <template #append>{{ $t('app.cpuCore') }}</template>
                            </el-input>
                            <span class="input-help">{{ $t('container.limitHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('container.memoryLimit')" prop="memoryLimit">
                            <el-input style="width: 40%" v-model.number="req.memoryLimit" maxlength="10">
                                <template #append>
                                    <el-select v-model="req.memoryUnit" placeholder="Select" style="width: 85px">
                                        <el-option label="KB" value="K" />
                                        <el-option label="MB" value="M" />
                                        <el-option label="GB" value="G" />
                                    </el-select>
                                </template>
                            </el-input>
                            <span class="input-help">{{ $t('container.limitHelper') }}</span>
                        </el-form-item>
                        <el-form-item prop="allowPort">
                            <el-checkbox v-model="req.allowPort" :label="$t('app.allowPort')" size="large" />
                            <span class="input-help">{{ $t('app.allowPortHelper') }}</span>
                        </el-form-item>
                    </div>
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
import { Rules, checkNumberRange } from '@/global/form-rules';
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
const open = ref(false);
const rules = ref<FormRules>({
    name: [Rules.appName],
    params: [],
    containerName: [Rules.containerName],
    cpuQuota: [checkNumberRange(0, 99999)],
    memoryLimit: [checkNumberRange(0, 9999999999)],
});
const loading = ref(false);
const paramForm = ref<FormInstance>();

const form = ref<{ [key: string]: any }>({});

const initData = () => ({
    appDetailId: 0,
    params: form.value,
    name: '',
    advanced: false,
    cpuQuota: 0,
    memoryLimit: 0,
    memoryUnit: 'MB',
    containerName: '',
    allowPort: false,
});

const req = reactive(initData());

const handleClose = () => {
    open.value = false;
    resetForm();
};

const resetForm = () => {
    if (paramForm.value) {
        paramForm.value.clearValidate();
        paramForm.value.resetFields();
    }
    Object.assign(req, initData());
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
