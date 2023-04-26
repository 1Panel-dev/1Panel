<template>
    <el-drawer v-model="open" :close-on-click-modal="false" size="40%" :before-close="handleClose">
        <template #header>
            <DrawerHeader :header="$t('website.' + proxy.operate + 'Proxy')" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form ref="proxyForm" label-position="top" :model="proxy" :rules="rules">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model.trim="proxy.name" :disabled="proxy.operate === 'edit'"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.modifier')" prop="modifier">
                        <el-input v-model.trim="proxy.modifier"></el-input>
                        <div>
                            <span class="input-help">{{ $t('website.modifierHelper') }}</span>
                        </div>
                    </el-form-item>
                    <el-form-item :label="$t('website.proxyPath')" prop="match">
                        <el-input v-model.trim="proxy.match"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.enableCache')" prop="cache">
                        <el-switch v-model="proxy.cache" @change="changeCache(proxy.cache)"></el-switch>
                    </el-form-item>
                    <el-form-item :label="$t('website.cacheTime')" prop="cacheTime" v-if="proxy.cache">
                        <el-input v-model.number="proxy.cacheTime" maxlength="15">
                            <template #append>
                                <el-select v-model="proxy.cacheUnit" style="width: 100px">
                                    <el-option
                                        v-for="(unit, index) in Units"
                                        :key="index"
                                        :label="unit.label"
                                        :value="unit.value"
                                    ></el-option>
                                </el-select>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-row :gutter="10">
                        <el-col :span="12">
                            <el-form-item :label="$t('website.proxyPass')" prop="proxyPass">
                                <el-input v-model.trim="proxy.proxyPass"></el-input>
                                <div>
                                    <span class="input-help">{{ $t('website.proxyPassHelper') }}</span>
                                </div>
                            </el-form-item>
                        </el-col>
                        <el-col :span="12">
                            <el-form-item :label="$t('website.proxyHost')" prop="proxyHost">
                                <el-input v-model.trim="proxy.proxyHost"></el-input>
                                <div>
                                    <span class="input-help">{{ $t('website.proxyHostHelper') }}</span>
                                </div>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-form-item :label="$t('website.replace')">
                        <div style="width: 100%" v-for="(replace, index) in replaces" :key="index">
                            <el-row :gutter="10">
                                <el-col :span="10">
                                    <el-input
                                        v-model.trim="replace.key"
                                        :placeholder="$t('website.replaced')"
                                    ></el-input>
                                </el-col>
                                <el-col :span="10">
                                    <el-input
                                        v-model.trim="replace.value"
                                        :placeholder="$t('website.replaceText')"
                                    ></el-input>
                                </el-col>
                                <el-col :span="2">
                                    <el-button link @click="removeReplace(index)" type="danger">
                                        {{ $t('commons.button.delete') }}
                                    </el-button>
                                </el-col>
                            </el-row>
                        </div>
                    </el-form-item>
                    <el-form-item>
                        <el-button type="primary" @click="addReplaces" :disabled="replaces.length >= 5">
                            {{ $t('website.addReplace') }}
                        </el-button>
                        <div>
                            <span class="input-help">{{ $t('website.replacementHelper') }}</span>
                        </div>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(proxyForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { OperateProxyConfig } from '@/api/modules/website';
import { checkNumberRange, Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Website } from '@/api/interface/website';
import { Units } from '@/global/mimetype';

const proxyForm = ref<FormInstance>();
const rules = ref({
    name: [Rules.requiredInput, Rules.appName],
    match: [Rules.requiredInput],
    modifier: [Rules.requiredInput],
    cacheTime: [Rules.requiredInput, checkNumberRange(1, 65535)],
    proxyPass: [Rules.requiredInput],
    proxyHost: [Rules.requiredInput],
});
const open = ref(false);
const loading = ref(false);

const initData = (): Website.ProxyConfig => ({
    id: 0,
    operate: 'create',
    enable: true,
    cache: false,
    cacheTime: 1,
    cacheUnit: 'm',
    name: '',
    modifier: '^~',
    match: '/',
    proxyPass: 'http://',
    proxyHost: '$host',
    filePath: '',
    replaces: {},
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
    } else {
        proxy.value.cacheTime = 0;
        proxy.value.cacheUnit = '';
    }
};

const addReplaces = () => {
    replaces.value.push({ key: '', value: '' });
};

const removeReplace = (index: number) => {
    replaces.value.splice(index, 1);
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
        OperateProxyConfig(proxy.value)
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

defineExpose({
    acceptParams,
});
</script>
