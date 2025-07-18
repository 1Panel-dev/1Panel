<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="$t('container.imageBuild')"
        @close="handleClose"
        size="large"
        :autoClose="false"
    >
        <el-form ref="formRef" label-position="top" :model="form" label-width="80px" :rules="rules">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input :placeholder="$t('container.imageNameHelper')" v-model.trim="form.name" clearable />
            </el-form-item>
            <el-form-item label="Dockerfile" prop="from">
                <el-radio-group v-model="form.from">
                    <el-radio value="edit">{{ $t('commons.button.edit') }}</el-radio>
                    <el-radio value="path">{{ $t('container.pathSelect') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="form.from === 'edit'" :rules="Rules.requiredInput">
                <CodemirrorPro
                    v-model="form.dockerfile"
                    mode="dockerfile"
                    placeholder="#Define or paste the content of your Dockerfile here"
                ></CodemirrorPro>
            </el-form-item>
            <el-form-item v-else :rules="Rules.requiredSelect" prop="dockerfile">
                <el-input clearable v-model="form.dockerfile">
                    <template #prepend>
                        <FileList @choose="loadBuildDir"></FileList>
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item :label="$t('container.tag')">
                <el-input :placeholder="$t('container.tagHelper')" type="textarea" :rows="3" v-model="form.tagStr" />
            </el-form-item>
        </el-form>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" width="70%" />
</template>

<script lang="ts" setup>
import FileList from '@/components/file-list/index.vue';
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { imageBuild } from '@/api/modules/container';
import TaskLog from '@/components/log/task/index.vue';
import { newUUID } from '@/utils/util';
import { MsgSuccess } from '@/utils/message';

const drawerVisible = ref(false);
const taskLogRef = ref();

const form = reactive({
    taskID: '',
    from: 'path',
    dockerfile: '',
    name: '',
    tagStr: '',
    tags: [] as Array<string>,
});

const rules = reactive({
    name: [Rules.requiredInput, Rules.imageName],
    from: [Rules.requiredSelect],
    dockerfile: [Rules.requiredInput],
});
const acceptParams = async () => {
    drawerVisible.value = true;
    form.from = 'path';
    form.dockerfile = '';
    form.tagStr = '';
    form.name = '';
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
    emit('search');
};

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (form.tagStr !== '') {
            form.tags = form.tagStr.split('\n');
        }
        form.taskID = newUUID();
        await imageBuild(form);
        openTaskLog(form.taskID);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const loadBuildDir = async (path: string) => {
    form.dockerfile = path;
};

defineExpose({
    acceptParams,
});
</script>
