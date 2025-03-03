<template>
    <DrawerPro v-model="drawerVisible" :header="$t('toolbox.device.syncSite')" @close="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('toolbox.device.syncSite')" prop="ntpSite" :rules="Rules.domain">
                <el-input v-model="form.ntpSite" />
                <el-button type="primary" link class="tagClass" @click="form.ntpSite = 'pool.ntp.org'">
                    {{ $t('commons.table.default') }}
                </el-button>
                <el-button type="primary" link class="tagClass" @click="form.ntpSite = 'ntp.aliyun.com'">
                    {{ $t('toolbox.device.ntpALi') }}
                </el-button>
                <el-button type="primary" link class="tagClass" @click="form.ntpSite = 'time.google.com'">
                    {{ $t('toolbox.device.ntpGoogle') }}
                </el-button>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSyncTime(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox, FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { updateDevice } from '@/api/modules/toolbox';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    ntpSite: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    ntpSite: '',
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.ntpSite = params.ntpSite;
    drawerVisible.value = true;
};

const onSyncTime = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(
            i18n.global.t('toolbox.device.syncSiteHelper', [form.ntpSite]),
            i18n.global.t('toolbox.device.syncSite'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        ).then(async () => {
            loading.value = true;
            await updateDevice('Ntp', form.ntpSite)
                .then(() => {
                    loading.value = false;
                    emit('search');
                    handleClose();
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        });
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.tagClass {
    margin-top: 5px;
}
</style>
