<template>
    <el-dialog
        v-model="open"
        :title="$t('website.renewSSL')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        width="30%"
        :before-close="handleClose"
    >
        <div style="text-align: center" v-loading="loading">
            <div v-if="websites.length > 0">
                <span>{{ $t('ssl.renewWebsite') }}</span>
                <div>
                    <br />
                    <span>
                        <span v-for="(website, index) in websites" :key="index">
                            <el-tag>{{ website.primaryDomain }}</el-tag>
                        </span>
                    </span>
                </div>
                <br />
            </div>
            <span>{{ $t('ssl.renewConfirm') }}</span>
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
import { RenewSSL } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { reactive, ref } from 'vue';

interface RenewProps {
    id: number;
    websites: Website.Website[];
}

let open = ref(false);
let loading = ref(false);
let renewReq = reactive({
    SSLId: 0,
});
const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    em('close', false);
};
const websites = ref([]);

const acceptParams = async (props: RenewProps) => {
    renewReq.SSLId = Number(props.id);
    websites.value = props.websites;
    open.value = true;
};

const submit = () => {
    loading.value = true;
    RenewSSL(renewReq)
        .then(() => {
            handleClose();
            MsgSuccess(i18n.global.t('ssl.renewSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
