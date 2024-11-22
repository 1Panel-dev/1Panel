<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            @close="handleClose"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="50%"
        >
            <template #header>
                <DrawerHeader :header="$t('setting.panelSSL')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" :rules="rules" v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.certType')">
                            <el-radio-group v-model="form.sslType">
                                <el-radio value="self">{{ $t('setting.selfSigned') }}</el-radio>
                                <el-radio value="select">{{ $t('setting.select') }}</el-radio>
                                <el-radio value="import">{{ $t('commons.button.import') }}</el-radio>
                            </el-radio-group>
                            <span class="input-help" v-if="form.sslType === 'self'">
                                {{ $t('setting.selfSignedHelper') }}
                            </span>
                        </el-form-item>
                        <el-form-item v-if="form.sslType === 'import'" :label="$t('commons.button.import')" prop="type">
                            <el-select v-model="form.itemSSLType">
                                <el-option :label="$t('website.pasteSSL')" value="paste"></el-option>
                                <el-option :label="$t('website.localSSL')" value="local"></el-option>
                            </el-select>
                        </el-form-item>

                        <el-form-item v-if="form.timeout">
                            <el-tag>{{ $t('setting.domainOrIP') }} {{ form.domain }}</el-tag>
                            <el-tag style="margin-left: 5px">
                                {{ $t('setting.timeOut') }} {{ dateFormat('', '', form.timeout) }}
                            </el-tag>
                            <el-button
                                @click="onDownload"
                                style="margin-left: 5px"
                                v-if="form.sslType === 'self'"
                                type="primary"
                                link
                                icon="Download"
                            >
                                {{ $t('setting.rootCrtDownload') }}
                            </el-button>
                        </el-form-item>

                        <div v-if="form.sslType === 'import' && form.itemSSLType === 'paste'">
                            <el-form-item :label="$t('website.privateKey')" prop="key">
                                <el-input v-model="form.key" :rows="5" type="textarea" />
                            </el-form-item>
                            <el-form-item class="marginTop" :label="$t('website.certificate')" prop="cert">
                                <el-input v-model="form.cert" :rows="5" type="textarea" />
                            </el-form-item>
                        </div>

                        <div v-if="form.sslType === 'import' && form.itemSSLType === 'local'">
                            <el-form-item :label="$t('website.privateKey')" prop="key">
                                <el-input v-model="form.key">
                                    <template #prepend>
                                        <FileList @choose="getKeyPath" :dir="false"></FileList>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item class="marginTop" :label="$t('website.certificate')" prop="cert">
                                <el-input v-model="form.cert">
                                    <template #prepend>
                                        <FileList @choose="getCertPath" :dir="false"></FileList>
                                    </template>
                                </el-input>
                            </el-form-item>
                        </div>

                        <div v-if="form.sslType === 'select'">
                            <el-form-item :label="$t('setting.certificate')" prop="sslID">
                                <el-select v-model="form.sslID" @change="changeSSl(form.sslID)">
                                    <el-option
                                        v-for="(item, index) in sslList"
                                        :key="index"
                                        :label="item.primaryDomain"
                                        :value="item.id"
                                    ></el-option>
                                </el-select>
                            </el-form-item>
                            <el-descriptions
                                class="marginTop"
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
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSaveSSL(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { dateFormatSimple, dateFormat, getProvider } from '@/utils/util';
import { ListSSL } from '@/api/modules/website';
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { downloadSSL, updateSSL } from '@/api/modules/setting';
import { Rules } from '@/global/form-rules';
import { FormInstance } from 'element-plus';
import { Setting } from '@/api/interface/setting';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const loading = ref();
const drawerVisible = ref();

const form = reactive({
    ssl: 'enable',
    domain: '',
    sslType: 'self',
    itemSSLType: 'paste',
    sslID: null as number,
    cert: '',
    key: '',
    rootPath: '',
    timeout: '',
});

const rules = reactive({
    cert: [Rules.requiredInput],
    key: [Rules.requiredInput],
    sslID: [Rules.requiredSelect],
});

const formRef = ref<FormInstance>();

const sslList = ref();
const itemSSL = ref();

interface DialogProps {
    sslType: string;
    sslInfo?: Setting.SSLInfo;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    if (params.sslType.indexOf('-') !== -1) {
        form.sslType = 'import';
        form.itemSSLType = params.sslType.split('-')[1];
    } else {
        form.sslType = params.sslType;
    }
    form.cert = params.sslInfo?.cert || '';
    form.key = params.sslInfo?.key || '';
    form.rootPath = params.sslInfo?.rootPath || '';
    form.domain = params.sslInfo?.domain || '';
    form.timeout = params.sslInfo?.timeout || '';

    if (params.sslInfo?.sslID) {
        form.sslID = params.sslInfo.sslID;
        const ssls = await ListSSL({});
        sslList.value = ssls.data || [];
        changeSSl(params.sslInfo?.sslID);
    } else {
        loadSSLs();
    }
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

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

const getKeyPath = (path: string) => {
    form.key = path;
};

const getCertPath = (path: string) => {
    form.cert = path;
};

const onDownload = async () => {
    await downloadSSL().then(async (file) => {
        const downloadUrl = window.URL.createObjectURL(new Blob([file]));
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = downloadUrl;
        a.download = 'server.crt';
        const event = new MouseEvent('click');
        a.dispatchEvent(event);
    });
};

const onSaveSSL = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let itemType = form.sslType;
        if (form.sslType === 'import') {
            itemType = form.itemSSLType === 'paste' ? 'import-paste' : 'import-local';
        }
        let param = {
            ssl: 'enable',
            sslType: itemType,
            domain: '',
            sslID: form.sslID,
            cert: form.cert,
            key: form.key,
        };
        let href = window.location.href;
        param.domain = href.split('//')[1].split(':')[0];
        await updateSSL(param).then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            let href = window.location.href;
            globalStore.isLogin = false;
            let address = href.split('://')[1];
            if (globalStore.entrance) {
                address = address.replaceAll('settings/safe', globalStore.entrance);
            } else {
                address = address.replaceAll('settings/safe', 'login');
            }
            window.open(`https://${address}`, '_self');
        });
    });
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.marginTop {
    margin-top: 10px;
}
</style>
