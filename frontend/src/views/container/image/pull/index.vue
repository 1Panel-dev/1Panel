<template>
    <el-drawer
        v-model="drawerVisiable"
        @close="onCloseLog"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.imagePull')" :back="onCloseLog" />
        </template>
        <el-row type="flex" justify="center">
            <el-col :span="22">
                <el-form ref="formRef" label-position="top" :model="form">
                    <el-form-item :label="$t('container.from')">
                        <el-checkbox v-model="form.fromRepo">{{ $t('container.imageRepo') }}</el-checkbox>
                    </el-form-item>
                    <el-form-item
                        v-if="form.fromRepo"
                        :label="$t('container.repoName')"
                        :rules="Rules.requiredSelect"
                        prop="repoID"
                    >
                        <el-select style="width: 100%" filterable v-model="form.repoID">
                            <el-option v-for="item in repos" :key="item.id" :value="item.id" :label="item.name" />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('container.imageName')" :rules="Rules.imageName" prop="imageName">
                        <el-input v-model.trim="form.imageName">
                            <template v-if="form.fromRepo" #prepend>{{ loadDetailInfo(form.repoID) }}/</template>
                        </el-input>
                    </el-form-item>
                </el-form>

                <codemirror
                    v-if="logVisiable"
                    :autofocus="true"
                    placeholder="Waiting for pull output..."
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
                    {{ $t('container.pull') }}
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
import { imagePull } from '@/api/modules/container';
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
    fromRepo: true,
    repoID: null as number,
    imageName: '',
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
}
const repos = ref();

const acceptParams = async (params: DialogProps): Promise<void> => {
    logVisiable.value = false;
    drawerVisiable.value = true;
    form.fromRepo = true;
    form.imageName = '';
    repos.value = params.repos;
    buttonDisabled.value = false;
    logInfo.value = '';
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
        const res = await imagePull(form);
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
            if (logInfo.value.endsWith('image pull failed!') || logInfo.value.endsWith('image pull successful!')) {
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
    for (const item of repos.value) {
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
