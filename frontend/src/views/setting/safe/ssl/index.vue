<template>
    <div>
        <el-card>
            <el-radio-group v-model="sslItemType">
                <el-radio label="self">{{ $t('setting.selfSigned') }}</el-radio>
                <el-radio label="select">{{ $t('setting.select') }}</el-radio>
                <el-radio label="import">{{ $t('setting.import') }}</el-radio>
            </el-radio-group>
            <span class="input-help" v-if="sslItemType === 'self'">{{ $t('setting.selfSignedHelper') }}</span>
            <div v-if="sslInfo.timeout">
                <el-tag>{{ $t('setting.domainOrIP') }} {{ sslInfo.domain }}</el-tag>
                <el-tag style="margin-left: 5px">{{ $t('setting.timeOut') }} {{ sslInfo.timeout }}</el-tag>
                <el-button
                    @click="onDownload"
                    style="margin-left: 5px"
                    v-if="sslItemType === 'self'"
                    type="primary"
                    link
                    icon="Download"
                >
                    {{ $t('setting.rootCrtDownload') }}
                </el-button>
            </div>

            <div v-if="sslItemType === 'import'">
                <span class="input-help">{{ $t('setting.primaryKey') }}</span>
                <el-input v-model="form.key" :autosize="{ minRows: 2, maxRows: 6 }" type="textarea" />
                <span class="input-help">{{ $t('setting.certificate') }}</span>
                <el-input v-model="form.cert" :autosize="{ minRows: 2, maxRows: 6 }" type="textarea" />
            </div>

            <div v-if="sslItemType === 'select'">
                <el-select style="margin-top: 10px" v-model="form.sslID" @change="changeSSl(form.sslID)">
                    <el-option
                        v-for="(item, index) in sslList"
                        :key="index"
                        :label="item.primaryDomain"
                        :value="item.id"
                    ></el-option>
                </el-select>
                <el-descriptions
                    class="margintop"
                    :column="5"
                    border
                    direction="vertical"
                    v-if="form.sslID > 0 && itemSSL"
                >
                    <el-descriptions-item :label="$t('website.primaryDomain')">
                        {{ itemSSL.primaryDomain }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('website.otherDomains')">
                        {{ itemSSL.domains }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('ssl.provider')">
                        {{ getProvider(itemSSL.provider) }}
                    </el-descriptions-item>
                    <el-descriptions-item
                        :label="$t('ssl.acmeAccount')"
                        v-if="itemSSL.acmeAccount?.email && itemSSL.provider !== 'manual'"
                    >
                        {{ itemSSL.acmeAccount.email }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('website.expireDate')">
                        {{ dateFormatSimple(itemSSL.expireDate) }}
                    </el-descriptions-item>
                </el-descriptions>
            </div>
            <el-button class="margintop" type="primary" @click="onSaveSSL">
                {{ $t('commons.button.saveAndEnable') }}
            </el-button>
        </el-card>
    </div>
</template>
<script lang="ts" setup>
import { Website } from '@/api/interface/Website';
import { loadSSLInfo } from '@/api/modules/setting';
import { dateFormatSimple, getProvider } from '@/utils/util';
import { ListSSL } from '@/api/modules/website';
import { nextTick, onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { updateSSL } from '@/api/modules/setting';
import { DownloadByPath } from '@/api/modules/files';

const form = reactive({
    ssl: 'enable',
    domain: '',
    sslType: 'self',
    sslID: null as number,
    cert: '',
    key: '',
    rootPath: '',
});

const props = defineProps({
    type: {
        type: String,
        default: 'self',
    },
});

const sslInfo = reactive({
    domain: '',
    timeout: '',
});
const sslList = ref();
const itemSSL = ref();
const sslItemType = ref('self');

const loadInfo = async () => {
    await loadSSLInfo().then(async (res) => {
        sslInfo.domain = res.data.domain || '';
        sslInfo.timeout = res.data.timeout || '';
        form.cert = res.data.cert;
        form.key = res.data.key;
        form.rootPath = res.data.rootPath;
        if (res.data.sslID) {
            form.sslID = res.data.sslID;
            const ssls = await ListSSL({});
            sslList.value = ssls.data || [];
            changeSSl(form.sslID);
        }
    });
};

const loadSSLs = async () => {
    const res = await ListSSL({});
    sslList.value = res.data || [];
};

const changeSSl = (sslid: number) => {
    const res = sslList.value.filter((element: Website.SSL) => {
        return element.id == sslid;
    });
    itemSSL.value = res[0];
};

const onDownload = async () => {
    const file = await DownloadByPath(form.rootPath);
    const downloadUrl = window.URL.createObjectURL(new Blob([file]));
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = downloadUrl;
    a.download = 'server.crt';
    const event = new MouseEvent('click');
    a.dispatchEvent(event);
};

const onSaveSSL = async () => {
    form.sslType = sslItemType.value;
    let href = window.location.href;
    form.domain = href.split('//')[1].split(':')[0];
    await updateSSL(form).then(() => {
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        let href = window.location.href;
        let address = href.split('://')[1];
        window.open(`https://${address}/`, '_self');
    });
};

onMounted(() => {
    nextTick(() => {
        sslItemType.value = props.type;
        loadInfo();
    });
    loadSSLs();
});
</script>

<style scoped lang="scss">
.margintop {
    margin-top: 10px;
}
</style>
