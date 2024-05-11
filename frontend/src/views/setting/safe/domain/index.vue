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
                <DrawerHeader :header="$t('setting.bindDomain')" :back="handleClose" />
            </template>
            <el-form
                ref="formRef"
                label-position="top"
                :rules="rules"
                :model="form"
                @submit.prevent
                v-loading="loading"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.bindDomain')" prop="bindDomain">
                            <el-input clearable v-model="form.bindDomain" />
                            <span class="input-help">{{ $t('setting.bindDomainHelper1') }}</span>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSavePort(formRef)">
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
import DrawerHeader from '@/components/drawer-header/index.vue';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    bindDomain: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    bindDomain: '',
});
const rules = reactive({
    bindDomain: [{ validator: checkSecurityEntrance, trigger: 'blur' }],
});

function checkSecurityEntrance(rule: any, value: any, callback: any) {
    if (form.bindDomain !== '') {
        const reg =
            /^([\w\u4e00-\u9fa5\-\*]{1,100}\.){1,10}([\w\u4e00-\u9fa5\-]{1,24}|[\w\u4e00-\u9fa5\-]{1,24}\.[\w\u4e00-\u9fa5\-]{1,24})$/;
        if (!reg.test(form.bindDomain)) {
            return callback(new Error(i18n.global.t('commons.rule.domain')));
        }
    }
    callback();
}

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.bindDomain = params.bindDomain;
    drawerVisible.value = true;
};

const onSavePort = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let title = form.bindDomain ? i18n.global.t('setting.bindDomain') : i18n.global.t('setting.unBindDomain');
        let helper = form.bindDomain
            ? i18n.global.t('setting.bindDomainWarning')
            : i18n.global.t('setting.unBindDomainHelper');
        ElMessageBox.confirm(helper, title, {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        }).then(async () => {
            loading.value = true;
            await updateSetting({ key: 'BindDomain', value: form.bindDomain })
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    if (form.bindDomain) {
                        let href = window.location.href;
                        let ipLocal = href.split('//')[1].split(':')[0];
                        href = href.replaceAll(ipLocal, form.bindDomain);
                        window.open(href, '_self');
                    } else {
                        handleClose();
                        emit('search');
                    }
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
