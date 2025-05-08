<template>
    <div>
        <DialogPro v-model="open" :title="$t('app.checkTitle')" size="small">
            <el-alert :closable="false" :title="$t('setting.systemIPWarning')" type="info">
                <el-link icon="Position" @click="jumpToPath(router, '/settings/panel')" type="primary">
                    {{ $t('firewall.quickJump') }}
                </el-link>
            </el-alert>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </DialogPro>
    </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { getAgentSettingInfo } from '@/api/modules/setting';
import i18n from '@/lang';
import { MsgError, MsgWarning } from '@/utils/message';
import { jumpToPath } from '@/utils/util';
import { useRouter } from 'vue-router';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();
const router = useRouter();

const open = ref();

interface DialogProps {
    port: any;
    ip: string;
    protocol: string;
}

const acceptParams = async (params: DialogProps): Promise<void> => {
    if (Number(params.port) === 0) {
        MsgError(i18n.global.t('commons.msg.errPort'));
        return;
    }
    let protocol = params.protocol === 'https' ? 'https' : 'http';
    const res = await getAgentSettingInfo();
    if (!res.data.systemIP) {
        if (!globalStore.currentNodeAddr) {
            open.value = true;
            return;
        } else {
            res.data.systemIP = globalStore.currentNodeAddr;
        }
    }
    if (res.data.systemIP.indexOf(':') === -1) {
        if (params.ip && params.ip === 'ipv6') {
            MsgWarning(i18n.global.t('setting.systemIPWarning1', ['IPv4']));
            return;
        }
        window.open(`${protocol}://${res.data.systemIP}:${params.port}`, '_blank');
    } else {
        if (params.ip && params.ip === 'ipv4') {
            MsgWarning(i18n.global.t('setting.systemIPWarning1', ['IPv6']));
            return;
        }
        window.open(`${protocol}://[${res.data.systemIP}]:${params.port}`, '_blank');
    }
};

defineExpose({ acceptParams });
</script>
