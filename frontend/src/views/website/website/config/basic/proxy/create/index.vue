<template>
    <DrawerPro v-model="open" :header="$t('commons.button.' + proxy.operate)" @close="handleClose" size="normal">
        <el-form ref="proxyForm" label-position="top" :model="proxy" :rules="rules" v-loading="loading">
            <el-tabs v-model="activeTab" type="border-card">
                <el-tab-pane :label="$t('website.proxySettings')" name="basic">
                    <el-divider content-position="left">{{ $t('website.proxy') }}</el-divider>

                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model.trim="proxy.name" :disabled="proxy.operate === 'edit'" />
                    </el-form-item>

                    <el-row :gutter="16">
                        <el-col :span="12">
                            <el-form-item :label="$t('website.modifier')" prop="modifier">
                                <el-input v-model.trim="proxy.modifier" />
                                <span class="input-help">{{ $t('website.modifierHelper') }}</span>
                            </el-form-item>
                        </el-col>
                        <el-col :span="12">
                            <el-form-item :label="$t('website.proxyPath')" prop="match">
                                <el-input v-model.trim="proxy.match" />
                            </el-form-item>
                        </el-col>
                    </el-row>

                    <el-row :gutter="16">
                        <el-col :span="12">
                            <el-form-item :label="$t('website.proxyPass')" prop="proxyAddress">
                                <el-input
                                    v-model.trim="proxy.proxyAddress"
                                    :placeholder="$t('website.proxyHelper')"
                                    @blur="getProxyHost"
                                >
                                    <template #prepend>
                                        <el-select v-model="proxy.proxyProtocol" class="!w-24">
                                            <el-option label="http" value="http://" />
                                            <el-option label="https" value="https://" />
                                        </el-select>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('website.proxyPassHelper') }}</span>
                            </el-form-item>
                        </el-col>
                        <el-col :span="12">
                            <el-form-item :label="$t('website.proxyHost')" prop="proxyHost">
                                <el-input v-model.trim="proxy.proxyHost" />
                                <span class="input-help">{{ $t('website.proxyHostHelper') }}</span>
                            </el-form-item>
                        </el-col>
                    </el-row>

                    <template v-if="proxy.proxyProtocol === 'https://'">
                        <el-divider content-position="left">{{ $t('website.sniSettings') }}</el-divider>

                        <div class="flex justify-between items-center py-3">
                            <div class="flex flex-col gap-1">
                                <span class="font-medium">{{ $t('website.sni') }}</span>
                                <span class="input-help">{{ $t('website.sniHelper') }}</span>
                            </div>
                            <el-switch v-model="proxy.sni" size="large" />
                        </div>

                        <el-form-item
                            :label="$t('website.proxySslName')"
                            prop="proxySSLName"
                            v-if="proxy.sni"
                            class="mt-4"
                        >
                            <el-input v-model.trim="proxy.proxySSLName" />
                        </el-form-item>
                    </template>
                </el-tab-pane>

                <el-tab-pane :label="$t('website.advancedSettings')" name="advanced">
                    <el-divider content-position="left">{{ $t('website.cacheSettings') }}</el-divider>

                    <div class="flex justify-between items-center py-3">
                        <div class="flex flex-col gap-1">
                            <span class="font-medium">{{ $t('website.enableCache') }}</span>
                        </div>
                        <el-switch v-model="proxy.cache" @change="changeCache(proxy.cache)" size="large" />
                    </div>

                    <el-collapse-transition>
                        <div v-if="proxy.cache" class="mt-4">
                            <el-row :gutter="16">
                                <el-col :span="12">
                                    <el-form-item :label="$t('website.browserCacheTime')" prop="cacheTime">
                                        <el-input v-model.number="proxy.cacheTime" maxlength="15">
                                            <template #append>
                                                <el-select v-model="proxy.cacheUnit" class="!w-24">
                                                    <el-option
                                                        v-for="(unit, index) in Units"
                                                        :key="index"
                                                        :label="unit.label"
                                                        :value="unit.value"
                                                    />
                                                </el-select>
                                            </template>
                                        </el-input>
                                        <span class="input-help">{{ $t('website.browserCacheTimeHelper') }}</span>
                                    </el-form-item>
                                </el-col>
                                <el-col :span="12">
                                    <el-form-item :label="$t('website.serverCacheTime')" prop="serverCacheTime">
                                        <el-input v-model.number="proxy.serverCacheTime" maxlength="15">
                                            <template #append>
                                                <el-select v-model="proxy.serverCacheUnit" class="!w-24">
                                                    <el-option
                                                        v-for="(unit, index) in Units"
                                                        :key="index"
                                                        :label="unit.label"
                                                        :value="unit.value"
                                                    />
                                                </el-select>
                                            </template>
                                        </el-input>
                                        <span class="input-help">{{ $t('website.serverCacheTimeHelper') }}</span>
                                    </el-form-item>
                                </el-col>
                            </el-row>
                        </div>
                    </el-collapse-transition>

                    <el-divider content-position="left">{{ $t('website.replace') }}</el-divider>

                    <div>
                        <div v-for="(replace, index) in replaces" :key="index" class="mb-3">
                            <el-row :gutter="16">
                                <el-col :span="10">
                                    <el-input
                                        v-model.trim="replace.key"
                                        :placeholder="$t('website.replaced')"
                                        size="default"
                                    ></el-input>
                                </el-col>
                                <el-col :span="10">
                                    <el-input
                                        v-model.trim="replace.value"
                                        :placeholder="$t('website.replaceText')"
                                        size="default"
                                    ></el-input>
                                </el-col>
                                <el-col :span="2">
                                    <el-button
                                        @click="removeReplace(index)"
                                        type="danger"
                                        link
                                        size="small"
                                        class="mt-1"
                                    >
                                        <el-icon><Delete /></el-icon>
                                        {{ $t('commons.button.delete') }}
                                    </el-button>
                                </el-col>
                            </el-row>
                        </div>
                        <div class="mt-4">
                            <el-button type="primary" @click="addReplaces" :icon="Plus">
                                {{ $t('website.addReplace') }}
                            </el-button>
                            <span class="input-help mt-4">{{ $t('website.replaceHelper') }}</span>
                        </div>
                    </div>
                </el-tab-pane>
            </el-tabs>
        </el-form>
        <template #footer>
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="submit(proxyForm)" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { operateProxyConfig } from '@/api/modules/website';
import { checkNumberRange, Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Website } from '@/api/interface/website';
import { Units } from '@/global/mimetype';
import { isDomain } from '@/utils/util';
import { Delete, Plus } from '@element-plus/icons-vue';

const proxyForm = ref<FormInstance>();
const rules = ref({
    name: [Rules.requiredInput, Rules.appName],
    match: [Rules.requiredInput],
    cacheTime: [Rules.requiredInput, checkNumberRange(1, 65535)],
    proxyPass: [Rules.requiredInput],
    proxyHost: [Rules.requiredInput],
    proxyAddress: [Rules.requiredInput],
});
const open = ref(false);
const loading = ref(false);
const activeTab = ref('basic');

const initData = (): Website.ProxyConfig => ({
    id: 0,
    operate: 'create',
    enable: true,
    cache: false,
    cacheTime: 1,
    cacheUnit: 'm',
    name: '',
    modifier: '',
    match: '/',
    proxyPass: 'http://127.0.0.1:8080',
    proxyHost: '$host',
    filePath: '',
    replaces: {},
    proxyAddress: '',
    proxyProtocol: 'http://',
    sni: false,
    proxySSLName: '',
    serverCacheTime: 10,
    serverCacheUnit: 'm',
});
let proxy = ref(initData());
const replaces = ref<any>([]);
const em = defineEmits(['close']);
const handleClose = () => {
    proxyForm.value?.resetFields();
    open.value = false;
    em('close', false);
};

const acceptParams = (proxyParam: Website.ProxyConfig) => {
    replaces.value = [];
    proxy.value = proxyParam;
    activeTab.value = 'basic';

    const res = getProtocolAndHost(proxyParam.proxyPass);
    if (res != null) {
        proxy.value.proxyProtocol = res.protocol;
        proxy.value.proxyAddress = res.host;
    } else {
        proxy.value.proxyProtocol = 'http://';
    }

    open.value = true;
    if (proxy.value.replaces) {
        for (const key in proxy.value.replaces) {
            replaces.value.push({ key: key, value: proxy.value.replaces[key] });
        }
    }
};

const changeCache = (cache: boolean) => {
    if (cache) {
        proxy.value.cacheTime = 1;
        proxy.value.cacheUnit = 'm';
        proxy.value.serverCacheTime = 10;
        proxy.value.serverCacheUnit = 'm';
    } else {
        proxy.value.cacheTime = 0;
        proxy.value.cacheUnit = '';
        proxy.value.serverCacheTime = 0;
        proxy.value.serverCacheUnit = '';
    }
};

const addReplaces = () => {
    replaces.value.push({ key: '', value: '' });
};

const removeReplace = (index: number) => {
    replaces.value.splice(index, 1);
};

const getProxyHost = () => {
    if (isDomain(proxy.value.proxyAddress)) {
        proxy.value.proxyHost = proxy.value.proxyAddress;
    } else {
        proxy.value.proxyHost = '$host';
    }
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }

        proxy.value.replaces = {};
        if (replaces.value.length > 0) {
            let keyMap = new Map();
            for (const rep of replaces.value) {
                if (keyMap.get(rep.key) != undefined) {
                    MsgError(i18n.global.t('website.replacedErr2'));
                    return;
                }
                keyMap.set(rep.key, '');
                if (rep.key === '') {
                    MsgError(i18n.global.t('website.replacedErr'));
                    return;
                }
                proxy.value.replaces[rep.key] = rep.value;
            }
        }
        loading.value = true;
        proxy.value.proxyPass = proxy.value.proxyProtocol + proxy.value.proxyAddress;
        operateProxyConfig(proxy.value)
            .then(() => {
                if (proxy.value.operate == 'create') {
                    MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                } else {
                    MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                }
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const getProtocolAndHost = (url: string): { protocol: string; host: string } | null => {
    if (url.startsWith('https://')) {
        return { protocol: 'https://', host: url.replaceAll('https://', '') };
    }
    if (url.startsWith('http://')) {
        return { protocol: 'http://', host: url.replaceAll('http://', '') };
    }
    return { protocol: '', host: url };
};

defineExpose({
    acceptParams,
});
</script>
