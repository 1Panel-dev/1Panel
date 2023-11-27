<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader :header="$t('ssl.detail')" :back="handleClose" />
        </template>
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
                        {{ ssl.domains }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('ssl.commonName')">
                        {{ ssl.type }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('website.brand')">
                        {{ ssl.organization }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('ssl.startDate')">
                        {{ dateFormatSimple(ssl.startDate) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('website.expireDate')">
                        {{ dateFormatSimple(ssl.expireDate) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('ssl.applyType')">
                        {{ getProvider(ssl.provider) }}
                    </el-descriptions-item>
                    <el-descriptions-item
                        :label="$t('website.dnsAccount')"
                        v-if="ssl.dnsAccount && ssl.dnsAccount.id > 0"
                    >
                        {{ ssl.dnsAccount.name }}
                        <el-tag type="info">{{ ssl.dnsAccount.type }}</el-tag>
                    </el-descriptions-item>
                    <el-descriptions-item
                        :label="$t('ssl.acmeAccount')"
                        v-if="ssl.acmeAccount && ssl.acmeAccount.id > 0"
                    >
                        {{ ssl.acmeAccount.email }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('ssl.pushDir')" v-if="ssl.pushDir">
                        {{ ssl.dir }}
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
    </el-drawer>
</template>
<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { GetSSL } from '@/api/modules/website';
import { ref } from 'vue';
import { dateFormatSimple, getProvider } from '@/utils/util';
import i18n from '@/lang';
import useClipboard from 'vue-clipboard3';
import { MsgError, MsgSuccess } from '@/utils/message';
const { toClipboard } = useClipboard();

const open = ref(false);
const id = ref(0);
const curr = ref('detail');
const ssl = ref<any>({});

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
        MsgSuccess(i18n.global.t('commons.msg.copySuccess'));
    } catch (e) {
        MsgError(i18n.global.t('commons.msg.copyFailed'));
    }
};

defineExpose({
    acceptParams,
});
</script>
