<template>
    <el-descriptions :column="7" border direction="vertical">
        <el-descriptions-item :label="$t('website.primaryDomain')">
            {{ websiteSSL.primaryDomain }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('website.otherDomains')">
            {{ websiteSSL.domains }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('website.brand')">
            {{ websiteSSL.organization }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('ssl.provider')">
            {{ getProvider(websiteSSL.provider) }}
        </el-descriptions-item>
        <el-descriptions-item
            :label="$t('ssl.acmeAccount')"
            v-if="websiteSSL.acmeAccount && websiteSSL.provider !== 'manual'"
        >
            {{ websiteSSL.acmeAccount.email }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('website.expireDate')">
            {{ dateFormatSimple(websiteSSL.expireDate) }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('website.remark')">
            {{ websiteSSL.description }}
        </el-descriptions-item>
    </el-descriptions>
</template>

<script lang="ts" setup>
import { dateFormatSimple, getProvider } from '@/utils/util';
defineProps({
    websiteSSL: {
        type: Object,
        required: true,
    },
});
</script>
