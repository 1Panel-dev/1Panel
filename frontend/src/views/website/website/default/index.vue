<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="40%">
        <template #header>
            <DrawerHeader :header="$t('website.defaulServer')" :back="handleClose"></DrawerHeader>
        </template>
        <el-row>
            <el-col :span="22" :offset="1">
                <el-form label-position="top">
                    <el-form-item :label="$t('website.defaulServer')">
                        <el-select v-model="defaultId" style="width: 100%">
                            <el-option :value="0" :key="-1" :label="$t('website.noDefaulServer')"></el-option>
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
                        <span style="white-space: pre-line">{{ $t('website.defaulServerHelper') }}</span>
                    </template>
                </el-alert>
            </el-col>
        </el-row>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>
<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { Website } from '@/api/interface/Website';
import { ChangeDefaultServer, ListWebsites } from '@/api/modules/website';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import { ref } from 'vue';

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
    const res = await ListWebsites();
    websites.value = res.data;
    websites.value.forEach((website: Website.WebsiteDTO) => {
        if (website.defaultServer) {
            defaultId.value = website.id;
        }
    });
};

const submit = () => {
    ChangeDefaultServer({ id: defaultId.value }).then(() => {
        ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
        handleClose();
    });
};

defineExpose({ acceptParams });
</script>
