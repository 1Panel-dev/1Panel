<template>
    <el-drawer
        v-model="drawerVisiable"
        :destroy-on-close="true"
        @close="handleClose"
        :close-on-click-modal="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.imageBuild')" :back="handleClose" />
        </template>
        <el-row type="flex" justify="center">
            <el-col :span="22">
                <el-form ref="formRef" label-position="top" :model="form" label-width="80px" :rules="rules">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input :placeholder="$t('container.imageNameHelper')" v-model.trim="form.name" clearable />
                    </el-form-item>
                    <el-form-item label="Dockerfile" prop="from">
                        <el-radio-group v-model="form.from">
                            <el-radio label="edit">{{ $t('commons.button.edit') }}</el-radio>
                            <el-radio label="path">{{ $t('container.pathSelect') }}</el-radio>
                        </el-radio-group>
                    </el-form-item>
                    <el-form-item v-if="form.from === 'edit'" :rules="Rules.requiredInput">
                        <codemirror
                            :autofocus="true"
                            placeholder="#Define or paste the content of your Dockerfile here"
                            :indent-with-tab="true"
                            :tabSize="4"
                            style="width: 100%; height: calc(100vh - 520px)"
                            :lineWrapping="true"
                            :matchBrackets="true"
                            theme="cobalt"
                            :styleActiveLine="true"
                            :extensions="extensions"
                            v-model="form.dockerfile"
                            :readOnly="true"
                        />
                    </el-form-item>
                    <el-form-item v-else :rules="Rules.requiredSelect" prop="dockerfile">
                        <el-input clearable v-model="form.dockerfile">
                            <template #prepend>
                                <FileList @choose="loadBuildDir"></FileList>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('container.tag')">
                        <el-input
                            :placeholder="$t('container.tagHelper')"
                            type="textarea"
                            :rows="3"
                            v-model="form.tagStr"
                        />
                    </el-form-item>
                </el-form>

                <LogFile
                    ref="logRef"
                    :config="logConfig"
                    :default-button="false"
                    v-if="logVisiable"
                    :style="'height: calc(100vh - 370px);min-height: 200px'"
                />
            </el-col>
        </el-row>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="buttonDisabled" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import FileList from '@/components/file-list/index.vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { nextTick, reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { imageBuild } from '@/api/modules/container';
import DrawerHeader from '@/components/drawer-header/index.vue';

const logVisiable = ref<boolean>(false);
const extensions = [javascript(), oneDark];
const buttonDisabled = ref(false);
const drawerVisiable = ref(false);
const logRef = ref();

const logConfig = reactive({
    type: 'image-build',
    name: '',
});
const form = reactive({
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
    logVisiable.value = false;
    drawerVisiable.value = true;
    form.from = 'path';
    form.dockerfile = '';
    form.tagStr = '';
    form.name = '';
    buttonDisabled.value = false;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisiable.value = false;
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
        const res = await imageBuild(form);
        buttonDisabled.value = true;
        logConfig.name = res.data;
        loadLogs();
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const loadLogs = () => {
    logVisiable.value = false;
    nextTick(() => {
        logVisiable.value = true;
        nextTick(() => {
            logRef.value.changeTail(true);
        });
    });
};

const loadBuildDir = async (path: string) => {
    form.dockerfile = path;
};

defineExpose({
    acceptParams,
});
</script>
