<template>
    <div>
        <el-alert
            v-if="websiteType == 'deployment'"
            :title="$t('website.websiteDeploymentHelper')"
            type="info"
            :closable="false"
        />
        <el-alert
            v-if="websiteType == 'static'"
            :title="$t('website.websiteStatictHelper')"
            type="info"
            :closable="false"
        />
        <el-alert
            v-if="websiteType == 'proxy'"
            :title="$t('website.websiteProxyHelper')"
            type="info"
            :closable="false"
        />
        <el-alert
            v-if="websiteType == 'runtime'"
            :title="$t('website.runtimeProxyHelper')"
            type="info"
            :closable="false"
        />
        <el-alert v-if="websiteType == 'subsite'" :title="$t('website.subsiteHelper')" type="info" :closable="false" />
    </div>
</template>

<script lang="ts" setup>
defineProps({
    websiteType: {
        type: String,
        default: 'deployment',
    },
});
</script>
