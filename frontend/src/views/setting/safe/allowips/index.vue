<template>
    <div>
        <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('setting.allowIPs')" :back="handleClose" />
            </template>
            <el-form label-position="top" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item>
                            <table style="width: 100%" class="tab-table">
                                <tr v-if="allowIPs.length !== 0">
                                    <th scope="col" width="90%" align="left">
                                        <label>IP</label>
                                    </th>
                                    <th align="left"></th>
                                </tr>
                                <tr v-for="(row, index) in allowIPs" :key="index">
                                    <td width="90%">
                                        <el-input
                                            :placeholder="$t('container.serverExample')"
                                            style="width: 100%"
                                            v-model="row.value"
                                        />
                                    </td>
                                    <td>
                                        <el-button link style="font-size: 10px" @click="handlePortsDelete(index)">
                                            {{ $t('commons.button.delete') }}
                                        </el-button>
                                    </td>
                                </tr>
                                <tr>
                                    <td align="left">
                                        <el-button @click="handlePortsAdd()">{{ $t('commons.button.add') }}</el-button>
                                    </td>
                                </tr>
                            </table>
                            <span class="input-help">{{ $t('setting.allowIPsHelper1') }}</span>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSavePort()">
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
import { checkIp } from '@/utils/util';

const emit = defineEmits<{ (e: 'search'): void }>();

const allowIPs = ref();
interface DialogProps {
    allowIPs: string;
}
const drawerVisiable = ref();
const loading = ref();

const acceptParams = (params: DialogProps): void => {
    allowIPs.value = [];
    if (params.allowIPs) {
        for (const ip of params.allowIPs.split(',')) {
            if (ip) {
                allowIPs.value.push({ value: ip });
            }
        }
    }
    drawerVisiable.value = true;
};

const handlePortsAdd = () => {
    let item = {
        value: '',
    };
    allowIPs.value.push(item);
};
const handlePortsDelete = (index: number) => {
    allowIPs.value.splice(index, 1);
};

const onSavePort = async () => {
    let allows = '';
    if (allowIPs.value.length !== 0) {
        for (const ip of allowIPs.value) {
            if (checkIp(ip.value)) {
                MsgError(i18n.global.t('firewall.addressFormatError'));
                return false;
            }
            allows += ip.value + ',';
        }
        allows = allows.substring(0, allows.length - 1);
    }

    ElMessageBox.confirm(i18n.global.t('setting.allowIPsHelper'), i18n.global.t('setting.allowIPs'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;

        await updateSetting({ key: 'AllowIPs', value: allows })
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
