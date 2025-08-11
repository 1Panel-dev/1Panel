<template>
    <DrawerPro v-model="open" :header="$t('commons.button.install')" @close="handleClose" size="large">
        <AppInstallForm ref="installFormRef" v-model="formData" :loading="loading" />
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="handleSubmit" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" />
</template>

<script lang="ts" setup name="AppInstallPage">
import { ref, reactive } from 'vue';
import { useRouter } from 'vue-router';
import AppInstallForm from '@/views/app-store/detail/form/index.vue';
import { installApp } from '@/api/modules/app';
import { MsgError } from '@/utils/message';
import { newUUID } from '@/utils/util';
import { routerToName } from '@/utils/router';
import TaskLog from '@/components/log/task/index.vue';
import i18n from '@/lang';

const router = useRouter();
const open = ref(false);
const loading = ref(false);
const installFormRef = ref<InstanceType<typeof AppInstallForm>>();
const taskLogRef = ref();

const formData = reactive({
    appDetailId: 0,
    params: {},
    name: '',
    advanced: true,
    cpuQuota: 0,
    memoryLimit: 0,
    memoryUnit: 'M',
    containerName: '',
    allowPort: false,
    editCompose: false,
    dockerCompose: '',
    version: '',
    appID: '',
    pullImage: true,
    taskID: '',
    gpuConfig: false,
    specifyIP: '',
});

const handleClose = () => {
    open.value = false;
    installFormRef.value?.resetForm();
    if (router.currentRoute.value.query.install) {
        routerToName('AppAll');
    }
};

const handleSubmit = async () => {
    const isValid = await installFormRef.value?.validate();
    if (!isValid) return;

    const submitData = installFormRef.value?.getFormData();

    if (submitData.editCompose && submitData.dockerCompose === '') {
        MsgError(i18n.global.t('app.composeNullErr'));
        return;
    }

    if (submitData.cpuQuota < 0) {
        submitData.cpuQuota = 0;
    }
    if (submitData.memoryLimit < 0) {
        submitData.memoryLimit = 0;
    }

    const isHostMode = installFormRef.value?.isHostMode();
    if (!isHostMode && !submitData.allowPort) {
        ElMessageBox.confirm(i18n.global.t('app.installWarn'), i18n.global.t('app.checkTitle'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(async () => {
            await install(submitData);
        });
    } else {
        await install(submitData);
    }
};

const install = async (submitData: any) => {
    loading.value = true;
    const taskID = newUUID();
    submitData.taskID = taskID;

    try {
        await installApp(submitData);
        handleClose();
        openTaskLog(taskID);
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const acceptParams = async (props: { app: any; params?: any }) => {
    open.value = true;
    await nextTick();
    installFormRef.value?.resetForm();
    installFormRef.value?.initForm(props.app.key);
};

defineExpose({
    acceptParams,
});
</script>
