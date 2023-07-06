<template>
    <el-drawer
        v-model="drawerVisiable"
        @close="handleClose"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.compose')" :back="handleClose" />
        </template>
        <div v-loading="loading">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form ref="formRef" @submit.prevent label-position="top" :model="form" :rules="rules">
                        <el-form-item :label="$t('container.from')">
                            <el-radio-group v-model="form.from" @change="changeFrom">
                                <el-radio label="edit">{{ $t('commons.button.edit') }}</el-radio>
                                <el-radio label="path">{{ $t('container.pathSelect') }}</el-radio>
                                <el-radio label="template">{{ $t('container.composeTemplate') }}</el-radio>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item v-if="form.from === 'path'" prop="path">
                            <el-input
                                :placeholder="$t('commons.example') + '/tmp/docker-compose.yml'"
                                v-model="form.path"
                            >
                                <template #prepend>
                                    <FileList @choose="loadDir" :dir="false"></FileList>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-row :gutter="20">
                            <el-col :span="12">
                                <el-form-item v-if="form.from === 'edit' || form.from === 'template'" prop="name">
                                    <el-input @input="changePath" v-model.trim="form.name">
                                        <template #prefix>
                                            <span style="margin-right: 8px">{{ $t('file.dir') }}</span>
                                        </template>
                                    </el-input>
                                    <span class="input-help">
                                        {{ $t('container.composePathHelper', [composeFile]) }}
                                    </span>
                                </el-form-item>
                            </el-col>
                            <el-col :span="12">
                                <el-form-item v-if="form.from === 'template'" prop="template">
                                    <el-select v-model="form.template" @change="changeTemplate">
                                        <template #prefix>{{ $t('container.template') }}</template>
                                        <el-option
                                            v-for="item in templateOptions"
                                            :key="item.id"
                                            :value="item.id"
                                            :label="item.name"
                                        />
                                    </el-select>
                                </el-form-item>
                            </el-col>
                        </el-row>
                        <el-form-item>
                            <div v-if="form.from === 'edit' || form.from === 'template'" style="width: 100%">
                                <el-radio-group v-model="mode" size="small">
                                    <el-radio-button label="edit">{{ $t('commons.button.edit') }}</el-radio-button>
                                    <el-radio-button label="log">{{ $t('commons.button.log') }}</el-radio-button>
                                </el-radio-group>
                                <codemirror
                                    v-if="mode === 'edit'"
                                    :autofocus="true"
                                    placeholder="#Define or paste the content of your docker-compose file here"
                                    :indent-with-tab="true"
                                    :tabSize="4"
                                    style="width: 100%; height: calc(100vh - 375px)"
                                    :lineWrapping="true"
                                    :matchBrackets="true"
                                    theme="cobalt"
                                    :styleActiveLine="true"
                                    :extensions="extensions"
                                    v-model="form.file"
                                />
                            </div>
                            <codemirror
                                v-if="mode === 'log'"
                                :autofocus="true"
                                placeholder="Waiting for docker-compose up output..."
                                :indent-with-tab="true"
                                :tabSize="4"
                                style="width: 100%; height: calc(100vh - 375px)"
                                :lineWrapping="true"
                                :matchBrackets="true"
                                theme="cobalt"
                                :styleActiveLine="true"
                                :extensions="extensions"
                                @ready="handleReady"
                                v-model="logInfo"
                                :disabled="true"
                            />
                        </el-form-item>
                    </el-form>
                </el-col>
            </el-row>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" :disabled="onCreating" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { nextTick, onBeforeUnmount, reactive, ref, shallowRef } from 'vue';
import FileList from '@/components/file-list/index.vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessageBox } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { listComposeTemplate, testCompose, upCompose } from '@/api/modules/container';
import { loadBaseDir } from '@/api/modules/setting';
import { LoadFile } from '@/api/modules/files';
import { formatImageStdout } from '@/utils/docker';
import { MsgError } from '@/utils/message';

const loading = ref();

const mode = ref('edit');
const onCreating = ref();
const oldFrom = ref('edit');

const extensions = [javascript(), oneDark];
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};
const logInfo = ref();

const drawerVisiable = ref(false);
const templateOptions = ref();

const baseDir = ref();
const composeFile = ref();

let timer: NodeJS.Timer | null = null;

const form = reactive({
    name: '',
    from: 'edit',
    path: '',
    file: '',
    template: null as number,
});
const rules = reactive({
    name: [Rules.requiredInput, Rules.imageName],
    path: [Rules.requiredInput],
    template: [Rules.requiredSelect],
});

const loadTemplates = async () => {
    const res = await listComposeTemplate();
    templateOptions.value = res.data;
};

const acceptParams = (): void => {
    mode.value = 'edit';
    drawerVisiable.value = true;
    form.name = '';
    form.from = 'edit';
    form.path = '';
    form.file = '';
    form.template = null;
    logInfo.value = '';
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
        ElMessageBox.confirm(i18n.global.t('container.fromChangeHelper'), i18n.global.t('container.from'), {
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
    clearInterval(Number(timer));
    timer = null;
    drawerVisiable.value = false;
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

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if ((form.from === 'edit' || form.from === 'template') && form.file.length === 0) {
            MsgError(i18n.global.t('container.contentEmpty'));
            return;
        }
        loading.value = true;
        logInfo.value = '';
        await testCompose(form)
            .then(async (res) => {
                loading.value = false;
                if (res.data) {
                    onCreating.value = true;
                    mode.value = 'log';
                    await upCompose(form)
                        .then((res) => {
                            logInfo.value = '';
                            loadLogs(res.data);
                        })
                        .catch(() => {
                            loading.value = false;
                            onCreating.value = false;
                        });
                }
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const loadLogs = async (path: string) => {
    timer = setInterval(async () => {
        const res = await LoadFile({ path: path });
        logInfo.value = formatImageStdout(res.data);
        nextTick(() => {
            const state = view.value.state;
            view.value.dispatch({
                selection: { anchor: state.doc.length, head: state.doc.length },
                scrollIntoView: true,
            });
        });
        if (
            logInfo.value.endsWith('docker-compose up failed!') ||
            logInfo.value.endsWith('docker-compose up successful!')
        ) {
            onCreating.value = false;
            clearInterval(Number(timer));
            timer = null;
        }
    }, 1000 * 3);
};

const loadDir = async (path: string) => {
    form.path = path;
};

onBeforeUnmount(() => {
    clearInterval(Number(timer));
    timer = null;
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss"></style>
