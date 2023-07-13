<template>
    <div>
        <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('setting.allowIPs')" :back="handleClose" />
            </template>
            <el-form label-position="top" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.allowIPs')">
                            <el-input
                                type="textarea"
                                :placeholder="$t('setting.allowIPEgs')"
                                :autosize="{ minRows: 8, maxRows: 10 }"
                                v-model="allowIPs"
                            />
                            <span class="input-help">{{ $t('setting.allowIPsHelper1') }}</span>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave()">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { updateSetting } from '@/api/modules/setting';
import { ElMessageBox } from 'element-plus';
import { checkIpV4V6 } from '@/utils/util';
import DrawerHeader from '@/components/drawer-header/index.vue';

const emit = defineEmits<{ (e: 'search'): void }>();

const allowIPs = ref();
interface DialogProps {
    allowIPs: string;
}
const drawerVisiable = ref();
const loading = ref();

const acceptParams = (params: DialogProps): void => {
    allowIPs.value = params.allowIPs;
    drawerVisiable.value = true;
};

const onSave = async () => {
    if (allowIPs.value) {
        let ips = allowIPs.value.split('\n');
        for (const ip of ips) {
            if (ip) {
                if (checkIpV4V6(ip) || ip === '0.0.0.0') {
                    MsgError(i18n.global.t('firewall.addressFormatError'));
                    return false;
                }
            }
        }
    }
    let title = allowIPs.value ? i18n.global.t('setting.allowIPs') : i18n.global.t('setting.unAllowIPs');
    let allow = allowIPs.value ? i18n.global.t('setting.allowIPsWarning') : i18n.global.t('setting.unAllowIPsWarning');
    ElMessageBox.confirm(allow, title, {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;

        await updateSetting({ key: 'AllowIPs', value: allowIPs.value.replaceAll('\n', ',') })
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
};

const handleClose = () => {
    drawerVisiable.value = false;
};

defineExpose({
    acceptParams,
});
</script>
