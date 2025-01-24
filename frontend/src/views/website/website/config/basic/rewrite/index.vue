<template>
    <div>
        <el-form-item :label="$t('website.rewriteMode')">
            <el-select v-model="req.name" filterable @change="getRewrite(req.name)" class="p-w-200">
                <el-option :label="$t('website.current')" :value="'current'"></el-option>
                <el-option
                    v-for="(rewrite, index) in rewrites"
                    :key="index"
                    :label="rewrite.name"
                    :value="rewrite.name"
                >
                    <span>{{ rewrite.name }}</span>
                    <el-button
                        class="float-right mt-1.5"
                        v-if="rewrite.resource == 'custom'"
                        link
                        icon="Close"
                        @click="deleteCustomRewrite(rewrite.name)"
                    ></el-button>
                </el-option>
            </el-select>
        </el-form-item>
        <el-text type="warning">{{ $t('website.rewriteHelper2') }}</el-text>
        <CodemirrorPro v-model="content" mode="nginx" :heightDiff="500"></CodemirrorPro>
        <div class="mt-2">
            <el-form-item>
                <el-alert :title="$t('website.rewriteHelper')" type="info" :closable="false" />
            </el-form-item>
            <el-button type="primary" @click="submit()">
                {{ $t('nginx.saveAndReload') }}
            </el-button>
            <el-button type="primary" @click="opCustomRewrite()" :disabled="content == ''">
                {{ $t('website.saveCustom') }}
            </el-button>
        </div>
        <CustomRewrite ref="customRef" @close="init()" />
        <OpDialog ref="deleteRef" @search="init()" />
    </div>
</template>

<script lang="ts" setup>
import { computed, nextTick, onMounted, reactive, ref } from 'vue';
import {
    getWebsite,
    getRewriteConfig,
    updateRewriteConfig,
    listCustomRewrite,
    operateCustomRewrite,
} from '@/api/modules/website';
import { Rewrites } from '@/global/mimetype';
import { MsgSuccess } from '@/utils/message';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import i18n from '@/lang';
import CustomRewrite from '@/views/website/website/config/basic/rewrite/custom/index.vue';

const loading = ref(false);
const content = ref(' ');
const codeRef = ref();
const customRef = ref();
const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const id = computed(() => {
    return props.id;
});
const req = reactive({
    websiteID: id.value,
    name: 'default',
});
const update = reactive({
    websiteID: id.value,
    content: 'd',
    name: '',
});
const rewrites = ref([]);
const deleteRef = ref();

const getRewrite = async (rewrite: string) => {
    loading.value = true;
    req.name = rewrite;
    req.websiteID = id.value;
    try {
        const res = await getRewriteConfig(req);
        content.value = res.data.content;
        if (res.data.content == '') {
            content.value = ' ';
        }
        setCursorPosition();
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const setCursorPosition = () => {
    nextTick(() => {
        const codeMirrorInstance = codeRef.value?.codemirror;
        codeMirrorInstance?.setCursor(0, 0);
    });
};

const submit = async () => {
    update.name = req.name;
    update.websiteID = id.value;
    update.content = content.value;
    loading.value = true;
    try {
        await updateRewriteConfig(update);
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const opCustomRewrite = async () => {
    customRef.value.acceptParams(content.value);
};

const deleteCustomRewrite = (name: string) => {
    deleteRef.value.acceptParams({
        title: i18n.global.t('commons.msg.deleteTitle'),
        names: [name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('container.template'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: operateCustomRewrite,
        params: { name: name, operate: 'delete' },
    });
};

const init = () => {
    listCustomRewrite().then((res) => {
        rewrites.value = [];
        if (res && res.data) {
            for (const d of res.data) {
                rewrites.value.push({
                    resource: 'custom',
                    name: d,
                });
            }
        }
        for (const r of Rewrites) {
            rewrites.value.push({
                resource: 'default',
                name: r,
            });
        }
    });
    getWebsite(id.value).then((res) => {
        const name = res.data.rewrite == '' ? 'default' : 'current';
        if (name === 'current') {
            req.name = 'current';
        }
        getRewrite(name);
    });
};

onMounted(() => {
    init();
});
</script>
