<template>
    <el-drawer
        v-model="open"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        :size="size"
        :before-close="handleClose"
    >
        <template #header>
            <DrawerHeader :header="$t('website.sourceFile')" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <codemirror
                    :autofocus="true"
                    placeholder=""
                    :indent-with-tab="true"
                    :tabSize="4"
                    style="margin-top: 10px; height: 600px; width: 100%"
                    :lineWrapping="true"
                    :matchBrackets="true"
                    theme="cobalt"
                    :styleActiveLine="true"
                    :extensions="extensions"
                    v-model="req.content"
                />
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
import { computed, reactive, ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { Codemirror } from 'vue-codemirror';
import { UpdateProxyConfigFile } from '@/api/modules/website';
import { StreamLanguage } from '@codemirror/language';
import { nginx } from '@codemirror/legacy-modes/mode/nginx';
import { oneDark } from '@codemirror/theme-one-dark';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

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
const size = ref('50%');

const acceptParams = async (proxyreq: any) => {
    req.name = proxyreq.name;
    req.websiteID = proxyreq.websiteID;
    req.content = proxyreq.content;
    open.value = true;

    if (mobile.value) {
        size.value = '100%';
    }
};

const submit = async () => {
    loading.value = true;
    UpdateProxyConfigFile(req)
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
