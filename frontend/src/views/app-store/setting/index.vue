<template>
    <LayoutContent :title="$t('commons.button.set')">
        <template #main>
            <el-form
                :model="config"
                label-position="left"
                label-width="180px"
                class="ml-2.5"
                v-loading="loading"
                :rules="rules"
                ref="configForm"
            >
                <el-row>
                    <el-col :xs="24" :sm="20" :md="15" :lg="12" :xl="12">
                        <el-form-item :label="$t('app.defaultWebDomain')" prop="defaultDomain">
                            <el-input v-model="config.defaultDomain" disabled>
                                <template #prepend>
                                    <el-select v-model="protocol" placeholder="Select" class="p-w-100" disabled>
                                        <el-option label="HTTP" value="http://" />
                                        <el-option label="HTTPS" value="https://" />
                                    </el-select>
                                </template>
                                <template #append>
                                    <el-button @click="setDefaultDomain()" icon="Setting">
                                        {{ $t('commons.button.set') }}
                                    </el-button>
                                </template>
                            </el-input>
                            <span class="input-help">{{ $t('app.defaultWebDomainHepler') }}</span>
                        </el-form-item>
                        <CustomSetting v-if="globalStore.isProductPro" />
                    </el-col>
                </el-row>
            </el-form>
        </template>
    </LayoutContent>
    <DefaultDomain ref="domainRef" @close="search" />
</template>

<script setup lang="ts">
import { getAppStoreConfig, getCurrentNodeCustomAppConfig } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import { FormRules } from 'element-plus';
import CustomSetting from '@/xpack/views/appstore/index.vue';
import DefaultDomain from './default-domain/index.vue';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const rules = ref<FormRules>({
    defaultDomain: [Rules.domainOrIP],
});
const config = ref({
    defaultDomain: '',
});
const loading = ref(false);
const configForm = ref();
const protocol = ref('http://');
const domainRef = ref();
const useCustomApp = ref(false);

function getUrl(url: string) {
    const regex = /^(https?:\/\/)(.*)/;
    const match = url.match(regex);
    if (match) {
        const protocol = match[1];
        const remainder = match[2];
        return {
            protocol: protocol,
            remainder: remainder,
        };
    } else {
        return null;
    }
}

const search = async () => {
    loading.value = true;
    try {
        const res = await getAppStoreConfig();
        if (res.data.defaultDomain != '') {
            const url = getUrl(res.data.defaultDomain);
            if (url) {
                config.value.defaultDomain = url.remainder;
                protocol.value = url.protocol;
            }
        }
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const setDefaultDomain = () => {
    domainRef.value.acceptParams({
        domain: config.value.defaultDomain,
        protocol: protocol.value,
    });
};

const getNodeConfig = async () => {
    if (globalStore.isMasterProductPro) {
        return;
    }
    const res = await getCurrentNodeCustomAppConfig();
    if (res && res.data) {
        useCustomApp.value = res.data.status === 'enable';
    }
};

onMounted(() => {
    search();
    getNodeConfig();
});
</script>
