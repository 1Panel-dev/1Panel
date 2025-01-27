<template>
    <DrawerPro v-model="drawerVisible" :header="$t('container.compose')" :back="handleClose" size="large">
        <el-form ref="formRef" @submit.prevent label-position="top" :model="form" :rules="rules" v-loading="loading">
            <el-form-item :label="$t('app.source')">
                <el-radio-group v-model="form.from" @change="onEdit('form')">
                    <el-radio value="edit">{{ $t('commons.button.edit') }}</el-radio>
                    <el-radio value="path">{{ $t('container.pathSelect') }}</el-radio>
                    <el-radio value="template">{{ $t('container.composeTemplate') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="form.from === 'path'" prop="path">
                <el-input
                    @change="onEdit('')"
                    :placeholder="$t('commons.example') + '/tmp/docker-compose.yml'"
                    v-model="form.path"
                >
                    <template #prepend>
                        <FileList @choose="loadDir" :dir="false"></FileList>
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item v-if="form.from === 'template'" prop="template">
                <el-select v-model="form.template" @change="onEdit('template')">
                    <template #prefix>{{ $t('container.template') }}</template>
                    <el-option v-for="item in templateOptions" :key="item.id" :value="item.id" :label="item.name" />
                </el-select>
            </el-form-item>
            <el-form-item v-if="form.from === 'edit' || form.from === 'template'" prop="name">
                <el-input @input="changePath" @change="onEdit('')" v-model.trim="form.name">
                    <template #prefix>
                        <span style="margin-right: 8px">{{ $t('file.dir') }}</span>
                    </template>
                </el-input>
                <span class="input-help">
                    {{ $t('container.composePathHelper', [composeFile]) }}
                </span>
            </el-form-item>
            <el-form-item>
                <div v-if="form.from === 'edit' || form.from === 'template'" class="w-full">
                    <CodemirrorPro
                        v-model="form.file"
                        placeholder="#Define or paste the content of your docker-compose file here"
                        mode="yaml"
                        :heightDiff="400"
                    ></CodemirrorPro>
                </div>
            </el-form-item>
            <el-form-item :label="$t('container.env')" prop="envStr">
                <el-input type="textarea" :placeholder="$t('container.tagHelper')" :rows="3" v-model="form.envStr" />
            </el-form-item>
            <span class="input-help">{{ $t('container.editComposeHelper') }}</span>
            <CodemirrorPro v-model="form.envFileContent" :height="45" :minHeight="45" disabled mode="yaml" />
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" width="70%" />
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import FileList from '@/components/file-list/index.vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessageBox } from 'element-plus';
import { loadBaseDir } from '@/api/modules/setting';
import { MsgError, MsgSuccess } from '@/utils/message';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import TaskLog from '@/components/task-log/index.vue';
import { listComposeTemplate, testCompose, upCompose } from '@/api/modules/container';
import { newUUID } from '@/utils/util';

const loading = ref();
const oldFrom = ref('edit');
const drawerVisible = ref(false);
const templateOptions = ref();
const baseDir = ref();
const composeFile = ref();
const taskLogRef = ref();

const form = reactive({
    taskID: '',
    name: '',
    from: 'edit',
    path: '',
    file: '',
    template: null as number,
    env: [],
    envStr: '',
    envFileContent: `env_file:\n  - 1panel.env`,
});
const rules = reactive({
    name: [Rules.requiredInput, Rules.composeName],
    path: [Rules.requiredInput],
    template: [Rules.requiredSelect],
});

const loadTemplates = async () => {
    const res = await listComposeTemplate();
    templateOptions.value = res.data;
};

const acceptParams = (): void => {
    drawerVisible.value = true;
    form.name = '';
    form.from = 'edit';
    form.path = '';
    form.file = '';
    form.template = null;
    form.env = [];
    form.envStr = '';
    loadTemplates();
    loadPath();
};
const emit = defineEmits<{ (e: 'search'): void }>();

const changeTemplate = () => {
    for (const item of templateOptions.value) {
        if (form.template === item.id) {
            form.file = item.content;
            break;
        }
    }
};

const changeFrom = () => {
    if ((oldFrom.value === 'edit' || oldFrom.value === 'template') && form.file) {
        ElMessageBox.confirm(i18n.global.t('container.fromChangeHelper'), i18n.global.t('app.source'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        })
            .then(() => {
                if (oldFrom.value === 'template') {
                    form.template = null;
                    form.file = '';
                }
                if (oldFrom.value === 'edit') {
                    form.file = '';
                }
                oldFrom.value = form.from;
            })
            .catch(() => {
                form.from = oldFrom.value;
            });
    } else {
        oldFrom.value = form.from;
    }
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

const loadPath = async () => {
    const pathRes = await loadBaseDir();
    baseDir.value = pathRes.data;
    changePath();
};

const changePath = async () => {
    composeFile.value = baseDir.value + '/docker/compose/' + form.name;
};

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onEdit = (item: string) => {
    if (item === 'template') {
        changeTemplate();
    }
    if (item === 'form') {
        changeFrom();
    }
};
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if ((form.from === 'edit' || form.from === 'template') && form.file.length === 0) {
            MsgError(i18n.global.t('container.contentEmpty'));
            return;
        }
        if (form.envStr) {
            form.env = form.envStr.split('\n');
        }
        loading.value = true;
        await testCompose(form)
            .then(async (res) => {
                loading.value = false;
                if (res.data) {
                    form.taskID = newUUID();
                    await upCompose(form);
                    openTaskLog(form.taskID);
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                }
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const loadDir = async (path: string) => {
    form.path = path;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss"></style>
