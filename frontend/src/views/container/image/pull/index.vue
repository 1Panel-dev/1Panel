<template>
    <DrawerPro v-model="drawerVisible" :header="$t('container.imagePull')" @close="onCloseLog" size="large">
        <el-form ref="formRef" label-position="top" :model="form">
            <el-form-item :label="$t('app.source')">
                <el-checkbox v-model="form.fromRepo">
                    {{ $t('container.imageRepo') }}
                </el-checkbox>
            </el-form-item>
            <el-form-item
                v-if="form.fromRepo"
                :label="$t('container.repoName')"
                :rules="Rules.requiredSelect"
                prop="repoID"
            >
                <el-select clearable style="width: 100%" filterable v-model="form.repoID">
                    <el-option v-for="item in repos" :key="item.id" :value="item.id" :label="item.name" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('container.imageName')" :rules="Rules.imageName" prop="imageName">
                <el-input v-model.trim="form.imageName">
                    <template v-if="form.fromRepo" #prepend>{{ loadDetailInfo(form.repoID) }}/</template>
                </el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('container.pull') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" width="70%" />
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { imagePull } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import TaskLog from '@/components/task-log/index.vue';
import { MsgSuccess } from '@/utils/message';
import { newUUID } from '@/utils/util';

const drawerVisible = ref(false);
const form = reactive({
    taskID: '',
    fromRepo: true,
    repoID: null as number,
    imageName: '',
});
const taskLogRef = ref();

interface DialogProps {
    repos: Array<Container.RepoOptions>;
}
const repos = ref();

const acceptParams = async (params: DialogProps): Promise<void> => {
    drawerVisible.value = true;
    form.fromRepo = true;
    form.imageName = '';
    repos.value = params.repos;
    form.repoID = 1;
    for (const item of repos.value) {
        if (item.name === 'Docker Hub' && item.downloadUrl === 'docker.io') {
            form.repoID = item.id;
            break;
        }
    }
};
const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!form.fromRepo) {
            form.repoID = 0;
        }
        form.taskID = newUUID();
        await imagePull(form);
        openTaskLog(form.taskID);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const onCloseLog = async () => {
    emit('search');
    drawerVisible.value = false;
};

function loadDetailInfo(id: number) {
    for (const item of repos.value) {
        if (item.id === id) {
            return item.downloadUrl;
        }
    }
    return '';
}

defineExpose({
    acceptParams,
});
</script>
