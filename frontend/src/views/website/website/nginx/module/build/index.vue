<template>
    <DrawerPro v-model="open" :header="$t('nginx.build')" size="normal" @close="handleClose">
        <el-form ref="buildForm" label-position="top" :model="build" :rules="rules">
            <el-form-item>
                <el-text type="info">{{ $t('nginx.buildPurposeHint') }}</el-text>
            </el-form-item>
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
                        value="http://mirrors.tuna.tsinghua.edu.cn/ubuntu/"
                        label="http://mirrors.tuna.tsinghua.edu.cn/ubuntu/"
                    ></el-option>
                    <el-option
                        value="http://mirrors.ustc.edu.cn/ubuntu/"
                        label="http://mirrors.ustc.edu.cn/ubuntu/"
                    ></el-option>
                </el-select>
            </el-form-item>
            <el-form-item v-if="moduleOptions.length > 0" :label="$t('nginx.modulesToBuild')" prop="modules">
                <el-checkbox-group v-model="build.modules">
                    <el-checkbox v-for="item in moduleOptions" :key="item.name" :value="item.name">
                        {{ item.name }}
                    </el-checkbox>
                </el-checkbox-group>
            </el-form-item>
            <el-form-item v-if="!dynamicSupported">
                <el-text type="warning">{{ $t('nginx.dynamicUnsupported') }}</el-text>
            </el-form-item>
            <el-form-item v-if="staticModules.length > 0" :label="$t('nginx.staticModules')">
                <div>
                    <div v-for="item in staticModules" :key="item.name" class="!mb-1">
                        <el-text>{{ item.name }}</el-text>
                    </div>
                    <el-text type="info" size="small">{{ $t('nginx.staticModulesHelper') }}</el-text>
                </div>
            </el-form-item>
            <el-form-item :label="$t('nginx.forceBuild')">
                <el-switch v-model="build.force" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button v-permission type="primary" @click="submit(buildForm)" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" />
</template>
<script setup lang="ts">
import { computed, ref } from 'vue';
import { ElMessageBox, FormInstance } from 'element-plus';
import { getNginxModules, buildNginx } from '@/api/modules/nginx';
import i18n from '@/lang';
import { newUUID } from '@/utils/id';
import TaskLog from '@/components/log/task/index.vue';
import { Rules } from '@/global/form-rules';
import { Nginx } from '@/api/interface/nginx';

const open = ref(false);
const loading = ref(false);
const buildForm = ref<FormInstance>();
const build = ref({
    mirror: 'http://archive.ubuntu.com/ubuntu',
    modules: [] as string[],
    force: false,
});
const moduleOptions = ref<Nginx.NginxModule[]>([]);
const staticModules = ref<Nginx.NginxModule[]>([]);
const dynamicSupported = ref(true);
const rules = computed(() => ({
    mirror: [Rules.requiredSelect],
    modules: staticModules.value.length === 0 ? [Rules.requiredSelect] : [],
}));
const taskLogRef = ref();

const acceptParams = async () => {
    getModules();
    open.value = true;
};

const getModules = async () => {
    try {
        const res = await getNginxModules();
        build.value.mirror = res.data.mirror;
        dynamicSupported.value = res.data.dynamicSupported;
        moduleOptions.value = res.data.dynamicSupported
            ? res.data.modules.filter((item) => item.enable && item.buildMode === 'dynamic')
            : [];
        staticModules.value = res.data.modules.filter((item) => item.enable && item.buildMode === 'static');
        build.value.modules = moduleOptions.value.map((item) => item.name);
    } catch (error) {}
};

const submit = async (form: FormInstance) => {
    await form.validate();
    await ElMessageBox.confirm(i18n.global.t('nginx.buildWarn'), i18n.global.t('nginx.build'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    });
    const taskID = newUUID();
    loading.value = true;
    try {
        await buildNginx({
            taskID: taskID,
            mirror: build.value.mirror,
            modules: build.value.modules,
            force: build.value.force,
        });
        handleClose();
        openTaskLog(taskID);
    } finally {
        loading.value = false;
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
