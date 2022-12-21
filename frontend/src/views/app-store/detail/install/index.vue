<template>
    <el-dialog
        v-model="open"
        :title="$t('app.install')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        width="40%"
        :before-close="handleClose"
        @opened="opened"
    >
        <el-form ref="paramForm" label-position="left" :model="form" label-width="150px" :rules="rules">
            <el-form-item :label="$t('app.name')" prop="NAME">
                <el-input v-model="form['NAME']"></el-input>
            </el-form-item>
            <Params v-model:form="form" v-model:params="installData.params" v-model:rules="rules"></Params>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(paramForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup name="appInstall">
import { App } from '@/api/interface/app';
import { InstallApp } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import { FormInstance, FormRules } from 'element-plus';
import { nextTick, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import Params from '../params/index.vue';
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
    NAME: [Rules.simpleName],
});
let loading = false;
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

const opened = () => {
    nextTick(() => {
        if (paramForm.value) {
            paramForm.value.clearValidate();
        }
    });
};

const resetForm = () => {
    if (paramForm.value) {
        paramForm.value.resetFields();
    }
};

const acceptParams = (props: InstallRrops): void => {
    installData.value = props;
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
        InstallApp(req).then(() => {
            handleClose();
            router.push({ path: '/apps/installed' });
        });
    });
};

defineExpose({
    acceptParams,
});
</script>
