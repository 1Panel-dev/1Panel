<template>
    <DrawerPro v-model="open" :header="$t('commons.button.' + proxy.operate)" @close="handleClose" size="large">
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
                                <div class="form-tip">
                                    <el-icon><QuestionFilled /></el-icon>
                                    <span>{{ $t('website.modifierHelper') }}</span>
                                </div>
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
                                        <el-select v-model="proxy.proxyProtocol" style="width: 100px">
                                            <el-option label="http" value="http://" />
                                            <el-option label="https" value="https://" />
                                            <el-option :label="$t('website.other')" value="" />
                                        </el-select>
                                    </template>
                                </el-input>
                                <div class="form-tip">
                                    <el-icon><QuestionFilled /></el-icon>
                                    <span>{{ $t('website.proxyPassHelper') }}</span>
                                </div>
                            </el-form-item>
                        </el-col>
                        <el-col :span="12">
                            <el-form-item :label="$t('website.proxyHost')" prop="proxyHost">
                                <el-input v-model.trim="proxy.proxyHost" />
                                <div class="form-tip">
                                    <el-icon><QuestionFilled /></el-icon>
                                    <span>{{ $t('website.proxyHostHelper') }}</span>
                                </div>
                            </el-form-item>
                        </el-col>
                    </el-row>

                    <el-divider content-position="left">{{ $t('website.sniSettings') }}</el-divider>

                    <div class="switch-item">
                        <div class="switch-label">
                            <span>{{ $t('website.sni') }}</span>
                            <div class="form-tip">
                                <el-icon><QuestionFilled /></el-icon>
                                <span>{{ $t('website.sniHelper') }}</span>
                            </div>
                        </div>
                        <el-switch v-model="proxy.sni" size="large" />
                    </div>

                    <el-form-item
                        :label="$t('website.proxySslName')"
                        prop="proxySSLName"
                        v-if="proxy.sni"
                        style="margin-top: 16px"
                    >
                        <el-input v-model.trim="proxy.proxySSLName" />
                    </el-form-item>
                </el-tab-pane>

                <el-tab-pane :label="$t('website.advancedSettings')" name="advanced">
                    <el-divider content-position="left">{{ $t('website.cacheSettings') }}</el-divider>

                    <div class="switch-item">
                        <div class="switch-label">
                            <span>{{ $t('website.enableCache') }}</span>
                        </div>
                        <el-switch v-model="proxy.cache" @change="changeCache(proxy.cache)" size="large" />
                    </div>

                    <el-collapse-transition>
                        <div v-if="proxy.cache" style="margin-top: 16px">
                            <el-row :gutter="16">
                                <el-col :span="12">
                                    <el-form-item :label="$t('website.browserCacheTime')" prop="cacheTime">
                                        <el-input v-model.number="proxy.cacheTime" maxlength="15">
                                            <template #append>
                                                <el-select v-model="proxy.cacheUnit" style="width: 80px">
                                                    <el-option
                                                        v-for="(unit, index) in Units"
                                                        :key="index"
                                                        :label="unit.label"
                                                        :value="unit.value"
                                                    />
                                                </el-select>
                                            </template>
                                        </el-input>
                                        <div class="form-tip">
                                            <el-icon><QuestionFilled /></el-icon>
                                            <span>{{ $t('website.browserCacheTimeHelper') }}</span>
                                        </div>
                                    </el-form-item>
                                </el-col>
                                <el-col :span="12">
                                    <el-form-item :label="$t('website.serverCacheTime')" prop="serverCacheTime">
                                        <el-input v-model.number="proxy.serverCacheTime" maxlength="15">
                                            <template #append>
                                                <el-select v-model="proxy.serverCacheUnit" style="width: 80px">
                                                    <el-option
                                                        v-for="(unit, index) in Units"
                                                        :key="index"
                                                        :label="unit.label"
                                                        :value="unit.value"
                                                    />
                                                </el-select>
                                            </template>
                                        </el-input>
                                        <div class="form-tip">
                                            <el-icon><QuestionFilled /></el-icon>
                                            <span>{{ $t('website.serverCacheTimeHelper') }}</span>
                                        </div>
                                    </el-form-item>
                                </el-col>
                            </el-row>
                        </div>
                    </el-collapse-transition>

                    <el-divider content-position="left">{{ $t('website.replace') }}</el-divider>

                    <div class="replace-section">
                        <div v-if="replaces.length === 0" class="empty-state">
                            <el-empty :image-size="60" :description="$t('website.replacedListEmpty')" />
                        </div>
                        <div v-else class="replace-list">
                            <div v-for="(replace, index) in replaces" :key="index" class="replace-item">
                                <div class="replace-item-header">
                                    <span class="replace-index">Rule {{ index + 1 }}</span>
                                    <el-button @click="removeReplace(index)" type="danger" link size="small">
                                        <el-icon><Delete /></el-icon>
                                        {{ $t('commons.button.delete') }}
                                    </el-button>
                                </div>
                                <el-row :gutter="16">
                                    <el-col :span="11">
                                        <el-input
                                            v-model.trim="replace.key"
                                            :placeholder="$t('website.replaced')"
                                            size="default"
                                        >
                                            <template #prepend>Search</template>
                                        </el-input>
                                    </el-col>
                                    <el-col
                                        :span="2"
                                        style="display: flex; align-items: center; justify-content: center"
                                    >
                                        <el-icon size="16"><ArrowRight /></el-icon>
                                    </el-col>
                                    <el-col :span="11">
                                        <el-input
                                            v-model.trim="replace.value"
                                            :placeholder="$t('website.replaceText')"
                                            size="default"
                                        >
                                            <template #prepend>Replace</template>
                                        </el-input>
                                    </el-col>
                                </el-row>
                            </div>
                        </div>
                        <div style="text-align: center; margin-top: 16px">
                            <el-button type="primary" @click="addReplaces" :icon="Plus">
                                {{ $t('website.addReplace') }}
                            </el-button>
                        </div>
                        <div class="form-tip">
                            <el-icon><QuestionFilled /></el-icon>
                            <span>{{ $t('website.replaceHelper') }}</span>
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
import { QuestionFilled, Delete, Plus, ArrowRight } from '@element-plus/icons-vue';

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
    cacheTime: 4,
    cacheUnit: 'h',
    name: '',
    modifier: '^~',
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

<style lang="scss" scoped>
.form-tip {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-top: 4px;
    font-size: 12px;

    .el-icon {
        font-size: 14px;
    }
}

.switch-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;

    .switch-label {
        display: flex;
        flex-direction: column;
        gap: 4px;

        > span {
            font-weight: 500;
        }
    }
}

.replace-section {
    .replace-list {
        .replace-item {
            margin-bottom: 12px;

            .replace-item-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 12px;

                .replace-index {
                    font-size: 14px;
                    font-weight: 500;
                }
            }

            :deep(.el-input-group__prepend) {
                font-weight: 500;
                min-width: 60px;
                text-align: center;
            }
        }
    }
}
</style>
