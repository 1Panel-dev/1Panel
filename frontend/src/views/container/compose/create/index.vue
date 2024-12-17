<template>
    <el-drawer
        v-model="drawerVisible"
        @close="handleClose"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
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
                                <el-option
                                    v-for="item in templateOptions"
                                    :key="item.id"
                                    :value="item.id"
                                    :label="item.name"
                                />
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
                            <div v-if="form.from === 'edit' || form.from === 'template'" style="width: 100%">
                                <el-radio-group v-model="mode" size="small">
                                    <el-radio-button label="edit">{{ $t('commons.button.edit') }}</el-radio-button>
                                    <el-radio-button label="log">{{ $t('commons.button.log') }}</el-radio-button>
                                </el-radio-group>
                                <codemirror
                                    @change="onEdit('')"
                                    v-if="mode === 'edit'"
                                    :autofocus="true"
                                    placeholder="#Define or paste the content of your docker-compose file here"
                                    :indent-with-tab="true"
                                    :tabSize="4"
                                    style="width: 100%; height: calc(100vh - 400px)"
                                    :lineWrapping="true"
                                    :matchBrackets="true"
                                    theme="cobalt"
                                    :styleActiveLine="true"
                                    :extensions="extensions"
                                    v-model="form.file"
                                />
                            </div>
                            <div style="width: 100%">
                                <LogFile
                                    ref="logRef"
                                    v-model:is-reading="isReading"
                                    :config="logConfig"
                                    :default-button="false"
                                    v-if="mode === 'log' && showLog"
                                    :style="'height: calc(100vh - 370px);min-height: 200px'"
                                />
                            </div>
                        </el-form-item>
                        <el-form-item :label="$t('container.env')" prop="envStr">
                            <el-input
                                type="textarea"
                                :placeholder="$t('container.tagHelper')"
                                :rows="3"
                                v-model="form.envStr"
                            />
                        </el-form-item>
                        <span class="input-help whitespace-break-spaces">{{ $t('container.editComposeHelper') }}</span>
                        <codemirror
                            v-model="form.envFileContent"
                            :autofocus="true"
                            :indent-with-tab="true"
                            :tabSize="4"
                            :lineWrapping="true"
                            :disabled="true"
                            :matchBrackets="true"
                            theme="cobalt"
                            :styleActiveLine="true"
                            :extensions="extensions"
                        ></codemirror>
                    </el-form>
                </el-col>
            </el-row>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" :disabled="isStartReading || isReading" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { nextTick, onBeforeUnmount, reactive, ref } from 'vue';
import FileList from '@/components/file-list/index.vue';
import { Codemirror } from 'vue-codemirror';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessageBox } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { listComposeTemplate, testCompose, upCompose } from '@/api/modules/container';
import { loadBaseDir } from '@/api/modules/setting';
import { MsgError } from '@/utils/message';
import { yaml } from '@codemirror/lang-yaml';
import { oneDark } from '@codemirror/theme-one-dark';

const extensions = [yaml(), oneDark];

const showLog = ref(false);
const loading = ref();
const mode = ref('edit');
const oldFrom = ref('edit');
const drawerVisible = ref(false);
const templateOptions = ref();
const baseDir = ref();
const composeFile = ref();
let timer: NodeJS.Timer | null = null;
const logRef = ref();
const isStartReading = ref(false);
const isReading = ref();

const logConfig = reactive({
    type: 'compose-create',
    name: '',
});

const form = reactive({
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
    mode.value = 'edit';
    drawerVisible.value = true;
    form.name = '';
    form.from = 'edit';
    form.path = '';
    form.file = '';
    form.template = null;
    form.envStr = '';
    form.env = [];
    loadTemplates();
    loadPath();
    isStartReading.value = false;
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
    if (!isReading.value && isStartReading.value) {
        isStartReading.value = false;
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
                    mode.value = 'log';
                    await upCompose(form)
                        .then((res) => {
                            logConfig.name = res.data;
                            loadLogs();
                            isStartReading.value = true;
                        })
                        .catch(() => {
                            loading.value = false;
                        });
                }
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const loadLogs = () => {
    showLog.value = false;
    nextTick(() => {
        showLog.value = true;
        nextTick(() => {
            logRef.value.changeTail(true);
        });
    });
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
