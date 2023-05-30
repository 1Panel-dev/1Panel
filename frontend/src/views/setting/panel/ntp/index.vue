<template>
    <div>
        <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('setting.syncTime')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.syncSite')" prop="ntpSite" :rules="Rules.domain">
                            <el-input v-model="form.ntpSite" />
                            <el-button type="primary" link class="tagClass" @click="form.ntpSite = 'pool.ntp.org'">
                                {{ $t('website.default') }}
                            </el-button>
                            <el-button type="primary" link class="tagClass" @click="form.ntpSite = 'ntp.aliyun.com'">
                                {{ $t('setting.ntpALi') }}
                            </el-button>
                            <el-button type="primary" link class="tagClass" @click="form.ntpSite = 'time.google.com'">
                                {{ $t('setting.ntpGoogle') }}
                            </el-button>
                        </el-form-item>

                        <el-form-item :label="$t('setting.syncTime')" prop="localTime">
                            <el-input v-model="form.localTime" disabled />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSyncTime(formRef)">
                        {{ $t('commons.button.sync') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { syncTime } from '@/api/modules/setting';
import { ElMessageBox, FormInstance } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { Rules } from '@/global/form-rules';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    localTime: string;
    ntpSite: string;
}
const drawerVisiable = ref();
const loading = ref();

const form = reactive({
    localTime: '',
    ntpSite: '',
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.localTime = params.localTime;
    form.ntpSite = params.ntpSite;
    drawerVisiable.value = true;
};

const onSyncTime = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(
            i18n.global.t('setting.syncSiteHelper', [form.ntpSite]),
            i18n.global.t('setting.syncSite'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        ).then(async () => {
            loading.value = true;
            await syncTime(form.ntpSite)
                .then((res) => {
                    loading.value = false;
                    form.localTime = res.data;
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
    drawerVisiable.value = false;
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
