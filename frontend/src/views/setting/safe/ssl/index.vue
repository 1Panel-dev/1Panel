<template>
    <div>
        <el-drawer
            v-model="drawerVisiable"
            :destroy-on-close="true"
            @close="handleClose"
            :close-on-click-modal="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader header="https" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" :rules="rules" v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.certType')">
                            <el-radio-group v-model="form.sslType">
                                <el-radio label="self">{{ $t('setting.selfSigned') }}</el-radio>
                                <el-radio label="select">{{ $t('setting.select') }}</el-radio>
                                <el-radio label="import">{{ $t('commons.button.import') }}</el-radio>
                            </el-radio-group>
                            <span class="input-help" v-if="form.sslType === 'self'">
                                {{ $t('setting.selfSignedHelper') }}
                            </span>
                        </el-form-item>

                        <el-form-item v-if="form.timeout">
                            <el-tag>{{ $t('setting.domainOrIP') }} {{ form.domain }}</el-tag>
                            <el-tag style="margin-left: 5px">{{ $t('setting.timeOut') }} {{ form.timeout }}</el-tag>
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

                        <div v-if="form.sslType === 'import'">
                            <el-form-item :label="$t('website.privateKey')" prop="key">
                                <el-input v-model="form.key" :autosize="{ minRows: 5, maxRows: 10 }" type="textarea" />
                            </el-form-item>
                            <el-form-item class="margintop" :label="$t('website.certificate')" prop="cert">
                                <el-input v-model="form.cert" :autosize="{ minRows: 5, maxRows: 10 }" type="textarea" />
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
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
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
import { dateFormatSimple, getProvider } from '@/utils/util';
import { ListSSL } from '@/api/modules/website';
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { updateSSL } from '@/api/modules/setting';
import { DownloadByPath } from '@/api/modules/files';
import { Rules } from '@/global/form-rules';
import { ElMessageBox, FormInstance } from 'element-plus';
import { Setting } from '@/api/interface/setting';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const loading = ref();
const drawerVisiable = ref();

const form = reactive({
    ssl: 'enable',
    domain: '',
    sslType: 'self',
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
    form.sslType = params.sslType;
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
    drawerVisiable.value = true;
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

const onSaveSSL = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('setting.sslChangeHelper'), 'https', {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        }).then(async () => {
            let param = {
                ssl: 'enable',
                sslType: form.sslType,
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
    });
};

const handleClose = () => {
    emit('search');
    drawerVisiable.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.margintop {
    margin-top: 10px;
}
</style>
