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
const domainRef = ref();
const useCustomApp = ref(false);

const search = async () => {
    loading.value = true;
    try {
        const res = await getAppStoreConfig();
        if (res.data.defaultDomain != '') {
            config.value.defaultDomain = res.data.defaultDomain;
        }
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const setDefaultDomain = () => {
    domainRef.value.acceptParams({
        domain: config.value.defaultDomain,
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
