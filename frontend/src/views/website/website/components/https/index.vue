<template>
    <div>
        <el-form-item :label="$t('website.HTTPConfig')" prop="httpConfig">
            <el-select v-model="formData.httpConfig" class="p-w-400">
                <el-option :label="$t('website.HTTPToHTTPS')" value="HTTPToHTTPS"></el-option>
                <el-option :label="$t('website.HTTPAlso')" value="HTTPAlso"></el-option>
                <el-option :label="$t('website.HTTPSOnly')" value="HTTPSOnly"></el-option>
            </el-select>
        </el-form-item>

        <el-form-item label="HSTS" prop="hsts">
            <el-checkbox v-model="formData.hsts">{{ $t('commons.button.enable') }}</el-checkbox>
            <span class="input-help">{{ $t('website.hstsHelper') }}</span>
        </el-form-item>

        <el-form-item
            :label="'HSTS ' + $t('website.includeSubDomains')"
            prop="hstsIncludeSubDomains"
            v-if="formData.hsts"
        >
            <el-checkbox v-model="formData.hstsIncludeSubDomains">
                {{ $t('commons.button.enable') }}
            </el-checkbox>
            <span class="input-help">{{ $t('website.hstsIncludeSubDomainsHelper') }}</span>
        </el-form-item>

        <el-form-item label="HTTP3" prop="http3">
            <el-checkbox v-model="formData.http3">{{ $t('commons.button.enable') }}</el-checkbox>
            <span class="input-help">{{ $t('website.http3Helper') }}</span>
        </el-form-item>

        <el-form-item :label="$t('website.sslConfig')" prop="type">
            <el-select v-model="formData.type" @change="handleTypeChange" class="p-w-400">
                <el-option :label="$t('website.oldSSL')" value="existed"></el-option>
                <el-option :label="$t('website.manualSSL')" value="manual"></el-option>
            </el-select>
        </el-form-item>

        <div v-if="formData.type === 'existed'">
            <el-form-item :label="$t('website.acmeAccountManage')" prop="acmeAccountID">
                <el-select
                    v-model="formData.acmeAccountID"
                    :placeholder="$t('website.selectAcme')"
                    @change="handleAcmeAccountChange"
                    class="p-w-400"
                >
                    <el-option :key="0" :label="$t('website.imported')" :value="0"></el-option>
                    <el-option v-for="(acme, index) in acmeAccounts" :key="index" :label="acme.email" :value="acme.id">
                        <span>
                            {{ acme.email }}
                            <el-tag class="ml-5">{{ getAccountName(acme.type) }}</el-tag>
                        </span>
                    </el-option>
                </el-select>
            </el-form-item>

            <el-form-item :label="$t('website.ssl')" prop="websiteSSLId" :hide-required-asterisk="true">
                <el-select
                    v-model="formData.websiteSSLId"
                    :placeholder="$t('website.selectSSL')"
                    @change="handleSSLChange"
                    class="p-w-400"
                >
                    <el-option
                        v-for="(ssl, index) in ssls"
                        :key="index"
                        :label="ssl.primaryDomain"
                        :value="ssl.id"
                        :disabled="ssl.pem === ''"
                    ></el-option>
                </el-select>
            </el-form-item>
        </div>

        <div v-if="formData.type === 'manual'">
            <el-form-item :label="$t('website.importType')" prop="importType">
                <el-select v-model="formData.importType">
                    <el-option :label="$t('website.pasteSSL')" value="paste"></el-option>
                    <el-option :label="$t('website.localSSL')" value="local"></el-option>
                </el-select>
            </el-form-item>

            <div v-if="formData.importType === 'paste'">
                <el-form-item :label="$t('website.privateKey')" prop="privateKey">
                    <el-input v-model="formData.privateKey" :rows="6" type="textarea" />
                </el-form-item>
                <el-form-item :label="$t('website.certificate')" prop="certificate">
                    <el-input v-model="formData.certificate" :rows="6" type="textarea" />
                </el-form-item>
            </div>

            <div v-if="formData.importType === 'local'">
                <el-form-item :label="$t('website.privateKeyPath')" prop="privateKeyPath">
                    <el-input v-model="formData.privateKeyPath">
                        <template #prepend>
                            <el-button icon="Folder" @click="keyFileRef.acceptParams({ dir: false })" />
                        </template>
                    </el-input>
                </el-form-item>
                <el-form-item :label="$t('website.certificatePath')" prop="certificatePath">
                    <el-input v-model="formData.certificatePath">
                        <template #prepend>
                            <el-button icon="Folder" @click="certFileRef.acceptParams({ dir: false })" />
                        </template>
                    </el-input>
                </el-form-item>
            </div>
        </div>

        <el-form-item :label="' '" v-if="websiteSSL && websiteSSL.id > 0">
            <slot name="ssl-info" :websiteSSL="websiteSSL">
                <WebsiteSSL :websiteSSL="websiteSSL" />
            </slot>
        </el-form-item>

        <el-divider content-position="left">{{ $t('website.SSLProConfig') }}</el-divider>

        <el-form-item :label="$t('website.supportProtocol')" prop="SSLProtocol">
            <el-checkbox-group v-model="formData.SSLProtocol">
                <el-checkbox value="TLSv1.3">TLS 1.3</el-checkbox>
                <el-checkbox value="TLSv1.2">TLS 1.2</el-checkbox>
                <el-checkbox value="TLSv1.1">
                    {{ 'TLS 1.1' + $t('website.notSecurity') }}
                </el-checkbox>
                <el-checkbox value="TLSv1">
                    {{ 'TLS 1.0' + $t('website.notSecurity') }}
                </el-checkbox>
            </el-checkbox-group>
        </el-form-item>

        <el-form-item prop="algorithm" :label="$t('website.encryptionAlgorithm')">
            <el-input type="textarea" :rows="3" v-model.trim="formData.algorithm"></el-input>
        </el-form-item>

        <FileList ref="keyFileRef" @choose="getPrivateKeyPath" />
        <FileList ref="certFileRef" @choose="getCertificatePath" />
    </div>
</template>

<script lang="ts" setup>
import { computed, watch, onMounted, ref } from 'vue';
import { getAccountName } from '@/utils/util';
import WebsiteSSL from '@/views/website/website/components/website-ssl/index.vue';
import { Website } from '@/api/interface/website';
import { listSSL, searchAcmeAccount } from '@/api/modules/website';

interface AcmeAccount {
    id: number;
    email: string;
    type: string;
}

interface Props {
    modelValue: {
        httpsPort: string;
        httpConfig: string;
        hsts: boolean;
        hstsIncludeSubDomains: boolean;
        http3: boolean;
        type: string;
        acmeAccountID: number;
        websiteSSLId?: number;
        importType: string;
        privateKey: string;
        certificate: string;
        privateKeyPath: string;
        certificatePath: string;
        SSLProtocol: string[];
        algorithm: string;
    };
    websiteSSL?: Website.SSL;
    autoLoad?: boolean;
}

interface Emits {
    (e: 'update:modelValue', value: Props['modelValue']): void;
    (e: 'type-change', type: string): void;
    (e: 'ssl-change', ssl: Website.SSL): void;
    (e: 'select-key-file'): void;
    (e: 'select-cert-file'): void;
}

const props = withDefaults(defineProps<Props>(), {
    autoLoad: true,
});

const emit = defineEmits<Emits>();

const acmeAccounts = ref<AcmeAccount[]>([]);
const ssls = ref<Website.SSL[]>([]);
const keyFileRef = ref();
const certFileRef = ref();

const formData = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value),
});

const listAcmeAccount = async () => {
    try {
        const res = await searchAcmeAccount({ page: 1, pageSize: 100 });
        acmeAccounts.value = res.data.items || [];
    } catch (error) {
        console.error('Failed to load ACME accounts:', error);
    }
};

const listSSLs = async () => {
    try {
        const res = await listSSL({
            acmeAccountID: String(formData.value.acmeAccountID),
        });
        ssls.value = res.data || [];

        if (ssls.value.length > 0) {
            let exist = false;
            for (const ssl of ssls.value) {
                if (ssl.id === formData.value.websiteSSLId) {
                    exist = true;
                    break;
                }
            }

            if (!exist) {
                for (const ssl of ssls.value) {
                    if (ssl.pem !== '') {
                        formData.value.websiteSSLId = ssl.id;
                        handleSSLChange(ssl.id);
                        break;
                    }
                }
            } else {
                handleSSLChange(formData.value.websiteSSLId);
            }
        } else {
            formData.value.websiteSSLId = undefined;
        }
    } catch (error) {
        console.error('Failed to load SSL certificates:', error);
    }
};

const handleTypeChange = (type: string) => {
    if (type !== 'existed') {
        formData.value.websiteSSLId = undefined;
    }
    emit('type-change', type);
};

const handleAcmeAccountChange = () => {
    listSSLs();
};

const handleSSLChange = (sslId: number) => {
    const selectedSSL = ssls.value.find((ssl: Website.SSL) => ssl.id === sslId);
    if (selectedSSL && selectedSSL.pem !== '') {
        emit('ssl-change', selectedSSL);
    }
};

watch(
    () => formData.value.acmeAccountID,
    () => {
        listSSLs();
    },
);

const refresh = () => {
    listAcmeAccount();
    listSSLs();
};

const getPrivateKeyPath = (path: string) => {
    formData.value.privateKeyPath = path;
};

const getCertificatePath = (path: string) => {
    formData.value.certificatePath = path;
};

defineExpose({
    refresh,
    listSSLs,
    listAcmeAccount,
});

onMounted(() => {
    if (props.autoLoad) {
        listAcmeAccount();
        listSSLs();
    }
});
</script>

<style scoped>
.input-help {
    margin-left: 8px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}
</style>
