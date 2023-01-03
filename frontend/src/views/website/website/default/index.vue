<template>
    <el-dialog
        v-model="open"
        :title="$t('website.defaulServer')"
        width="20%"
        @close="handleClose"
        :close-on-click-modal="false"
    >
        <div style="text-align: center">
            <el-select v-model="defaultId">
                <el-option :value="0" :key="-1" :label="$t('website.noDefaulServer')"></el-option>
                <el-option
                    v-for="(website, key) in websites"
                    :key="key"
                    :value="website.id"
                    :label="website.primaryDomain"
                ></el-option>
            </el-select>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>
<script lang="ts" setup>
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
