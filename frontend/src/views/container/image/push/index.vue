<template>
    <el-drawer
        v-model="drawerVisiable"
        :destroy-on-close="true"
        @close="onCloseLog"
        :close-on-click-modal="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.imagePush')" :back="onCloseLog" />
        </template>
        <el-row type="flex" justify="center">
            <el-col :span="22">
                <el-form ref="formRef" label-position="top" :model="form" label-width="80px">
                    <el-form-item :label="$t('container.tag')" :rules="Rules.requiredSelect" prop="tagName">
                        <el-select filterable v-model="form.tagName" @change="form.name = form.tagName">
                            <el-option v-for="item in form.tags" :key="item" :value="item" :label="item" />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('container.repoName')" :rules="Rules.requiredSelect" prop="repoID">
                        <el-select style="width: 100%" filterable v-model="form.repoID">
                            <el-option
                                v-for="item in dialogData.repos"
                                :key="item.id"
                                :value="item.id"
                                :label="item.name"
                            />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('container.image')" :rules="Rules.imageName" prop="name">
                        <el-input v-model.trim="form.name">
                            <template #prepend>{{ loadDetailInfo(form.repoID) }}/</template>
                        </el-input>
                    </el-form-item>
                </el-form>

                <codemirror
                    v-if="logVisiable"
                    :autofocus="true"
                    placeholder="Waiting for push output..."
                    :indent-with-tab="true"
                    :tabSize="4"
                    style="height: calc(100vh - 415px)"
                    :lineWrapping="true"
                    :matchBrackets="true"
                    theme="cobalt"
                    :styleActiveLine="true"
                    :extensions="extensions"
                    @ready="handleReady"
                    v-model="logInfo"
                    :disabled="true"
                />
            </el-col>
        </el-row>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="buttonDisabled" type="primary" @click="onSubmit(formRef)">
                    {{ $t('container.push') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { nextTick, onBeforeUnmount, reactive, ref, shallowRef } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { imagePush } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { LoadFile } from '@/api/modules/files';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { formatImageStdout } from '@/utils/docker';
import { MsgSuccess } from '@/utils/message';

const drawerVisiable = ref(false);
const form = reactive({
    tags: [] as Array<string>,
    tagName: '',
    repoID: 1,
    name: '',
});

const buttonDisabled = ref(false);

const logVisiable = ref(false);
const logInfo = ref();
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};
const extensions = [javascript(), oneDark];
let timer: NodeJS.Timer | null = null;

interface DialogProps {
    repos: Array<Container.RepoOptions>;
    tags: Array<string>;
}
const dialogData = ref<DialogProps>({
    repos: [] as Array<Container.RepoOptions>,
    tags: [] as Array<string>,
});

const acceptParams = async (params: DialogProps): Promise<void> => {
    logVisiable.value = false;
    drawerVisiable.value = true;
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
        const res = await imagePush(form);
        logVisiable.value = true;
        buttonDisabled.value = true;
        loadLogs(res.data);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
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
            if (logInfo.value.endsWith('image push failed!') || logInfo.value.endsWith('image push successful!')) {
                clearInterval(Number(timer));
                timer = null;
                buttonDisabled.value = false;
            }
        }
    }, 1000 * 3);
};
const onCloseLog = async () => {
    emit('search');
    clearInterval(Number(timer));
    timer = null;
    drawerVisiable.value = false;
};

function loadDetailInfo(id: number) {
    for (const item of dialogData.value.repos) {
        if (item.id === id) {
            return item.downloadUrl;
        }
    }
    return '';
}

onBeforeUnmount(() => {
    clearInterval(Number(timer));
    timer = null;
});

defineExpose({
    acceptParams,
});
</script>
