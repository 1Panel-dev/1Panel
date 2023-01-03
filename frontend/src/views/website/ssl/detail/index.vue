<template>
    <el-dialog v-model="open" :before-close="handleClose" :title="$t('ssl.detail')" width="30%">
        <div>
            <el-radio-group v-model="curr">
                <el-radio-button label="detail">{{ $t('ssl.msg') }}</el-radio-button>
                <el-radio-button label="ssl">{{ $t('ssl.ssl') }}</el-radio-button>
                <el-radio-button label="key">{{ $t('ssl.key') }}</el-radio-button>
            </el-radio-group>
            <br />
            <br />
            <div v-if="curr === 'detail'">
                <el-descriptions border :column="1">
                    <el-descriptions-item :label="$t('website.primaryDomain')">
                        {{ ssl.primaryDomain }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('website.otherDomains')">
                        {{ ssl.otherDomains }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('ssl.provider')">
                        {{ getProvider(ssl.provider) }}
                    </el-descriptions-item>
                    <el-descriptions-item
                        :label="$t('ssl.acmeAccount')"
                        v-if="ssl.acmeAccount && ssl.provider !== 'manual'"
                    >
                        {{ ssl.acmeAccount.email }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('website.brand')">
                        {{ ssl.type }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('ssl.startDate')">
                        {{ dateFromat(0, 0, ssl.startDate) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('website.expireDate')">
                        {{ dateFromat(0, 0, ssl.expireDate) }}
                    </el-descriptions-item>
                </el-descriptions>
            </div>
            <div v-else-if="curr === 'ssl'">
                <el-input v-model="ssl.pem" :autosize="{ minRows: 10, maxRows: 15 }" type="textarea" id="textArea" />
                <div>
                    <br />
                    <el-button type="primary" @click="copyText(ssl.pem)">{{ $t('file.copy') }}</el-button>
                </div>
            </div>
            <div v-else>
                <el-input
                    v-model="ssl.privateKey"
                    :autosize="{ minRows: 10, maxRows: 15 }"
                    type="textarea"
                    id="textArea"
                />
                <div>
                    <br />
                    <el-button type="primary" @click="copyText(ssl.privateKey)">{{ $t('file.copy') }}</el-button>
                </div>
            </div>
        </div>
    </el-dialog>
</template>
<script lang="ts" setup>
import { GetSSL } from '@/api/modules/website';
import { ref } from 'vue';
import { dateFromat, getProvider } from '@/utils/util';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';
import useClipboard from 'vue-clipboard3';
const { toClipboard } = useClipboard();

let open = ref(false);
let id = ref(0);
let curr = ref('detail');
let ssl = ref<any>({});

const handleClose = () => {
    open.value = false;
};

const acceptParams = (sslId: number) => {
    ssl.value = {};
    id.value = sslId;
    curr.value = 'detail';
    get();
    open.value = true;
};

const get = async () => {
    const res = await GetSSL(id.value);
    ssl.value = res.data;
};

const copyText = async (msg) => {
    try {
        await toClipboard(msg);
        ElMessage.success(i18n.global.t('commons.msg.copySuccess'));
    } catch (e) {
        ElMessage.error(i18n.global.t('commons.msg.copyfailed'));
    }
};

// const copyText = async (text: string) => {
//     try {
//         try {
//             await navigator.clipboard.writeText(text);
//             ElMessage.success(i18n.global.t('commons.msg.copySuccess'));
//             return await Promise.resolve();
//         } catch (err) {
//             return await Promise.reject(err);
//         }
//     } catch (e) {
//         let input = document.createElement('input');
//         input.style.position = 'fixed';
//         input.style.top = '-10000px';
//         input.style.zIndex = '-999';
//         document.body.appendChild(input);
//         console.log('input', input);
//         input.value = text;
//         input.focus();
//         input.select();
//         try {
//             let result = document.execCommand('copy');
//             document.body.removeChild(input);
//             if (!result) {
//                 ElMessage.error(i18n.global.t('commons.msg.copyfailed'));
//                 return Promise.reject();
//             } else {
//                 ElMessage.success(i18n.global.t('commons.msg.copySuccess'));
//                 return Promise.resolve();
//             }
//         } catch (e) {
//             document.body.removeChild(input);
//             ElMessage.error(i18n.global.t('commons.msg.copyfailed'));
//             return Promise.reject();
//         }
//     }
// };

defineExpose({
    acceptParams,
});
</script>
