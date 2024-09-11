<template>
    <div v-loading="loading">
        <el-row>
            <el-col :xs="20" :sm="12" :md="10" :lg="10" :xl="8" :offset="1">
                <el-form>
                    <div v-if="website.type === 'static'">
                        <el-text type="info">{{ $t('website.staticChangePHPHelper') }}</el-text>
                    </div>
                    <el-form-item :label="$t('website.changeVersion')">
                        <el-select v-model="versionReq.runtimeID" style="width: 100%">
                            <el-option :key="-1" :label="$t('website.static')" :value="0"></el-option>
                            <el-option
                                v-for="(item, index) in versions"
                                :key="index"
                                :label="item.label"
                                :value="item.value"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item>
                        <el-button type="primary" @click="submit()" :disabled="versionReq.runtimeID === oldRuntimeID">
                            {{ $t('commons.button.save') }}
                        </el-button>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
    </div>
</template>

<script setup lang="ts">
import { SearchRuntimes } from '@/api/modules/runtime';
import { onMounted, reactive, ref } from 'vue';
import { Runtime } from '@/api/interface/runtime';
import { Website } from '@/api/interface/website';
import { ChangePHPVersion } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
const props = defineProps({
    website: {
        type: Object,
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

const getRuntimes = async () => {
    try {
        loading.value = true;
        const res = await SearchRuntimes(runtimeReq);
        const items = res.data.items || [];
        for (const item of items) {
            versions.value.push({
                value: item.id,
                label: item.name + ' [' + i18n.global.t('runtime.version') + ':' + item.params['PHP_VERSION'] + ']',
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
                await ChangePHPVersion(versionReq);
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            } catch (error) {}
            loading.value = false;
        });
    } catch (error) {}
};

onMounted(() => {
    versionReq.runtimeID = props.website.runtimeID;
    versionReq.websiteID = props.website.id;
    oldRuntimeID.value = props.website.runtimeID;
    getRuntimes();
});
</script>
