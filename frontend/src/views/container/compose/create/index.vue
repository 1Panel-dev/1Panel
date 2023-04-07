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
                    <el-form ref="formRef" label-position="top" :model="form" :rules="rules" label-width="80px">
                        <el-form-item :label="$t('container.from')">
                            <el-radio-group v-model="form.from" @change="hasChecked = false">
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
                        <el-form-item v-if="form.from === 'edit' || form.from === 'template'" prop="name">
                            <el-input @input="changePath" v-model.trim="form.name">
                                <template #prepend>{{ $t('file.dir') }}</template>
                            </el-input>
                            <span class="input-help">{{ $t('container.composePathHelper', [composeFile]) }}</span>
                        </el-form-item>
                        <el-form-item v-if="form.from === 'template'" prop="template">
                            <el-select v-model="form.template" @change="hasChecked = false">
                                <el-option
                                    v-for="item in templateOptions"
                                    :key="item.id"
                                    :value="item.id"
                                    :label="item.name"
                                />
                            </el-select>
                        </el-form-item>
                        <el-form-item v-if="form.from === 'edit'">
                            <codemirror
                                :autofocus="true"
                                placeholder="#Define or paste the content of your docker-compose file here"
                                :indent-with-tab="true"
                                :tabSize="4"
                                style="width: 100%; height: 250px"
                                :lineWrapping="true"
                                :matchBrackets="true"
                                theme="cobalt"
                                @change="hasChecked = false"
                                :styleActiveLine="true"
                                :extensions="extensions"
                                v-model="form.file"
                            />
                        </el-form-item>
                    </el-form>
                    <codemirror
                        v-if="logVisiable && form.from !== 'edit'"
                        :autofocus="true"
                        placeholder="Waiting for docker-compose up output..."
                        :indent-with-tab="true"
                        :tabSize="4"
                        style="height: calc(100vh - 370px)"
                        :lineWrapping="true"
                        :matchBrackets="true"
                        theme="cobalt"
                        :styleActiveLine="true"
                        :extensions="extensions"
                        @ready="handleReady"
                        v-model="logInfo"
                        :readOnly="true"
                    />
                    <codemirror
                        v-if="logVisiable && form.from === 'edit'"
                        :autofocus="true"
                        placeholder="Waiting for docker-compose up output..."
                        :indent-with-tab="true"
                        :tabSize="4"
                        style="height: calc(100vh - 590px)"
                        :lineWrapping="true"
                        :matchBrackets="true"
                        theme="cobalt"
                        :styleActiveLine="true"
                        :extensions="extensions"
                        @ready="handleReady"
                        v-model="logInfo"
                        :readOnly="true"
                    />
                </el-col>
            </el-row>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="buttonDisabled" @click="onTest(formRef)">
                    {{ $t('commons.button.verify') }}
                </el-button>
                <el-button type="primary" :disabled="buttonDisabled || !hasChecked" @click="onSubmit(formRef)">
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
import { ElForm } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { listComposeTemplate, testCompose, upCompose } from '@/api/modules/container';
import { loadBaseDir } from '@/api/modules/setting';
import { LoadFile } from '@/api/modules/files';
import { formatImageStdout } from '@/utils/docker';
import { MsgSuccess } from '@/utils/message';

const loading = ref();

const extensions = [javascript(), oneDark];
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};
const logVisiable = ref();
const logInfo = ref();

const drawerVisiable = ref(false);
const templateOptions = ref();
const buttonDisabled = ref(false);

const baseDir = ref();
const composeFile = ref();

const hasChecked = ref();

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
    path: [Rules.requiredSelect],
});

const loadTemplates = async () => {
    const res = await listComposeTemplate();
    templateOptions.value = res.data;
    if (templateOptions.value && templateOptions.value.length !== 0) {
        form.template = templateOptions.value[0].id;
    }
};

const acceptParams = (): void => {
    drawerVisiable.value = true;
    form.name = '';
    form.from = 'edit';
    form.path = '';
    form.file = '';
    logVisiable.value = false;
    hasChecked.value = false;
    logInfo.value = '';
    loadTemplates();
    loadPath();
};
const emit = defineEmits<{ (e: 'search'): void }>();

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
    composeFile.value = baseDir.value + '/docker/compose/' + form.name + '/docker-compose.yml';
};

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onTest = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await testCompose(form)
            .then((res) => {
                loading.value = false;
                if (res.data) {
                    MsgSuccess(i18n.global.t('container.composeHelper'));
                    hasChecked.value = true;
                }
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        const res = await upCompose(form);
        logInfo.value = '';
        buttonDisabled.value = true;
        logVisiable.value = true;
        loadLogs(res.data);
    });
};

const loadLogs = async (path: string) => {
    timer = setInterval(async () => {
        if (logVisiable.value) {
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
                clearInterval(Number(timer));
                timer = null;
                buttonDisabled.value = false;
            }
        }
    }, 1000 * 3);
};

const loadDir = async (path: string) => {
    form.path = path;
    hasChecked.value = false;
};

onBeforeUnmount(() => {
    clearInterval(Number(timer));
    timer = null;
});

defineExpose({
    acceptParams,
});
</script>
