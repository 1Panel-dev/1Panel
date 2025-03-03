<template>
    <DrawerPro v-model="open" :header="$t('nginx.build')" size="normal" @close="handleClose">
        <el-form ref="buildForm" label-position="top" :model="build" :rules="rules">
            <el-form-item :label="$t('nginx.mirrorUrl')" prop="mirror">
                <el-select v-model="build.mirror">
                    <el-option
                        value="http://archive.ubuntu.com/ubuntu/"
                        label="http://archive.ubuntu.com/ubuntu/"
                    ></el-option>
                    <el-option
                        value="http://mirrors.aliyun.com/ubuntu/"
                        label="http://mirrors.aliyun.com/ubuntu/"
                    ></el-option>
                    <el-option
                        value="https://mirrors.tuna.tsinghua.edu.cn/ubuntu/"
                        label="https://mirrors.tuna.tsinghua.edu.cn/ubuntu/"
                    ></el-option>
                    <el-option
                        value="https://mirrors.ustc.edu.cn/ubuntu/"
                        label="https://mirrors.ustc.edu.cn/ubuntu/"
                    ></el-option>
                </el-select>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="submit(buildForm)" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" />
</template>
<script setup lang="ts">
import { ref } from 'vue';
import { FormInstance } from 'element-plus';
import { getNginxModules, buildNginx } from '@/api/modules/nginx';
import i18n from '@/lang';
import { newUUID } from '@/utils/util';
import TaskLog from '@/components/log/task/index.vue';
import { Rules } from '@/global/form-rules';

const open = ref(false);
const loading = ref(false);
const buildForm = ref<FormInstance>();
const build = ref({
    mirror: 'http://archive.ubuntu.com/ubuntu',
});
const rules = {
    mirror: [Rules.requiredSelect],
};
const taskLogRef = ref();

const acceptParams = async () => {
    getModules();
    open.value = true;
};

const getModules = async () => {
    try {
        const res = await getNginxModules();
        build.value.mirror = res.data.mirror;
    } catch (error) {}
};

const submit = async (form: FormInstance) => {
    await form.validate();
    if (form.validate()) {
        ElMessageBox.confirm(i18n.global.t('nginx.buildWarn'), i18n.global.t('nginx.build'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(async () => {
            const taskID = newUUID();
            try {
                await buildNginx({
                    taskID: taskID,
                    mirror: build.value.mirror,
                });
                handleClose();
                openTaskLog(taskID);
            } catch (error) {}
        });
    }
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const handleClose = () => {
    open.value = false;
};

defineExpose({
    acceptParams,
});
</script>
