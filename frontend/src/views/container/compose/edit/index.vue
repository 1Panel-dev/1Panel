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
                                style="width: 100%; height: calc(100vh - 175px)"
                                :lineWrapping="true"
                                :matchBrackets="true"
                                theme="cobalt"
                                :styleActiveLine="true"
                                :extensions="extensions"
                                v-model="content"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('container.env')" prop="environmentStr">
                            <el-input
                                type="textarea"
                                :placeholder="$t('container.tagHelper')"
                                :rows="3"
                                v-model="environmentStr"
                            />
                        </el-form-item>
                        <span class="input-help">{{ $t('container.editComposeHelper') }}</span>
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
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { ref } from 'vue';
import { composeUpdate } from '@/api/modules/container';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { ElForm } from 'element-plus';

const loading = ref(false);
const composeVisible = ref(false);
const extensions = [javascript(), oneDark];
const path = ref();
const content = ref();
const name = ref();
const environmentStr = ref();

const onSubmitEdit = async () => {
    const param = {
        name: name.value,
        path: path.value,
        content: content.value,
        env: environmentStr.value,
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
        })
        .catch(() => {
            loading.value = false;
        });
};

interface DialogProps {
    name: string;
    path: string;
    content: string;
}

const acceptParams = (props: DialogProps): void => {
    composeVisible.value = true;
    path.value = props.path;
    name.value = props.name;
    content.value = props.content;
    environmentStr.value = '';
};
const handleClose = () => {
    composeVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
