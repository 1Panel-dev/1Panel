<template>
    <el-drawer :close-on-click-modal="false" :close-on-press-escape="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader :header="$t('ssl.detail')" :back="handleClose" />
        </template>
        <div v-loading="loading">
            <el-radio-group v-model="curr">
                <el-radio-button value="detail">{{ $t('ssl.msg') }}</el-radio-button>
                <el-radio-button value="ssl">{{ $t('ssl.ssl') }}</el-radio-button>
                <el-radio-button value="key">{{ $t('ssl.key') }}</el-radio-button>
            </el-radio-group>
            <div v-if="curr === 'detail'" class="mt-5">
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
                        <el-tag type="info">{{ getDNSName(ssl.dnsAccount.type) }}</el-tag>
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
            <div v-else-if="curr === 'ssl'" class="mt-5">
                <el-input v-model="ssl.pem" :rows="15" type="textarea" id="textArea" />
                <div>
                    <br />
                    <CopyButton :content="ssl.pem" />
                </div>
            </div>
            <div v-else class="mt-5">
                <el-input v-model="ssl.privateKey" :rows="15" type="textarea" id="textArea" />
                <div>
                    <br />
                    <CopyButton :content="ssl.privateKey" />
                </div>
            </div>
        </div>
    </el-drawer>
</template>
<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { GetSSL } from '@/api/modules/website';
import { ref } from 'vue';
import { dateFormatSimple, getProvider, getDNSName } from '@/utils/util';

const open = ref(false);
const id = ref(0);
const curr = ref('detail');
const ssl = ref<any>({});
const loading = ref(false);

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

defineExpose({
    acceptParams,
});
</script>
