<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('setting.user')" :back="handleClose" />
            </template>
            <el-form
                ref="formRef"
                label-position="top"
                :model="form"
                @submit.prevent
                v-loading="loading"
                :rules="rules"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.user')" prop="userName">
                            <el-input clearable v-model.trim="form.userName" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSaveUserName(formRef)">
                        {{ $t('commons.button.confirm') }}
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
import { updateSetting } from '@/api/modules/setting';
import { ElMessageBox, FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { logOutApi } from '@/api/modules/auth';
import router from '@/routers';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

interface DialogProps {
    userName: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    userName: '',
});
const rules = reactive({
    userName: [Rules.userName, Rules.noSpace],
});
const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.userName = params.userName;
    drawerVisible.value = true;
};

const onSaveUserName = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('setting.userChangeHelper'), i18n.global.t('setting.userChange'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        }).then(async () => {
            await updateSetting({ key: 'UserName', value: form.userName })
                .then(async () => {
                    await logOutApi();
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    router.push({ name: 'entrance', params: { code: globalStore.entrance } });
                    globalStore.setLogStatus(false);
                    return;
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
