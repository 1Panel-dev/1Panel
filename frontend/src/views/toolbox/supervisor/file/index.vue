<template>
    <DrawerPro
        v-model="open"
        :header="$t('website.source')"
        @close="handleClose"
        :size="globalStore.isFullScreen ? 'full' : 'large'"
        :fullScreen="true"
    >
        <template #content>
            <div v-loading="loading">
                <CodemirrorPro class="mt-5" v-model="content" :heightDiff="400"></CodemirrorPro>
            </div>
        </template>

        <template #footer>
            <span>
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" :disabled="loading" @click="submit()">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { onUnmounted, reactive, ref } from 'vue';
import { operateSupervisorProcessFile } from '@/api/modules/host-tool';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const loading = ref(false);
const content = ref('');
const tailLog = ref(false);
const open = ref(false);
const req = reactive({
    name: '',
    file: 'conf',
    operate: '',
    content: '',
});

const em = defineEmits(['search']);

const getContent = () => {
    loading.value = true;
    operateSupervisorProcessFile(req)
        .then((res) => {
            content.value = res.data;
        })
        .finally(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    content.value = '';
    open.value = false;
};

const submit = () => {
    const updateReq = {
        name: req.name,
        operate: 'update',
        file: req.file,
        content: content.value,
    };
    loading.value = true;
    operateSupervisorProcessFile(updateReq)
        .then(() => {
            em('search');
            open.value = false;
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const acceptParams = (name: string, file: string, operate: string) => {
    req.name = name;
    req.file = file;
    req.operate = operate;

    getContent();
    open.value = true;
};

const onCloseLog = async () => {
    tailLog.value = false;
};

onUnmounted(() => {
    onCloseLog();
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.fullScreen {
    border: none;
}
</style>
