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
                        <el-form-item :label="$t('app.uninstallDeleteBackup')" prop="uninstallDeleteBackup">
                            <el-switch
                                v-model="config.uninstallDeleteBackup"
                                active-value="True"
                                inactive-value="False"
                                :loading="loading"
                                @change="updateConfig('UninstallDeleteBackup', config.uninstallDeleteBackup)"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('app.uninstallDeleteImage')" prop="uninstallDeleteImage">
                            <el-switch
                                v-model="config.uninstallDeleteImage"
                                active-value="True"
                                inactive-value="False"
                                :loading="loading"
                                @change="updateConfig('UninstallDeleteImage', config.uninstallDeleteImage)"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('app.upgradeBackup')" prop="upgradeBackup">
                            <el-switch
                                v-model="config.upgradeBackup"
                                active-value="True"
                                inactive-value="False"
                                :loading="loading"
                                @change="updateConfig('UpgradeBackup', config.upgradeBackup)"
                            />
                        </el-form-item>
                        <CustomSetting v-if="globalStore.isProductPro" />
                    </el-col>
                </el-row>
            </el-form>
        </template>
    </LayoutContent>
</template>

<script setup lang="ts">
import { getAppStoreConfig, getCurrentNodeCustomAppConfig, updateAppStoreConfig } from '@/api/modules/app';
import { FormRules } from 'element-plus';
import CustomSetting from '@/xpack/views/appstore/index.vue';
import { GlobalStore } from '@/store';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const globalStore = GlobalStore();
const rules = ref<FormRules>({});
const config = ref({
    uninstallDeleteImage: '',
    uninstallDeleteBackup: '',
    upgradeBackup: '',
});
const loading = ref(false);
const configForm = ref();
const useCustomApp = ref(false);
const isInitializing = ref(true);

const search = async () => {
    loading.value = true;
    try {
        const res = await getAppStoreConfig();
        if (res && res.data) {
            isInitializing.value = true;
            config.value = res.data;
            setTimeout(() => {
                isInitializing.value = false;
            }, 0);
        }
    } catch (error) {
    } finally {
        loading.value = false;
    }
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

const updateConfig = async (scope: string, value: string) => {
    if (isInitializing.value) {
        return;
    }
    loading.value = true;
    try {
        const req = {
            scope: scope,
            status: value,
        };
        await updateAppStoreConfig(req);
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        search();
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

onMounted(() => {
    search();
    getNodeConfig();
});
</script>
