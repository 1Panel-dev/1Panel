<template>
    <DrawerPro v-model="open" :header="$t('website.proxyFile')" @close="handleClose" size="normal">
        <CodemirrorPro v-model="req.content" mode="nginx"></CodemirrorPro>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { updateRedirectConfigFile } from '@/api/modules/website';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';

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
    updateRedirectConfigFile(req)
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
