<template>
    <div>
        <DialogPro v-model="open" :title="$t('container.createByCommand')" @close="handleClose" size="w-70">
            <el-form
                @submit.prevent
                ref="formRef"
                :rules="rules"
                :model="form"
                label-position="top"
                v-loading="loading"
            >
                <el-form-item prop="command">
                    <CodemirrorPro
                        :lineWrapping="true"
                        v-model="form.command"
                        :height="300"
                        :minHeight="50"
                        mode="shell"
                        placeholder="e.g. docker run -p 80:80 --name my-nginx nginx"
                    />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="open = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </DialogPro>
        <TaskLog ref="taskLogRef" width="70%" />
    </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { ElForm } from 'element-plus';
import i18n from '@/lang';
import TaskLog from '@/components/log/task/index.vue';
import { createContainerByCommand } from '@/api/modules/container';
import { MsgSuccess } from '@/utils/message';
import { newUUID } from '@/utils/util';

const open = ref<boolean>(false);
const emit = defineEmits<{ (e: 'search'): void }>();
const loading = ref(false);
const form = reactive({
    command: '',
});
const taskLogRef = ref();

const acceptParams = (): void => {
    form.command = '';
    open.value = true;
};

const formRef = ref<FormInstance>();
type FormInstance = InstanceType<typeof ElForm>;

const verifyCommand = (rule: any, value: any, callback: any) => {
    if (!form.command || !form.command.startsWith('docker run')) {
        callback(new Error(i18n.global.t('container.commandRule')));
        return;
    }
    callback();
};
const rules = reactive({
    command: [{ validator: verifyCommand, trigger: 'blur', required: true }],
});

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('container.commandHelper'), i18n.global.t('container.createByCommand'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        }).then(async () => {
            loading.value = true;
            let taskID = newUUID();
            await createContainerByCommand(form.command, taskID)
                .then(() => {
                    loading.value = false;
                    emit('search');
                    openTaskLog(taskID);
                    open.value = false;
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
    open.value = false;
    emit('search');
};

defineExpose({
    acceptParams,
});
</script>
