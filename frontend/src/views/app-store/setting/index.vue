<template>
    <LayoutContent :title="$t('commons.button.set')">
        <template #main>
            <el-form
                :model="config"
                label-position="left"
                label-width="150px"
                class="ml-2.5"
                v-loading="loading"
                :rules="rules"
                ref="configForm"
            >
                <el-row>
                    <el-col :xs="24" :sm="20" :md="15" :lg="12" :xl="12">
                        <el-form-item :label="$t('app.defaultWebDomain')" prop="defaultDomain">
                            <el-input v-model="config.defaultDomain">
                                <template #prepend>
                                    <el-select v-model="protocol" placeholder="Select" class="p-w-100">
                                        <el-option label="HTTP" value="http://" />
                                        <el-option label="HTTPS" value="https://" />
                                    </el-select>
                                </template>
                            </el-input>
                            <span class="input-help">{{ $t('app.defaultWebDomainHepler') }}</span>
                        </el-form-item>
                        <el-form-item>
                            <el-button type="primary" :disabled="loading" @click="submit()">
                                {{ $t('commons.button.confirm') }}
                            </el-button>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
        </template>
    </LayoutContent>
</template>

<script setup lang="ts">
import { GetAppStoreConfig, UpdateAppStoreConfig } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormRules } from 'element-plus';

const rules = ref<FormRules>({
    defaultDomain: [Rules.domainOrIP],
});
const config = ref({
    defaultDomain: '',
});
const loading = ref(false);
const configForm = ref();
const protocol = ref('http://');

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
        const res = await GetAppStoreConfig();
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

const submit = async () => {
    if (!configForm.value) return;
    await configForm.value.validate(async (valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        try {
            let defaultDomain = '';
            if (config.value.defaultDomain) {
                defaultDomain = protocol.value + config.value.defaultDomain;
            }
            const req = {
                defaultDomain: defaultDomain,
            };
            await UpdateAppStoreConfig(req);
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        } catch (error) {
        } finally {
            loading.value = false;
            search();
        }
    });
};

onMounted(() => {
    search();
});
</script>
