<template>
    <div v-loading="loading">
        <el-tabs type="border-card" v-model="tabIndex">
            <el-tab-pane :label="$t('website.changeVersion')">
                <el-form label-position="right" label-width="100px" v-if="tabIndex == '0'">
                    <el-form-item v-if="website.type === 'static'">
                        <el-text type="info">
                            {{ $t('website.staticChangePHPHelper') }}
                        </el-text>
                    </el-form-item>
                    <el-form-item :label="$t('website.changeVersion')">
                        <el-row :gutter="20">
                            <el-col :span="20">
                                <el-select v-model="versionReq.runtimeID" class="p-w-200">
                                    <el-option :key="-1" :label="$t('website.static')" :value="0"></el-option>
                                    <el-option
                                        v-for="(item, index) in versions"
                                        :key="index"
                                        :label="item.label"
                                        :value="item.value"
                                    ></el-option>
                                </el-select>
                            </el-col>
                            <el-col :span="4">
                                <el-button
                                    type="primary"
                                    @click="submit()"
                                    :disabled="versionReq.runtimeID === oldRuntimeID"
                                >
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </el-col>
                        </el-row>
                    </el-form-item>
                </el-form>
            </el-tab-pane>
            <el-tab-pane
                :label="$t('website.openBaseDir')"
                v-if="website.type === 'runtime' && website.runtimeType == 'php'"
            >
                <el-form label-position="right" label-width="100px" v-if="tabIndex == '1'">
                    <el-form-item :label="$t('website.openBaseDir')">
                        <el-switch v-model="openBaseDir" @change="operateCrossSite"></el-switch>
                        <span class="input-help">{{ $t('website.openBaseDirHelper') }}</span>
                    </el-form-item>
                </el-form>
            </el-tab-pane>
            <el-tab-pane :label="'Composer'" v-if="website.type === 'runtime' && website.runtimeType == 'php'">
                <Composer :websiteID="id" v-if="tabIndex == '2'" />
            </el-tab-pane>
        </el-tabs>
    </div>
</template>

<script setup lang="ts">
import { SearchRuntimes } from '@/api/modules/runtime';
import { onMounted, reactive, ref } from 'vue';
import { Runtime } from '@/api/interface/runtime';
import { Website } from '@/api/interface/website';
import { changePHPVersion, getWebsite, operateCrossSiteAccess } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import Composer from './composer/index.vue';
const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const runtimeReq = reactive<Runtime.RuntimeReq>({ page: 1, pageSize: 200, type: 'php' });
const versionReq = reactive<Website.PHPVersionChange>({
    websiteID: undefined,
    runtimeID: undefined,
});
const versions = ref([]);
const loading = ref(false);
const oldRuntimeID = ref(0);
const website = ref({
    type: '',
    openBaseDir: false,
    runtimeType: '',
});
const openBaseDir = ref(false);
const tabIndex = ref('0');

const getRuntimes = async () => {
    try {
        loading.value = true;
        const res = await SearchRuntimes(runtimeReq);
        const items = res.data.items || [];
        for (const item of items) {
            versions.value.push({
                value: item.id,
                label: item.name + ' [' + i18n.global.t('app.version') + ':' + item.params['PHP_VERSION'] + ']',
            });
        }
    } catch (error) {}
    loading.value = false;
};

const submit = async () => {
    try {
        ElMessageBox.confirm(i18n.global.t('website.changePHPVersionWarn'), i18n.global.t('website.changeVersion'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(async () => {
            loading.value = true;
            try {
                await changePHPVersion(versionReq);
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                getWebsiteDetail();
            } catch (error) {}
            loading.value = false;
        });
    } catch (error) {}
};

const getWebsiteDetail = async () => {
    const res = await getWebsite(props.id);
    versionReq.runtimeID = res.data.runtimeID;
    oldRuntimeID.value = res.data.runtimeID;
    website.value = res.data;
    openBaseDir.value = res.data.openBaseDir || false;
};

const operateCrossSite = async () => {
    try {
        await operateCrossSiteAccess({
            websiteID: props.id,
            operation: openBaseDir.value ? 'Enable' : 'Disable',
        });
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        getWebsiteDetail();
    } catch (error) {}
};

onMounted(() => {
    versionReq.websiteID = props.id;
    getWebsiteDetail();
    getRuntimes();
});
</script>
