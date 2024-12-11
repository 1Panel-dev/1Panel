<template>
    <el-drawer
        v-model="composeVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('commons.button.edit')" :resource="name" :back="handleClose" />
        </template>
        <div v-loading="loading" style="padding-bottom: 20px">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form ref="formRef" @submit.prevent label-position="top">
                        <el-form-item>
                            <codemirror
                                :autofocus="true"
                                placeholder="#Define or paste the content of your docker-compose file here"
                                :indent-with-tab="true"
                                :tabSize="4"
                                :style="{ width: '100%', height: `calc(100vh - ${loadHeight()})` }"
                                :lineWrapping="true"
                                :matchBrackets="true"
                                theme="cobalt"
                                :styleActiveLine="true"
                                :extensions="extensions"
                                v-model="content"
                            />
                        </el-form-item>
                        <div v-if="createdBy === '1Panel'">
                            <el-form-item
                                :label="$t('container.env')"
                                prop="environmentStr"
                                v-if="createdBy === '1Panel'"
                            >
                                <el-input
                                    type="textarea"
                                    :placeholder="$t('container.tagHelper')"
                                    :rows="3"
                                    v-model="environmentStr"
                                />
                            </el-form-item>
                            <span class="input-help whitespace-break-spaces">
                                {{ $t('container.editComposeHelper') }}
                            </span>
                            <codemirror
                                v-model="envFileContent"
                                :autofocus="true"
                                :indent-with-tab="true"
                                :tabSize="4"
                                :lineWrapping="true"
                                :matchBrackets="true"
                                :disabled="true"
                                theme="cobalt"
                                :styleActiveLine="true"
                                :extensions="extensions"
                            ></codemirror>
                        </div>
                    </el-form>
                </el-col>
            </el-row>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="composeVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmitEdit()">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { yaml } from '@codemirror/lang-yaml';
import { oneDark } from '@codemirror/theme-one-dark';
import { ref } from 'vue';
import { composeUpdate } from '@/api/modules/container';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { ElForm } from 'element-plus';

const emit = defineEmits<{ (e: 'search'): void }>();
const loading = ref(false);
const composeVisible = ref(false);
const extensions = [yaml(), oneDark];
const path = ref();
const content = ref();
const name = ref();
const environmentStr = ref();
const environmentEnv = ref();
const createdBy = ref();
const envFileContent = ref(`env_file:\n  - 1panel.env`);

const onSubmitEdit = async () => {
    const param = {
        name: name.value,
        path: path.value,
        content: content.value,
        env: environmentStr.value,
        createdBy: createdBy.value,
    };
    if (environmentStr.value != undefined) {
        param.env = environmentStr.value.split('\n');
    }
    loading.value = true;
    await composeUpdate(param)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            composeVisible.value = false;
            emit('search');
        })
        .catch(() => {
            loading.value = false;
        });
};

interface DialogProps {
    name: string;
    path: string;
    content: string;
    env: Array<string>;
    envStr: string;
    createdBy: string;
}

const loadHeight = () => {
    return createdBy.value === '1Panel' ? '300px' : '200px';
};
const acceptParams = (props: DialogProps): void => {
    composeVisible.value = true;
    path.value = props.path;
    name.value = props.name;
    content.value = props.content;
    createdBy.value = props.createdBy;
    environmentEnv.value = props.env || [];
    environmentStr.value = environmentEnv.value.join('\n');
};
const handleClose = () => {
    composeVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
