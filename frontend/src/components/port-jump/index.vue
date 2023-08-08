<template>
    <div>
        <el-dialog
            v-model="dialogVisiable"
            :title="$t('app.checkTitle')"
            width="30%"
            :close-on-click-modal="false"
            :destroy-on-close="true"
        >
            <el-alert :closable="false" :title="$t('setting.systemIPWarning')" type="info">
                <el-link icon="Position" @click="goRouter('/settings/panel')" type="primary">
                    {{ $t('firewall.quickJump') }}
                </el-link>
            </el-alert>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="dialogVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { getSettingInfo } from '@/api/modules/setting';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
import { useRouter } from 'vue-router';
import { checkDomain, checkIp } from '@/utils/util';
const router = useRouter();

const dialogVisiable = ref();

interface DialogProps {
    port: any;
    protocol: string;
}

const acceptParams = async (params: DialogProps): Promise<void> => {
    if (Number(params.port) === 0) {
        MsgError(i18n.global.t('setting.errPort'));
        return;
    }
    let protocol = params.protocol === 'https' ? 'https' : 'http';
    const res = await getSettingInfo();
    if (!res.data.systemIP) {
        dialogVisiable.value = true;
        return;
    }
    if (!checkIp(res.data.systemIP) || !checkDomain(res.data.systemIP)) {
        window.open(`${protocol}://${res.data.systemIP}:${params.port}`, '_blank');
        return;
    }
    window.open(`${protocol}://[${res.data.systemIP}]:${params.port}`, '_blank');
};

const goRouter = async (path: string) => {
    router.push({ path: path });
};

defineExpose({ acceptParams });
</script>
