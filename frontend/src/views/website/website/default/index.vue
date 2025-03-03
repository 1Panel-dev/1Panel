<template>
    <DrawerPro v-model="open" :header="$t('website.defaultServer')" size="normal" @close="handleClose">
        <el-form @submit.prevent label-position="top" v-loading="loading">
            <el-form-item :label="$t('website.defaultServer')">
                <el-select v-model="defaultId">
                    <el-option :value="0" :key="-1" :label="$t('website.noDefaultServer')"></el-option>
                    <el-option
                        v-for="(website, key) in websites"
                        :key="key"
                        :value="website.id"
                        :label="website.primaryDomain"
                    ></el-option>
                </el-select>
            </el-form-item>
        </el-form>
        <el-alert :closable="false">
            <template #default>
                <span class="whitespace-pre-line">{{ $t('website.defaultServerHelper') }}</span>
            </template>
        </el-alert>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { changeDefaultServer, listWebsites } from '@/api/modules/website';
import i18n from '@/lang';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';

let open = ref(false);
let websites = ref<any>();
let defaultId = ref(-1);
let loading = ref(false);

const acceptParams = () => {
    defaultId.value = 0;
    get();
    open.value = true;
};

const handleClose = () => {
    open.value = false;
};

const get = async () => {
    const res = await listWebsites();
    websites.value = res.data;
    websites.value.forEach((website: Website.WebsiteDTO) => {
        if (website.defaultServer) {
            defaultId.value = website.id;
        }
    });
};

const submit = () => {
    loading.value = true;
    changeDefaultServer({ id: defaultId.value })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            handleClose();
        })
        .finally(() => {
            loading.value = false;
        });
};
defineExpose({ acceptParams });
</script>
