<template>
    <DrawerPro v-model="loadVisible" :header="$t('container.importImage')" @close="handleClose" size="small">
        <el-form @submit.prevent v-loading="loading" ref="formRef" :model="form" label-position="top">
            <el-form-item :label="$t('container.path')" :rules="Rules.requiredInput" prop="path">
                <el-input v-model="form.path">
                    <template #prepend>
                        <el-button icon="Folder" @click="fileRef.acceptParams({ dir: false })" />
                    </template>
                </el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="loadVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.import') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <FileList ref="fileRef" @choose="loadLoadDir" />
    <TaskLog ref="taskLogRef" width="70%" />
</template>

<script lang="ts" setup>
import FileList from '@/components/file-list/index.vue';
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import TaskLog from '@/components/log/task/index.vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { imageLoad } from '@/api/modules/container';
import { MsgSuccess } from '@/utils/message';
import { newUUID } from '@/utils/util';

const loading = ref(false);
const fileRef = ref();
const taskLogRef = ref();

const loadVisible = ref(false);
const form = reactive({
    path: '',
    taskID: '',
});

const acceptParams = () => {
    loadVisible.value = true;
    form.path = '';
};
const handleClose = () => {
    loadVisible.value = false;
};

const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        form.taskID = newUUID();
        await imageLoad(form)
            .then(() => {
                loading.value = false;
                loadVisible.value = false;
                openTaskLog(form.taskID);
                emit('search');
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const loadLoadDir = async (path: string) => {
    form.path = path;
};

defineExpose({
    acceptParams,
});
</script>
