<template>
    <DrawerPro v-model="open" :header="$t('website.proxyFile')" :back="handleClose" :size="mobile ? 'full' : 'normal'">
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
import { computed, reactive, ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { updateLoadBalanceFile } from '@/api/modules/website';
import { GlobalStore } from '@/store';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import { Website } from '@/api/interface/website';
const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

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

const acceptParams = async (ups: Website.NginxUpstreamFile) => {
    req.name = ups.name;
    req.websiteID = ups.websiteID;
    req.content = ups.content;
    open.value = true;
};

const submit = async () => {
    loading.value = true;
    updateLoadBalanceFile(req)
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
