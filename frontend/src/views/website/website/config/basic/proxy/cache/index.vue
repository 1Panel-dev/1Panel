<template>
    <DrawerPro v-model="open" :header="$t('website.proxyCache')" size="normal" @close="handleClose">
        <el-form
            v-loading="loading"
            @submit.prevent
            ref="proxyForm"
            label-position="top"
            :model="req"
            :rules="rules"
            :validate-on-rule-change="false"
        >
            <el-form-item :label="$t('commons.button.start')" prop="open">
                <el-switch v-model="req.open" :disabled="hasCache"></el-switch>
                <span class="input-help" v-if="hasCache">{{ $t('website.cacheWarn') }}</span>
            </el-form-item>
            <el-form-item :label="$t('website.cacheLimit')" prop="cacheLimit">
                <el-input v-model.number="req.cacheLimit" class="p-w-200">
                    <template #append>
                        <el-select v-model="req.cacheLimitUnit" class="p-w-100">
                            <el-option
                                v-for="(unit, index) in sizeUnits"
                                :key="index"
                                :label="unit.label"
                                :value="unit.value"
                            ></el-option>
                        </el-select>
                    </template>
                </el-input>
                <span class="input-help">{{ $t('website.cacheLimitHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('website.shareCahe')" prop="shareCache">
                <el-input v-model.number="req.shareCache" class="p-w-200">
                    <template #append>
                        <el-select v-model="req.shareCacheUnit" class="p-w-100">
                            <el-option
                                v-for="(unit, index) in sizeUnits"
                                :key="index"
                                :label="unit.label"
                                :value="unit.value"
                            ></el-option>
                        </el-select>
                    </template>
                </el-input>
                <span class="input-help">{{ $t('website.shareCaheHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('website.cacheExpire')" prop="cacheExpire">
                <el-input v-model.number="req.cacheExpire" class="p-w-200">
                    <template #append>
                        <el-select v-model="req.cacheExpireUnit" class="p-w-100">
                            <el-option
                                v-for="(unit, index) in Units"
                                :key="index"
                                :label="unit.label"
                                :value="unit.value"
                            ></el-option>
                        </el-select>
                    </template>
                </el-input>
                <span class="input-help">{{ $t('website.cacheExpireJHelper') }}</span>
            </el-form-item>
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
import { ref } from 'vue';
import { Units, sizeUnits } from '@/global/mimetype';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { FormInstance } from 'element-plus';
import { getCacheConfig, updateCacheConfig } from '@/api/modules/website';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const open = ref(false);
const loading = ref(false);
const proxyForm = ref<FormInstance>();
const hasCache = ref(false);

const req = reactive({
    open: false,
    cacheLimit: 1,
    cacheLimitUnit: 'g',
    shareCache: 5,
    shareCacheUnit: 'm',
    cacheExpire: 24,
    cacheExpireUnit: 'h',
    websiteID: 0,
});

const rules = {
    cacheLimit: [Rules.requiredInput, checkNumberRange(0, 9999)],
    shareCache: [Rules.requiredInput, checkNumberRange(0, 9999)],
    cacheExpire: [Rules.requiredInput, checkNumberRange(0, 9999)],
};

const handleClose = () => {
    open.value = false;
};

const acceptParams = (websiteID: number, cache: boolean) => {
    req.websiteID = websiteID;
    hasCache.value = cache;
    get();
    open.value = true;
};

const get = async () => {
    try {
        const res = await getCacheConfig(req.websiteID);
        req.open = res.data.open;
        if (req.open) {
            req.cacheLimit = res.data.cacheLimit;
            req.cacheLimitUnit = res.data.cacheLimitUnit;
            req.shareCache = res.data.shareCache;
            req.shareCacheUnit = res.data.shareCacheUnit;
            req.cacheExpire = res.data.cacheExpire;
            req.cacheExpireUnit = res.data.cacheExpireUnit;
        }
    } catch (error) {}
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate(async (valid) => {
        if (!valid) {
            return;
        }
        try {
            await updateCacheConfig(req);
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            handleClose();
        } catch (error) {}
    });
};

defineExpose({
    acceptParams,
});
</script>
