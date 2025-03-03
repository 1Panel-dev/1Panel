<template>
    <DrawerPro v-model="drawerVisible" :header="$t('container.imagePush')" @close="onCloseLog" size="large">
        <el-form ref="formRef" label-position="top" :model="form" label-width="80px">
            <el-form-item :label="$t('container.tag')" :rules="Rules.requiredSelect" prop="tagName">
                <el-select filterable v-model="form.tagName">
                    <el-option v-for="item in form.tags" :key="item" :value="item" :label="item" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('container.repoName')" :rules="Rules.requiredSelect" prop="repoID">
                <el-select clearable style="width: 100%" filterable v-model="form.repoID">
                    <el-option v-for="item in dialogData.repos" :key="item.id" :value="item.id" :label="item.name" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('container.image')" :rules="Rules.imageName" prop="name">
                <el-input v-model.trim="form.name">
                    <template #prepend>{{ loadDetailInfo(form.repoID) }}/</template>
                </el-input>
            </el-form-item>
        </el-form>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('container.push') }}
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
import { imagePush } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { MsgSuccess } from '@/utils/message';
import TaskLog from '@/components/task-log/index.vue';
import { newUUID } from '@/utils/util';

const drawerVisible = ref(false);
const taskLogRef = ref();
const form = reactive({
    taskID: '',
    tags: [] as Array<string>,
    tagName: '',
    repoID: 1,
    name: '',
});

interface DialogProps {
    repos: Array<Container.RepoOptions>;
    tags: Array<string>;
}
const dialogData = ref<DialogProps>({
    repos: [] as Array<Container.RepoOptions>,
    tags: [] as Array<string>,
});

const acceptParams = async (params: DialogProps): Promise<void> => {
    drawerVisible.value = true;
    form.tags = params.tags;
    form.repoID = 1;
    form.tagName = form.tags.length !== 0 ? form.tags[0] : '';
    form.name = form.tags.length !== 0 ? form.tags[0] : '';
    dialogData.value.repos = params.repos;
};
const emit = defineEmits<{ (e: 'search'): void }>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        form.taskID = newUUID();
        await imagePush(form);
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
    for (const item of dialogData.value.repos) {
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
