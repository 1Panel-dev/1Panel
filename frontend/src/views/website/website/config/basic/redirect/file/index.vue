<template>
    <el-drawer
        v-model="open"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="40%"
        :before-close="handleClose"
    >
        <template #header>
            <DrawerHeader :header="$t('website.sourceFile')" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <div class="redirect-editor">
                    <codemirror
                        :autofocus="true"
                        placeholder=""
                        :indent-with-tab="true"
                        :tabSize="4"
                        :lineWrapping="true"
                        :matchBrackets="true"
                        style="height: 600px"
                        theme="cobalt"
                        :styleActiveLine="true"
                        :extensions="extensions"
                        v-model="req.content"
                    />
                </div>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { Codemirror } from 'vue-codemirror';
import { UpdateRedirectConfigFile } from '@/api/modules/website';
import { StreamLanguage } from '@codemirror/language';
import { nginx } from '@codemirror/legacy-modes/mode/nginx';
import { oneDark } from '@codemirror/theme-one-dark';

const extensions = [StreamLanguage.define(nginx), oneDark];
const proxyForm = ref<FormInstance>();
const open = ref(false);
const loading = ref(false);
const em = defineEmits(['close']);
const handleClose = () => {
    proxyForm.value?.resetFields();
    open.value = false;
    em('close', false);
};
const req = reactive({
    name: '',
    websiteID: 0,
    content: '',
});

const acceptParams = async (proxyreq: any) => {
    req.name = proxyreq.name;
    req.websiteID = proxyreq.websiteID;
    req.content = proxyreq.content;
    open.value = true;
};

const submit = async () => {
    loading.value = true;
    UpdateRedirectConfigFile(req)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            handleClose();
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>

<style scoped>
.redirect-editor {
    margin-top: 10px;
    width: 100%;
}
</style>
