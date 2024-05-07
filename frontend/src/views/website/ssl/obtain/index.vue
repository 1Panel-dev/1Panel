<template>
    <el-dialog
        v-model="open"
        :title="$t('ssl.apply')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        width="30%"
        :before-close="handleClose"
    >
        <div class="text-center" v-loading="loading">
            <div v-if="ssl.websites && ssl.websites.length > 0">
                <span>{{ $t('ssl.renewWebsite') }}</span>
                <div>
                    <br />
                    <span>
                        <span v-for="(website, index) in ssl.websites" :key="index">
                            <el-tag type="info">{{ website.primaryDomain }}</el-tag>
                        </span>
                    </span>
                </div>
                <br />
            </div>
            <span>{{ $t('ssl.renewConfirm', [ssl.primaryDomain]) }}</span>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { ObtainSSL, RenewSSLByCA } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ref } from 'vue';

interface RenewProps {
    ssl: Website.SSL;
}

const open = ref(false);
const loading = ref(false);
const em = defineEmits(['close', 'submit']);
const handleClose = () => {
    open.value = false;
    em('close', false);
};
const ssl = ref();

const acceptParams = async (props: RenewProps) => {
    ssl.value = props.ssl;
    open.value = true;
};

const submit = async () => {
    loading.value = true;
    try {
        if (ssl.value.provider == 'selfSigned') {
            await RenewSSLByCA({ SSLID: ssl.value.id });
        } else {
            await ObtainSSL({ ID: ssl.value.id });
        }
        handleClose();
        MsgSuccess(i18n.global.t('ssl.applyStart'));
        loading.value = false;
        em('submit', ssl.value.id);
    } catch (error) {}
};

defineExpose({
    acceptParams,
});
</script>
