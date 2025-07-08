<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="$t('container.makeImage')"
        @close="handleClose"
        :resource="form.containerName"
        size="large"
    >
        <el-form @submit.prevent ref="formRef" :model="form" label-position="top" v-loading="loading">
            <el-form-item prop="newImageName" :rules="Rules.imageName">
                <template #label>
                    {{ $t('container.newImageName') }}
                </template>
                <el-input v-model="form.newImageName" />
            </el-form-item>
            <el-form-item prop="comment">
                <template #label>
                    {{ $t('container.commitMessage') }}
                </template>
                <el-input v-model="form.comment" />
            </el-form-item>
            <el-form-item prop="author">
                <template #label>
                    {{ $t('container.author') }}
                </template>
                <el-input v-model="form.author" />
            </el-form-item>
            <el-form-item prop="pause">
                <el-checkbox v-model="form.pause">
                    {{ $t('container.ifPause') }}
                </el-checkbox>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" width="70%" />
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { commitContainer } from '@/api/modules/container';
import TaskLog from '@/components/log/task/index.vue';
import { MsgSuccess } from '@/utils/message';
import { newUUID } from '@/utils/util';

const drawerVisible = ref<boolean>(false);
const emit = defineEmits<{ (e: 'search'): void }>();
const loading = ref(false);
const form = reactive({
    containerID: '',
    containerName: '',
    newImageName: '',
    comment: '',
    author: '',
    pause: false,
    taskID: '',
});

const taskLogRef = ref();

interface DialogProps {
    containerID: string;
    containerName: string;
}
const acceptParams = (props: DialogProps): void => {
    form.newImageName = '';
    form.comment = '';
    form.author = '';
    form.containerID = props.containerID;
    form.containerName = props.containerName;
    drawerVisible.value = true;
};

const formRef = ref<FormInstance>();
type FormInstance = InstanceType<typeof ElForm>;

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('container.ifMakeImageWithContainer'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(async () => {
            loading.value = true;
            form.taskID = newUUID();
            await commitContainer(form)
                .then(() => {
                    loading.value = false;
                    emit('search');
                    drawerVisible.value = false;
                    openTaskLog(form.taskID);
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        });
    });
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const handleClose = async () => {
    drawerVisible.value = false;
    emit('search');
};

defineExpose({
    acceptParams,
});
</script>
