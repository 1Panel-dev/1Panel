<template>
    <div>
        <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('setting.allowIPs')" :back="handleClose" />
            </template>
            <el-form
                ref="formRef"
                label-position="top"
                @submit.prevent
                :model="form"
                :rules="rules"
                v-loading="loading"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.allowIPs')" prop="allowIPs">
                            <el-input
                                type="textarea"
                                :placeholder="$t('setting.allowIPEgs')"
                                :autosize="{ minRows: 8, maxRows: 10 }"
                                v-model="form.allowIPs"
                            />
                            <span class="input-help">{{ $t('setting.allowIPsHelper1') }}</span>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
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
import { checkCidr, checkIpV4V6 } from '@/utils/util';
import DrawerHeader from '@/components/drawer-header/index.vue';

const emit = defineEmits<{ (e: 'search'): void }>();

const form = reactive({
    allowIPs: '',
});
const rules = reactive({
    allowIPs: [{ validator: checkAddress, trigger: 'blur' }],
});
function checkAddress(rule: any, value: any, callback: any) {
    if (form.allowIPs !== '') {
        let addrs = form.allowIPs.split('\n');
        for (const item of addrs) {
            if (item === '0.0.0.0') {
                return callback(new Error(i18n.global.t('firewall.addressFormatError')));
            }
            if (item.indexOf('/') !== -1) {
                if (checkCidr(item)) {
                    return callback(new Error(i18n.global.t('firewall.addressFormatError')));
                }
            } else {
                if (checkIpV4V6(item)) {
                    return callback(new Error(i18n.global.t('firewall.addressFormatError')));
                }
            }
        }
        callback();
    }
}
const formRef = ref<FormInstance>();

interface DialogProps {
    allowIPs: string;
}

const drawerVisiable = ref();
const loading = ref();

const acceptParams = (params: DialogProps): void => {
    form.allowIPs = params.allowIPs;
    drawerVisiable.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let title = form.allowIPs ? i18n.global.t('setting.allowIPs') : i18n.global.t('setting.unAllowIPs');
        let allow = form.allowIPs
            ? i18n.global.t('setting.allowIPsWarning')
            : i18n.global.t('setting.unAllowIPsWarning');
        ElMessageBox.confirm(allow, title, {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        }).then(async () => {
            loading.value = true;

            await updateSetting({ key: 'AllowIPs', value: form.allowIPs.replaceAll('\n', ',') })
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    handleClose();
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
