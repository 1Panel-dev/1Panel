<template>
    <el-row :gutter="20" v-loading="loading">
        <el-col :xs="24" :sm="18" :md="18" :lg="14" :xl="14">
            <el-form
                :model="form"
                :rules="rules"
                ref="leechRef"
                label-position="right"
                label-width="150px"
                class="moblie-form"
            >
                <el-form-item :label="$t('website.extends')" prop="extends" class="mt-2">
                    <el-input v-model="form.extends" class="p-w-600"></el-input>
                </el-form-item>
                <el-divider content-position="left">{{ $t('website.antiLeech') }}</el-divider>
                <el-form-item :label="$t('website.enableOrNot')" prop="enable">
                    <el-switch v-model="form.enable" @change="changeEnable"></el-switch>
                </el-form-item>
                <template v-if="form.enable">
                    <el-form-item :label="$t('website.accessDomain')" prop="domains">
                        <div class="domain-list-container">
                            <div v-for="(_, index) in domainList" :key="index" class="flex items-center mb-2">
                                <el-input
                                    v-model="domainList[index]"
                                    @input="updateDomainsString"
                                    class="flex-1 mr-2"
                                ></el-input>
                                <el-button
                                    type="danger"
                                    size="small"
                                    :icon="Delete"
                                    @click="removeDomain(index)"
                                    v-if="domainList.length > 1"
                                ></el-button>
                            </div>
                            <el-button type="primary" size="small" :icon="Plus" @click="addDomain" plain>
                                {{ $t('commons.button.add') }}
                            </el-button>
                        </div>
                    </el-form-item>
                    <el-row :gutter="15">
                        <el-col>
                            <el-form-item :label="$t('website.noneRef')" prop="noneRef">
                                <el-switch v-model="form.noneRef" />
                            </el-form-item>
                        </el-col>

                        <el-col>
                            <el-form-item :label="$t('website.blockedRef')" prop="blocked">
                                <el-switch v-model="form.blocked" />
                                <span class="input-help">
                                    {{ $t('website.leechSpecialValidHelper') }}
                                </span>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-form-item :label="$t('website.leechReturn')" prop="return">
                        <el-select v-model="form.return" class="p-w-200">
                            <el-option
                                v-for="option in returnOptions"
                                :key="option.value"
                                :label="option.label"
                                :value="option.value"
                            ></el-option>
                        </el-select>
                        <span class="input-help">
                            {{ $t('website.leechInvalidReturnHelper') }}
                        </span>
                    </el-form-item>
                </template>
                <el-divider content-position="left">{{ $t('website.leechcacheControl') }}</el-divider>
                <el-form-item :label="$t('website.browserCache')" prop="cache">
                    <el-switch v-model="form.cache" />
                </el-form-item>
                <el-form-item :label="$t('website.cacheTime')" prop="cacheTime" v-if="form.cache">
                    <el-input v-model.number="form.cacheTime" maxlength="15" class="p-w-300">
                        <template #append>
                            <el-select v-model="form.cacheUint" class="w-s-button p-w-100">
                                <el-option
                                    v-for="(unit, index) in Units"
                                    :key="index"
                                    :label="unit.label"
                                    :value="unit.value"
                                ></el-option>
                            </el-select>
                        </template>
                    </el-input>
                    <span class="input-help">{{ $t('website.browserCacheTimeHelper') }}</span>
                </el-form-item>
                <el-form-item :label="$t('website.logEnableControl')" prop="logEnable" v-if="form.cache || form.enable">
                    <el-switch v-model="form.logEnable" />
                    <span class="input-help">{{ $t('website.leechlogControlHelper') }}</span>
                </el-form-item>
                <div class="flex items-center gap-4 mt-2">
                    <el-button type="primary" @click="submit(leechRef, form.enable)" :disabled="loading">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </div>
            </el-form>
        </el-col>
    </el-row>
</template>

<script setup lang="ts">
import { getAntiLeech, listDomains, updateAntiLeech } from '@/api/modules/website';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { FormInstance } from 'element-plus';
import { computed, onMounted, reactive } from 'vue';
import { ref } from 'vue';
import { Units } from '@/global/mimetype';
import { MsgSuccess, MsgError } from '@/utils/message';
import i18n from '@/lang';
import { Plus, Delete } from '@element-plus/icons-vue';

const loading = ref(false);
const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const id = computed(() => {
    return props.id;
});
const leechRef = ref<FormInstance>();
const resData = ref({
    enable: false,
});
const returnOptions = [
    { label: '400 Bad Request', value: '400' },
    { label: '403 Forbidden', value: '403' },
    { label: '404 Not Found', value: '404' },
];

const domainList = ref(['']);

const form = reactive({
    enable: false,
    cache: false,
    cacheTime: 30,
    cacheUint: 'd',
    extends: 'js,css,png,jpg,jpeg,gif,webp,webm,avif,ico,bmp,swf,eot,svg,ttf,woff,woff2',
    return: '404',
    domains: '',
    noneRef: true,
    logEnable: false,
    blocked: false,
    serverNames: [],
    websiteID: 0,
});

const rules = ref({
    extends: [Rules.requiredInput, Rules.leechExts],
    cacheTime: [Rules.requiredInput, checkNumberRange(1, 65535)],
    return: [Rules.requiredInput],
    domains: [Rules.requiredInput],
});

const addDomain = () => {
    domainList.value.push('');
};

const removeDomain = (index: number) => {
    domainList.value.splice(index, 1);
    updateDomainsString();
};

const updateDomainsString = () => {
    form.domains = domainList.value.filter((domain) => domain.trim() !== '').join('\n');
};

const initDomainList = (domainsStr: string) => {
    if (domainsStr) {
        domainList.value = domainsStr.split('\n').filter((domain) => domain.trim() !== '');
        if (domainList.value.length === 0) {
            domainList.value = [''];
        }
    } else {
        domainList.value = [''];
    }
};

const changeEnable = (enable: boolean) => {
    if (enable) {
        listDomains(id.value)
            .then((res) => {
                const domains = res.data || [];
                let serverNameStr = '';
                for (const param of domains) {
                    serverNameStr = serverNameStr + param.domain + '\n';
                }
                form.domains = serverNameStr;
                initDomainList(serverNameStr);
            })
            .finally(() => {});
    }
};

const search = async () => {
    loading.value = true;
    const res = await getAntiLeech({ websiteID: id.value });
    loading.value = false;
    if (!res.data.enable && !res.data.cache) {
        return;
    }
    resData.value = res.data;
    form.blocked = res.data.blocked;
    form.cache = res.data.cache;
    form.enable = res.data.enable;
    if (res.data.cache) {
        form.cacheTime = res.data.cacheTime;
        form.cacheUint = res.data.cacheUint;
    }
    form.extends = res.data.extends;
    form.return = res.data.return;
    form.logEnable = res.data.logEnable;
    form.noneRef = res.data.noneRef;

    const serverNames = res.data.serverNames;
    let serverNameStr = '';
    for (const param of serverNames) {
        serverNameStr = serverNameStr + param + '\n';
    }
    form.domains = serverNameStr;
    initDomainList(serverNameStr);
};

const submit = async (formEl: FormInstance | undefined, enable: boolean) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        update(enable);
    });
};

const update = async (enable: boolean) => {
    if (enable) {
        updateDomainsString();
        form.serverNames = form.domains.split('\n');
    }
    if (!checkReturn()) {
        return;
    }
    form.enable = enable;
    loading.value = true;
    form.websiteID = id.value;
    await updateAntiLeech(form)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const checkReturn = (): boolean => {
    let returns = form.return.split(' ');
    if (returns[0]) {
        if (isHttpStatusCode(returns[0])) {
            return true;
        } else {
            MsgError(i18n.global.t('website.leechReturnError'));
            return false;
        }
    } else {
        return false;
    }
};

function isHttpStatusCode(input: string): boolean {
    const statusCodeRegex = /^[1-5][0-9]{2}$/;
    return statusCodeRegex.test(input);
}

onMounted(() => {
    search();
});
</script>
